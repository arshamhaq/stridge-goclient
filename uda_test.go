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

func TestCreateUDASuccess(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{name: "created", statusCode: http.StatusCreated},
		{name: "existing", statusCode: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantRequest := validCreateUDARequest()
			client := newUDATestClient(t, "api-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("request method = %q, want %q", r.Method, http.MethodPost)
				}
				if r.URL.Path != "/v1/uda" {
					t.Errorf("request path = %q, want %q", r.URL.Path, "/v1/uda")
				}
				if got := r.Header.Get("Accept"); got != "application/json" {
					t.Errorf("Accept header = %q, want %q", got, "application/json")
				}
				if got := r.Header.Get("Content-Type"); got != "application/json" {
					t.Errorf("Content-Type header = %q, want %q", got, "application/json")
				}
				if got := r.Header.Get("X-API-Key"); got != "api-key" {
					t.Errorf("X-API-Key header = %q, want %q", got, "api-key")
				}
				if got := r.Header.Get("X-Gateway-Key"); got != "" {
					t.Errorf("X-Gateway-Key header = %q, want empty", got)
				}

				var gotRequest CreateUDARequest
				if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
					t.Fatalf("decode request body: %v", err)
				}
				if !reflect.DeepEqual(gotRequest, wantRequest) {
					t.Errorf("request body = %#v, want %#v", gotRequest, wantRequest)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = fmt.Fprint(w, `{
					"id":"uda_123",
					"owner":"alice",
					"status":"active",
					"created_at":"2026-05-20T14:03:12Z",
					"destination":{
						"address":"0xDestination",
						"network_id":"9001",
						"eip155_id":"42161",
						"network_name":"arbitrum",
						"asset_address":"0xUSDC",
						"asset_symbol":"USDC",
						"asset_decimals":6
					},
					"deposit_addresses":[{
						"address":"0xDeposit",
						"network_id":"60",
						"eip155_id":"1",
						"network_name":"ethereum",
						"accepted_assets":[{
							"symbol":"USDC",
							"address":"0xUSDCOnEthereum",
							"decimals":6,
							"logo":"https://assets.stridge.com/usdc.png"
						}]
					}],
					"routing_rules":[{
						"match":{
							"source_network_id":"195",
							"source_asset_symbol":"TRX",
							"source_token_address":"native"
						},
						"destination":{
							"address":"TTronDestination",
							"network_id":"195",
							"network_name":"tron",
							"asset_symbol":"TRX",
							"asset_decimals":6
						}
					}]
				}`)
			}))

			response, err := client.CreateUDA(context.Background(), wantRequest)
			if err != nil {
				t.Fatalf("CreateUDA() error = %v", err)
			}
			if response == nil {
				t.Fatal("CreateUDA() response = nil")
			}
			if response.ID != "uda_123" || response.Owner != "alice" || response.Status != "active" {
				t.Errorf("CreateUDA() response = %#v", response)
			}
			wantCreatedAt := time.Date(2026, time.May, 20, 14, 3, 12, 0, time.UTC)
			if !response.CreatedAt.Equal(wantCreatedAt) {
				t.Errorf("CreatedAt = %v, want %v", response.CreatedAt, wantCreatedAt)
			}
			if response.Destination.AssetSymbol != "USDC" || response.Destination.AssetDecimals != 6 {
				t.Errorf("Destination = %#v", response.Destination)
			}
			if len(response.DepositAddresses) != 1 {
				t.Fatalf("DepositAddresses = %#v, want one address", response.DepositAddresses)
			}
			if len(response.DepositAddresses[0].AcceptedAssets) != 1 {
				t.Fatalf("AcceptedAssets = %#v, want one asset", response.DepositAddresses[0].AcceptedAssets)
			}
			if response.DepositAddresses[0].AcceptedAssets[0].Symbol != "USDC" {
				t.Errorf("AcceptedAssets = %#v", response.DepositAddresses[0].AcceptedAssets)
			}
			if len(response.RoutingRules) != 1 || response.RoutingRules[0].Match.SourceTokenAddress != "native" {
				t.Errorf("RoutingRules = %#v", response.RoutingRules)
			}
		})
	}
}

