package engine

import (
	"net/http"
	"strings"
	"testing"
)

func TestGithubErrorRateLimitGuidanceWithoutToken(t *testing.T) {
	t.Setenv("XGET_GITHUB_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("EGET_GITHUB_TOKEN", "")

	err := (&GithubError{Code: http.StatusTooManyRequests, Status: "429 Too Many Requests"}).Error()
	for _, name := range []string{"XGET_GITHUB_TOKEN", "GITHUB_TOKEN", "EGET_GITHUB_TOKEN"} {
		if !strings.Contains(err, name) {
			t.Fatalf("error %q missing %s", err, name)
		}
	}
}
