package noitadata

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var ErrNotFound = errors.New("noita install not found")

// FindInstall locates the Noita install directory.
//
// Resolution order:
//  1. $NOITA_PATH if set
//  2. Platform-specific Steam default locations
//
// Returns the absolute path to the Noita directory (the one containing
// noita.exe, data/, tools_modding/, etc).
func FindInstall() (string, error) {
	if env := os.Getenv("NOITA_PATH"); env != "" {
		if ok, err := looksLikeNoita(env); err != nil {
			return "", err
		} else if ok {
			return filepath.Clean(env), nil
		}
		return "", fmt.Errorf("%w: NOITA_PATH=%q does not contain data/data.wak", ErrNotFound, env)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home dir: %w", err)
	}

	var candidates []string
	switch runtime.GOOS {
	case "linux":
		candidates = []string{
			filepath.Join(home, ".steam/steam/steamapps/common/Noita"),
			filepath.Join(home, ".local/share/Steam/steamapps/common/Noita"),
			filepath.Join(home, ".var/app/com.valvesoftware.Steam/data/Steam/steamapps/common/Noita"),
		}
	case "darwin":
		candidates = []string{
			filepath.Join(home, "Library/Application Support/Steam/steamapps/common/Noita"),
		}
	case "windows":
		candidates = []string{
			`C:\Program Files (x86)\Steam\steamapps\common\Noita`,
			`C:\Program Files\Steam\steamapps\common\Noita`,
		}
	}

	for _, c := range candidates {
		if ok, _ := looksLikeNoita(c); ok {
			return filepath.Clean(c), nil
		}
	}
	return "", fmt.Errorf("%w: checked %v", ErrNotFound, candidates)
}

func looksLikeNoita(dir string) (bool, error) {
	info, err := os.Stat(filepath.Join(dir, "data", "data.wak"))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}
