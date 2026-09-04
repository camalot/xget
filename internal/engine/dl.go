package engine

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/camalot/xget/internal/home"
	pb "github.com/schollz/progressbar/v3"
)

var runtimeDisableSSL bool

func SetDisableSSL(disable bool) {
	runtimeDisableSSL = disable
}

func readValidatedFile(p string) ([]byte, error) {
	clean := filepath.Clean(p)
	if clean == "" || clean == "." {
		return nil, fmt.Errorf("invalid file path %q", p)
	}

	info, err := os.Stat(clean)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", clean)
	}

	// #nosec G304 -- path is normalized and validated to a regular file above.
	return os.ReadFile(clean)
}

func openValidatedFile(p string) (*os.File, error) {
	clean := filepath.Clean(p)
	if clean == "" || clean == "." {
		return nil, fmt.Errorf("invalid file path %q", p)
	}

	info, err := os.Stat(clean)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", clean)
	}

	// #nosec G304 -- path is normalized and validated to a regular file above.
	return os.Open(clean)
}

func tokenFrom(s string) (string, error) {
	if strings.HasPrefix(s, "@") {
		f, err := home.Expand(s[1:])
		if err != nil {
			return "", err
		}
		b, err := readValidatedFile(f)
		if err != nil {
			return "", err
		}
		return strings.TrimRight(string(b), "\r\n"), nil
	}
	return s, nil
}

var ErrNoToken = errors.New("no github token")

func getGithubToken() (string, error) {
	// support for EGET_GITHUB_TOKEN is kept for backwards compatibility, but XGET_GITHUB_TOKEN is preferred
	if os.Getenv("EGET_GITHUB_TOKEN") != "" {
		return tokenFrom(os.Getenv("EGET_GITHUB_TOKEN"))
	}
	if os.Getenv("XGET_GITHUB_TOKEN") != "" {
		return tokenFrom(os.Getenv("XGET_GITHUB_TOKEN"))
	}
	if os.Getenv("GITHUB_TOKEN") != "" {
		return tokenFrom(os.Getenv("GITHUB_TOKEN"))
	}
	return "", ErrNoToken
}

func GithubTokenConfigured() bool {
	_, err := getGithubToken()
	return err == nil
}

func SetAuthHeader(req *http.Request) *http.Request {
	token, err := getGithubToken()
	if err != nil && !errors.Is(err, ErrNoToken) {
		fmt.Fprintln(os.Stderr, "warning: not using github token:", err)
	}

	if req.URL.Scheme == "https" && req.Host == "api.github.com" && err == nil {
		if runtimeDisableSSL {
			fmt.Fprintln(os.Stderr, "warning: not using GitHub token while SSL verification is disabled")
			return req
		}
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	return req
}

func Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req = SetAuthHeader(req)

	transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
	if runtimeDisableSSL {
		// #nosec G402 -- explicit user opt-in via --disable-ssl.
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	proxyClient := &http.Client{Transport: transport}

	return proxyClient.Do(req)
}

type RateLimitJson struct {
	Resources map[string]RateLimit
}

type RateLimit struct {
	Limit     int
	Remaining int
	Reset     int64
}

func (r RateLimit) ResetTime() time.Time {
	return time.Unix(r.Reset, 0)
}

func (r RateLimit) String() string {
	now := time.Now()
	rtime := r.ResetTime()
	if rtime.Before(now) {
		return fmt.Sprintf("Limit: %d, Remaining: %d, Reset: %v", r.Limit, r.Remaining, rtime)
	} else {
		return fmt.Sprintf(
			"Limit: %d, Remaining: %d, Reset: %v (%v)",
			r.Limit, r.Remaining, rtime, rtime.Sub(now).Round(time.Second),
		)
	}
}

func GetRateLimit() (RateLimit, error) {
	url := "https://api.github.com/rate_limit"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return RateLimit{}, err
	}

	req = SetAuthHeader(req)

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return RateLimit{}, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("error closing response body:", err)
		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return RateLimit{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return RateLimit{}, &GithubError{Status: resp.Status, Code: resp.StatusCode, Body: b, Url: url}
	}

	var parsed RateLimitJson
	err = json.Unmarshal(b, &parsed)

	return parsed.Resources["core"], err
}

// Download the file at 'url' and write the http response body to 'out'. The
// 'getbar' function allows the caller to construct a progress bar given the
// size of the file being downloaded, and the download will write to the
// returned progress bar.
func Download(url string, out io.Writer, getbar func(size int64) *pb.ProgressBar) error {
	if IsLocalFile(url) {
		f, err := openValidatedFile(url)
		if err != nil {
			return err
		}
		defer func() {
			if err := f.Close(); err != nil {
				fmt.Println("error closing file:", err)
			}
		}()
		_, err = io.Copy(out, f)
		return err
	}

	resp, err := Get(url)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("error closing response body:", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("download error: %d: %s", resp.StatusCode, body)
	}

	bar := getbar(resp.ContentLength)
	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	return err
}
