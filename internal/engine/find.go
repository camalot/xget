package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// A Finder returns a list of URLs making up a project's assets.
type Finder interface {
	Find() ([]string, error)
}

// A GithubRelease matches the Assets portion of Github's release API json.
type GithubRelease struct {
	Assets []struct {
		DownloadURL string `json:"browser_download_url"`
		Digest      string `json:"digest"`
	} `json:"assets"`

	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	Name        string    `json:"name"`
	Tag         string    `json:"tag_name"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type Release struct {
	Name        string
	Tag         string
	PublishedAt time.Time
}

type GithubError struct {
	Code   int
	Status string
	Body   []byte
	Url    string
}
type errResponse struct {
	Message string `json:"message"`
	Doc     string `json:"documentation_url"`
}

func (ge *GithubError) Error() string {
	var msg errResponse
	_ = json.Unmarshal(ge.Body, &msg)

	if ge.Code == http.StatusTooManyRequests && !GithubTokenConfigured() {
		return fmt.Sprintf("%s: GitHub API rate limit exceeded. Set XGET_GITHUB_TOKEN, GITHUB_TOKEN, or EGET_GITHUB_TOKEN to a GitHub token and try again.", ge.Status)
	}
	if ge.Code == http.StatusForbidden {
		return fmt.Sprintf("%s: %s: %s", ge.Status, msg.Message, msg.Doc)
	}
	return fmt.Sprintf("%s (URL: %s)", ge.Status, ge.Url)
}

// A GithubAssetFinder finds assets for the given Repo at the given tag. Tags
// must be given as 'tag/<tag>'. Use 'latest' to get the latest release.
type GithubAssetFinder struct {
	Repo       string
	Tag        string
	Prerelease bool
	MinTime    time.Time // release must be after MinTime to be found
	Digests    map[string]string
	ReleaseTag string
}

var ErrNoUpgrade = errors.New("requested release is not more recent than current version")

func ListReleases(repo string, includePrereleases bool) ([]Release, error) {
	if strings.Count(repo, "/") != 1 {
		return nil, fmt.Errorf("invalid argument (must be of the form user/repo)")
	}

	const limit = 10
	releases := make([]Release, 0, limit)
	for page := 1; len(releases) < limit; page++ {
		url := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=100&page=%d", repo, page)
		resp, err := Get(url)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			body, readErr := io.ReadAll(resp.Body)
			closeErr := resp.Body.Close()
			if readErr != nil {
				return nil, readErr
			}
			if closeErr != nil {
				return nil, closeErr
			}
			return nil, &GithubError{Status: resp.Status, Code: resp.StatusCode, Body: body, Url: url}
		}
		body, err := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}

		var pageReleases []GithubRelease
		if err := json.Unmarshal(body, &pageReleases); err != nil {
			return nil, err
		}
		for _, release := range pageReleases {
			if release.Draft || (!includePrereleases && release.Prerelease) {
				continue
			}
			publishedAt := release.PublishedAt
			if publishedAt.IsZero() {
				publishedAt = release.CreatedAt
			}
			releases = append(releases, Release{Name: release.Name, Tag: release.Tag, PublishedAt: publishedAt})
			if len(releases) == limit {
				return releases, nil
			}
		}
		if len(pageReleases) < 100 {
			break
		}
	}
	return releases, nil
}

func (f *GithubAssetFinder) Find() ([]string, error) {
	if f.Prerelease && f.Tag == "latest" {
		tag, err := f.getLatestTag()
		if err != nil {
			return nil, err
		}
		f.Tag = fmt.Sprintf("tags/%s", tag)
	}

	// query github's API for this repo/tag pair.
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/%s", f.Repo, f.Tag)
	resp, err := Get(url)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("error closing response body:", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(f.Tag, "tags/") && resp.StatusCode == http.StatusNotFound {
			return f.FindMatch()
		}
		return nil, &GithubError{
			Status: resp.Status,
			Code:   resp.StatusCode,
			Body:   body,
			Url:    url,
		}
	}

	// read and unmarshal the resulting json
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var release GithubRelease
	err = json.Unmarshal(body, &release)
	if err != nil {
		return nil, err
	}

	if release.CreatedAt.Before(f.MinTime) {
		return nil, ErrNoUpgrade
	}
	f.ReleaseTag = release.Tag

	// accumulate all assets from the json into a slice
	assets := make([]string, 0, len(release.Assets))
	if f.Digests == nil {
		f.Digests = map[string]string{}
	}
	for _, a := range release.Assets {
		assets = append(assets, a.DownloadURL)
		if strings.HasPrefix(a.Digest, "sha256:") {
			f.Digests[a.DownloadURL] = strings.TrimPrefix(a.Digest, "sha256:")
		}
	}

	return assets, nil
}

func (f *GithubAssetFinder) FindMatch() ([]string, error) {
	tag := f.Tag[len("tags/"):]

	for page := 1; ; page++ {
		url := fmt.Sprintf("https://api.github.com/repos/%s/releases?page=%d", f.Repo, page)
		resp, err := Get(url)
		if err != nil {
			return nil, err
		}

		defer func() {
			if err := resp.Body.Close(); err != nil {
				fmt.Println("error closing response body:", err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			return nil, &GithubError{
				Status: resp.Status,
				Code:   resp.StatusCode,
				Body:   body,
				Url:    url,
			}
		}

		// read and unmarshal the resulting json
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var releases []GithubRelease
		err = json.Unmarshal(body, &releases)
		if err != nil {
			return nil, err
		}

		for _, r := range releases {
			if !f.Prerelease && r.Prerelease {
				continue
			}
			if strings.Contains(r.Tag, tag) && !r.CreatedAt.Before(f.MinTime) {
				f.ReleaseTag = r.Tag
				// we have a winner
				assets := make([]string, 0, len(r.Assets))
				if f.Digests == nil {
					f.Digests = map[string]string{}
				}
				for _, a := range r.Assets {
					assets = append(assets, a.DownloadURL)
					if strings.HasPrefix(a.Digest, "sha256:") {
						f.Digests[a.DownloadURL] = strings.TrimPrefix(a.Digest, "sha256:")
					}
				}
				return assets, nil
			}
		}

		if len(releases) < 30 {
			break
		}
	}

	return nil, fmt.Errorf("no matching tag for '%s'", tag)
}

// finds the latest pre-release and returns the tag
func (f *GithubAssetFinder) getLatestTag() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", f.Repo)
	resp, err := Get(url)
	if err != nil {
		return "", fmt.Errorf("pre-release finder: %w", err)
	}

	var releases []GithubRelease

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("pre-release finder: %w", err)
	}
	err = json.Unmarshal(body, &releases)
	if err != nil {
		return "", fmt.Errorf("pre-release finder: %w", err)
	}

	if len(releases) <= 0 {
		return "", fmt.Errorf("no releases found")
	}

	for _, release := range releases {
		if !release.Draft {
			return release.Tag, nil
		}
	}

	return "", fmt.Errorf("no published releases found")
}

// A DirectAssetFinder returns the embedded URL directly as the only asset.
type DirectAssetFinder struct {
	URL string
}

func (f *DirectAssetFinder) Find() ([]string, error) {
	return []string{f.URL}, nil
}

type GithubSourceFinder struct {
	Tool string
	Repo string
	Tag  string
}

func (f *GithubSourceFinder) Find() ([]string, error) {
	return []string{fmt.Sprintf("https://github.com/%s/tarball/%s/%s.tar.gz", f.Repo, f.Tag, f.Tool)}, nil
}
