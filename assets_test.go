package stridge

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSupportedAssetsSuccess(t *testing.T) {
	client := newAssetsTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("request method = %q, want %q", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v1/uda/supported-assets" {
			t.Errorf("request path = %q, want %q", r.URL.Path, "/v1/uda/supported-assets")
		}
		if got := r.URL.Query(); len(got) != 0 {
			t.Errorf("request query = %#v, want no query parameters", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept header = %q, want %q", got, "application/json")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{
			"assets": [{
				"network_id": "60",
				"network_name": "Ethereum",
				"network_symbol": "ETH",
				"eip155_id": 1,
				"chain_type": "EVM",
				"native_currency": {
					"symbol": "ETH",
					"name": "Ether",
					"decimals": 18,
					"logo": "https://assets.stridge.com/tokens/eth.png",
					"min_deposit_usd": "1.00",
					"price_impact": "0"
				},
				"assets": [{
					"symbol": "USDC",
					"name": "USD Coin",
					"address": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
					"decimals": 6,
					"logo": "https://assets.stridge.com/tokens/usdc.png",
					"min_deposit_usd": "1.00",
					"price_impact": "0"
				}]
			}]
		}`)
	}))

	response, err := client.SupportedAssets(context.Background())
	if err != nil {
		t.Fatalf("SupportedAssets() error = %v", err)
	}
	if response == nil || len(response.Assets) != 1 {
		t.Fatalf("SupportedAssets() response = %#v, want one network", response)
	}

	network := response.Assets[0]
	if network.NetworkID != "60" || network.NetworkName != "Ethereum" || network.EIP155ID != 1 {
		t.Errorf("network = %#v", network)
	}
	if network.ChainType != "EVM" || network.NativeCurrency.Symbol != "ETH" || network.NativeCurrency.Decimals != 18 {
		t.Errorf("network native currency = %#v", network.NativeCurrency)
	}
	if len(network.Assets) != 1 {
		t.Fatalf("network assets = %#v, want one asset", network.Assets)
	}
	asset := network.Assets[0]
	if asset.Symbol != "USDC" || asset.Decimals != 6 || asset.MinDepositUSD != "1.00" {
		t.Errorf("asset = %#v", asset)
	}
}

func TestSupportedAssetsAPIError(t *testing.T) {
	client := newAssetsTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, `{"success":false,"code":500,"error":"internal server error"}`)
	}))

	response, err := client.SupportedAssets(context.Background())
	if err == nil {
		t.Fatal("SupportedAssets() error = nil, want APIError")
	}
	if response != nil {
		t.Fatalf("SupportedAssets() response = %#v, want nil", response)
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("SupportedAssets() error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError || apiErr.Code != http.StatusInternalServerError {
		t.Errorf("APIError = %#v, want HTTP status and code 500", apiErr)
	}
	if apiErr.Message != "internal server error" {
		t.Errorf("APIError.Message = %q", apiErr.Message)
	}
}

func TestSupportedAssetsRejectsMalformedJSON(t *testing.T) {
	client := newAssetsTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"assets":`)
	}))

	response, err := client.SupportedAssets(context.Background())
	if err == nil || !strings.Contains(err.Error(), "decode Stridge response") {
		t.Fatalf("SupportedAssets() error = %v, want response decoding error", err)
	}
	if response != nil {
		t.Fatalf("SupportedAssets() response = %#v, want nil", response)
	}
}

func TestSupportedAssetsReturnsTransportError(t *testing.T) {
	wantErr := errors.New("network unavailable")
	client, err := NewClient(Config{
		BaseURL: "http://example.invalid/v1",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, wantErr
		})},
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	response, err := client.SupportedAssets(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("SupportedAssets() error = %v, want wrapped %v", err, wantErr)
	}
	if response != nil {
		t.Fatalf("SupportedAssets() response = %#v, want nil", response)
	}
}

func TestSupportedAssetsHonorsContextCancellation(t *testing.T) {
	requestStarted := make(chan struct{})
	releaseHandler := make(chan struct{})
	defer close(releaseHandler)
	client := newAssetsTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-releaseHandler
	}))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		_, err := client.SupportedAssets(ctx)
		errCh <- err
	}()

	select {
	case <-requestStarted:
		cancel()
	case <-time.After(time.Second):
		cancel()
		t.Fatal("SupportedAssets() did not reach test server")
	}

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("SupportedAssets() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("SupportedAssets() did not return after context cancellation")
	}
}

func newAssetsTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := NewClient(Config{
		BaseURL:    server.URL + "/v1",
		HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client
}
