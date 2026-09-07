package stridge

import "time"

// QuoteRequest contains the query parameters for GET /uda/quote. Quote encodes
// these fields into the URL; this endpoint does not send a JSON request body.
type QuoteRequest struct {
	FromNetworkID int64
	FromAsset     string
	ToNetworkID   int64
	ToAsset       string
	Amount        string

	// FromAddress is the source signer EOA, required for an executable quote.
	FromAddress string
	// ToAddress is the recipient; the provider defaults it to FromAddress.
	ToAddress string
}

// Quote is the response envelope returned by GET /uda/quote.
//
// On success, Message and Data are populated. On failure, Code and Error are
// populated instead. Data is a pointer so an omitted data field can be
// distinguished from a present but empty object.
type Quote struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Data    *QuoteData `json:"data,omitempty"`
	Code    int        `json:"code,omitempty"`
	Error   string     `json:"error,omitempty"`
}

// QuoteData describes the route and amounts returned for a successful quote.
type QuoteData struct {
	From         QuoteAmount `json:"from"`
	To           QuoteAmount `json:"to"`
	ExchangeRate string      `json:"exchange_rate"`
	ExpiresAt    time.Time   `json:"expires_at"`
	Fees         QuoteFees   `json:"fees"`
	Route        QuoteRoute  `json:"route"`
}

// QuoteAmount describes one side of a quote. Amount is expressed in the
// asset's smallest unit and remains a string to preserve integer precision.
type QuoteAmount struct {
	NetworkID    int64  `json:"network_id"`
	AssetAddress string `json:"asset_address"`
	Amount       string `json:"amount"`
}

// QuoteFees contains the fee totals and their individual components.
// Monetary values remain strings because they are integer token amounts.
type QuoteFees struct {
	TotalFee       string         `json:"total_fee"`
	GasFee         string         `json:"gas_fee"`
	NetworkReserve string         `json:"network_reserve"`
	PlatformFee    string         `json:"platform_fee"`
	ProviderFee    string         `json:"provider_fee"`
	Items          []QuoteFeeItem `json:"items"`
}

// QuoteFeeItem describes one component of the total quote fee.
type QuoteFeeItem struct {
	Kind      string `json:"kind"`
	Amount    string `json:"amount"`
	Source    string `json:"source"`
	Recipient string `json:"recipient"`
}

// QuoteRoute identifies the provider route selected for the quote.
type QuoteRoute struct {
	Provider             string `json:"provider"`
	Scenario             string `json:"scenario"`
	EstimatedTimeSeconds int64  `json:"estimated_time_seconds"`
}

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
