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

	for _, githubError := range []GithubError{
		{Code: http.StatusTooManyRequests, Status: "429 Too Many Requests"},
		{Code: http.StatusForbidden, Status: "403 Forbidden", Body: []byte(`{"message":"API rate limit exceeded"}`)},
	} {
		err := githubError.Error()
		for _, name := range []string{"XGET_GITHUB_TOKEN", "GITHUB_TOKEN", "EGET_GITHUB_TOKEN"} {
			if !strings.Contains(err, name) {
				t.Fatalf("error %q missing %s", err, name)
			}
		}
	}
}
