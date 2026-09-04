package stridge

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	SandboxBaseURL     = "https://api.stridge.dev/v1"
	ProductionBaseURL  = "https://api.stridge.com/v1"
	DefaultHTTPTimeout = 30 * time.Second
)

type Config struct {
	BaseURL    string
	APIKey     string
	GatewayKey string
	HTTPClient *http.Client
}

func parseBaseURL(rawURL string) (*url.URL, error) {
	if rawURL == "" {
		rawURL = SandboxBaseURL
	}

	baseURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse Stridge base URL: %w", err)
	}

	if (baseURL.Scheme != "http" && baseURL.Scheme != "https") || baseURL.Host == "" {
		return nil, fmt.Errorf("invalid Stridge base URL %q: an absolute HTTP(S) URL is required", rawURL)
	}
	if baseURL.User != nil || baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return nil, fmt.Errorf("invalid Stridge base URL %q: user info, query strings, and fragments are not allowed", rawURL)
	}

	// A stable base path makes joining future endpoint paths predictable.
	baseURL.Path = strings.TrimRight(baseURL.Path, "/")
	return baseURL, nil
}
