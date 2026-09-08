package stridge

import (
	"context"
	"fmt"
	"net/http"
)

// GatewayStart creates or fetches the UDA for an owner and destination.
func (c *Client) GatewayStart(ctx context.Context, req GatewayStartRequest) (*GatewayStartResponse, error) {
	endpointURL := c.baseURL.JoinPath("gateway", "start")
	headers := make(http.Header)
	headers.Set("X-Gateway-Key", c.gatewayKey)

	var response GatewayStartResponse
	if err := c.do(ctx, http.MethodPost, endpointURL.String(), headers, req, &response); err != nil {
		return nil, err
	}
	if response.Data == nil {
		return nil, fmt.Errorf("decode gateway-start response: successful response is missing data")
	}

	return &response, nil
}

// GatewayPoll will retrieve gateway state for owner.
func (c *Client) GatewayPoll(ctx context.Context, owner string) (*GatewayPollResponse, error) {
	// TODO: Implement the gateway-poll endpoint using c.do.
	return nil, ErrNotImplemented
}
