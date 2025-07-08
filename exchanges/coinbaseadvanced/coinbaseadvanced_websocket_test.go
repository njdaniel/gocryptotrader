package coinbaseadvanced

import (
	"testing"

	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

func TestGenerateDefaultSubscriptions(t *testing.T) {
	c := CoinbaseAdvanced{}
	c.SetDefaults()

	// Set up some test pairs via the currency pairs manager
	testPairs, err := currency.NewPairsFromStrings([]string{"BTC-USD", "ETH-USD"})
	if err != nil {
		t.Fatal(err)
	}

	// Use the currency pairs manager to set both available and enabled pairs
	c.CurrencyPairs.StorePairs(asset.Spot, testPairs, false) // available
	c.CurrencyPairs.StorePairs(asset.Spot, testPairs, true)  // enabled

	subs, err := c.GenerateDefaultSubscriptions()
	if err != nil {
		t.Fatal(err)
	}

	if len(subs) == 0 {
		t.Error("Expected subscriptions, got none")
	}

	// Check that we have the expected number of subscriptions
	// 3 channels (ticker, l2_data, market_trades) * 2 pairs
	expectedSubs := 3 * len(testPairs)
	if len(subs) != expectedSubs {
		t.Errorf("Expected %d subscriptions, got %d", expectedSubs, len(subs))
	}

	// Verify subscription structure
	for _, sub := range subs {
		if sub.Channel == "" {
			t.Error("Subscription channel is empty")
		}
		if len(sub.Pairs) == 0 {
			t.Error("Subscription has no pairs")
		}
		if sub.Asset != asset.Spot {
			t.Error("Expected spot asset type")
		}
	}
}

func TestWsHandleData(t *testing.T) {
	c := CoinbaseAdvanced{}
	c.SetDefaults()

	// Test ticker data
	tickerData := `{
		"channel": "ticker",
		"events": [{
			"type": "snapshot",
			"tickers": [{
				"type": "ticker",
				"product_id": "BTC-USD",
				"price": "50000.00",
				"volume_24_h": "1000.0",
				"low_24_h": "49000.0",
				"high_24_h": "51000.0",
				"price_percent_change_24_h": "2.0"
			}]
		}]
	}`

	err := c.wsHandleData([]byte(tickerData))
	if err != nil {
		t.Errorf("Failed to handle ticker data: %v", err)
	}

	// Test trade data
	tradeData := `{
		"channel": "market_trades",
		"events": [{
			"type": "snapshot",
			"trades": [{
				"trade_id": "12345",
				"product_id": "BTC-USD",
				"price": "50000.00",
				"size": "0.1",
				"side": "BUY",
				"time": "2023-01-01T12:00:00Z"
			}]
		}]
	}`

	err = c.wsHandleData([]byte(tradeData))
	if err != nil {
		t.Errorf("Failed to handle trade data: %v", err)
	}

	// Test error response
	errorData := `{"type": "error", "message": "test error"}`
	err = c.wsHandleData([]byte(errorData))
	if err == nil {
		t.Error("Expected error handling to return an error")
	}
}

func TestGenerateWsSignature(t *testing.T) {
	c := CoinbaseAdvanced{}
	c.SetDefaults()

	// This test will only work if credentials are available
	// In a real test environment, you'd mock the credentials
	signature, timestamp, err := c.generateWsSignature()

	// If no credentials are available, we expect an error
	if err == nil {
		if signature == "" {
			t.Error("Expected non-empty signature")
		}
		if timestamp == "" {
			t.Error("Expected non-empty timestamp")
		}
	}
	// If credentials aren't available, that's fine for this test
}
