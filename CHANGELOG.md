## [v2.0.0](https://github.com/camalot/xget/releases/tag/v2.0.0) - 2026-09-04

### 🚀 FEATURES

#### _GENERAL_

- Android specific build -[@camalot](https://github.com/camalot)

- Implement xget config commands to manage configuration similar to git config commands -[@camalot](https://github.com/camalot)

- Xget upgrade functionality -[@camalot](https://github.com/camalot)

- Install and uninstall command paths -[@camalot](https://github.com/camalot)

- Support for org/reop@tag -[@camalot](https://github.com/camalot)

- Support xget rate command as well as xget --rate -[@camalot](https://github.com/camalot)

- Load .env / .secrets file if they exist -[@camalot](https://github.com/camalot)


### 🐛 BUG FIXES

#### _GENERAL_

- Update install scripts to validate their checksums when running -[@camalot](https://github.com/camalot)

- Update ps1 check of checksum -[@camalot](https://github.com/camalot)

- Update bash installer script to validate its checksum before running -[@camalot](https://github.com/camalot)

- Update to clear downloaded sha file -[@camalot](https://github.com/camalot)

- Treat MINGW and MSYS as windows -[@camalot](https://github.com/camalot)

- Provide feedback info when checksum fails -[@camalot](https://github.com/camalot)

- Better binary and man file install -[@camalot](https://github.com/camalot)

- Improving powershell install script -[@camalot](https://github.com/camalot)

- Improving powershell install script -[@camalot](https://github.com/camalot)

- Added filter for bat -[@camalot](https://github.com/camalot)

- Verify sha256 of scripts on commit -[@camalot](https://github.com/camalot)

- Check all files before exiting -[@camalot](https://github.com/camalot)

- Update sha for install scripts and run verify -[@camalot](https://github.com/camalot)

- Setup xget action test -[@camalot](https://github.com/camalot)

- Merge branch -[@camalot](https://github.com/camalot)

- Xget check version fix -[@camalot](https://github.com/camalot)

- Filters needed for dog -[@camalot](https://github.com/camalot)

- Dog does not have sha256 so need to skip verify -[@camalot](https://github.com/camalot)

- Do not fail if no sha256 missing from github. skip verifiy -[@camalot](https://github.com/camalot)

- Use different tool to test -[@camalot](https://github.com/camalot)

- Prerelease should not be global config -[@camalot](https://github.com/camalot)

- Alignment of output for `list --installed` -[@camalot](https://github.com/camalot)

- Output information when ratelimit hit to stderr -[@camalot](https://github.com/camalot)

- Set archive type for android package -[@camalot](https://github.com/camalot)


### 💼 OTHER

#### _GENERAL_

- Merge pull request  from camalot/commit-checks[#8](https://github.com/camalot/xget/issues/8)  [#8](https://github.com/camalot/xget/pull/8) -[@camalot](https://github.com/camalot)

- Commit checks [#8](https://github.com/camalot/xget/pull/8) -[@camalot](https://github.com/camalot)


### 📚 DOCUMENTATION

#### _GENERAL_

- Added info about the xget-action -[@camalot](https://github.com/camalot)


## GitHub

### 💛 Contributors


- [@camalot](https://github.com/camalot)
## 📈 Commit Statistics


- `34` commits contributed to the release.
- `2` days have passed between the first and last commit.
- `32` commits parsed as conventional.
- `1` linked issue detected in commits.
  - [#8](https://github.com/camalot/xget/issues/8) (referenced 1 time)
- `2` days  have passed between releases.


![Statistics](https://quickchart.io/chart?c={type:'bar',data:{labels:['Commits','Contributors','Days%20Between%20Commits','Conventional%20Commits','Referenced%20Links','Days%20Since%20Last%20Release'],datasets:[{label:'Release',data:[34,1,2,32,1,2]}]}})



---


**Full Changelog**: https://github.com/camalot/xget/compare/v2.0.0-beta...v2.0.0

## [v2.0.0-beta](https://github.com/camalot/xget/releases/tag/v2.0.0-beta) - 2026-09-02

### 🚀 FEATURES


#### _CLI_
- Display target when no upgrade is available ()[#87](https://github.com/camalot/xget/issues/87)  -[@hhromic](https://github.com/hhromic)


#### _CONFIG_
- Use system defaults for config fallback ()[#66](https://github.com/camalot/xget/issues/66)  -[@daylinmorgan](https://github.com/daylinmorgan)

- Use system defaults for config fallback -[@daylinmorgan](https://github.com/daylinmorgan)

- Report config load errors when files exist ()[#85](https://github.com/camalot/xget/issues/85)  -[@hhromic](https://github.com/hhromic)

- Report config load errors when files exist -[@hhromic](https://github.com/hhromic)


#### _DOWNLOAD_
- Report non-OK HTTP status codes when downloading ()[#90](https://github.com/camalot/xget/issues/90)  -[@hhromic](https://github.com/hhromic)


#### _RATE_
- Also display reset time as duration from now when available ()[#86](https://github.com/camalot/xget/issues/86)  -[@hhromic](https://github.com/hhromic)

#### _GENERAL_

- Added --ignore flag -[@camalot](https://github.com/camalot)

- Added regex support for --asset -[@camalot](https://github.com/camalot)

- Added re:/not:/text: syntax for --asset and --ignore -[@camalot](https://github.com/camalot)

- Add template variable support in asset_filters and ignore -[@camalot](https://github.com/camalot)

- Package install tracking -[@camalot](https://github.com/camalot)


### 🐛 BUG FIXES


#### _CONFIG_
- Add blank eget.toml file to override config in tests -[@daylinmorgan](https://github.com/daylinmorgan)

- Eget needs to try to load a config file -[@daylinmorgan](https://github.com/daylinmorgan)

- Expand home in GitHub token file read and use token for rate ()[#88](https://github.com/camalot/xget/issues/88)  -[@hhromic](https://github.com/hhromic)


#### _REPO_
- Remove unused and incorrect steps -[@camalot](https://github.com/camalot)

#### _GENERAL_

- Fix some man page formatting -[@zyedidia](https://github.com/zyedidia)

- Fix renaming -[@zyedidia](https://github.com/zyedidia)

- Fix test env variable -[@zyedidia](https://github.com/zyedidia)

- Fix test -[@zyedidia](https://github.com/zyedidia)

- Fixes [#9](https://github.com/camalot/xget/issues/9)  -[@zyedidia](https://github.com/zyedidia)

- Fixes [#13](https://github.com/camalot/xget/issues/13)  -[@zyedidia](https://github.com/zyedidia)

- Fix bug with zip archive -[@zyedidia](https://github.com/zyedidia)

- Fix bug with directory extraction -[@zyedidia](https://github.com/zyedidia)

- Fix off-by-one error in zip extraction -[@zyedidia](https://github.com/zyedidia)

- Fixes [#18](https://github.com/camalot/xget/issues/18)  -[@zyedidia](https://github.com/zyedidia)

- Fixes [#30](https://github.com/camalot/xget/issues/30)  -[@zyedidia](https://github.com/zyedidia)

- Fixes [#34](https://github.com/camalot/xget/issues/34)  -[@dufferzafar](https://github.com/dufferzafar)

- Fix txz -[@zyedidia](https://github.com/zyedidia)

- Fixes [#29](https://github.com/camalot/xget/issues/29)  -[@dufferzafar](https://github.com/dufferzafar)

- Fix GitHub repo detection regex ()[#41](https://github.com/camalot/xget/issues/41)  -[@dufferzafar](https://github.com/dufferzafar)

- Fixes [#45](https://github.com/camalot/xget/issues/45)  -[@zyedidia](https://github.com/zyedidia)

- Fixes [#43](https://github.com/camalot/xget/issues/43)  -[@zyedidia](https://github.com/zyedidia)

- Fixes [#38](https://github.com/camalot/xget/issues/38)  -[@zyedidia](https://github.com/zyedidia)

- Fixes [#40](https://github.com/camalot/xget/issues/40)  -[@zyedidia](https://github.com/zyedidia)

- Fixes [#48](https://github.com/camalot/xget/issues/48)  -[@zyedidia](https://github.com/zyedidia)

- Fix link and empty directory extraction -[@zyedidia](https://github.com/zyedidia)

- Fix build when dist does not exist -[@zyedidia](https://github.com/zyedidia)

- Fix readme -[@zyedidia](https://github.com/zyedidia)

- Fixes [#60](https://github.com/camalot/xget/issues/60)  -[@zyedidia](https://github.com/zyedidia)

- Fixes [#60](https://github.com/camalot/xget/issues/60)  -[@zyedidia](https://github.com/zyedidia)

- Fixes [#71](https://github.com/camalot/xget/issues/71)  -[@zyedidia](https://github.com/zyedidia)

- Fixes [#70](https://github.com/camalot/xget/issues/70)  -[@zyedidia](https://github.com/zyedidia)

- Fix for github token reading -[@zyedidia](https://github.com/zyedidia)

- Fixes [#102](https://github.com/camalot/xget/issues/102)  -[@zyedidia](https://github.com/zyedidia)

- Fixes [#100](https://github.com/camalot/xget/issues/100)  -[@zyedidia](https://github.com/zyedidia)

- Fixes [#74](https://github.com/camalot/xget/issues/74)  -[@zyedidia](https://github.com/zyedidia)

- Fixes [#95](https://github.com/camalot/xget/issues/95)  -[@zyedidia](https://github.com/zyedidia)

- Fix CGO disable in build script -[@zyedidia](https://github.com/zyedidia)

- Move --version from flag to subcommand -[@camalot](https://github.com/camalot)

- Gofmt fixes -[@camalot](https://github.com/camalot)

- Resolve checksum parsing issue -[@camalot](https://github.com/camalot)


### 💼 OTHER

#### _GENERAL_

- Initial commit -[@zyedidia](https://github.com/zyedidia)

- Better support for different compression -[@zyedidia](https://github.com/zyedidia)

- Prompt for multiple candidates -[@zyedidia](https://github.com/zyedidia)

- Prompt for downloading and better candidates -[@zyedidia](https://github.com/zyedidia)

- Download progress bar -[@zyedidia](https://github.com/zyedidia)

- More systems -[@zyedidia](https://github.com/zyedidia)

- Better target file extraction -[@zyedidia](https://github.com/zyedidia)

- Add tools and man file -[@zyedidia](https://github.com/zyedidia)

- Use posix flags -[@zyedidia](https://github.com/zyedidia)

- Add stub readme -[@zyedidia](https://github.com/zyedidia)

- Add license -[@zyedidia](https://github.com/zyedidia)

- Change message -[@zyedidia](https://github.com/zyedidia)

- Add download-only and url flags -[@zyedidia](https://github.com/zyedidia)

- Better extraction and comments -[@zyedidia](https://github.com/zyedidia)

- Add sha256 hash option -[@zyedidia](https://github.com/zyedidia)

- Quiet option and better progress bar -[@zyedidia](https://github.com/zyedidia)

- Add man file -[@zyedidia](https://github.com/zyedidia)

- X and --asset and basic readme -[@zyedidia](https://github.com/zyedidia)

- Dit -[@zyedidia](https://github.com/zyedidia)

- Remove -y option -[@zyedidia](https://github.com/zyedidia)

- Add --rename option -[@zyedidia](https://github.com/zyedidia)

- Update readme -[@zyedidia](https://github.com/zyedidia)

- Add demo to readme -[@zyedidia](https://github.com/zyedidia)

- Update demo link -[@zyedidia](https://github.com/zyedidia)

- Expand readme and fix bug -[@zyedidia](https://github.com/zyedidia)

- Packaging and cross compilation -[@zyedidia](https://github.com/zyedidia)

- Rename to eget -[@zyedidia](https://github.com/zyedidia)

- Improve name references -[@zyedidia](https://github.com/zyedidia)

- Remove --url option -[@zyedidia](https://github.com/zyedidia)

- Improve -[@zyedidia](https://github.com/zyedidia)

- V0.1.0 -[@zyedidia](https://github.com/zyedidia)

- Add quick install -[@zyedidia](https://github.com/zyedidia)

- Minor -[@zyedidia](https://github.com/zyedidia)

- Add release badge -[@zyedidia](https://github.com/zyedidia)

- Update readme -[@zyedidia](https://github.com/zyedidia)

- Update readme -[@zyedidia](https://github.com/zyedidia)

- Ignore build tools and add go reportcard -[@zyedidia](https://github.com/zyedidia)

- Update readme -[@zyedidia](https://github.com/zyedidia)

- Version 0.1.1 -[@zyedidia](https://github.com/zyedidia)

- Update readme -[@zyedidia](https://github.com/zyedidia)

- Simplify download function -[@zyedidia](https://github.com/zyedidia)

- Remove --rename (subsumed by --to) -[@zyedidia](https://github.com/zyedidia)

- Update readme -[@zyedidia](https://github.com/zyedidia)

- Add tests -[@zyedidia](https://github.com/zyedidia)

- Relax --asset to be more useful -[@zyedidia](https://github.com/zyedidia)

- V0.1.2 -[@zyedidia](https://github.com/zyedidia)

- Update readme -[@zyedidia](https://github.com/zyedidia)

- Merge -[@zyedidia](https://github.com/zyedidia)

- Add eget target to makefile -[@zyedidia](https://github.com/zyedidia)

- Improve --asset -[@zyedidia](https://github.com/zyedidia)

- V0.1.3 -[@zyedidia](https://github.com/zyedidia)

- Better error message for invalid repo argument -[@zyedidia](https://github.com/zyedidia)

- Better detection for windows .exe files -[@zyedidia](https://github.com/zyedidia)

- Simplify AllDetector -[@zyedidia](https://github.com/zyedidia)

- Merge -[@zyedidia](https://github.com/zyedidia)

- Specify system for pandoc test -[@zyedidia](https://github.com/zyedidia)

- Update readme -[@zyedidia](https://github.com/zyedidia)

- Update readme regarding curl to bash install -[@zyedidia](https://github.com/zyedidia)

- Auto-download single asset even if it is not a match -[@zyedidia](https://github.com/zyedidia)

- Mark single files executable (no extract) -[@zyedidia](https://github.com/zyedidia)

- Better macos and x86 detection -[@zyedidia](https://github.com/zyedidia)

- Add documentation -[@zyedidia](https://github.com/zyedidia)

- Ref [#3](https://github.com/camalot/xget/issues/3)  -[@zyedidia](https://github.com/zyedidia)

- Escape markdown pipes -[@zyedidia](https://github.com/zyedidia)

- Forgot one escape -[@zyedidia](https://github.com/zyedidia)

- Update readme -[@zyedidia](https://github.com/zyedidia)

- Better executable finder and extraction renamer -[@zyedidia](https://github.com/zyedidia)

- Better renaming and appimage priority -[@zyedidia](https://github.com/zyedidia)

- Ref [#1](https://github.com/camalot/xget/issues/1)  -[@zyedidia](https://github.com/zyedidia)

- Update readme -[@zyedidia](https://github.com/zyedidia)

- Update readme -[@zyedidia](https://github.com/zyedidia)

- Use $EGET_BIN for installation if available -[@zyedidia](https://github.com/zyedidia)

- If `$EGET_BIN` is non-empty, eget will place extracted executables into -[@zyedidia](https://github.com/zyedidia)

- That directory instead of the current directory. -[@zyedidia](https://github.com/zyedidia)

- Closes [#5](https://github.com/camalot/xget/issues/5)  -[@zyedidia](https://github.com/zyedidia)

- Some main function cleanup -[@zyedidia](https://github.com/zyedidia)

- Remove -x option thanks to better exec detection -[@zyedidia](https://github.com/zyedidia)

- Automatically verify sha256 checksums if present -[@zyedidia](https://github.com/zyedidia)

- Closes [#4](https://github.com/camalot/xget/issues/4)  -[@zyedidia](https://github.com/zyedidia)

- Update readme and man page -[@zyedidia](https://github.com/zyedidia)

- Option for including pre-releases in search -[@zyedidia](https://github.com/zyedidia)

- Use `--pre-release` to include pre-releases when getting the latest release. -[@zyedidia](https://github.com/zyedidia)

- Closes [#2](https://github.com/camalot/xget/issues/2)  -[@zyedidia](https://github.com/zyedidia)

- Update docs about EGET_BIN -[@zyedidia](https://github.com/zyedidia)

- Update docs and bz2 extension -[@zyedidia](https://github.com/zyedidia)

- Update readme -[@zyedidia](https://github.com/zyedidia)

- V0.2.0 -[@zyedidia](https://github.com/zyedidia)

- Support xz compression -[@zyedidia](https://github.com/zyedidia)

- Closes [#6](https://github.com/camalot/xget/issues/6)  -[@zyedidia](https://github.com/zyedidia)

- Detect aarch64 and use bufio for xz -[@zyedidia](https://github.com/zyedidia)

- Support directory names in filename extraction -[@zyedidia](https://github.com/zyedidia)

- Specify rename field for DLOnly option -[@zyedidia](https://github.com/zyedidia)

- Allow asset chains -[@zyedidia](https://github.com/zyedidia)

- Update docs and fix issue -[@zyedidia](https://github.com/zyedidia)

- V0.3.0 -[@zyedidia](https://github.com/zyedidia)

- Allow GitHub authorization with `GITHUB_TOKEN` -[@zyedidia](https://github.com/zyedidia)

- Closes [#7](https://github.com/camalot/xget/issues/7)  -[@zyedidia](https://github.com/zyedidia)

- Allow "downloading" local files -[@zyedidia](https://github.com/zyedidia)

- Closes [#12](https://github.com/camalot/xget/issues/12)  -[@zyedidia](https://github.com/zyedidia)

- Add rate limiting information -[@zyedidia](https://github.com/zyedidia)

- Ref [#7](https://github.com/camalot/xget/issues/7)  -[@zyedidia](https://github.com/zyedidia)

- Version 0.3.1 -[@zyedidia](https://github.com/zyedidia)

- Remove target file if it exists -[@zyedidia](https://github.com/zyedidia)

- Add .tgz support ()[#17](https://github.com/camalot/xget/issues/17)  -[@daylinmorgan](https://github.com/daylinmorgan)

- Add .tgz support -[@daylinmorgan](https://github.com/daylinmorgan)

- Simplify matching switch case condition -[@daylinmorgan](https://github.com/daylinmorgan)

- Unify archive access -[@zyedidia](https://github.com/zyedidia)

- Directory extraction (needs testing) -[@zyedidia](https://github.com/zyedidia)

- Update docs -[@zyedidia](https://github.com/zyedidia)

- Add --all option -[@zyedidia](https://github.com/zyedidia)

- Allow directory extraction -[@zyedidia](https://github.com/zyedidia)

- Allow specifying all at prompt -[@zyedidia](https://github.com/zyedidia)

- Auto mkdir when using --all -[@zyedidia](https://github.com/zyedidia)

- Update docs -[@zyedidia](https://github.com/zyedidia)

- Don't select dirs with binary chooser -[@zyedidia](https://github.com/zyedidia)

- Clean tests -[@zyedidia](https://github.com/zyedidia)

- Don't put extracted dirs in eget bin -[@zyedidia](https://github.com/zyedidia)

- Update docs -[@zyedidia](https://github.com/zyedidia)

- Update readme -[@zyedidia](https://github.com/zyedidia)

- Version 1.0.0 -[@zyedidia](https://github.com/zyedidia)

- Version 1.0.1 -[@zyedidia](https://github.com/zyedidia)

- Write info messages to stderr instead of stdout -[@zyedidia](https://github.com/zyedidia)

- Ref [#23](https://github.com/camalot/xget/issues/23)  -[@zyedidia](https://github.com/zyedidia)

- Allow full github URLs -[@zyedidia](https://github.com/zyedidia)

- Closes [#27](https://github.com/camalot/xget/issues/27)  -[@zyedidia](https://github.com/zyedidia)

- Enable building additional platforms ()[#28](https://github.com/camalot/xget/issues/28)  -[@neersighted](https://github.com/neersighted)

- Enable darwin-arm64 for native Apple Silicon support, as well as FreeBSD -[@neersighted](https://github.com/neersighted)

- (a very common server platform). -[@neersighted](https://github.com/neersighted)

- Update README.md ()[#26](https://github.com/camalot/xget/issues/26)  -[@masta99wrfvshsa](https://github.com/masta99wrfvshsa)

- Update readme -[@zyedidia](https://github.com/zyedidia)

- Support option to only download on upgrade -[@zyedidia](https://github.com/zyedidia)

- Closes [#14](https://github.com/camalot/xget/issues/14)  -[@zyedidia](https://github.com/zyedidia)

- Exclude android from linux matches -[@zyedidia](https://github.com/zyedidia)

- Prefer --to with --upgrade-only ()[#35](https://github.com/camalot/xget/issues/35)  -[@dufferzafar](https://github.com/dufferzafar)

- Impove --to interaction with --upgrade-only -[@zyedidia](https://github.com/zyedidia)

- Handle archives with .tbz, .txz extensions ()[#37](https://github.com/camalot/xget/issues/37)  -[@dufferzafar](https://github.com/dufferzafar)

- This is the first time I've even heard of "tbz -[@dufferzafar](https://github.com/dufferzafar)

- But apparently, it is a thing: -[@dufferzafar](https://github.com/dufferzafar)

- //github.com/aristocratos/btop/releases/tag/v1.2.7 -[@dufferzafar](https://github.com/dufferzafar)

- Use ProxyFromEnvironment for downloading ()[#36](https://github.com/camalot/xget/issues/36)  -[@dufferzafar](https://github.com/dufferzafar)

- Taken from: https://stackoverflow.com/a/25102190 -[@dufferzafar](https://github.com/dufferzafar)

- Https://www.debuggex.com/r/fFggA8Uc4YYKjl34 -[@dufferzafar](https://github.com/dufferzafar)

- Add homebrew install section to readme -[@zyedidia](https://github.com/zyedidia)

- Update version.go to 1.1.0 -[@zyedidia](https://github.com/zyedidia)

- Ref [#44](https://github.com/camalot/xget/issues/44)  -[@zyedidia](https://github.com/zyedidia)

- A complete resolution for  will require a new release.[#44](https://github.com/camalot/xget/issues/44)  -[@zyedidia](https://github.com/zyedidia)

- Support symbolic and hard links in extracted dirs -[@zyedidia](https://github.com/zyedidia)

- Add --remove flag for removing binaries -[@zyedidia](https://github.com/zyedidia)

- Apply the system detector to --asset if applicable -[@zyedidia](https://github.com/zyedidia)

- Add ^ syntax for anti-match -[@zyedidia](https://github.com/zyedidia)

- Hopefully there aren't any assets that actually begin with a ^. -[@zyedidia](https://github.com/zyedidia)

- V1.2.0 -[@zyedidia](https://github.com/zyedidia)

- Add --source flag for downloading repo source code -[@zyedidia](https://github.com/zyedidia)

- Add openbsd to build all -[@zyedidia](https://github.com/zyedidia)

- Allow --to - to send to stdout -[@zyedidia](https://github.com/zyedidia)

- Recognize ubuntu as linux -[@zyedidia](https://github.com/zyedidia)

- Merge branch 'master' of https://github.com/zyedidia/eget -[@zyedidia](https://github.com/zyedidia)

- Update readme -[@zyedidia](https://github.com/zyedidia)

- Allow trailing slashes in github urls -[@zyedidia](https://github.com/zyedidia)

- Merge branch 'master' of https://github.com/zyedidia/eget -[@zyedidia](https://github.com/zyedidia)

- Improve mac detection -[@zyedidia](https://github.com/zyedidia)

- Add concurrent builds ()[#52](https://github.com/camalot/xget/issues/52)  -[@patinthehat](https://github.com/patinthehat)

- Add build-dist task, package task now uses the dist directory for all outputs, eget.1 task now outputs to dist directory -[@patinthehat](https://github.com/patinthehat)

- Use goroutines to build all targets concurrently, significantly decreasing overall build time on multi-core cpus -[@patinthehat](https://github.com/patinthehat)

- Add support for eget environment variable ()[#51](https://github.com/camalot/xget/issues/51)  -[@patinthehat](https://github.com/patinthehat)

- Add dotenv file loading -[@patinthehat](https://github.com/patinthehat)

- Add dotenv package -[@patinthehat](https://github.com/patinthehat)

- Add support for either EGET_GITHUB_TOKEN or GITHUB_TOKEN env vars -[@patinthehat](https://github.com/patinthehat)

- Update docs -[@patinthehat](https://github.com/patinthehat)

- Wip -[@patinthehat](https://github.com/patinthehat)

- Remove dotenv support -[@patinthehat](https://github.com/patinthehat)

- Update docs -[@patinthehat](https://github.com/patinthehat)

- Improve extraction performance for * -[@zyedidia](https://github.com/zyedidia)

- Update src version number -[@zyedidia](https://github.com/zyedidia)

- Update README.md ()[#54](https://github.com/camalot/xget/issues/54)  -[@ksz16](https://github.com/ksz16)

- Exit with an error if scanf fails ()[#59](https://github.com/camalot/xget/issues/59)  -[@larsks](https://github.com/larsks)

- Exit if scanf returns an EOF error, rather than going into an infinite loop -[@larsks](https://github.com/larsks)

- If we are unable to read from stdin. -[@larsks](https://github.com/larsks)

- Closes [#58](https://github.com/camalot/xget/issues/58)  -[@larsks](https://github.com/larsks)

- Add config file support ()[#53](https://github.com/camalot/xget/issues/53)  -[@patinthehat](https://github.com/patinthehat)

- Add support for toml config file -[@patinthehat](https://github.com/patinthehat)

- Ignore test config file -[@patinthehat](https://github.com/patinthehat)

- Start config filename with a dot, check both home dir and eget bin dir for config file -[@patinthehat](https://github.com/patinthehat)

- Wip -[@patinthehat](https://github.com/patinthehat)

- Remove debug code -[@patinthehat](https://github.com/patinthehat)

- Wip -[@patinthehat](https://github.com/patinthehat)

- Wip -[@patinthehat](https://github.com/patinthehat)

- Update readme with info about configuration file -[@patinthehat](https://github.com/patinthehat)

- Update readme -[@patinthehat](https://github.com/patinthehat)

- Update manual docs -[@patinthehat](https://github.com/patinthehat)

- Add additional config<->flags support -[@patinthehat](https://github.com/patinthehat)

- Update readme -[@patinthehat](https://github.com/patinthehat)

- Update readme -[@patinthehat](https://github.com/patinthehat)

- Update readme -[@patinthehat](https://github.com/patinthehat)

- If direct tag is not found, look for tags that contain it -[@zyedidia](https://github.com/zyedidia)

- Ref [#56](https://github.com/camalot/xget/issues/56)  -[@zyedidia](https://github.com/zyedidia)

- Override config options with CLI flags if set -[@zyedidia](https://github.com/zyedidia)

- Update man page -[@zyedidia](https://github.com/zyedidia)

- V1.3.0 -[@zyedidia](https://github.com/zyedidia)

- Bump src version -[@zyedidia](https://github.com/zyedidia)

- Disable cgo in build script except on darwin -[@zyedidia](https://github.com/zyedidia)

- Ref [#60](https://github.com/camalot/xget/issues/60)  -[@zyedidia](https://github.com/zyedidia)

- Expand ~ in configuration -[@zyedidia](https://github.com/zyedidia)

- Actually fix tilde expansion in config file -[@zyedidia](https://github.com/zyedidia)

- Add downloading of configuration-defined projects ()[#62](https://github.com/camalot/xget/issues/62)  -[@patinthehat](https://github.com/patinthehat)

- Add download config flag -[@patinthehat](https://github.com/patinthehat)

- Add implementation of download config flag functionality -[@patinthehat](https://github.com/patinthehat)

- Update docs with info on -D/--download-config flag -[@patinthehat](https://github.com/patinthehat)

- Update flag name to --download-all -[@patinthehat](https://github.com/patinthehat)

- Use os.Executable() to get current binary -[@patinthehat](https://github.com/patinthehat)

- Redirect cmd.stderr to os.stderr instead of using a scanner -[@patinthehat](https://github.com/patinthehat)

- Wip -[@patinthehat](https://github.com/patinthehat)

- Display specific errors that occurred during batch download -[@patinthehat](https://github.com/patinthehat)

- Code cleanup -[@patinthehat](https://github.com/patinthehat)

- Update docs -[@patinthehat](https://github.com/patinthehat)

- V1.3.1 -[@zyedidia](https://github.com/zyedidia)

- If err is still nil because it skipped the first check then it -[@daylinmorgan](https://github.com/daylinmorgan)

- Would never generate the default config and cause segmentation fault. -[@daylinmorgan](https://github.com/daylinmorgan)

- Add FAQ section about monorepos -[@zyedidia](https://github.com/zyedidia)

- Closes [#67](https://github.com/camalot/xget/issues/67)  -[@zyedidia](https://github.com/zyedidia)

- V1.3.2 -[@zyedidia](https://github.com/zyedidia)

- Zstd support -[@zyedidia](https://github.com/zyedidia)

- Closes [#69](https://github.com/camalot/xget/issues/69)  -[@zyedidia](https://github.com/zyedidia)

- Zstd -> zst -[@zyedidia](https://github.com/zyedidia)

- Better detection for arm32 -[@zyedidia](https://github.com/zyedidia)

- Ref [#72](https://github.com/camalot/xget/issues/72)  -[@zyedidia](https://github.com/zyedidia)

- Add verify_sha256 option to configuration -[@zyedidia](https://github.com/zyedidia)

- Update readme -[@zyedidia](https://github.com/zyedidia)

- Add option to disable SSL certificate verification -[@zyedidia](https://github.com/zyedidia)

- Support reading github token from file with @path -[@zyedidia](https://github.com/zyedidia)

- Closes [#82](https://github.com/camalot/xget/issues/82)  -[@zyedidia](https://github.com/zyedidia)

- Simplify if statements and error messages -[@hhromic](https://github.com/hhromic)

- Also trim newline and carriage-return from the read token file -[@hhromic](https://github.com/hhromic)

- Replace 'to' with 'target' in man page -[@zyedidia](https://github.com/zyedidia)

- Proper fallback to EGET_BIN if target not set -[@zyedidia](https://github.com/zyedidia)

- Always expand target -[@zyedidia](https://github.com/zyedidia)

- Update version -[@zyedidia](https://github.com/zyedidia)

- Ref [#105](https://github.com/camalot/xget/issues/105)  -[@zyedidia](https://github.com/zyedidia)


### 🚜 REFACTOR


#### _CONFIG_
- Make .config the default for non-windows -[@daylinmorgan](https://github.com/daylinmorgan)

#### _GENERAL_

- Refactor to drop viper and use a toml-specific library -[@patinthehat](https://github.com/patinthehat)

- Restructor to follow cobra/viper architecture -[@camalot](https://github.com/camalot)


### 📚 DOCUMENTATION

#### _GENERAL_

- Update readme documentation -[@camalot](https://github.com/camalot)

- Updated docs on how to install -[@camalot](https://github.com/camalot)


### ⚙️ MISCELLANEOUS TASKS


#### _DOCS_
- Document that GitHub tokens can be read from files ()[#89](https://github.com/camalot/xget/issues/89)  -[@hhromic](https://github.com/hhromic)


#### _REPO_
- Setup auto for prerelease in goreleaser -[@camalot](https://github.com/camalot)

- Correct action for push to main -[@camalot](https://github.com/camalot)

#### _GENERAL_

- First commit. WIP. resolving gosec issues still pending -[@camalot](https://github.com/camalot)

- Added test for what needs to be supported -[@camalot](https://github.com/camalot)


## GitHub

### ❤️ New Contributors

- [@camalot](https://github.com/camalot)
- [@zyedidia](https://github.com/zyedidia)
- [@hhromic](https://github.com/hhromic)
- [@daylinmorgan](https://github.com/daylinmorgan)
- [@patinthehat](https://github.com/patinthehat)
- [@larsks](https://github.com/larsks)
- [@ksz16](https://github.com/ksz16)
- [@dufferzafar](https://github.com/dufferzafar)
- [@masta99wrfvshsa](https://github.com/masta99wrfvshsa)
- [@neersighted](https://github.com/neersighted)
## 📈 Commit Statistics


- `300` commits contributed to the release.
- `1843` days have passed between the first and last commit.
- `30` commits parsed as conventional.
- `57` linked issues detected in commits.
  - [#60](https://github.com/camalot/xget/issues/60) (referenced 3 times)
  - [#44](https://github.com/camalot/xget/issues/44) (referenced 2 times)
  - [#7](https://github.com/camalot/xget/issues/7) (referenced 2 times)
  - [#1](https://github.com/camalot/xget/issues/1) (referenced 1 time)
  - [#100](https://github.com/camalot/xget/issues/100) (referenced 1 time)
  - [#102](https://github.com/camalot/xget/issues/102) (referenced 1 time)
  - [#105](https://github.com/camalot/xget/issues/105) (referenced 1 time)
  - [#12](https://github.com/camalot/xget/issues/12) (referenced 1 time)
  - [#13](https://github.com/camalot/xget/issues/13) (referenced 1 time)
  - [#14](https://github.com/camalot/xget/issues/14) (referenced 1 time)
  - [#17](https://github.com/camalot/xget/issues/17) (referenced 1 time)
  - [#18](https://github.com/camalot/xget/issues/18) (referenced 1 time)
  - [#2](https://github.com/camalot/xget/issues/2) (referenced 1 time)
  - [#23](https://github.com/camalot/xget/issues/23) (referenced 1 time)
  - [#26](https://github.com/camalot/xget/issues/26) (referenced 1 time)
  - [#27](https://github.com/camalot/xget/issues/27) (referenced 1 time)
  - [#28](https://github.com/camalot/xget/issues/28) (referenced 1 time)
  - [#29](https://github.com/camalot/xget/issues/29) (referenced 1 time)
  - [#3](https://github.com/camalot/xget/issues/3) (referenced 1 time)
  - [#30](https://github.com/camalot/xget/issues/30) (referenced 1 time)
  - [#34](https://github.com/camalot/xget/issues/34) (referenced 1 time)
  - [#35](https://github.com/camalot/xget/issues/35) (referenced 1 time)
  - [#36](https://github.com/camalot/xget/issues/36) (referenced 1 time)
  - [#37](https://github.com/camalot/xget/issues/37) (referenced 1 time)
  - [#38](https://github.com/camalot/xget/issues/38) (referenced 1 time)
  - [#4](https://github.com/camalot/xget/issues/4) (referenced 1 time)
  - [#40](https://github.com/camalot/xget/issues/40) (referenced 1 time)
  - [#41](https://github.com/camalot/xget/issues/41) (referenced 1 time)
  - [#43](https://github.com/camalot/xget/issues/43) (referenced 1 time)
  - [#45](https://github.com/camalot/xget/issues/45) (referenced 1 time)
  - [#48](https://github.com/camalot/xget/issues/48) (referenced 1 time)
  - [#5](https://github.com/camalot/xget/issues/5) (referenced 1 time)
  - [#51](https://github.com/camalot/xget/issues/51) (referenced 1 time)
  - [#52](https://github.com/camalot/xget/issues/52) (referenced 1 time)
  - [#53](https://github.com/camalot/xget/issues/53) (referenced 1 time)
  - [#54](https://github.com/camalot/xget/issues/54) (referenced 1 time)
  - [#56](https://github.com/camalot/xget/issues/56) (referenced 1 time)
  - [#58](https://github.com/camalot/xget/issues/58) (referenced 1 time)
  - [#59](https://github.com/camalot/xget/issues/59) (referenced 1 time)
  - [#6](https://github.com/camalot/xget/issues/6) (referenced 1 time)
  - [#62](https://github.com/camalot/xget/issues/62) (referenced 1 time)
  - [#66](https://github.com/camalot/xget/issues/66) (referenced 1 time)
  - [#67](https://github.com/camalot/xget/issues/67) (referenced 1 time)
  - [#69](https://github.com/camalot/xget/issues/69) (referenced 1 time)
  - [#70](https://github.com/camalot/xget/issues/70) (referenced 1 time)
  - [#71](https://github.com/camalot/xget/issues/71) (referenced 1 time)
  - [#72](https://github.com/camalot/xget/issues/72) (referenced 1 time)
  - [#74](https://github.com/camalot/xget/issues/74) (referenced 1 time)
  - [#82](https://github.com/camalot/xget/issues/82) (referenced 1 time)
  - [#85](https://github.com/camalot/xget/issues/85) (referenced 1 time)
  - [#86](https://github.com/camalot/xget/issues/86) (referenced 1 time)
  - [#87](https://github.com/camalot/xget/issues/87) (referenced 1 time)
  - [#88](https://github.com/camalot/xget/issues/88) (referenced 1 time)
  - [#89](https://github.com/camalot/xget/issues/89) (referenced 1 time)
  - [#9](https://github.com/camalot/xget/issues/9) (referenced 1 time)
  - [#90](https://github.com/camalot/xget/issues/90) (referenced 1 time)
  - [#95](https://github.com/camalot/xget/issues/95) (referenced 1 time)


![Statistics](https://quickchart.io/chart?c={type:'bar',data:{labels:['Commits','Contributors','Days%20Between%20Commits','Conventional%20Commits','Referenced%20Links','Days%20Since%20Last%20Release'],datasets:[{label:'Release',data:[300,10,1843,30,57,0]}]}})



---


**Full Changelog**: https://github.com/camalot/xget/compare/...v2.0.0-beta
<!-- https://cliff-notes.dev#new=true&template=cliff-notes-remote.toml -->
_💙 [cliff.toml](https://git-cliff.org) designed in [cliff-notes](https://cliff-notes.dev)_
