package coinbaseadvanced

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

var testExchange *CoinbaseAdvanced

// setupTestExchange creates a test exchange instance
func setupTestExchange(t *testing.T) *CoinbaseAdvanced {
	t.Helper()
	c := &CoinbaseAdvanced{}
	c.SetDefaults()

	exchConfig := &config.Exchange{
		Name:                          "CoinbaseAdvanced",
		Enabled:                       true,
		HTTPTimeout:                   time.Duration(15) * time.Second,
		HTTPUserAgent:                 "GoCryptoTrader",
		HTTPDebugging:                 false,
		WebsocketResponseMaxLimit:     100,
		WebsocketResponseCheckTimeout: time.Second * 5,
		WebsocketTrafficTimeout:       time.Second * 30,
		BaseCurrencies:                currency.Currencies{currency.USD},
		API: config.APIConfig{
			AuthenticatedSupport: true,
			Credentials: config.APICredentialsConfig{
				Key:      "test-key",
				Secret:   "test-secret",
				ClientID: "test-client-id",
			},
		},
	}

	err := c.Setup(exchConfig)
	if err != nil {
		t.Fatalf("Failed to setup exchange: %v", err)
	}

	return c
}

// createMockServer creates a mock HTTP server for testing
func createMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func TestGetProducts(t *testing.T) {
	c := setupTestExchange(t)

	mockResponse := ProductsResponse{
		Products: []Product{
			{
				ProductID:     "BTC-USD",
				Price:         "50000.00",
				BaseCurrency:  "BTC",
				QuoteCurrency: "USD",
				Status:        "online",
			},
			{
				ProductID:     "ETH-USD",
				Price:         "3000.00",
				BaseCurrency:  "ETH",
				QuoteCurrency: "USD",
				Status:        "online",
			},
		},
	}

	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/brokerage/products" {
			t.Errorf("Expected path /api/v3/brokerage/products, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	// Update the API endpoint to use the mock server
	c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: server.URL + "/api/v3/brokerage/",
	})

	products, err := c.GetProducts(context.Background())
	if err != nil {
		t.Fatalf("GetProducts failed: %v", err)
	}

	if len(products) != 2 {
		t.Errorf("Expected 2 products, got %d", len(products))
	}

	if products[0].ProductID != "BTC-USD" {
		t.Errorf("Expected first product to be BTC-USD, got %s", products[0].ProductID)
	}

	if products[1].ProductID != "ETH-USD" {
		t.Errorf("Expected second product to be ETH-USD, got %s", products[1].ProductID)
	}
}

func TestGetProduct(t *testing.T) {
	c := setupTestExchange(t)

	mockProduct := Product{
		ProductID:     "BTC-USD",
		Price:         "50000.00",
		BaseCurrency:  "BTC",
		QuoteCurrency: "USD",
		Status:        "online",
		Volume24h:     "1000.0",
	}

	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v3/brokerage/products/BTC-USD"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockProduct)
	})

	c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: server.URL + "/api/v3/brokerage/",
	})

	product, err := c.GetProduct(context.Background(), "BTC-USD")
	if err != nil {
		t.Fatalf("GetProduct failed: %v", err)
	}

	if product.ProductID != "BTC-USD" {
		t.Errorf("Expected product ID BTC-USD, got %s", product.ProductID)
	}

	if product.Price != "50000.00" {
		t.Errorf("Expected price 50000.00, got %s", product.Price)
	}
}

func TestGetTicker(t *testing.T) {
	c := setupTestExchange(t)

	mockTicker := TickerData{
		Bid:     "49950.00",
		Ask:     "50050.00",
		Volume:  "1000.0",
		TradeID: "12345",
		Price:   "50000.00",
		Size:    "0.1",
		Time:    time.Now(),
	}

	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v3/brokerage/products/BTC-USD/ticker"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockTicker)
	})

	c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: server.URL + "/api/v3/brokerage/",
	})

	ticker, err := c.GetTicker(context.Background(), "BTC-USD")
	if err != nil {
		t.Fatalf("GetTicker failed: %v", err)
	}

	if ticker.Price != "50000.00" {
		t.Errorf("Expected price 50000.00, got %s", ticker.Price)
	}

	if ticker.Bid != "49950.00" {
		t.Errorf("Expected bid 49950.00, got %s", ticker.Bid)
	}

	if ticker.Ask != "50050.00" {
		t.Errorf("Expected ask 50050.00, got %s", ticker.Ask)
	}
}

