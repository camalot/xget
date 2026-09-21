package engine

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/camalot/xget/internal/config"
)

func TestGetSourceTokenUsesConfiguredEnvironmentOrderThenToken(t *testing.T) {
	t.Setenv("FIRST_TOKEN", "first")
	t.Setenv("SECOND_TOKEN", "second")
	source := config.Source{TokenEnv: []string{"FIRST_TOKEN", "SECOND_TOKEN"}, Token: "configured"}

	token, err := getSourceToken(source)
	if err != nil {
		t.Fatal(err)
	}
	if token != "first" {
		t.Fatalf("token = %q, want first", token)
	}

	t.Setenv("FIRST_TOKEN", "")
	t.Setenv("SECOND_TOKEN", "")
	token, err = getSourceToken(source)
	if err != nil {
		t.Fatal(err)
	}
	if token != "configured" {
		t.Fatalf("token = %q, want configured", token)
	}
}

func TestGetSourceTokenReadsConfiguredTokenFile(t *testing.T) {
	tokenPath := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(tokenPath, []byte("file-token\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	token, err := getSourceToken(config.Source{Token: "@" + tokenPath})
	if err != nil {
		t.Fatal(err)
	}
	if token != "file-token" {
		t.Fatalf("token = %q, want file-token", token)
	}
}

func TestSetSourceAuthHeaderUsesProviderHeaderOnlyForConfiguredHosts(t *testing.T) {
	t.Setenv("TEST_TOKEN", "secret")
	source := config.Source{
		Type:     "gitlab",
		Host:     "gitlab.example.com",
		APIURL:   "https://api.gitlab.example.com/v4",
		TokenEnv: []string{"TEST_TOKEN"},
	}

	for _, test := range []struct {
		url      string
		wantAuth bool
	}{
		{url: "https://gitlab.example.com/group/project/-/releases/download/v1/tool", wantAuth: true},
		{url: "https://api.gitlab.example.com/v4/projects/group%2Fproject/releases", wantAuth: true},
		{url: "https://example.com/tool", wantAuth: false},
		{url: "http://gitlab.example.com/tool", wantAuth: false},
	} {
		req, err := http.NewRequest(http.MethodGet, test.url, nil)
		if err != nil {
			t.Fatal(err)
		}
		setSourceAuthHeader(req, source)
		if got := req.Header.Get("PRIVATE-TOKEN") != ""; got != test.wantAuth {
			t.Errorf("PRIVATE-TOKEN for %s = %t, want %t", test.url, got, test.wantAuth)
		}
	}
}

func TestSourceRedirectPolicyStripsTokenForUnrelatedHost(t *testing.T) {
	source := config.Source{Type: "gitlab", Host: "gitlab.example.com", APIURL: "https://gitlab.example.com/api/v4"}
	request, err := http.NewRequest(http.MethodGet, "https://objects.example.net/tool", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("PRIVATE-TOKEN", "secret")
	request.Header.Set("Authorization", "Bearer secret")

	if err := sourceRedirectPolicy(source)(request, nil); err != nil {
		t.Fatal(err)
	}
	if request.Header.Get("PRIVATE-TOKEN") != "" || request.Header.Get("Authorization") != "" {
		t.Fatalf("credentials were retained for unrelated redirect: %#v", request.Header)
	}
}
