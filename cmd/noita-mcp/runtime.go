// Runtime tools attach to the running noita.exe process and mirror the
// introspection surface of cmd/cli. Each tool opens a fresh Reader per
// call; there is no shared state between tools.
package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"noitrainer/noita"

	"github.com/vitaminmoo/memtools/process"
)

// connectRuntime finds the running noita.exe and returns a Reader bound to
// it. Returns an error when noita is not running or cannot be attached.
func connectRuntime() (*noita.Reader, *process.Process, error) {
	procs, err := process.FromName("noita.exe")
	if err != nil {
		return nil, nil, fmt.Errorf("noita.exe not found: %w", err)
	}
	if len(procs) == 0 {
		return nil, nil, fmt.Errorf("noita.exe not running")
	}
	proc := procs[0]
	return noita.NewReader(proc), proc, nil
}

// parseRuntimeAddr accepts decimal or 0x-prefixed hex as a 32-bit address.
// Mirrors cmd/cli/main.go:parseAddr.
func parseRuntimeAddr(s string) (uint32, error) {
	s = strings.TrimSpace(s)
	base := 10
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
		base = 16
	}
	v, err := strconv.ParseUint(s, base, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid address %q: %w", s, err)
	}
	return uint32(v), nil
}

// truncateStr mirrors cmd/cli/main.go:truncate.
func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-2] + ".."
}

// registerRuntimeTools registers every runtime introspection tool.
// Implementations are added in subsequent plan tasks.
func registerRuntimeTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "ngplus",
		Description: "Read the NG+ count from the running game.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		v, err := noita.ReadGNgPlusCount(reader.Ctx)
		if err != nil {
			return toolErr(fmt.Errorf("read ng+ count: %w", err))
		}
		return textResult(fmt.Sprintf("NG+ count: %d\n", v)), nil, nil
	})
}