func TestGetOrderbook(t *testing.T) {
	c := setupTestExchange(t)

	mockOrderbook := OrderbookData{
		ProductID: "BTC-USD",
		Bids: [][]string{
			{"49950.00", "0.5"},
			{"49900.00", "1.0"},
		},
		Asks: [][]string{
			{"50050.00", "0.3"},
			{"50100.00", "0.8"},
		},
		Time: time.Now(),
	}

	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v3/brokerage/products/BTC-USD/book"
		if !strings.HasPrefix(r.URL.Path, expectedPath) {
			t.Errorf("Expected path to start with %s, got %s", expectedPath, r.URL.Path)
		}

		// Check level parameter
		level := r.URL.Query().Get("level")
		if level != "2" {
			t.Errorf("Expected level=2, got level=%s", level)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockOrderbook)
	})

	c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: server.URL + "/api/v3/brokerage/",
	})

	orderbook, err := c.GetOrderbook(context.Background(), "BTC-USD", 2)
	if err != nil {
		t.Fatalf("GetOrderbook failed: %v", err)
	}

	if orderbook.ProductID != "BTC-USD" {
		t.Errorf("Expected product ID BTC-USD, got %s", orderbook.ProductID)
	}

	if len(orderbook.Bids) != 2 {
		t.Errorf("Expected 2 bids, got %d", len(orderbook.Bids))
	}

	if len(orderbook.Asks) != 2 {
		t.Errorf("Expected 2 asks, got %d", len(orderbook.Asks))
	}

	if orderbook.Bids[0][0] != "49950.00" {
		t.Errorf("Expected first bid price 49950.00, got %s", orderbook.Bids[0][0])
	}
}

func TestGetMarketTrades(t *testing.T) {
	c := setupTestExchange(t)

	mockTrades := TradesResponse{
		Trades: []TradeData{
			{
				TradeID:   "12345",
				ProductID: "BTC-USD",
				Price:     "50000.00",
				Size:      "0.1",
				Time:      time.Now(),
				Side:      "buy",
			},
			{
				TradeID:   "12346",
				ProductID: "BTC-USD",
				Price:     "50010.00",
				Size:      "0.05",
				Time:      time.Now(),
				Side:      "sell",
			},
		},
	}

	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v3/brokerage/products/BTC-USD/trades"
		if !strings.HasPrefix(r.URL.Path, expectedPath) {
			t.Errorf("Expected path to start with %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockTrades)
	})

	c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: server.URL + "/api/v3/brokerage/",
	})

	trades, err := c.GetMarketTrades(context.Background(), "BTC-USD", 10, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("GetMarketTrades failed: %v", err)
	}

	if len(trades) != 2 {
		t.Errorf("Expected 2 trades, got %d", len(trades))
	}

	if trades[0].TradeID != "12345" {
		t.Errorf("Expected first trade ID 12345, got %s", trades[0].TradeID)
	}

	if trades[0].Price != "50000.00" {
		t.Errorf("Expected first trade price 50000.00, got %s", trades[0].Price)
	}
}

func TestGetCandles(t *testing.T) {
	c := setupTestExchange(t)

	mockCandles := CandlesResponse{
		Candles: []CandleData{
			{
				Start:  time.Now().Add(-time.Hour),
				Low:    49000.0,
				High:   51000.0,
				Open:   50000.0,
				Close:  50500.0,
				Volume: 100.0,
			},
			{
				Start:  time.Now(),
				Low:    50000.0,
				High:   52000.0,
				Open:   50500.0,
				Close:  51000.0,
				Volume: 150.0,
			},
		},
	}

	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v3/brokerage/products/BTC-USD/candles"
		if !strings.HasPrefix(r.URL.Path, expectedPath) {
			t.Errorf("Expected path to start with %s, got %s", expectedPath, r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if !query.Has("start") {
			t.Error("Expected start parameter")
		}
		if !query.Has("end") {
			t.Error("Expected end parameter")
		}
		if !query.Has("granularity") {
			t.Error("Expected granularity parameter")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockCandles)
	})

	c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: server.URL + "/api/v3/brokerage/",
	})

	start := time.Now().Add(-2 * time.Hour)
	end := time.Now()
	granularity := time.Hour

	candles, err := c.GetCandles(context.Background(), "BTC-USD", start, end, granularity)
	if err != nil {
		t.Fatalf("GetCandles failed: %v", err)
	}

	if len(candles) != 2 {
		t.Errorf("Expected 2 candles, got %d", len(candles))
	}

	if candles[0].Open != 50000.0 {
		t.Errorf("Expected first candle open 50000.0, got %f", candles[0].Open)
	}

	if candles[0].Close != 50500.0 {
		t.Errorf("Expected first candle close 50500.0, got %f", candles[0].Close)
	}
}

