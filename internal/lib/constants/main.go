// Package constants defines shared string constants.
package constants

// Shared runtime, provider, scheme, and tag constants.
const (
	RuntimeWindows = "windows"
	RuntimeLinux   = "linux"
	RuntimeDarwin  = "darwin"
	RuntimeNetBSD  = "netbsd"
	RuntimeFreeBSD = "freebsd"
	RuntimeOpenBSD = "openbsd"
	RuntimeAndroid = "android"
	RuntimeIllumos = "illumos"
	RuntimeSolaris = "solaris"
	RuntimePlan9   = "plan9"

	SystemAll = "all"

	ProviderGithub = "github"
	ProviderGitlab = "gitlab"
	ProviderURL    = "url"

	SchemeHTTPS = "https"
	SchemeHTTP  = "http"

	TagLatest = "latest"
	TagHead   = "HEAD"
)
