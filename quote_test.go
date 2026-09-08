package stridge

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestQuoteSuccess(t *testing.T) {
	wantQuery := url.Values{
		"amount":          {"1000000000000000000"},
		"from_asset":      {"0xfrom"},
		"from_network_id": {"1"},
		"to_asset":        {"0xto"},
		"to_network_id":   {"56"},
	}

	client := newQuoteTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("request method = %q, want %q", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v1/uda/quote" {
			t.Errorf("request path = %q, want %q", r.URL.Path, "/v1/uda/quote")
		}
		if got := r.URL.Query(); !reflect.DeepEqual(got, wantQuery) {
			t.Errorf("request query = %#v, want %#v", got, wantQuery)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept header = %q, want %q", got, "application/json")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{
			"success": true,
			"message": "operation completed successfully",
			"data": {
				"from": {
					"network_id": 1,
					"asset_address": "0xfrom",
					"amount": "1000000000000000000"
				},
				"to": {
					"network_id": 56,
					"asset_address": "0xto",
					"amount": "998500000000000000"
				},
				"exchange_rate": "0.9985",
				"expires_at": "2026-05-29T12:05:00Z",
				"fees": {
					"total_fee": "3500000000000000",
					"gas_fee": "1000000000000000",
					"network_reserve": "0",
					"platform_fee": "2500000000000000",
					"provider_fee": "0",
					"items": [{
						"kind": "platform",
						"amount": "2500000000000000",
						"source": "stridge",
						"recipient": "0xfee"
					}]
				},
				"route": {
					"provider": "lifi",
					"scenario": "cross_chain_swap",
					"estimated_time_seconds": 45
				}
			}
		}`)
	}))

	quote, err := client.Quote(context.Background(), validQuoteRequest())
	if err != nil {
		t.Fatalf("Quote() error = %v", err)
	}
	if quote == nil || quote.Data == nil {
		t.Fatal("Quote() returned nil quote data")
	}
	if !quote.Success {
		t.Fatal("Quote().Success = false, want true")
	}
	if quote.Message != "operation completed successfully" {
		t.Errorf("Quote().Message = %q", quote.Message)
	}
	if quote.Data.From.NetworkID != 1 || quote.Data.From.Amount != "1000000000000000000" {
		t.Errorf("Quote().Data.From = %#v", quote.Data.From)
	}
	if quote.Data.To.NetworkID != 56 || quote.Data.To.Amount != "998500000000000000" {
		t.Errorf("Quote().Data.To = %#v", quote.Data.To)
	}
	if quote.Data.ExchangeRate != "0.9985" {
		t.Errorf("Quote().Data.ExchangeRate = %q", quote.Data.ExchangeRate)
	}
	wantExpiry := time.Date(2026, time.May, 29, 12, 5, 0, 0, time.UTC)
	if !quote.Data.ExpiresAt.Equal(wantExpiry) {
		t.Errorf("Quote().Data.ExpiresAt = %v, want %v", quote.Data.ExpiresAt, wantExpiry)
	}
	if quote.Data.Fees.TotalFee != "3500000000000000" || len(quote.Data.Fees.Items) != 1 {
		t.Errorf("Quote().Data.Fees = %#v", quote.Data.Fees)
	}
	if quote.Data.Route.Provider != "lifi" || quote.Data.Route.EstimatedTimeSeconds != 45 {
		t.Errorf("Quote().Data.Route = %#v", quote.Data.Route)
	}
}

func TestQuoteIncludesOptionalAddresses(t *testing.T) {
	client := newQuoteTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if got := query.Get("from_address"); got != "sender+tag/1" {
			t.Errorf("from_address = %q, want %q", got, "sender+tag/1")
		}
		if got := query.Get("to_address"); got != "recipient address" {
			t.Errorf("to_address = %q, want %q", got, "recipient address")
		}
		_, _ = fmt.Fprint(w, `{"success":true,"data":{}}`)
	}))

	request := validQuoteRequest()
	request.FromAddress = "sender+tag/1"
	request.ToAddress = "recipient address"

	if _, err := client.Quote(context.Background(), request); err != nil {
		t.Fatalf("Quote() error = %v", err)
	}
}

func TestQuoteAPIErrors(t *testing.T) {
	tests := []struct {
		status  int
		message string
	}{
		{status: http.StatusBadRequest, message: "invalid request"},
		{status: http.StatusUnauthorized, message: "unauthorized"},
		{status: http.StatusNotFound, message: "not found"},
		{status: http.StatusTooManyRequests, message: "rate limited"},
		{status: http.StatusServiceUnavailable, message: "provider unavailable"},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			client := newQuoteTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = fmt.Fprintf(w, `{"success":false,"code":%d,"error":%q}`, tt.status, tt.message)
			}))

			quote, err := client.Quote(context.Background(), validQuoteRequest())
			if err == nil {
				t.Fatal("Quote() error = nil, want APIError")
			}
			if quote != nil {
				t.Fatalf("Quote() quote = %#v, want nil", quote)
			}

			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("Quote() error type = %T, want *APIError", err)
			}
			if apiErr.StatusCode != tt.status {
				t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, tt.status)
			}
			if apiErr.Code != tt.status {
				t.Errorf("APIError.Code = %d, want %d", apiErr.Code, tt.status)
			}
			if apiErr.Message != tt.message {
				t.Errorf("APIError.Message = %q, want %q", apiErr.Message, tt.message)
			}
		})
	}
}

func TestQuoteAPIErrorUsesHTTPFallbacks(t *testing.T) {
	client := newQuoteTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, `{}`)
	}))

	_, err := client.Quote(context.Background(), validQuoteRequest())
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Quote() error = %v, want *APIError", err)
	}
	if apiErr.Code != http.StatusServiceUnavailable {
		t.Errorf("APIError.Code = %d, want %d", apiErr.Code, http.StatusServiceUnavailable)
	}
	if apiErr.Message != http.StatusText(http.StatusServiceUnavailable) {
		t.Errorf("APIError.Message = %q, want %q", apiErr.Message, http.StatusText(http.StatusServiceUnavailable))
	}
}

func TestQuoteRejectsMalformedSuccessJSON(t *testing.T) {
	client := newQuoteTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"success":`)
	}))

	quote, err := client.Quote(context.Background(), validQuoteRequest())
	if err == nil || !strings.Contains(err.Error(), "decode Stridge response") {
		t.Fatalf("Quote() error = %v, want response decoding error", err)
	}
	if quote != nil {
		t.Fatalf("Quote() quote = %#v, want nil", quote)
	}
}

