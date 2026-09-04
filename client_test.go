package stridge

import (
	"net/http"
	"testing"
)

func TestNewClientCreatesDefaultHTTPClient(t *testing.T) {
	client, err := NewClient(Config{})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client.httpClient == nil {
		t.Fatal("client HTTP client = nil, want a default client")
	}
	if client.httpClient.Timeout <= 0 {
		t.Fatalf("default HTTP client timeout = %v, want a non-zero timeout", client.httpClient.Timeout)
	}
}

func TestNewClientPreservesCallerHTTPClient(t *testing.T) {
	provided := &http.Client{}

	client, err := NewClient(Config{HTTPClient: provided})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client.httpClient != provided {
		t.Fatal("client did not preserve the caller-provided HTTP client")
	}
	if provided.Timeout != 0 {
		t.Fatalf("caller-provided HTTP client timeout = %v, want it unchanged", provided.Timeout)
	}
}