func TestPlaceMarketOrder(t *testing.T) {
	c := setupTestExchange(t)

	mockOrderData := &OrderData{
		OrderID:   "test-order-123",
		ProductID: "BTC-USD",
		Side:      "buy",
		Status:    "filled",
	}

	mockResponse := OrderResponse{
		Success:         true,
		OrderID:         "test-order-123",
		SuccessResponse: mockOrderData,
	}

	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v3/brokerage/orders"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Check request body
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if reqBody["product_id"] != "BTC-USD" {
			t.Errorf("Expected product_id BTC-USD, got %v", reqBody["product_id"])
		}

		if reqBody["side"] != "buy" {
			t.Errorf("Expected side buy, got %v", reqBody["side"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: server.URL + "/api/v3/brokerage/",
	})

	// Test buy order with quote size
	orderData, err := c.PlaceMarketOrder(context.Background(), "BTC-USD", "buy", 0, 1000.0)
	if err != nil {
		t.Fatalf("PlaceMarketOrder failed: %v", err)
	}

	if orderData.OrderID != "test-order-123" {
		t.Errorf("Expected order ID test-order-123, got %s", orderData.OrderID)
	}

	if orderData.ProductID != "BTC-USD" {
		t.Errorf("Expected product ID BTC-USD, got %s", orderData.ProductID)
	}
}

func TestPlaceLimitOrder(t *testing.T) {
	c := setupTestExchange(t)

	mockOrderData := &OrderData{
		OrderID:   "test-limit-order-456",
		ProductID: "BTC-USD",
		Side:      "sell",
		Status:    "open",
	}

	mockResponse := OrderResponse{
		Success:         true,
		OrderID:         "test-limit-order-456",
		SuccessResponse: mockOrderData,
	}

	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v3/brokerage/orders"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Check request body
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		orderConfig, ok := reqBody["order_configuration"].(map[string]interface{})
		if !ok {
			t.Error("Expected order_configuration in request")
		}

		limitConfig, ok := orderConfig["limit_limit_gtc"].(map[string]interface{})
		if !ok {
			t.Error("Expected limit_limit_gtc in order_configuration")
		}

		if limitConfig["limit_price"] != "51000.00" {
			t.Errorf("Expected limit_price 51000.00, got %v", limitConfig["limit_price"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: server.URL + "/api/v3/brokerage/",
	})

	orderData, err := c.PlaceLimitOrder(context.Background(), "BTC-USD", "sell", 0.1, 51000.0, "GTC")
	if err != nil {
		t.Fatalf("PlaceLimitOrder failed: %v", err)
	}

	if orderData.OrderID != "test-limit-order-456" {
		t.Errorf("Expected order ID test-limit-order-456, got %s", orderData.OrderID)
	}

	if orderData.Side != "sell" {
		t.Errorf("Expected side sell, got %s", orderData.Side)
	}
}

func TestParseOrderData(t *testing.T) {
	c := setupTestExchange(t)

	orderData := &OrderData{
		OrderID:            "test-order-789",
		ProductID:          "BTC-USD",
		Side:               "buy",
		Status:             "filled",
		OrderType:          "limit",
		CreatedTime:        time.Now(),
		FilledSize:         "0.1",
		AverageFilledPrice: "50000.00",
		Fee:                "25.00",
	}

	orderDetail, err := c.parseOrderData(orderData, asset.Spot)
	if err != nil {
		t.Fatalf("parseOrderData failed: %v", err)
	}

	if orderDetail.OrderID != "test-order-789" {
		t.Errorf("Expected order ID test-order-789, got %s", orderDetail.OrderID)
	}

	if orderDetail.Side != order.Buy {
		t.Errorf("Expected side Buy, got %s", orderDetail.Side)
	}

	if orderDetail.Type != order.Limit {
		t.Errorf("Expected type Limit, got %s", orderDetail.Type)
	}

	if orderDetail.Price != 50000.0 {
		t.Errorf("Expected price 50000.0, got %f", orderDetail.Price)
	}

	if orderDetail.Amount != 0.1 {
		t.Errorf("Expected amount 0.1, got %f", orderDetail.Amount)
	}

	if orderDetail.Fee != 25.0 {
		t.Errorf("Expected fee 25.0, got %f", orderDetail.Fee)
	}

	expectedPair, _ := currency.NewPairFromString("BTC-USD")
	if !orderDetail.Pair.Equal(expectedPair) {
		t.Errorf("Expected pair %s, got %s", expectedPair, orderDetail.Pair)
	}
}

func TestPlaceMarketOrderErrors(t *testing.T) {
	c := setupTestExchange(t)

	// Test invalid parameters
	_, err := c.PlaceMarketOrder(context.Background(), "BTC-USD", "buy", 0, 0)
	if err == nil {
		t.Error("Expected error for invalid parameters")
	}

	_, err = c.PlaceMarketOrder(context.Background(), "BTC-USD", "sell", 0, 0)
	if err == nil {
		t.Error("Expected error for invalid parameters")
	}

	// Test failed order response
	mockResponse := OrderResponse{
		Success:       false,
		FailureReason: "Insufficient funds",
	}

	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: server.URL + "/api/v3/brokerage/",
	})

	_, err = c.PlaceMarketOrder(context.Background(), "BTC-USD", "buy", 0, 1000.0)
	if err == nil {
		t.Error("Expected error for failed order")
	}

	expectedError := "order failed: Insufficient funds"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// Additional tests for remaining functions

func TestGetAccounts(t *testing.T) {
	c := setupTestExchange(t)

	mockAccounts := AccountsResponse{
		Accounts: []AccountData{
			{
				UUID:     "account-1",
				Name:     "BTC Wallet",
				Currency: "BTC",
				AvailableBalance: Balance{
					Value:    "1.0",
					Currency: "BTC",
				},
				Active: true,
			},
			{
				UUID:     "account-2",
				Name:     "USD Wallet",
				Currency: "USD",
				AvailableBalance: Balance{
					Value:    "10000.0",
					Currency: "USD",
				},
				Active: true,
			},
		},
		HasNext: false,
		Size:    2,
	}

	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v3/brokerage/accounts"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockAccounts)
	})

	c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: server.URL + "/api/v3/brokerage/",
	})

	accounts, err := c.GetAccounts(context.Background())
	if err != nil {
		t.Fatalf("GetAccounts failed: %v", err)
	}

	if len(accounts) != 2 {
		t.Errorf("Expected 2 accounts, got %d", len(accounts))
	}

	if accounts[0].Currency != "BTC" {
		t.Errorf("Expected first account currency BTC, got %s", accounts[0].Currency)
	}

	if accounts[1].Currency != "USD" {
		t.Errorf("Expected second account currency USD, got %s", accounts[1].Currency)
	}
}

