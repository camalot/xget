package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/camalot/xget/internal/config"
)

type GitlabRelease struct {
	Name            string    `json:"name"`
	Tag             string    `json:"tag_name"`
	CreatedAt       time.Time `json:"created_at"`
	ReleasedAt      time.Time `json:"released_at"`
	UpcomingRelease bool      `json:"upcoming_release"`
	Assets          struct {
		Links []struct {
			URL            string `json:"url"`
			DirectAssetURL string `json:"direct_asset_url"`
		} `json:"links"`
	} `json:"assets"`
}

type GitlabAssetFinder struct {
	Repo       string
	Tag        string
	Prerelease bool
	MinTime    time.Time
	ReleaseTag string
	Source     config.Source
}

func (f *GitlabAssetFinder) Find() ([]string, error) {
	project := url.PathEscape(f.Repo)
	endpoint := fmt.Sprintf("%s/projects/%s/releases", f.Source.APIURL, project)
	if f.Tag != "" && f.Tag != "latest" {
		endpoint += "/" + url.PathEscape(f.Tag)
	}

	resp, err := GetWithSource(endpoint, f.Source)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("error closing response body:", err)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API: %s (URL: %s): %s", resp.Status, endpoint, strings.TrimSpace(string(body)))
	}

	var release GitlabRelease
	if f.Tag != "" && f.Tag != "latest" {
		if err := json.Unmarshal(body, &release); err != nil {
			return nil, err
		}
	} else {
		var releases []GitlabRelease
		if err := json.Unmarshal(body, &releases); err != nil {
			return nil, err
		}
		now := time.Now()
		for _, candidate := range releases {
			if candidate.UpcomingRelease || candidate.ReleasedAt.After(now) {
				continue
			}
			release = candidate
			break
		}
		if release.Tag == "" {
			return nil, fmt.Errorf("no published GitLab releases found for %s", f.Repo)
		}
	}

	publishedAt := release.ReleasedAt
	if publishedAt.IsZero() {
		publishedAt = release.CreatedAt
	}
	if publishedAt.Before(f.MinTime) {
		return nil, ErrNoUpgrade
	}
	f.ReleaseTag = release.Tag
	assets := make([]string, 0, len(release.Assets.Links))
	for _, link := range release.Assets.Links {
		assetURL := link.DirectAssetURL
		if assetURL == "" {
			assetURL = link.URL
		}
		if assetURL != "" {
			assets = append(assets, assetURL)
		}
	}
	return assets, nil
}

type GitlabSourceFinder struct {
	Tool   string
	Repo   string
	Tag    string
	Source config.Source
}

func (f *GitlabSourceFinder) Find() ([]string, error) {
	return []string{fmt.Sprintf("%s/projects/%s/repository/archive.tar.gz?sha=%s", f.Source.APIURL, url.PathEscape(f.Repo), url.QueryEscape(f.Tag))}, nil
}