func TestCreateUDAAPIErrors(t *testing.T) {
	tests := []struct {
		status  int
		message string
	}{
		{status: http.StatusBadRequest, message: "invalid destination"},
		{status: http.StatusUnauthorized, message: "invalid API key"},
		{status: http.StatusTooManyRequests, message: "rate limited"},
		{status: http.StatusServiceUnavailable, message: "service unavailable"},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			client := newUDATestClient(t, "api-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = fmt.Fprintf(w, `{"code":%d,"error":%q}`, tt.status, tt.message)
			}))

			response, err := client.CreateUDA(context.Background(), validCreateUDARequest())
			if response != nil {
				t.Fatalf("CreateUDA() response = %#v, want nil", response)
			}
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("CreateUDA() error = %v, want *APIError", err)
			}
			if apiErr.StatusCode != tt.status || apiErr.Code != tt.status || apiErr.Message != tt.message {
				t.Errorf("APIError = %#v", apiErr)
			}
		})
	}
}

func TestCreateUDARejectsMalformedJSON(t *testing.T) {
	client := newUDATestClient(t, "api-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"id":`)
	}))

	response, err := client.CreateUDA(context.Background(), validCreateUDARequest())
	if err == nil || !strings.Contains(err.Error(), "decode Stridge response") {
		t.Fatalf("CreateUDA() error = %v, want response decoding error", err)
	}
	if response != nil {
		t.Fatalf("CreateUDA() response = %#v, want nil", response)
	}
}

func TestCreateUDAReturnsTransportError(t *testing.T) {
	wantErr := errors.New("network unavailable")
	client, err := NewClient(Config{
		BaseURL: "http://example.invalid/v1",
		APIKey:  "api-key",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, wantErr
		})},
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	response, err := client.CreateUDA(context.Background(), validCreateUDARequest())
	if !errors.Is(err, wantErr) {
		t.Fatalf("CreateUDA() error = %v, want wrapped %v", err, wantErr)
	}
	if response != nil {
		t.Fatalf("CreateUDA() response = %#v, want nil", response)
	}
}

func TestCreateUDAHonorsContextCancellation(t *testing.T) {
	requestStarted := make(chan struct{})
	releaseHandler := make(chan struct{})
	defer close(releaseHandler)
	client := newUDATestClient(t, "api-key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-releaseHandler
	}))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		_, err := client.CreateUDA(ctx, validCreateUDARequest())
		errCh <- err
	}()

	select {
	case <-requestStarted:
		cancel()
	case <-time.After(time.Second):
		cancel()
		t.Fatal("CreateUDA() did not reach test server")
	}

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("CreateUDA() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("CreateUDA() did not return after context cancellation")
	}
}

func newUDATestClient(t *testing.T, apiKey string, handler http.Handler) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := NewClient(Config{
		BaseURL:    server.URL + "/v1",
		APIKey:     apiKey,
		GatewayKey: "gateway-key-that-must-not-be-sent",
		HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client
}

func validCreateUDARequest() CreateUDARequest {
	return CreateUDARequest{
		Owner: "alice",
		Destination: CreateUDADestination{
			Address:     "0xDestination",
			NetworkID:   "9001",
			AssetSymbol: "USDC",
		},
		AcceptedAssets: []string{"USDC", "USDT"},
		RoutingRules: []CreateUDARoutingRule{{
			Match: CreateUDARoutingMatch{
				SourceNetworkID:   "195",
				SourceAssetSymbol: "TRX",
			},
			Destination: CreateUDADestination{
				Address:     "TTronDestination",
				NetworkID:   "195",
				AssetSymbol: "TRX",
			},
		}},
	}
}
