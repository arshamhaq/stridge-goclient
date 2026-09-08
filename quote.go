package stridge

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// Quote will request a cross-chain quote. It's public; No Auth is required.
func (c *Client) Quote(ctx context.Context, req QuoteRequest) (*Quote, error) {
	endpointURL := c.baseURL.JoinPath("uda", "quote")
	query := endpointURL.Query()
	query.Set("from_network_id", strconv.FormatInt(req.FromNetworkID, 10))
	query.Set("from_asset", req.FromAsset)
	query.Set("to_network_id", strconv.FormatInt(req.ToNetworkID, 10))
	query.Set("to_asset", req.ToAsset)
	query.Set("amount", req.Amount)
	if req.FromAddress != "" {
		query.Set("from_address", req.FromAddress)
	}
	if req.ToAddress != "" {
		query.Set("to_address", req.ToAddress)
	}
	endpointURL.RawQuery = query.Encode()

	var quote Quote
	if err := c.do(ctx, http.MethodGet, endpointURL.String(), nil, nil, &quote); err != nil {
		return nil, err
	}

	if !quote.Success {
		message := quote.Error
		if message == "" {
			message = "response reported an unsuccessful quote"
		}
		return nil, &APIError{StatusCode: http.StatusOK, Code: quote.Code, Message: message}
	}

	if quote.Data == nil {
		return nil, fmt.Errorf("decode quote response: successful response is missing data")
	}

	return &quote, nil
}