func TestGetOrderByID(t *testing.T) {
	c := setupTestExchange(t)

	mockOrder := OrderData{
		OrderID:            "test-order-123",
		ProductID:          "BTC-USD",
		Side:               "buy",
		Status:             "filled",
		OrderType:          "market",
		CreatedTime:        time.Now(),
		FilledSize:         "0.1",
		AverageFilledPrice: "50000.00",
	}

	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v3/brokerage/orders/historical/test-order-123"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockOrder)
	})

	c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: server.URL + "/api/v3/brokerage/",
	})

	order, err := c.GetOrderByID(context.Background(), "test-order-123")
	if err != nil {
		t.Fatalf("GetOrderByID failed: %v", err)
	}

	if order.OrderID != "test-order-123" {
		t.Errorf("Expected order ID test-order-123, got %s", order.OrderID)
	}

	if order.ProductID != "BTC-USD" {
		t.Errorf("Expected product ID BTC-USD, got %s", order.ProductID)
	}
}

func TestListOrders(t *testing.T) {
	c := setupTestExchange(t)

	mockOrders := OrdersResponse{
		Orders: []OrderData{
			{
				OrderID:   "order-1",
				ProductID: "BTC-USD",
				Side:      "buy",
				Status:    "open",
			},
			{
				OrderID:   "order-2",
				ProductID: "ETH-USD",
				Side:      "sell",
				Status:    "filled",
			},
		},
		HasNext: false,
	}

	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v3/brokerage/orders/historical/batch"
		if !strings.HasPrefix(r.URL.Path, expectedPath) {
			t.Errorf("Expected path to start with %s, got %s", expectedPath, r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("product_id") != "BTC-USD" {
			t.Errorf("Expected product_id BTC-USD, got %s", query.Get("product_id"))
		}
		if query.Get("order_status") != "OPEN" {
			t.Errorf("Expected order_status OPEN, got %s", query.Get("order_status"))
		}
		if query.Get("limit") != "10" {
			t.Errorf("Expected limit 10, got %s", query.Get("limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockOrders)
	})

	c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: server.URL + "/api/v3/brokerage/",
	})

	orders, err := c.ListOrders(context.Background(), "BTC-USD", "OPEN", "", "", 10)
	if err != nil {
		t.Fatalf("ListOrders failed: %v", err)
	}

	if len(orders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(orders))
	}

	if orders[0].OrderID != "order-1" {
		t.Errorf("Expected first order ID order-1, got %s", orders[0].OrderID)
	}
}

