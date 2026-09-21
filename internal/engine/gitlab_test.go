package engine

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/camalot/xget/internal/config"
	"github.com/camalot/xget/internal/options"
)

func TestGitlabAssetFinderLatestUsesReleaseLinks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.EscapedPath() != "/api/v4/projects/group%2Fsubgroup%2Fproject/releases" {
			t.Errorf("path = %q", request.URL.EscapedPath())
		}
		_, _ = fmt.Fprint(response, `[
          {"tag_name":"future","released_at":"2999-01-01T00:00:00Z","upcoming_release":true,"assets":{"links":[]}},
          {"tag_name":"v1.2.3","released_at":"2025-01-01T00:00:00Z","assets":{"links":[
            {"url":"https://gitlab.example/tool.zip","direct_asset_url":"https://gitlab.example/direct/tool.zip"},
            {"url":"https://cdn.example/tool.sha256"}
          ]}}
        ]`)
	}))
	defer server.Close()

	source := config.Source{Name: "gitlab", Type: "gitlab", Host: "gitlab.example", APIURL: server.URL + "/api/v4"}
	finder := &GitlabAssetFinder{Repo: "group/subgroup/project", Source: source}
	assets, err := finder.Find()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"https://gitlab.example/direct/tool.zip", "https://cdn.example/tool.sha256"}
	if !reflect.DeepEqual(assets, want) {
		t.Fatalf("assets = %#v, want %#v", assets, want)
	}
	if finder.ReleaseTag != "v1.2.3" {
		t.Fatalf("release tag = %q", finder.ReleaseTag)
	}
}

func TestGetFinderSupportsNestedGitlabProjectAndSourceArchive(t *testing.T) {
	source, err := config.Default().ResolveSource("gitlab")
	if err != nil {
		t.Fatal(err)
	}
	opts := &options.Flags{Source: true, Tag: "v1.0.0", SourceType: "gitlab", SourceConfig: source}
	finder, tool, err := getFinder("group/subgroup/project", opts)
	if err != nil {
		t.Fatal(err)
	}
	if tool != "project" {
		t.Fatalf("tool = %q", tool)
	}
	gitlabFinder, ok := finder.(*GitlabSourceFinder)
	if !ok {
		t.Fatalf("finder = %T", finder)
	}
	assets, err := gitlabFinder.Find()
	if err != nil {
		t.Fatal(err)
	}
	want := "https://gitlab.com/api/v4/projects/group%2Fsubgroup%2Fproject/repository/archive.tar.gz?sha=v1.0.0"
	if len(assets) != 1 || assets[0] != want {
		t.Fatalf("assets = %#v, want %q", assets, want)
	}
}

func TestGitlabAssetFinderUsesTaggedReleaseEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.EscapedPath() != "/api/v4/projects/group%2Fproject/releases/v1.2.3" {
			t.Errorf("path = %q", request.URL.EscapedPath())
		}
		_, _ = fmt.Fprint(response, `{"tag_name":"v1.2.3","released_at":"2025-01-01T00:00:00Z","assets":{"links":[{"url":"https://cdn.example/tool.zip"}]}}`)
	}))
	defer server.Close()

	source := config.Source{Name: "gitlab", Type: "gitlab", Host: "gitlab.example", APIURL: server.URL + "/api/v4"}
	finder := &GitlabAssetFinder{Repo: "group/project", Tag: "v1.2.3", Source: source}
	assets, err := finder.Find()
	if err != nil {
		t.Fatal(err)
	}
	if finder.ReleaseTag != "v1.2.3" || !reflect.DeepEqual(assets, []string{"https://cdn.example/tool.zip"}) {
		t.Fatalf("tag = %q, assets = %#v", finder.ReleaseTag, assets)
	}
}
