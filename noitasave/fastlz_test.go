package noitasave

import (
	"bytes"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestFastlzRoundTrip covers both levels across sizes that straddle the
// level-selection threshold and the literal-run/match boundaries.
func TestFastlzRoundTrip(t *testing.T) {
	sizes := []int{0, 1, 2, 3, 4, 5, 31, 32, 33, 63, 64, 100, 1000,
		65535, 65536, 65537, 200000}

	for _, n := range sizes {
		for _, name := range []string{"zeros", "random", "compressible"} {
			in := make([]byte, n)
			rng := rand.New(rand.NewSource(int64(n)))
			switch name {
			case "random":
				rng.Read(in)
			case "compressible":
				// Long runs plus repeated phrases: exercises run encoding,
				// near matches and (at size) far matches.
				phrase := []byte("the quick brown fox jumps over the lazy dog")
				for i := range in {
					if i%97 < 40 {
						in[i] = phrase[i%len(phrase)]
					} else {
						in[i] = byte(i / 256)
					}
				}
			}

			comp := fastlzCompress(in)
			if n == 0 {
				if len(comp) != 0 {
					t.Errorf("size 0: expected empty output, got %d bytes", len(comp))
				}
				continue
			}

			got, err := fastlzDecompress(comp, len(in))
			if err != nil {
				t.Errorf("%s/%d: decompress: %v", name, n, err)
				continue
			}
			if !bytes.Equal(got, in) {
				t.Errorf("%s/%d: round-trip mismatch at byte %d", name, n, firstDiff(got, in))
			}
		}
	}
}

// TestFastlzMatchesReferenceEncoder compares our encoder against the original
// C implementation byte for byte, so a re-written save is indistinguishable
// from one Noita wrote itself.
//
// Build the harness from the reference sources and point the env var at it:
//
//	cc -O2 -o /tmp/flz_cli flz_cli.c fastlz.c
//	NOITA_FASTLZ_REF=/tmp/flz_cli go test ./noitasave/ -run Reference
//
// where flz_cli.c reads stdin and writes fastlz_compress(...) to stdout.
func TestFastlzMatchesReferenceEncoder(t *testing.T) {
	ref := os.Getenv("NOITA_FASTLZ_REF")
	if ref == "" {
		t.Skip("set NOITA_FASTLZ_REF to a reference fastlz_compress harness")
	}

	inputs := map[string][]byte{}

	// Synthetic shapes.
	for _, n := range []int{4, 33, 1000, 65535, 65536, 300000} {
		rng := rand.New(rand.NewSource(int64(n)))
		random := make([]byte, n)
		rng.Read(random)
		inputs[padName("random", n)] = random

		mixed := make([]byte, n)
		for i := range mixed {
			if i%97 < 40 {
				mixed[i] = "the quick brown fox"[i%19]
			} else {
				mixed[i] = byte(i / 256)
			}
		}
		inputs[padName("mixed", n)] = mixed
		inputs[padName("zeros", n)] = make([]byte, n)
	}

	// Real payloads, which is what actually matters.
	if dir := os.Getenv("NOITA_SAVE_WORLD"); dir != "" {
		paths, _ := filepath.Glob(filepath.Join(dir, "*.png_petri"))
		for _, p := range paths {
			raw, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			payload, err := DecodeContainer(raw)
			if err != nil {
				continue
			}
			inputs["save:"+filepath.Base(p)] = payload
		}
	}

	for name, in := range inputs {
		cmd := exec.Command(ref)
		cmd.Stdin = bytes.NewReader(in)
		want, err := cmd.Output()
		if err != nil {
			t.Fatalf("%s: reference encoder failed: %v", name, err)
		}
		got := fastlzCompress(in)
		if !bytes.Equal(got, want) {
			t.Errorf("%s (%d bytes in): encoder differs from reference; got %d bytes, want %d, first diff at %d",
				name, len(in), len(got), len(want), firstDiff(got, want))
		}
	}
}

func padName(kind string, n int) string {
	return kind + ":" + itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
