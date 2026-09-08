package stridge

import (
	"context"
	"net/http"
)

// CreateUDA creates a Universal Deposit Address or returns the existing UDA
// for the same owner.
func (c *Client) CreateUDA(ctx context.Context, req CreateUDARequest) (*CreateUDAResponse, error) {
	endpointURL := c.baseURL.JoinPath("uda")
	headers := make(http.Header)
	headers.Set("X-API-Key", c.apiKey)

	var response CreateUDAResponse
	if err := c.do(ctx, http.MethodPost, endpointURL.String(), headers, req, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
