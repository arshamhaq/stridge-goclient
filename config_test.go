package stridge

import "testing"

func TestNewClientUsesSandboxByDefault(t *testing.T) {
	client, err := NewClient(Config{})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if got := client.baseURL.String(); got != SandboxBaseURL {
		t.Fatalf("client base URL = %q, want %q", got, SandboxBaseURL)
	}
}

func TestNewClientUsesCustomBaseURL(t *testing.T) {
	const customBaseURL = "http://localhost:8080/test/v1/"

	client, err := NewClient(Config{BaseURL: customBaseURL})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	const want = "http://localhost:8080/test/v1"
	if got := client.baseURL.String(); got != want {
		t.Fatalf("client base URL = %q, want %q", got, want)
	}
}

func TestNewClientRejectsMalformedBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
	}{
		{name: "missing scheme", baseURL: "api.stridge.dev/v1"},
		{name: "missing host", baseURL: "https:///v1"},
		{name: "unsupported scheme", baseURL: "ftp://api.stridge.dev/v1"},
		{name: "bad escape", baseURL: "https://api.stridge.dev/%zz"},
		{name: "query string", baseURL: "https://api.stridge.dev/v1?debug=true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewClient(Config{BaseURL: tt.baseURL}); err == nil {
				t.Fatalf("NewClient(Config{BaseURL: %q}) error = nil, want an error", tt.baseURL)
			}
		})
	}
}
