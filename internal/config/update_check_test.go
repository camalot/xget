package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestXgetUpdateCheckDefaultsToTrue(t *testing.T) {
	for name, content := range map[string]string{
		"unset":    "global:\n  quiet: true\n",
		"disabled": "global:\n  xget_update_check: false\n",
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), ".xget.yml")
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			cfg, err := LoadQuiet(path)
			if err != nil {
				t.Fatal(err)
			}
			if want := name == "unset"; cfg.Global.XgetUpdateCheck != want {
				t.Fatalf("xget_update_check = %v, want %v", cfg.Global.XgetUpdateCheck, want)
			}
		})
	}
	if !Default().Global.XgetUpdateCheck {
		t.Fatal("Default() xget_update_check = false, want true")
	}
}
