package stridge

// QuoteRequest contains the initial placeholder fields for a quote request.
// Token amounts use strings so their precision is not lost.
// TODO: Verify field names, JSON tags, and required fields against the official
// Stridge API documentation when implementing Quote.
type QuoteRequest struct {
	FromNetworkID string `json:"fromNetworkId"`
	FromAsset     string `json:"fromAsset"`
	ToNetworkID   string `json:"toNetworkId"`
	ToAsset       string `json:"toAsset"`
	Amount        string `json:"amount"`
}

// Quote is a placeholder for a quote response.
// TODO: Define this model from the official Stridge API documentation.
type Quote struct{}

// SupportedAssetsResponse is a placeholder for the supported-assets response.
// TODO: Define this model from the official Stridge API documentation.
type SupportedAssetsResponse struct{}

// GatewayStartRequest is a placeholder for a gateway-start request.
// TODO: Define this model from the official Stridge API documentation.
type GatewayStartRequest struct{}

// GatewayStartResponse is a placeholder for a gateway-start response.
// TODO: Define this model from the official Stridge API documentation.
type GatewayStartResponse struct{}

// GatewayPollResponse is a placeholder for a gateway-poll response.
// TODO: Define this model from the official Stridge API documentation.
type GatewayPollResponse struct{}
