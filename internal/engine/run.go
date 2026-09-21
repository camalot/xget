package engine

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/camalot/xget/internal/config"
	"github.com/camalot/xget/internal/installed"
	"github.com/camalot/xget/internal/options"
	pb "github.com/schollz/progressbar/v3"
)

func cleanLocalPath(p string) (string, error) {
	clean := filepath.Clean(p)
	if clean == "" || clean == "." {
		return "", fmt.Errorf("invalid path %q", p)
	}
	if !filepath.IsAbs(clean) && (clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator))) {
		return "", fmt.Errorf("invalid relative path %q", p)
	}
	return clean, nil
}

func safeBaseName(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty file name")
	}
	base := filepath.Base(name)
	if base == "." || base == string(os.PathSeparator) || base == ".." {
		return "", fmt.Errorf("invalid file name %q", name)
	}
	if strings.ContainsRune(base, os.PathSeparator) {
		return "", fmt.Errorf("invalid file name %q", name)
	}
	return base, nil
}

// IsUrl returns true if s is a valid URL.
func IsUrl(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// Cut is strings.Cut
func Cut(s, sep string) (before, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}

var ghrgx = regexp.MustCompile(`^(http(s)?://)?github\.com/[\w,\-,_]+/[\w,\-,_]+(.git)?(/)?$`)

// IsGithubUrl returns true if s is a URL with github.com as the host.
func IsGithubUrl(s string) bool {
	return ghrgx.MatchString(s)
}

func IsLocalFile(s string) bool {
	_, err := os.Stat(s)
	return err == nil
}

// IsDirectory returns true if the file at path is a directory.
func IsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func isDirectoryDestination(path string) bool {
	return strings.HasSuffix(path, "/") || strings.HasSuffix(path, "\\") || IsDirectory(path)
}

// searches for an asset that has the same name as the requested one but
// ending with .sha256 or .sha256sum
func checksumAsset(asset string, assets []string) string {
	for _, a := range assets {
		if a == asset+".sha256sum" || a == asset+".sha256" {
			return a
		}
	}
	return ""
}

// isProviderRepoRootPath reports whether urlPath names a repository root
// (e.g. "/owner/repo") rather than a release asset, archive, or other
// provider-hosted subpath, which must stay on the DirectAssetFinder path.
func isProviderRepoRootPath(source config.Source, urlPath string) bool {
	trimmed := strings.TrimSuffix(strings.Trim(urlPath, "/"), ".git")
	if trimmed == "" {
		return false
	}
	parts := strings.Split(trimmed, "/")
	for _, part := range parts {
		if part == "" {
			return false
		}
	}
	if source.Type == "gitlab" {
		for _, part := range parts {
			// GitLab prefixes non-root project pages (releases, archives,
			// blobs, etc.) with a literal "-" path segment.
			if part == "-" {
				return false
			}
		}
		return len(parts) >= 2
	}
	return len(parts) == 2
}

// Determine the appropriate Finder to use. If url is a local/direct URL we use
// a DirectAssetFinder. Otherwise we use a GithubAssetFinder.
func getFinder(project string, opts *options.Flags) (finder Finder, tool string, err error) {
	source := opts.SourceConfig
	if source.Type == "" {
		profile := strings.ToLower(opts.SourceType)
		source, err = config.Default().ResolveSource(profile)
		if err != nil {
			return nil, "", err
		}
		opts.SourceConfig = source
		opts.SourceType = source.Name
	}

	if IsLocalFile(project) {
		finder = &DirectAssetFinder{URL: project}
		tool = filepath.Base(project)
		opts.SourceType = "URL"
		opts.System = "all"
		return finder, tool, nil
	}

	if IsUrl(project) {
		parsed, parseErr := url.Parse(project)
		if parseErr != nil || !strings.EqualFold(parsed.Hostname(), source.Host) || !isProviderRepoRootPath(source, parsed.Path) {
			finder = &DirectAssetFinder{URL: project}
			tool = path.Base(parsed.Path)
			opts.SourceType = "URL"
			opts.System = "all"
			return finder, tool, nil
		}
		project = strings.TrimSuffix(strings.Trim(parsed.Path, "/"), ".git")
	}

	repo := project
	parts := strings.Split(repo, "/")
	if len(parts) < 2 || (source.Type == "github" && len(parts) != 2) {
		return nil, "", fmt.Errorf("invalid %s repository %q", source.Type, repo)
	}
	for _, part := range parts {
		if part == "" {
			return nil, "", fmt.Errorf("invalid %s repository %q", source.Type, repo)
		}
	}
	tool = parts[len(parts)-1]

	if opts.Source {
		tag := "master"
		if opts.Tag != "" {
			tag = opts.Tag
		}
		if source.Type == "gitlab" {
			finder = &GitlabSourceFinder{Repo: repo, Tag: tag, Tool: tool, Source: source}
		} else {
			finder = &GithubSourceFinder{Repo: repo, Tag: tag, Tool: tool, Source: source}
		}
		return finder, tool, nil
	}

	tag := "latest"
	if opts.Tag != "" && opts.Tag != "latest" {
		tag = fmt.Sprintf("tags/%s", opts.Tag)
	}

	var mint time.Time
	if opts.UpgradeOnly {
		last := parts[len(parts)-1]
		mint = bintime(last, opts.Output)
	}

	if source.Type == "gitlab" {
		finder = &GitlabAssetFinder{Repo: repo, Tag: opts.Tag, Prerelease: opts.Prerelease, MinTime: mint, Source: source}
	} else {
		finder = &GithubAssetFinder{
			Repo:       repo,
			Tag:        tag,
			Prerelease: opts.Prerelease,
			MinTime:    mint,
			Source:     source,
		}
	}
	return finder, tool, nil
}

func getVerifier(sumAsset, githubDigest string, opts *options.Flags) (verifier Verifier, err error) {
	if opts.Verify != "" {
		if opts.Verify == "auto" {
			if githubDigest == "" {
				// return nil, fmt.Errorf("no SHA256 digest available for this asset")
				// skip verification if no SHA256 digest is available
				return &NoVerifier{}, nil
			}
			verifier, err = NewSha256Verifier(githubDigest)
		} else {
			verifier, err = NewSha256Verifier(opts.Verify)
		}
	} else if sumAsset != "" {
		verifier = &Sha256AssetVerifier{AssetURL: sumAsset, Source: opts.SourceConfig}
	} else if githubDigest != "" {
		verifier, err = NewSha256Verifier(githubDigest)
	} else if opts.Hash {
		verifier = &Sha256Printer{}
	} else {
		verifier = &NoVerifier{}
	}
	return verifier, err
}

// Determine the appropriate detector.
func getDetector(opts *options.Flags) (detector Detector, err error) {
	var system Detector
	if opts.System == "all" {
		system = &AllDetector{}
	} else if opts.System != "" {
		split := strings.Split(opts.System, "/")
		if len(split) < 2 {
			return nil, fmt.Errorf("system descriptor must be os/arch")
		}
		system, err = NewSystemDetector(split[0], split[1])
	} else {
		system, err = NewSystemDetector(runtime.GOOS, runtime.GOARCH)
	}
	if err != nil {
		return nil, err
	}

	var detectors []Detector

	for _, raw := range opts.Ignore {
		asset, anti, rx, perr := parseAssetMatcher(raw)
		if perr != nil {
			return nil, fmt.Errorf("invalid ignore matcher %q: %w", raw, perr)
		}
		// ignore excludes matches by default; negative ignore matchers invert this.
		detectors = append(detectors, &SingleAssetDetector{Asset: asset, Anti: !anti, Regex: rx})
	}

	for _, raw := range opts.Asset {
		asset, anti, rx, perr := parseAssetMatcher(raw)
		if perr != nil {
			return nil, fmt.Errorf("invalid asset matcher %q: %w", raw, perr)
		}
		detectors = append(detectors, &SingleAssetDetector{Asset: asset, Anti: anti, Regex: rx})
	}

	if len(detectors) == 0 {
		return system, nil
	}

	return &DetectorChain{detectors: detectors, system: system}, nil
}

func parseAssetMatcher(raw string) (asset string, anti bool, rx *regexp.Regexp, err error) {
	asset = raw

	if strings.HasPrefix(asset, "^^") {
		asset = asset[1:]
	} else if strings.HasPrefix(asset, "not:") {
		anti = true
		asset = strings.TrimPrefix(asset, "not:")
	} else if strings.HasPrefix(asset, "^") {
		anti = true
		asset = asset[1:]
	}

	if strings.TrimSpace(asset) == "" {
		return "", anti, nil, fmt.Errorf("matcher cannot be empty")
	}

	if strings.HasPrefix(asset, "text:") {
		literal := strings.TrimPrefix(asset, "text:")
		if strings.TrimSpace(literal) == "" {
			return "", anti, nil, fmt.Errorf("text matcher cannot be empty")
		}
		return literal, anti, nil, nil
	}

	if strings.HasPrefix(asset, "~~") {
		asset = asset[1:]
		return asset, anti, nil, nil
	}

	if strings.HasPrefix(asset, "~") || strings.HasPrefix(asset, "=~") || strings.HasPrefix(asset, "re:") {
		pattern := asset
		switch {
		case strings.HasPrefix(pattern, "=~"):
			pattern = strings.TrimPrefix(pattern, "=~")
		case strings.HasPrefix(pattern, "re:"):
			pattern = strings.TrimPrefix(pattern, "re:")
		default:
			pattern = strings.TrimPrefix(pattern, "~")
		}
		if strings.TrimSpace(pattern) == "" {
			return "", anti, nil, fmt.Errorf("regex pattern cannot be empty")
		}
		rx, err = regexp.Compile(pattern)
		if err != nil {
			return "", anti, nil, err
		}
		return asset, anti, rx, nil
	}
	return asset, anti, nil, nil
}

// Determine which extractor to use.
func getExtractor(url, tool string, opts *options.Flags) (extractor Extractor, err error) {
	filename := extractorFilename(url)
	if opts.DLOnly {
		extractor = &SingleFileExtractor{
			Name:   filename,
			Rename: filename,
			Decompress: func(r io.Reader) (io.Reader, error) {
				return r, nil
			},
		}
	} else if opts.ExtractFile != "" {
		gc, err := NewGlobChooser(opts.ExtractFile)
		if err != nil {
			return nil, err
		}
		extractor = NewExtractor(filename, tool, gc)
	} else {
		extractor = NewExtractor(filename, tool, &BinaryChooser{Tool: tool})
	}
	return extractor, nil
}

// extractorFilename returns the base filename used for archive-suffix
// detection, stripping any query string (e.g. GitLab's
// "archive.tar.gz?sha=...") that would otherwise defeat NewExtractor's
// ".tar.gz"-style suffix checks.
func extractorFilename(rawURL string) string {
	if u, err := url.Parse(rawURL); err == nil && u.Path != "" {
		return path.Base(u.Path)
	}
	return path.Base(rawURL)
}

// Write an extracted file to disk with a new name.
func writeFile(data []byte, rename string, mode fs.FileMode) error {
	if rename == "" {
		return fmt.Errorf("invalid output path")
	}
	if rename[0] == '-' {
		_, err := os.Stdout.Write(data)
		return err
	}

	safeRename, err := cleanLocalPath(rename)
	if err != nil {
		return err
	}

	// #nosec G703 -- cleanLocalPath rejects traversal segments for relative paths.
	err = os.Remove(safeRename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	// #nosec G703 -- destination path is normalized and validated via cleanLocalPath.
	_ = os.MkdirAll(filepath.Dir(safeRename), 0750)

	// #nosec G304,G703 -- path is normalized and validated via cleanLocalPath above.
	f, err := os.OpenFile(safeRename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("error closing file:", err)
		}
	}()
	_, err = f.Write(data)
	return err
}

// ErrNonInteractive is returned when user input is required but xget is
// running in non-interactive mode.
var ErrNonInteractive = errors.New("interactive user request while execution is in non-interactive mode")

var runtimeNonInteractive bool

// SetNonInteractive toggles non-interactive mode for the current process.
func SetNonInteractive(disable bool) {
	runtimeNonInteractive = disable
}

// NonInteractive reports whether prompting the user is forbidden.
func NonInteractive() bool {
	return runtimeNonInteractive
}

func userSelect(choices []interface{}) (int, error) {
	for i, c := range choices {
		fmt.Fprintf(os.Stderr, "(%d) %v\n", i+1, c)
	}
	if runtimeNonInteractive {
		return 0, ErrNonInteractive
	}
	var choice int
	for {
		fmt.Fprint(os.Stderr, "Enter selection number: ")
		_, err := fmt.Scanf("%d", &choice)
		if err == nil && (choice <= 0 || choice > len(choices)) {
			err = fmt.Errorf("%d is out of bounds", choice)
		}
		if err == nil {
			break
		}
		if errors.Is(err, io.EOF) {
			return 0, fmt.Errorf("error reading selection")
		}
		fmt.Fprintf(os.Stderr, "Invalid selection: %v\n", err)
	}
	return choice, nil
}

func bintime(bin string, to string) (t time.Time) {
	file := ""
	dir := "."
	if to != "" && IsDirectory(to) {
		dir = to
	} else if xbin := os.Getenv("EGET_BIN"); xbin != "" {
		dir = xbin
	} else if xbin := os.Getenv("XGET_BIN"); xbin != "" {
		dir = xbin
	}

	if to != "" && !strings.ContainsRune(to, os.PathSeparator) {
		bin = to
	} else if to != "" && !IsDirectory(to) {
		file = to
	}

	if file == "" {
		file = filepath.Join(dir, bin)
	}
	safeFile, err := cleanLocalPath(file)
	if err != nil {
		return
	}
	// #nosec G703 -- cleanLocalPath rejects traversal segments for relative paths.
	fi, err := os.Stat(safeFile)
	if err != nil {
		return
	}
	return fi.ModTime()
}

func DownloadConfigRepositories(cfg *config.Config) error {
	hasError := false
	errorList := []error{}

	binary, err := os.Executable()
	if err != nil {
		binary = os.Args[0]
	}

	for name := range cfg.Repositories {
		// #nosec G204,G702 -- executes this binary with repository names as plain arguments.
		cmd := exec.Command(binary, name)
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			hasError = true
			errorList = append(errorList, err)
		}
	}

	if hasError {
		return fmt.Errorf("one or more errors occurred while downloading: %v", errorList)
	}
	return nil
}

