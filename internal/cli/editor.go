package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// editorEnvVars lists the environment variables consulted, in order, when
// resolving the editor for `xget config edit`.
var editorEnvVars = []string{"XGET_EDITOR", "VISUAL", "EDITOR"}

// lookPath is indirected for testing.
var lookPath = exec.LookPath

func fallbackEditors() []string {
	if runtime.GOOS == "windows" {
		return []string{"nano", "notepad"}
	}
	return []string{"nano"}
}

// resolveEditor returns the editor executable and any leading arguments.
func resolveEditor() (string, []string, error) {
	for _, name := range editorEnvVars {
		value := strings.TrimSpace(os.Getenv(name))
		if value == "" {
			continue
		}
		fields := strings.Fields(value)
		return fields[0], fields[1:], nil
	}

	var last string
	for _, candidate := range fallbackEditors() {
		last = candidate
		if _, err := lookPath(candidate); err == nil {
			return candidate, nil, nil
		}
	}
	if runtime.GOOS == "windows" {
		// notepad is always present; use it even if LookPath failed.
		return "notepad", nil, nil
	}
	return "", nil, fmt.Errorf("no editor found: set XGET_EDITOR, VISUAL, or EDITOR (tried %s)", last)
}

// runEditor is indirected for testing.
var runEditor = func(name string, args []string) error {
	resolved, err := lookPath(name)
	if err != nil {
		return fmt.Errorf("editor %q not found: %w", name, err)
	}
	cmd := exec.Command(resolved, args...) // #nosec G204 -- editor comes from user configuration by design
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
