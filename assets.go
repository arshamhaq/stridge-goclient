package stridge

import (
	"context"
	"net/http"
)

// SupportedAssets retrieves the networks and assets supported by Stridge UDA.
// This endpoint is public and does not require an API key.
func (c *Client) SupportedAssets(ctx context.Context) (*SupportedAssetsResponse, error) {
	endpointURL := c.baseURL.JoinPath("uda", "supported-assets")

	var response SupportedAssetsResponse
	if err := c.do(ctx, http.MethodGet, endpointURL.String(), nil, nil, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
