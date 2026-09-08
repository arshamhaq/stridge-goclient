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

// SupportedAssetsResponse is returned by GET /uda/supported-assets.
type SupportedAssetsResponse struct {
	Assets []SupportedAssetNetwork `json:"assets"`
}

// SupportedAssetNetwork describes one UDA-enabled blockchain and its assets.
type SupportedAssetNetwork struct {
	NetworkID      string           `json:"network_id"`
	NetworkName    string           `json:"network_name"`
	NetworkSymbol  string           `json:"network_symbol"`
	EIP155ID       int64            `json:"eip155_id"`
	ChainType      string           `json:"chain_type"`
	NativeCurrency SupportedAsset   `json:"native_currency"`
	Assets         []SupportedAsset `json:"assets"`
}

// SupportedAsset describes a native currency or configured token contract.
// Decimal and USD-related values use the API's documented representation.
type SupportedAsset struct {
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	Address       string `json:"address"`
	Decimals      int64  `json:"decimals"`
	Logo          string `json:"logo"`
	MinDepositUSD string `json:"min_deposit_usd"`
	PriceImpact   string `json:"price_impact"`
}

// CreateUDARequest describes POST /uda. Destination may be either an inline
// destination or a named treasury.
type CreateUDARequest struct {
	Owner          string                 `json:"owner"`
	Destination    CreateUDADestination   `json:"destination"`
	AcceptedAssets []string               `json:"accepted_assets,omitempty"`
	RoutingRules   []CreateUDARoutingRule `json:"routing_rules,omitempty"`
}

// CreateUDADestination selects where deposits ultimately settle. Treasury is
// an alternative to the inline Address, NetworkID, and AssetSymbol fields.
type CreateUDADestination struct {
	Address     string `json:"address,omitempty"`
	NetworkID   string `json:"network_id,omitempty"`
	AssetSymbol string `json:"asset_symbol,omitempty"`
	Treasury    string `json:"treasury,omitempty"`
}

// CreateUDARoutingRule overrides the default destination for a source match.
type CreateUDARoutingRule struct {
	Match       CreateUDARoutingMatch `json:"match"`
	Destination CreateUDADestination  `json:"destination"`
}

// CreateUDARoutingMatch selects deposits to which a routing rule applies.
type CreateUDARoutingMatch struct {
	SourceNetworkID   string `json:"source_network_id"`
	SourceAssetSymbol string `json:"source_asset_symbol,omitempty"`
}

// CreateUDAResponse is returned by POST /uda for both newly created and
// existing UDAs.
type CreateUDAResponse struct {
	ID               string              `json:"id"`
	Owner            string              `json:"owner"`
	Status           string              `json:"status"`
	CreatedAt        time.Time           `json:"created_at"`
	Destination      UDADestination      `json:"destination"`
	DepositAddresses []UDADepositAddress `json:"deposit_addresses"`
	RoutingRules     []UDARoutingRule    `json:"routing_rules,omitempty"`
}

// UDADestination is the resolved settlement destination returned by Stridge.
type UDADestination struct {
	Address       string `json:"address"`
	NetworkID     string `json:"network_id"`
	EIP155ID      string `json:"eip155_id"`
	NetworkName   string `json:"network_name"`
	AssetAddress  string `json:"asset_address"`
	AssetSymbol   string `json:"asset_symbol"`
	AssetDecimals int64  `json:"asset_decimals"`
	Treasury      string `json:"treasury,omitempty"`
}

// UDADepositAddress describes a source-chain address and the assets it accepts.
type UDADepositAddress struct {
	Address        string             `json:"address"`
	NetworkID      string             `json:"network_id,omitempty"`
	EIP155ID       string             `json:"eip155_id"`
	NetworkName    string             `json:"network_name"`
	AcceptedAssets []UDAAcceptedAsset `json:"accepted_assets"`
}

// UDAAcceptedAsset describes a token accepted at a UDA deposit address.
type UDAAcceptedAsset struct {
	Symbol   string `json:"symbol"`
	Address  string `json:"address"`
	Decimals int64  `json:"decimals"`
	Logo     string `json:"logo,omitempty"`
}

// UDARoutingRule is a resolved routing rule returned by Stridge.
type UDARoutingRule struct {
	Match       UDARoutingMatch `json:"match"`
	Destination UDADestination  `json:"destination"`
}

// UDARoutingMatch is the resolved source selector returned by Stridge.
type UDARoutingMatch struct {
	SourceNetworkID    string `json:"source_network_id"`
	SourceAssetSymbol  string `json:"source_asset_symbol,omitempty"`
	SourceTokenAddress string `json:"source_token_address,omitempty"`
}

// GatewayStartRequest describes POST /gateway/start.
type GatewayStartRequest struct {
	Owner       string                  `json:"owner"`
	Destination GatewayStartDestination `json:"destination"`
	Metadata    map[string]any          `json:"metadata,omitempty"`
}

// GatewayStartDestination selects the asset and address that receive funds.
type GatewayStartDestination struct {
	ToAddress   string `json:"to_address"`
	NetworkID   string `json:"network_id"`
	AssetSymbol string `json:"asset_symbol"`
}

// GatewayStartResponse is the response envelope returned by POST
// /gateway/start.
type GatewayStartResponse struct {
	Data *GatewayStartData `json:"data"`
}

// GatewayStartData contains the provisioned UDA and its deposit addresses.
type GatewayStartData struct {
	UDAID            string                  `json:"uda_id"`
	Owner            string                  `json:"owner"`
	Status           string                  `json:"status"`
	Destination      GatewayStartDestination `json:"destination"`
	DepositAddresses []UDADepositAddress     `json:"deposit_addresses"`
	Metadata         map[string]any          `json:"metadata,omitempty"`
}

// GatewayPollResponse is a placeholder for a gateway-poll response.
// TODO: Define this model from the official Stridge API documentation.
type GatewayPollResponse struct{}