func TestCancelOrderByID(t *testing.T) {
	c := setupTestExchange(t)

	mockResponse := CancelOrderResponse{
		Success: true,
		Results: []struct {
			Success       bool   `json:"success"`
			FailureReason string `json:"failure_reason"`
			OrderID       string `json:"order_id"`
		}{
			{
				Success: true,
				OrderID: "test-order-123",
			},
		},
	}

	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v3/brokerage/orders/batch_cancel"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Check request body
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		orderIDs, ok := reqBody["order_ids"].([]interface{})
		if !ok || len(orderIDs) != 1 {
			t.Error("Expected order_ids array with 1 element")
		}

		if orderIDs[0] != "test-order-123" {
			t.Errorf("Expected order ID test-order-123, got %v", orderIDs[0])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: server.URL + "/api/v3/brokerage/",
	})

	response, err := c.CancelOrderByID(context.Background(), "test-order-123")
	if err != nil {
		t.Fatalf("CancelOrderByID failed: %v", err)
	}

	if !response.Success {
		t.Error("Expected successful cancel response")
	}

	if len(response.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(response.Results))
	}

	if response.Results[0].OrderID != "test-order-123" {
		t.Errorf("Expected order ID test-order-123, got %s", response.Results[0].OrderID)
	}
}

func TestCancelAllOrdersByProductID(t *testing.T) {
	c := setupTestExchange(t)

	// Mock response for listing orders
	mockOrders := OrdersResponse{
		Orders: []OrderData{
			{
				OrderID:   "order-1",
				ProductID: "BTC-USD",
				Status:    "open",
			},
			{
				OrderID:   "order-2",
				ProductID: "BTC-USD",
				Status:    "open",
			},
		},
	}

	// Mock response for cancel orders
	mockCancelResponse := CancelOrderResponse{
		Success: true,
		Results: []struct {
			Success       bool   `json:"success"`
			FailureReason string `json:"failure_reason"`
			OrderID       string `json:"order_id"`
		}{
			{Success: true, OrderID: "order-1"},
			{Success: true, OrderID: "order-2"},
		},
	}

	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "historical/batch") {
			// List orders request
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockOrders)
		} else if strings.Contains(r.URL.Path, "batch_cancel") {
			// Cancel orders request
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockCancelResponse)
		} else {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	})

	c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: server.URL + "/api/v3/brokerage/",
	})

	orders, err := c.CancelAllOrdersByProductID(context.Background(), "BTC-USD")
	if err != nil {
		t.Fatalf("CancelAllOrdersByProductID failed: %v", err)
	}

	if len(orders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(orders))
	}

	if orders[0].OrderID != "order-1" {
		t.Errorf("Expected first order ID order-1, got %s", orders[0].OrderID)
	}
}

func TestParseOrderDataErrors(t *testing.T) {
	c := setupTestExchange(t)

	// Test invalid product ID
	invalidOrder := &OrderData{
		OrderID:            "test-order",
		ProductID:          "", // Empty product ID should cause error
		Side:               "buy",
		Status:             "filled",
		OrderType:          "market",
		CreatedTime:        time.Now(),
		FilledSize:         "0.1",
		AverageFilledPrice: "50000.00",
		Fee:                "25.00",
	}

	_, err := c.parseOrderData(invalidOrder, asset.Spot)
	if err == nil {
		t.Error("Expected error for invalid product ID")
	}

	// Test invalid side
	invalidSideOrder := &OrderData{
		OrderID:            "test-order",
		ProductID:          "BTC-USD",
		Side:               "invalid-side",
		Status:             "filled",
		OrderType:          "market",
		CreatedTime:        time.Now(),
		FilledSize:         "0.1",
		AverageFilledPrice: "50000.00",
		Fee:                "25.00",
	}

	_, err = c.parseOrderData(invalidSideOrder, asset.Spot)
	if err == nil {
		t.Error("Expected error for invalid side")
	}

	// Test invalid filled size
	invalidFilledSizeOrder := &OrderData{
		OrderID:            "test-order",
		ProductID:          "BTC-USD",
		Side:               "buy",
		Status:             "filled",
		OrderType:          "market",
		CreatedTime:        time.Now(),
		FilledSize:         "invalid-size",
		AverageFilledPrice: "50000.00",
		Fee:                "25.00",
	}

	_, err = c.parseOrderData(invalidFilledSizeOrder, asset.Spot)
	if err == nil {
		t.Error("Expected error for invalid filled size")
	}
}
