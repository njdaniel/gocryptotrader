package coinbaseadvanced

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
)

// GetProducts retrieves product information from Coinbase Advanced
func (c *CoinbaseAdvanced) GetProducts(ctx context.Context) ([]Product, error) {
	var resp ProductsResponse
	err := c.SendHTTPRequest(ctx, exchange.RestSpot, "products", &resp)
	if err != nil {
		return nil, err
	}
	return resp.Products, nil
}

// GetProduct retrieves specific product information
func (c *CoinbaseAdvanced) GetProduct(ctx context.Context, productID string) (*Product, error) {
	var resp Product
	path := fmt.Sprintf("products/%s", productID)
	err := c.SendHTTPRequest(ctx, exchange.RestSpot, path, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAccounts retrieves account information
func (c *CoinbaseAdvanced) GetAccounts(ctx context.Context) ([]AccountData, error) {
	var resp AccountsResponse
	err := c.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, "accounts", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Accounts, nil
}

// GetTicker retrieves ticker information for a product
func (c *CoinbaseAdvanced) GetTicker(ctx context.Context, productID string) (*TickerData, error) {
	var resp TickerData
	path := fmt.Sprintf("products/%s/ticker", productID)
	err := c.SendHTTPRequest(ctx, exchange.RestSpot, path, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetOrderbook retrieves orderbook information for a product
func (c *CoinbaseAdvanced) GetOrderbook(ctx context.Context, productID string, level int) (*OrderbookData, error) {
	var resp OrderbookData
	path := fmt.Sprintf("products/%s/book?level=%d", productID, level)
	err := c.SendHTTPRequest(ctx, exchange.RestSpot, path, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMarketTrades retrieves market trades for a product
func (c *CoinbaseAdvanced) GetMarketTrades(ctx context.Context, productID string, limit int, start, end time.Time) ([]TradeData, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if !start.IsZero() {
		params.Set("start", start.Format(time.RFC3339))
	}
	if !end.IsZero() {
		params.Set("end", end.Format(time.RFC3339))
	}

	path := fmt.Sprintf("products/%s/trades", productID)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var resp TradesResponse
	err := c.SendHTTPRequest(ctx, exchange.RestSpot, path, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Trades, nil
}

// GetCandles retrieves candle data for a product
func (c *CoinbaseAdvanced) GetCandles(ctx context.Context, productID string, start, end time.Time, granularity time.Duration) ([]CandleData, error) {
	params := url.Values{}
	params.Set("start", start.Format(time.RFC3339))
	params.Set("end", end.Format(time.RFC3339))
	params.Set("granularity", strconv.Itoa(int(granularity.Seconds())))

	path := fmt.Sprintf("products/%s/candles?%s", productID, params.Encode())

	var resp CandlesResponse
	err := c.SendHTTPRequest(ctx, exchange.RestSpot, path, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Candles, nil
}

// PlaceMarketOrder places a market order
func (c *CoinbaseAdvanced) PlaceMarketOrder(ctx context.Context, productID, side string, baseSize, quoteSize float64) (*OrderData, error) {
	req := map[string]interface{}{
		"product_id": productID,
		"side":       side,
	}

	if side == "buy" && quoteSize > 0 {
		req["order_configuration"] = map[string]interface{}{
			"market_market_ioc": map[string]interface{}{
				"quote_size": fmt.Sprintf("%.8f", quoteSize),
			},
		}
	} else if side == "sell" && baseSize > 0 {
		req["order_configuration"] = map[string]interface{}{
			"market_market_ioc": map[string]interface{}{
				"base_size": fmt.Sprintf("%.8f", baseSize),
			},
		}
	} else {
		return nil, fmt.Errorf("invalid market order parameters")
	}

	var resp OrderResponse
	err := c.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, "orders", req, &resp)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("order failed: %s", resp.FailureReason)
	}

	return resp.SuccessResponse, nil
}

// PlaceLimitOrder places a limit order
func (c *CoinbaseAdvanced) PlaceLimitOrder(ctx context.Context, productID, side string, baseSize, limitPrice float64, timeInForce string) (*OrderData, error) {
	req := map[string]interface{}{
		"product_id": productID,
		"side":       side,
		"order_configuration": map[string]interface{}{
			"limit_limit_gtc": map[string]interface{}{
				"base_size":   fmt.Sprintf("%.8f", baseSize),
				"limit_price": fmt.Sprintf("%.2f", limitPrice),
				"post_only":   false,
			},
		},
	}

	var resp OrderResponse
	err := c.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, "orders", req, &resp)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("order failed: %s", resp.FailureReason)
	}

	return resp.SuccessResponse, nil
}

// GetOrderByID retrieves order information by ID
func (c *CoinbaseAdvanced) GetOrderByID(ctx context.Context, orderID string) (*OrderData, error) {
	var resp OrderData
	path := fmt.Sprintf("orders/historical/%s", orderID)
	err := c.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListOrders retrieves a list of orders
func (c *CoinbaseAdvanced) ListOrders(ctx context.Context, productID, orderStatus, orderType, orderSide string, limit int) ([]OrderData, error) {
	params := url.Values{}
	if productID != "" {
		params.Set("product_id", productID)
	}
	if orderStatus != "" {
		params.Set("order_status", orderStatus)
	}
	if orderType != "" {
		params.Set("order_type", orderType)
	}
	if orderSide != "" {
		params.Set("order_side", orderSide)
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	path := "orders/historical/batch"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var resp OrdersResponse
	err := c.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Orders, nil
}

// CancelOrderByID cancels an order by ID
func (c *CoinbaseAdvanced) CancelOrderByID(ctx context.Context, orderID string) (*CancelOrderResponse, error) {
	req := map[string]interface{}{
		"order_ids": []string{orderID},
	}

	var resp CancelOrderResponse
	err := c.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, "orders/batch_cancel", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CancelAllOrdersByProductID cancels all orders for a product
func (c *CoinbaseAdvanced) CancelAllOrdersByProductID(ctx context.Context, productID string) ([]OrderData, error) {
	// First get all open orders for the product
	orders, err := c.ListOrders(ctx, productID, "OPEN", "", "", 0)
	if err != nil {
		return nil, err
	}

	var orderIDs []string
	for _, order := range orders {
		orderIDs = append(orderIDs, order.OrderID)
	}

	if len(orderIDs) == 0 {
		return nil, nil
	}

	req := map[string]interface{}{
		"order_ids": orderIDs,
	}

	var resp CancelOrderResponse
	err = c.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, "orders/batch_cancel", req, &resp)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (c *CoinbaseAdvanced) SendHTTPRequest(ctx context.Context, ep exchange.URL, path string, result interface{}) error {
	endpoint, err := c.API.Endpoints.GetURL(ep)
	if err != nil {
		return err
	}

	item := &request.Item{
		Method:        http.MethodGet,
		Path:          endpoint + path,
		Result:        result,
		Verbose:       c.Verbose,
		HTTPDebugging: c.HTTPDebugging,
		HTTPRecording: c.HTTPRecording,
	}

	return c.SendPayload(ctx, request.Unset, func() (*request.Item, error) {
		return item, nil
	}, request.UnauthenticatedRequest)
}

// SendAuthenticatedHTTPRequest sends an authenticated HTTP request
func (c *CoinbaseAdvanced) SendAuthenticatedHTTPRequest(ctx context.Context, ep exchange.URL, path string, params interface{}, result interface{}) error {
	creds, err := c.GetCredentials(ctx)
	if err != nil {
		return err
	}

	endpoint, err := c.API.Endpoints.GetURL(ep)
	if err != nil {
		return err
	}

	item := &request.Item{
		Method:        http.MethodGet,
		Path:          endpoint + path,
		Result:        result,
		Verbose:       c.Verbose,
		HTTPDebugging: c.HTTPDebugging,
		HTTPRecording: c.HTTPRecording,
	}

	if params != nil {
		payload, err := json.Marshal(params)
		if err != nil {
			return err
		}
		item.Method = http.MethodPost
		item.Body = bytes.NewReader(payload)
		item.Headers = map[string]string{
			"Content-Type": "application/json",
		}
	}

	// TODO: Add proper ECDSA signing here
	// This is a placeholder - real implementation would use Coinbase's JWT signing
	item.Headers = map[string]string{
		"Authorization": "Bearer " + creds.Key,
	}

	return c.SendPayload(ctx, request.Auth, func() (*request.Item, error) {
		return item, nil
	}, request.AuthenticatedRequest)
}

// parseOrderData converts OrderData to order.Detail
func (c *CoinbaseAdvanced) parseOrderData(orderData *OrderData, assetType asset.Item) (*order.Detail, error) {
	pair, err := currency.NewPairFromString(orderData.ProductID)
	if err != nil {
		return nil, err
	}

	side, err := order.StringToOrderSide(orderData.Side)
	if err != nil {
		return nil, err
	}

	status, err := order.StringToOrderStatus(orderData.Status)
	if err != nil {
		return nil, err
	}

	var orderType order.Type
	switch orderData.OrderType {
	case "market":
		orderType = order.Market
	case "limit":
		orderType = order.Limit
	default:
		orderType = order.UnknownType
	}

	filledSize, err := strconv.ParseFloat(orderData.FilledSize, 64)
	if err != nil {
		return nil, err
	}

	averagePrice, err := strconv.ParseFloat(orderData.AverageFilledPrice, 64)
	if err != nil {
		return nil, err
	}

	fee, err := strconv.ParseFloat(orderData.Fee, 64)
	if err != nil {
		return nil, err
	}

	return &order.Detail{
		Exchange:        c.Name,
		OrderID:         orderData.OrderID,
		Pair:            pair,
		Side:            side,
		Type:            orderType,
		Date:            orderData.CreatedTime,
		Status:          status,
		Price:           averagePrice,
		Amount:          filledSize,
		ExecutedAmount:  filledSize,
		RemainingAmount: 0, // TODO: Calculate remaining amount
		Fee:             fee,
		AssetType:       assetType,
	}, nil
}