func ListAvailable(target string, opts options.Flags) ([]string, error) {
	SetDisableSSL(opts.DisableSSL)
	finder, _, err := getFinder(target, &opts)
	if err != nil {
		return nil, err
	}
	return finder.Find()
}

// shouldRecordInstall reports whether a completed install should be written to
// the installed package store. --untracked keeps the run out of the store
// without touching any entry that already exists for the target.
func shouldRecordInstall(opts options.Flags) bool {
	return opts.Output != "-" && !opts.Untracked
}

func installedOptions(opts options.Flags) installed.Options {
	return installed.Options{
		Tag:            opts.Tag,
		Prerelease:     opts.Prerelease,
		DownloadSource: opts.Source,
		Output:         opts.Output,
		System:         opts.System,
		ExtractFile:    opts.ExtractFile,
		All:            opts.All,
		DownloadOnly:   opts.DLOnly,
		UpgradeOnly:    opts.UpgradeOnly,
		Asset:          opts.Asset,
		Ignore:         opts.Ignore,
		Verify:         opts.Verify,
	}
}

func finderVersion(finder Finder, opts options.Flags) string {
	switch f := finder.(type) {
	case *GithubAssetFinder:
		return f.ReleaseTag
	case *GithubSourceFinder:
		return f.Tag
	case *GitlabAssetFinder:
		return f.ReleaseTag
	case *GitlabSourceFinder:
		return f.Tag
	default:
		return opts.Tag
	}
}

