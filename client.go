package stridge

import (
	"net/http"
	"net/url"
)

// Client is a client for the Stridge REST API.
type Client struct {
	baseURL    *url.URL
	apiKey     string
	gatewayKey string
	httpClient *http.Client
}

// NewClient constructs a Stridge client from cfg. Empty BaseURL values use the
// sandbox. A caller-provided HTTPClient is preserved without modification.
func NewClient(cfg Config) (*Client, error) {
	baseURL, err := parseBaseURL(cfg.BaseURL)
	if err != nil {
		return nil, err
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: DefaultHTTPTimeout}
	}

	return &Client{
		baseURL:    baseURL,
		apiKey:     cfg.APIKey,
		gatewayKey: cfg.GatewayKey,
		httpClient: httpClient,
	}, nil
}
