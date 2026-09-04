package stridge

import "context"

// do will hold the shared HTTP request lifecycle for Stridge endpoints.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	// TODO: Implement this method as a learning exercise. Suggested stages:
	//   1. JSON-encode a non-nil request body.
	//   2. Create the request with http.NewRequestWithContext.
	//   3. Attach the appropriate content, API, and gateway headers.
	//   4. Execute the request through c.httpClient.Do.
	//   5. Defer closing the response body.
	//   6. Inspect the HTTP status code.
	//   7. Decode non-success responses into APIError.
	//   8. Decode successful JSON into out when out is non-nil.
	return ErrNotImplemented
}