func packageName(target string, finder Finder) string {
	switch f := finder.(type) {
	case *GithubAssetFinder:
		return f.Repo
	case *GithubSourceFinder:
		return f.Repo
	case *GitlabAssetFinder:
		return f.Repo
	case *GitlabSourceFinder:
		return f.Repo
	default:
		return target
	}
}

func packageSource(finder Finder, fallback string) string {
	switch f := finder.(type) {
	case *GithubAssetFinder:
		return f.Source.Name
	case *GithubSourceFinder:
		return f.Source.Name
	case *GitlabAssetFinder:
		return f.Source.Name
	case *GitlabSourceFinder:
		return f.Source.Name
	case *DirectAssetFinder:
		return "URL"
	default:
		return fallback
	}
}

// RefreshInstalledPackage looks up the newest release for pkg. The caller
// supplies the resolved options; Tag and UpgradeOnly are cleared here because
// either would prevent the newest release from being reported.
func RefreshInstalledPackage(pkg installed.Package, opts options.Flags) (installed.Package, error) {
	if strings.EqualFold(pkg.Source, "URL") {
		return pkg, nil
	}
	opts.Tag = ""
	opts.UpgradeOnly = false
	opts.Source = false
	if opts.SourceType == "" {
		opts.SourceType = pkg.Source
	}
	finder, _, err := getFinder(pkg.Name, &opts)
	if err != nil {
		return pkg, err
	}
	if _, err := finder.Find(); err != nil {
		return pkg, err
	}
	version := finderVersion(finder, opts)
	if version != "" {
		pkg.CurrentTag = version
	}
	pkg.RefreshedAt = time.Now().UTC()
	return pkg, nil
}

