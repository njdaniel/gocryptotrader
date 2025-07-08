package coinbaseadvanced

import (
	"time"

	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
)

const (
	coinbaseAdvancedAPIURL = "https://api.coinbase.com/api/v3/brokerage/"
	coinbaseAdvancedWSURL  = "wss://advanced-trade-ws.coinbase.com"

	// Rate limits
	defaultRateInterval = time.Second
	defaultRequestRate  = 10
)

// CoinbaseAdvanced is the overarching type across the CoinbaseAdvanced package
type CoinbaseAdvanced struct {
	exchange.Base
}

// Product represents a product on Coinbase Advanced
type Product struct {
	ProductID        string `json:"product_id"`
	Price            string `json:"price"`
	PricePercent24h  string `json:"price_percentage_change_24h"`
	Volume24h        string `json:"volume_24h"`
	VolumePercent24h string `json:"volume_percentage_change_24h"`
	BaseCurrency     string `json:"base_currency"`
	QuoteCurrency    string `json:"quote_currency"`
	QuoteIncrement   string `json:"quote_increment"`
	BaseIncrement    string `json:"base_increment"`
	Status           string `json:"status"`
	TradingDisabled  bool   `json:"trading_disabled"`
	CancelOnly       bool   `json:"cancel_only"`
	PostOnly         bool   `json:"post_only"`
	LimitOnly        bool   `json:"limit_only"`
	AuctionMode      bool   `json:"auction_mode"`
}

// ProductsResponse represents the response from the products endpoint
type ProductsResponse struct {
	Products []Product `json:"products"`
}

// AccountData represents account information
type AccountData struct {
	UUID             string  `json:"uuid"`
	Name             string  `json:"name"`
	Currency         string  `json:"currency"`
	AvailableBalance Balance `json:"available_balance"`
	Hold             Balance `json:"hold"`
	Default          bool    `json:"default"`
	Active           bool    `json:"active"`
	Type             string  `json:"type"`
}

// Balance represents a balance value
type Balance struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

// AccountsResponse represents the response from the accounts endpoint
type AccountsResponse struct {
	Accounts []AccountData `json:"accounts"`
	HasNext  bool          `json:"has_next"`
	Cursor   string        `json:"cursor"`
	Size     int           `json:"size"`
}

// OrderData represents order information
type OrderData struct {
	OrderID             string      `json:"order_id"`
	ProductID           string      `json:"product_id"`
	UserID              string      `json:"user_id"`
	OrderConfiguration  OrderConfig `json:"order_configuration"`
	Side                string      `json:"side"`
	ClientOrderID       string      `json:"client_order_id"`
	Status              string      `json:"status"`
	TimeInForce         string      `json:"time_in_force"`
	CreatedTime         time.Time   `json:"created_time"`
	CompletionPercent   string      `json:"completion_percentage"`
	FilledSize          string      `json:"filled_size"`
	AverageFilledPrice  string      `json:"average_filled_price"`
	Fee                 string      `json:"fee"`
	NumberOfFills       string      `json:"number_of_fills"`
	FilledValue         string      `json:"filled_value"`
	PendingCancel       bool        `json:"pending_cancel"`
	SizeInQuote         bool        `json:"size_in_quote"`
	TotalFees           string      `json:"total_fees"`
	SizeInclusiveOfFees bool        `json:"size_inclusive_of_fees"`
	TotalValueAfterFees string      `json:"total_value_after_fees"`
	TriggerStatus       string      `json:"trigger_status"`
	OrderType           string      `json:"order_type"`
	RejectReason        string      `json:"reject_reason"`
	Settled             bool        `json:"settled"`
	ProductType         string      `json:"product_type"`
	RejectMessage       string      `json:"reject_message"`
	CancelMessage       string      `json:"cancel_message"`
}

// OrderConfig represents order configuration
type OrderConfig struct {
	MarketMarketIOC struct {
		QuoteSize string `json:"quote_size,omitempty"`
		BaseSize  string `json:"base_size,omitempty"`
	} `json:"market_market_ioc,omitempty"`
	LimitLimitGTC struct {
		BaseSize   string `json:"base_size,omitempty"`
		LimitPrice string `json:"limit_price,omitempty"`
		PostOnly   bool   `json:"post_only,omitempty"`
	} `json:"limit_limit_gtc,omitempty"`
}

