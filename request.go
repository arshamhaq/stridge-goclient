package stridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// do executes one Stridge HTTP request and handles behavior shared by all
// endpoints. Endpoint methods remain responsible for constructing their URL
// and choosing their request and response models.
func (c *Client) do(ctx context.Context, method, requestURL string, body, out any) error {
	var requestBody io.Reader
	if body != nil {
		encodedBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode Stridge request body: %w", err)
		}
		requestBody = bytes.NewReader(encodedBody)
	}

	request, err := http.NewRequestWithContext(ctx, method, requestURL, requestBody)
	if err != nil {
		return fmt.Errorf("create Stridge request: %w", err)
	}

	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := c.httpClient.Do(request) //this is a network call and the error corresponds to the network
	if err != nil {
		return fmt.Errorf("send Stridge request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices { //this error is gor when the HTTP goes through the network, but doesn't have a 200-299 status code
		apiErr := &APIError{StatusCode: response.StatusCode}
		if err := json.NewDecoder(response.Body).Decode(apiErr); err != nil {
			return fmt.Errorf("decode Stridge API error response (status %d): %w", response.StatusCode, err)
		}
		if apiErr.Code == 0 {
			apiErr.Code = response.StatusCode
		}
		if apiErr.Message == "" {
			apiErr.Message = http.StatusText(response.StatusCode)
		}
		return apiErr
	}

	if out == nil {
		if _, err := io.Copy(io.Discard, response.Body); err != nil {
			return fmt.Errorf("read Stridge response body: %w", err)
		}
		return nil
	}

	if err := json.NewDecoder(response.Body).Decode(out); err != nil {
		return fmt.Errorf("decode Stridge response (status %d): %w", response.StatusCode, err)
	}

	return nil
}
