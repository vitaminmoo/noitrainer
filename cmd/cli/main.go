package main

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"noitrainer/noita"

	"github.com/vitaminmoo/memtools/process"
)

func main() {
	fmt.Println("Looking for noita.exe...")
	procs, err := process.FromName("noita.exe")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	proc := procs[0]
	fmt.Printf("Found noita.exe (PID %d)\n\n", proc.PID)

	reader := noita.NewReader(proc)

	// First read: print everything
	state := reader.ReadState()
	if !state.Connected {
		fmt.Fprintf(os.Stderr, "Failed to connect: %s\n", state.Error)
		os.Exit(1)
	}

	prev := flattenState(state, reader)
	fmt.Println("=== Initial State ===")
	keys := sortedKeys(prev)
	for _, k := range keys {
		fmt.Printf("  %-40s %s\n", k, prev[k])
	}
	fmt.Printf("\n(%d values)\n", len(prev))
	fmt.Println("\n=== Watching for changes (1s interval) ===")

	for {
		time.Sleep(1 * time.Second)
		state = reader.ReadState()
		if !state.Connected {
			fmt.Println("[disconnected]")
			continue
		}
		curr := flattenState(state, reader)
		changes := diff(prev, curr)
		if len(changes) > 0 {
			ts := time.Now().Format("15:04:05")
			fmt.Printf("\n[%s] %d changed:\n", ts, len(changes))
			for _, k := range sortedKeys(changes) {
				fmt.Printf("  %-40s %s -> %s\n", k, prev[k], changes[k])
			}
		}
		prev = curr
	}
}

func flattenState(gs *noita.GameState, reader *noita.Reader) map[string]string {
	m := make(map[string]string)

	m["WorldSeed"] = fmt.Sprintf("%d", gs.WorldSeed)
	m["DeathCount"] = fmt.Sprintf("%d", gs.DeathCount)
	m["NumOrbsTotal"] = fmt.Sprintf("%d", gs.NumOrbsTotal)
	m["Camera.X"] = fmt.Sprintf("%.1f", gs.CameraX)
	m["Camera.Y"] = fmt.Sprintf("%.1f", gs.CameraY)
	m["Camera.ViewW"] = fmt.Sprintf("%.0f", gs.ViewW)
	m["Camera.ViewH"] = fmt.Sprintf("%.0f", gs.ViewH)

	if g := gs.Globals; g != nil {
		flattenStruct(m, "Globals", g, reader)
	}
	if ws := gs.WorldState; ws != nil {
		flattenStruct(m, "WorldState", ws, reader)
	}
	if e := gs.PlayerEntity; e != nil {
		flattenStruct(m, "Player", e, reader)
	}
	if d := gs.PlayerHP; d != nil {
		flattenStruct(m, "HP", d, reader)
	}
	if w := gs.PlayerWallet; w != nil {
		flattenStruct(m, "Wallet", w, reader)
	}
	if c := gs.PlayerChar; c != nil {
		flattenStruct(m, "Char", c, reader)
	}
	if inv := gs.PlayerInv; inv != nil {
		flattenStruct(m, "Inv", inv, reader)
	}
	for i, item := range gs.Wands {
		m[fmt.Sprintf("Wand%d.Name", i)] = fmt.Sprintf("%q", item.Name(reader.Ctx))
		flattenStruct(m, fmt.Sprintf("Wand%d", i), item.Ability, reader)
	}
	for i, item := range gs.Items {
		m[fmt.Sprintf("Item%d.Name", i)] = fmt.Sprintf("%q", item.Name(reader.Ctx))
		for _, mat := range item.Contents {
			m[fmt.Sprintf("Item%d.%s", i, mat.Name)] = fmt.Sprintf("%.0f", mat.Amount)
		}
	}

	return m
}

func flattenStruct(m map[string]string, prefix string, v any, reader *noita.Reader) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return
		}
		rv = rv.Elem()
	}
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fv := rv.Field(i)
		key := prefix + "." + field.Name

		switch fv.Kind() {
		case reflect.Struct:
			if field.Type == reflect.TypeOf(noita.MsvcString{}) {
				ms := fv.Addr().Interface().(*noita.MsvcString)
				m[key] = fmt.Sprintf("%q", noita.MsvcStringValue(ms, reader.Ctx))
			} else if field.Type == reflect.TypeOf(noita.ComponentHeader{}) {
				hdr := fv.Interface().(noita.ComponentHeader)
				m[key+".TypeId"] = fmt.Sprintf("%d", hdr.TypeId)
				m[key+".Active"] = fmt.Sprintf("%v", hdr.Active)
			} else {
				flattenStruct(m, key, fv.Addr().Interface(), reader)
			}
		case reflect.Array:
			if field.Type.Elem().Kind() == reflect.Uint8 && field.Type.Len() > 16 {
				m[key] = fmt.Sprintf("[%d bytes]", field.Type.Len())
			} else {
				m[key] = fmt.Sprintf("%v", fv.Interface())
			}
		case reflect.Bool:
			m[key] = fmt.Sprintf("%v", fv.Bool())
		case reflect.Float32, reflect.Float64:
			m[key] = fmt.Sprintf("%g", fv.Float())
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			m[key] = fmt.Sprintf("%d", fv.Int())
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val := fv.Uint()
			name := strings.ToLower(field.Name)
			if val > 0xFFFF && (strings.Contains(name, "ptr") || strings.HasPrefix(name, "p") || strings.Contains(name, "vtable")) {
				m[key] = fmt.Sprintf("0x%08X", val)
			} else {
				m[key] = fmt.Sprintf("%d", val)
			}
		default:
			m[key] = fmt.Sprintf("%v", fv.Interface())
		}
	}
}

func diff(prev, curr map[string]string) map[string]string {
	changes := make(map[string]string)
	for k, v := range curr {
		if pv, ok := prev[k]; !ok || pv != v {
			changes[k] = v
		}
	}
	for k := range prev {
		if _, ok := curr[k]; !ok {
			changes[k] = "<removed>"
		}
	}
	return changes
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