// OrderResponse represents the response from placing an order
type OrderResponse struct {
	Success         bool           `json:"success"`
	FailureReason   string         `json:"failure_reason"`
	OrderID         string         `json:"order_id"`
	SuccessResponse *OrderData     `json:"success_response,omitempty"`
	ErrorResponse   *ErrorResponse `json:"error_response,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error           string `json:"error"`
	Message         string `json:"message"`
	ErrorDetails    string `json:"error_details"`
	PreviewFailures []any  `json:"preview_failures"`
}

// OrdersResponse represents the response from listing orders
type OrdersResponse struct {
	Orders  []OrderData `json:"orders"`
	HasNext bool        `json:"has_next"`
	Cursor  string      `json:"cursor"`
}

// CancelOrderResponse represents the response from canceling an order
type CancelOrderResponse struct {
	Success       bool   `json:"success"`
	FailureReason string `json:"failure_reason"`
	Results       []struct {
		Success       bool   `json:"success"`
		FailureReason string `json:"failure_reason"`
		OrderID       string `json:"order_id"`
	} `json:"results"`
}

// TradeData represents trade information
type TradeData struct {
	TradeID   string    `json:"trade_id"`
	ProductID string    `json:"product_id"`
	Price     string    `json:"price"`
	Size      string    `json:"size"`
	Time      time.Time `json:"time"`
	Side      string    `json:"side"`
	Bid       string    `json:"bid"`
	Ask       string    `json:"ask"`
}

// TradesResponse represents the response from market trades endpoint
type TradesResponse struct {
	Trades []TradeData `json:"trades"`
}

// TickerData represents ticker information
type TickerData struct {
	Bid     string    `json:"bid"`
	Ask     string    `json:"ask"`
	Volume  string    `json:"volume"`
	TradeID string    `json:"trade_id"`
	Price   string    `json:"price"`
	Size    string    `json:"size"`
	Time    time.Time `json:"time"`
}

// OrderbookData represents orderbook information
type OrderbookData struct {
	ProductID string     `json:"product_id"`
	Bids      [][]string `json:"bids"`
	Asks      [][]string `json:"asks"`
	Time      time.Time  `json:"time"`
}

// CandleData represents candle/kline information
type CandleData struct {
	Start  time.Time `json:"start"`
	Low    float64   `json:"low,string"`
	High   float64   `json:"high,string"`
	Open   float64   `json:"open,string"`
	Close  float64   `json:"close,string"`
	Volume float64   `json:"volume,string"`
}

// CandlesResponse represents the response from candles endpoint
type CandlesResponse struct {
	Candles []CandleData `json:"candles"`
}

// WebSocketResponse represents a generic websocket response
type WebSocketResponse struct {
	Channel     string    `json:"channel"`
	ClientID    string    `json:"client_id"`
	Timestamp   time.Time `json:"timestamp"`
	SequenceNum int64     `json:"sequence_num"`
	Events      []WSEvent `json:"events"`
}

// WSEvent represents a websocket event
type WSEvent struct {
	Type      string     `json:"type"`
	ProductID string     `json:"product_id,omitempty"`
	Updates   []WSUpdate `json:"updates,omitempty"`
	Tickers   []WSTicker `json:"tickers,omitempty"`
}

// WSUpdate represents a websocket update
type WSUpdate struct {
	Side        string    `json:"side"`
	EventTime   time.Time `json:"event_time"`
	PriceLevel  string    `json:"price_level"`
	NewQuantity string    `json:"new_quantity"`
}

// WSTicker represents a websocket ticker
type WSTicker struct {
	Type            string `json:"type"`
	ProductID       string `json:"product_id"`
	Price           string `json:"price"`
	Volume24h       string `json:"volume_24_h"`
	Low24h          string `json:"low_24_h"`
	High24h         string `json:"high_24_h"`
	Low52w          string `json:"low_52_w"`
	High52w         string `json:"high_52_w"`
	PricePercent24h string `json:"price_percent_chg_24_h"`
}

// WSSubscribeMessage represents a websocket subscription message
type WSSubscribeMessage struct {
	Type       string   `json:"type"`
	ProductIDs []string `json:"product_ids"`
	Channel    string   `json:"channel"`
	JWT        string   `json:"jwt,omitempty"`
	Timestamp  string   `json:"timestamp,omitempty"`
}
