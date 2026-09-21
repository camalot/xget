package engine

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/camalot/xget/internal/config"
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
	source, _ := config.Default().ResolveSource("github")
	return getSourceToken(source)
}

func getSourceToken(source config.Source) (string, error) {
	for _, name := range source.TokenEnv {
		if value := os.Getenv(name); value != "" {
			return tokenFrom(value)
		}
	}
	if source.Token != "" {
		return tokenFrom(source.Token)
	}
	return "", ErrNoToken
}

func GithubTokenConfigured() bool {
	_, err := getGithubToken()
	return err == nil
}

func SetAuthHeader(req *http.Request) *http.Request {
	source, _ := config.Default().ResolveSource("github")
	return setSourceAuthHeader(req, source)
}

func setSourceAuthHeader(req *http.Request, source config.Source) *http.Request {
	token, err := getSourceToken(source)
	if err != nil && !errors.Is(err, ErrNoToken) {
		fmt.Fprintf(os.Stderr, "warning: not using %s token: %v\n", source.Type, err)
	}

	if req.URL.Scheme == "https" && sourceMatchesRequest(source, req) && err == nil {
		if runtimeDisableSSL {
			fmt.Fprintf(os.Stderr, "warning: not using %s token while SSL verification is disabled\n", source.Type)
			return req
		}
		if source.Type == "gitlab" {
			req.Header.Set("PRIVATE-TOKEN", token)
		} else {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		}
	}

	return req
}

func sourceMatchesRequest(source config.Source, req *http.Request) bool {
	if req.URL.Scheme != "https" {
		return false
	}
	reqHost, reqPort := splitHostPort(req.URL.Host, defaultPortForScheme(req.URL.Scheme))
	if source.Host != "" {
		srcHost, srcPort := splitHostPort(source.Host, "443")
		if reqHost == srcHost && reqPort == srcPort {
			return true
		}
	}
	apiURL, err := url.Parse(source.APIURL)
	if err != nil || apiURL.Scheme != "https" || apiURL.Host == "" {
		return false
	}
	apiHost, apiPort := splitHostPort(apiURL.Host, defaultPortForScheme(apiURL.Scheme))
	return reqHost == apiHost && reqPort == apiPort
}

// splitHostPort normalizes hostport to a lowercase host and an explicit port,
// falling back to defaultPort when hostport has none.
func splitHostPort(hostport, defaultPort string) (string, string) {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		host, port = hostport, defaultPort
	}
	return strings.ToLower(host), port
}

// defaultPortForScheme returns the implicit port for a URL scheme lacking one.
func defaultPortForScheme(scheme string) string {
	if scheme == "http" {
		return "80"
	}
	return "443"
}

func sourceRedirectPolicy(source config.Source) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, _ []*http.Request) error {
		if req.URL.Scheme != "https" || !sourceMatchesRequest(source, req) {
			req.Header.Del("Authorization")
			req.Header.Del("PRIVATE-TOKEN")
		}
		return nil
	}
}

func Get(url string) (*http.Response, error) {
	source, _ := config.Default().ResolveSource("github")
	return GetWithSource(url, source)
}

func GetWithSource(url string, source config.Source) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req = setSourceAuthHeader(req, source)

	transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
	if runtimeDisableSSL {
		// #nosec G402 -- explicit user opt-in via --disable-ssl.
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	proxyClient := &http.Client{Transport: transport, CheckRedirect: sourceRedirectPolicy(source)}

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
	source, _ := config.Default().ResolveSource("github")
	return DownloadWithSource(url, out, getbar, source)
}

func DownloadWithSource(url string, out io.Writer, getbar func(size int64) *pb.ProgressBar, source config.Source) error {
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

	resp, err := GetWithSource(url, source)
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