func TestQuoteRejectsMalformedAPIErrorJSON(t *testing.T) {
	client := newQuoteTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, `<html>unavailable</html>`)
	}))

	_, err := client.Quote(context.Background(), validQuoteRequest())
	if err == nil || !strings.Contains(err.Error(), "decode Stridge API error response (status 503)") {
		t.Fatalf("Quote() error = %v, want API error response decoding error", err)
	}
}

func TestQuoteRejectsSuccessWithoutData(t *testing.T) {
	client := newQuoteTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"success":true,"message":"ok"}`)
	}))

	quote, err := client.Quote(context.Background(), validQuoteRequest())
	if err == nil || !strings.Contains(err.Error(), "missing data") {
		t.Fatalf("Quote() error = %v, want missing data error", err)
	}
	if quote != nil {
		t.Fatalf("Quote() quote = %#v, want nil", quote)
	}
}

func TestQuoteRejectsUnsuccessfulTwoHundredResponse(t *testing.T) {
	client := newQuoteTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"success":false,"code":503,"error":"provider unavailable"}`)
	}))

	_, err := client.Quote(context.Background(), validQuoteRequest())
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Quote() error = %v, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusOK || apiErr.Code != http.StatusServiceUnavailable {
		t.Errorf("APIError = %#v, want HTTP status 200 and API code 503", apiErr)
	}
}

func TestQuoteReturnsTransportError(t *testing.T) {
	wantErr := errors.New("network unavailable")
	httpClient := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, wantErr
		}),
	}
	client, err := NewClient(Config{
		BaseURL:    "http://example.invalid/v1",
		HTTPClient: httpClient,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	quote, err := client.Quote(context.Background(), validQuoteRequest())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Quote() error = %v, want wrapped %v", err, wantErr)
	}
	if quote != nil {
		t.Fatalf("Quote() quote = %#v, want nil", quote)
	}
}

func TestQuoteHonorsContextCancellation(t *testing.T) {
	requestStarted := make(chan struct{})
	releaseHandler := make(chan struct{})
	defer close(releaseHandler)
	client := newQuoteTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-releaseHandler
	}))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		_, err := client.Quote(ctx, validQuoteRequest())
		errCh <- err
	}()

	select {
	case <-requestStarted:
		cancel()
	case <-time.After(time.Second):
		cancel()
		t.Fatal("Quote() did not reach test server")
	}

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Quote() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Quote() did not return after context cancellation")
	}
}

func TestQuoteHonorsHTTPClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)

	httpClient := server.Client()
	httpClient.Timeout = 25 * time.Millisecond
	client, err := NewClient(Config{
		BaseURL:    server.URL + "/v1",
		HTTPClient: httpClient,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.Quote(context.Background(), validQuoteRequest())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Quote() error = %v, want context deadline exceeded", err)
	}
}

func newQuoteTestClient(t *testing.T, handler http.Handler) *Client {
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

func validQuoteRequest() QuoteRequest {
	return QuoteRequest{
		FromNetworkID: 1,
		FromAsset:     "0xfrom",
		ToNetworkID:   56,
		ToAsset:       "0xto",
		Amount:        "1000000000000000000",
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