func Run(target string, opts options.Flags) error {
	SetDisableSSL(opts.DisableSSL)

	if opts.DisableSSL {
		fmt.Fprintln(os.Stderr, "warning: SSL verification is disabled")
	}

	if opts.Remove {
		xbin := os.Getenv("XGET_BIN")
		if xbin == "" {
			xbin = "."
		}
		targetName, err := safeBaseName(target)
		if err != nil {
			return err
		}
		removePath := filepath.Join(xbin, targetName)
		removePath, err = cleanLocalPath(removePath)
		if err != nil {
			return err
		}
		// #nosec G703 -- removePath uses validated basename joined to configured bin directory.
		err = os.Remove(removePath)
		if err != nil {
			return err
		}
		fmt.Printf("Removed `%s`\n", removePath)
		return nil
	}

	var output io.Writer = os.Stderr
	if opts.Quiet {
		output = io.Discard
	}

	finder, tool, err := getFinder(target, &opts)
	if err != nil {
		return err
	}
	assets, err := finder.Find()
	if err != nil {
		if errors.Is(err, ErrNoUpgrade) {
			_, _ = fmt.Fprintf(output, "%s: %v\n", target, err)
			return nil
		}
		return err
	}

	detector, err := getDetector(&opts)
	if err != nil {
		return err
	}

	url, candidates, err := detector.Detect(assets)
	if len(candidates) != 0 && err != nil {
		fmt.Fprintf(os.Stderr, "%v: please select manually\n", err)
		choices := make([]interface{}, len(candidates))
		for i := range candidates {
			choices[i] = path.Base(candidates[i])
		}
		choice, err := userSelect(choices)
		if err != nil {
			return err
		}
		url = candidates[choice-1]
	} else if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(output, "%s\n", url); err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	err = DownloadWithSource(url, buf, func(size int64) *pb.ProgressBar {
		var pbout io.Writer = os.Stderr
		if opts.Quiet {
			pbout = io.Discard
		}
		return pb.NewOptions64(size,
			pb.OptionSetWriter(pbout),
			pb.OptionShowBytes(true),
			pb.OptionSetWidth(10),
			pb.OptionThrottle(65*time.Millisecond),
			pb.OptionShowCount(),
			pb.OptionSpinnerType(14),
			pb.OptionFullWidth(),
			pb.OptionSetDescription("Downloading"),
			pb.OptionOnCompletion(func() {
				_, _ = fmt.Fprint(pbout, "\n")
			}),
			pb.OptionSetTheme(pb.Theme{
				Saucer:        "=",
				SaucerHead:    ">",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}))
	}, opts.SourceConfig)
	if err != nil {
		return fmt.Errorf("%s (URL: %s)", err, url)
	}

	body := buf.Bytes()
	sumAsset := checksumAsset(url, assets)
	githubDigest := ""
	if finder, ok := finder.(*GithubAssetFinder); ok {
		githubDigest = finder.Digests[url]
	}
	assetSHA256 := githubDigest
	if assetSHA256 == "" {
		sum := sha256.Sum256(body)
		assetSHA256 = hex.EncodeToString(sum[:])
	}
	verifier, err := getVerifier(sumAsset, githubDigest, &opts)
	if err != nil {
		return err
	}
	if err = verifier.Verify(body); err != nil {
		return err
	}
	if opts.Verify == "" && sumAsset != "" {
		_, _ = fmt.Fprintf(output, "Checksum verified with %s\n", path.Base(sumAsset))
	} else if opts.Verify != "" {
		_, _ = fmt.Fprintln(output, "Checksum verified")
	}

	extractor, err := getExtractor(url, tool, &opts)
	if err != nil {
		return err
	}

	bin, bins, err := extractor.Extract(body, opts.All)
	if len(bins) != 0 && err != nil && !opts.All {
		fmt.Fprintf(os.Stderr, "%v: please select manually\n", err)
		choices := make([]interface{}, len(bins)+1)
		for i := range bins {
			choices[i] = bins[i]
		}
		choices[len(bins)] = "all"
		choice, err := userSelect(choices)
		if err != nil {
			return err
		}
		if choice == len(bins)+1 {
			opts.All = true
		} else {
			bin = bins[choice-1]
		}
	} else if err != nil && len(bins) == 0 {
		return err
	}
	if len(bins) == 0 {
		bins = []ExtractedFile{bin}
	}
	extractedFiles := []string{}

	extract := func(bin ExtractedFile) error {
		mode := bin.Mode()
		out := filepath.Base(bin.Name)
		if opts.Output == "-" {
			out = "-"
		} else if opts.Output != "" && isDirectoryDestination(opts.Output) {
			out = filepath.Join(opts.Output, out)
		} else if opts.Output != "" && opts.All {
			err := os.MkdirAll(opts.Output, 0750)
			if err != nil {
				return err
			}
			out = filepath.Join(opts.Output, out)
		} else {
			if opts.Output != "" {
				out = opts.Output
			}
			if os.Getenv("EGET_BIN") != "" && !strings.ContainsRune(out, os.PathSeparator) && mode&0111 != 0 && !bin.Dir {
				out = filepath.Join(os.Getenv("EGET_BIN"), out)
			}
			if os.Getenv("XGET_BIN") != "" && !strings.ContainsRune(out, os.PathSeparator) && mode&0111 != 0 && !bin.Dir {
				out = filepath.Join(os.Getenv("XGET_BIN"), out)
			}
		}

		executable, executableErr := os.Executable()
		selfUpdate := executableErr == nil && isRunningExecutableDestination(out, executable, runtime.GOOS == "windows")
		extractionTarget := out
		if selfUpdate {
			extractionTarget += ".new"
		}
		if err := bin.Extract(extractionTarget); err != nil {
			return err
		}
		if selfUpdate {
			if err := replaceStagedExecutable(out); err != nil {
				return err
			}
		}
		extractedFiles = append(extractedFiles, out)
		_, err := fmt.Fprintf(output, "Extracted `%s` to `%s`\n", bin.ArchiveName, out)
		return err
	}

	recordInstall := func() error {
		if !shouldRecordInstall(opts) {
			return nil
		}
		storePath, err := installed.DefaultPath()
		if err != nil {
			return err
		}
		version := finderVersion(finder, opts)
		now := time.Now().UTC()
		installLocation, err := resolvedInstallLocation(extractedFiles)
		if err != nil {
			return err
		}
		pkg := installed.Package{
			Name:            packageName(target, finder),
			InstallLocation: installLocation,
			InstalledAt:     now,
			DownloadURL:     url,
			Asset:           path.Base(url),
			ExtractedFiles:  extractedFiles,
			Options:         installedOptions(opts),
			RefreshedAt:     now,
			CurrentTag:      version,
			InstalledTag:    version,
			Source:          packageSource(finder, opts.SourceType),
			SHA256:          assetSHA256,
		}
		return installed.Upsert(storePath, pkg)
	}
	if opts.All {
		for _, eb := range bins {
			if err := extract(eb); err != nil {
				return err
			}
		}
		return recordInstall()
	}
	if err := extract(bin); err != nil {
		return err
	}
	return recordInstall()
}

func resolvedInstallLocation(extractedFiles []string) (string, error) {
	if len(extractedFiles) == 0 {
		return "", nil
	}
	return filepath.Abs(filepath.Dir(extractedFiles[0]))
}
