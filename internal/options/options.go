package options

import "github.com/camalot/xget/internal/config"

type Flags struct {
	Tag          string
	Prerelease   bool
	Source       bool
	Output       string
	System       string
	ExtractFile  string
	All          bool
	Quiet        bool
	DLOnly       bool
	UpgradeOnly  bool
	Asset        []string
	Ignore       []string
	Hash         bool
	Verify       string
	Remove       bool
	DisableSSL   bool
	SourceType   string
	SourceConfig config.Source
	// Untracked skips recording this run in the installed package store.
	Untracked bool
}
