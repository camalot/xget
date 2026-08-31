package options

type Flags struct {
	Tag         string
	Prerelease  bool
	Source      bool
	Output      string
	System      string
	ExtractFile string
	All         bool
	Quiet       bool
	DLOnly      bool
	UpgradeOnly bool
	Asset       []string
	Ignore      []string
	Hash        bool
	Verify      string
	Remove      bool
	DisableSSL  bool
}
