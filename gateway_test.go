package stridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGatewayStartSuccess(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{name: "created", statusCode: http.StatusCreated},
		{name: "existing", statusCode: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantRequest := validGatewayStartRequest()
			client := newGatewayTestClient(t, "gateway-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("request method = %q, want %q", r.Method, http.MethodPost)
				}
				if r.URL.Path != "/v1/gateway/start" {
					t.Errorf("request path = %q, want %q", r.URL.Path, "/v1/gateway/start")
				}
				if got := r.Header.Get("Accept"); got != "application/json" {
					t.Errorf("Accept header = %q, want %q", got, "application/json")
				}
				if got := r.Header.Get("Content-Type"); got != "application/json" {
					t.Errorf("Content-Type header = %q, want %q", got, "application/json")
				}
				if got := r.Header.Get("X-Gateway-Key"); got != "gateway-key" {
					t.Errorf("X-Gateway-Key header = %q, want %q", got, "gateway-key")
				}
				if got := r.Header.Get("X-API-Key"); got != "" {
					t.Errorf("X-API-Key header = %q, want empty", got)
				}

				var gotRequest GatewayStartRequest
				if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
					t.Fatalf("decode request body: %v", err)
				}
				if !reflect.DeepEqual(gotRequest, wantRequest) {
					t.Errorf("request body = %#v, want %#v", gotRequest, wantRequest)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = fmt.Fprint(w, `{
					"data":{
						"uda_id":"uda_123",
						"owner":"0xalice",
						"status":"active",
						"destination":{
							"to_address":"0xAppWallet",
							"network_id":"0",
							"asset_symbol":"BTC"
						},
						"metadata":{"checkout_id":"order-123"},
						"deposit_addresses":[{
							"eip155_id":"1",
							"network_name":"ethereum",
							"address":"0xAliceDeposit",
							"accepted_assets":[{
								"symbol":"USDC",
								"address":"0xUSDC",
								"decimals":6
							}]
						}]
					}
				}`)
			}))

			response, err := client.GatewayStart(context.Background(), wantRequest)
			if err != nil {
				t.Fatalf("GatewayStart() error = %v", err)
			}
			if response == nil {
				t.Fatal("GatewayStart() response = nil")
			}
			if response.Data == nil {
				t.Fatal("GatewayStart().Data = nil")
			}
			if response.Data.UDAID != "uda_123" || response.Data.Owner != "0xalice" || response.Data.Status != "active" {
				t.Errorf("GatewayStart().Data = %#v", response.Data)
			}
			if response.Data.Destination.AssetSymbol != "BTC" {
				t.Errorf("Destination = %#v", response.Data.Destination)
			}
			if len(response.Data.DepositAddresses) != 1 {
				t.Fatalf("DepositAddresses = %#v, want one address", response.Data.DepositAddresses)
			}
			if len(response.Data.DepositAddresses[0].AcceptedAssets) != 1 {
				t.Fatalf("AcceptedAssets = %#v, want one asset", response.Data.DepositAddresses[0].AcceptedAssets)
			}
			if response.Data.DepositAddresses[0].AcceptedAssets[0].Symbol != "USDC" {
				t.Errorf("AcceptedAssets = %#v", response.Data.DepositAddresses[0].AcceptedAssets)
			}
			if response.Data.Metadata["checkout_id"] != "order-123" {
				t.Errorf("Metadata = %#v", response.Data.Metadata)
			}
		})
	}
}

func TestGatewayStartAPIErrors(t *testing.T) {
	tests := []struct {
		status  int
		message string
	}{
		{status: http.StatusBadRequest, message: "invalid destination"},
		{status: http.StatusUnauthorized, message: "invalid gateway key"},
		{status: http.StatusTooManyRequests, message: "rate limited"},
		{status: http.StatusServiceUnavailable, message: "service unavailable"},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			client := newGatewayTestClient(t, "gateway-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = fmt.Fprintf(w, `{"code":%d,"error":%q}`, tt.status, tt.message)
			}))

			response, err := client.GatewayStart(context.Background(), validGatewayStartRequest())
			if response != nil {
				t.Fatalf("GatewayStart() response = %#v, want nil", response)
			}
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("GatewayStart() error = %v, want *APIError", err)
			}
			if apiErr.StatusCode != tt.status || apiErr.Code != tt.status || apiErr.Message != tt.message {
				t.Errorf("APIError = %#v", apiErr)
			}
		})
	}
}

func TestGatewayStartRejectsMalformedJSON(t *testing.T) {
	client := newGatewayTestClient(t, "gateway-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"data":`)
	}))

	response, err := client.GatewayStart(context.Background(), validGatewayStartRequest())
	if err == nil || !strings.Contains(err.Error(), "decode Stridge response") {
		t.Fatalf("GatewayStart() error = %v, want response decoding error", err)
	}
	if response != nil {
		t.Fatalf("GatewayStart() response = %#v, want nil", response)
	}
}

func TestGatewayStartRejectsSuccessWithoutData(t *testing.T) {
	client := newGatewayTestClient(t, "gateway-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{}`)
	}))

	response, err := client.GatewayStart(context.Background(), validGatewayStartRequest())
	if err == nil || !strings.Contains(err.Error(), "missing data") {
		t.Fatalf("GatewayStart() error = %v, want missing data error", err)
	}
	if response != nil {
		t.Fatalf("GatewayStart() response = %#v, want nil", response)
	}
}

func TestGatewayStartReturnsTransportError(t *testing.T) {
	wantErr := errors.New("network unavailable")
	client, err := NewClient(Config{
		BaseURL:    "http://example.invalid/v1",
		GatewayKey: "gateway-key",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, wantErr
		})},
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	response, err := client.GatewayStart(context.Background(), validGatewayStartRequest())
	if !errors.Is(err, wantErr) {
		t.Fatalf("GatewayStart() error = %v, want wrapped %v", err, wantErr)
	}
	if response != nil {
		t.Fatalf("GatewayStart() response = %#v, want nil", response)
	}
}

func TestGatewayStartHonorsContextCancellation(t *testing.T) {
	requestStarted := make(chan struct{})
	releaseHandler := make(chan struct{})
	defer close(releaseHandler)
	client := newGatewayTestClient(t, "gateway-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-releaseHandler
	}))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		_, err := client.GatewayStart(ctx, validGatewayStartRequest())
		errCh <- err
	}()

	select {
	case <-requestStarted:
		cancel()
	case <-time.After(time.Second):
		cancel()
		t.Fatal("GatewayStart() did not reach test server")
	}

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("GatewayStart() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("GatewayStart() did not return after context cancellation")
	}
}

func newGatewayTestClient(t *testing.T, gatewayKey string, handler http.Handler) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := NewClient(Config{
		BaseURL:    server.URL + "/v1",
		APIKey:     "api-key-that-must-not-be-sent",
		GatewayKey: gatewayKey,
		HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client
}

func validGatewayStartRequest() GatewayStartRequest {
	return GatewayStartRequest{
		Owner: "0xAlice",
		Destination: GatewayStartDestination{
			ToAddress:   "0xAppWallet",
			NetworkID:   "0",
			AssetSymbol: "BTC",
		},
		Metadata: map[string]any{"checkout_id": "order-123"},
	}
}
