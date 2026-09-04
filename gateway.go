package stridge

import "context"

// GatewayStart will begin a gateway flow.
func (c *Client) GatewayStart(ctx context.Context, req GatewayStartRequest) (*GatewayStartResponse, error) {
	// TODO: Implement the gateway-start endpoint using c.do.
	return nil, ErrNotImplemented
}

// GatewayPoll will retrieve gateway state for owner.
func (c *Client) GatewayPoll(ctx context.Context, owner string) (*GatewayPollResponse, error) {
	// TODO: Implement the gateway-poll endpoint using c.do.
	return nil, ErrNotImplemented
}
