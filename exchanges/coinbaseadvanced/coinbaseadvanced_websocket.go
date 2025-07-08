package coinbaseadvanced

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thrasher-corp/gocryptotrader/common/crypto"
	"github.com/thrasher-corp/gocryptotrader/currency"
	gws "github.com/thrasher-corp/gocryptotrader/exchange/websocket"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/exchanges/subscription"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/trade"
	"github.com/thrasher-corp/gocryptotrader/log"
)

const (
	// WebSocket channels
	wsTickerChannel    = "ticker"
	wsOrderbookChannel = "l2_data"
	wsTradesChannel    = "market_trades"
	wsUserChannel      = "user"
)

// WsConnect initiates a websocket connection
func (c *CoinbaseAdvanced) WsConnect() error {
	if !c.Websocket.IsEnabled() || !c.IsEnabled() {
		return gws.ErrWebsocketNotEnabled
	}
	var dialer websocket.Dialer
	err := c.Websocket.Conn.Dial(context.TODO(), &dialer, http.Header{})
	if err != nil {
		return err
	}

	c.Websocket.Wg.Add(1)
	go c.wsReadData()

	return nil
}

// wsReadData receives and processes data from websocket connection
func (c *CoinbaseAdvanced) wsReadData() {
	defer c.Websocket.Wg.Done()

	for {
		resp := c.Websocket.Conn.ReadMessage()
		if resp.Raw == nil {
			return
		}

		err := c.wsHandleData(resp.Raw)
		if err != nil {
			log.Errorln(log.ExchangeSys, err)
		}
	}
}

// wsHandleData processes websocket messages
func (c *CoinbaseAdvanced) wsHandleData(respRaw []byte) error {
	var wsResp WsResponse
	err := json.Unmarshal(respRaw, &wsResp)
	if err != nil {
		return err
	}

	switch wsResp.Channel {
	case wsTickerChannel:
		return c.wsProcessTicker(respRaw)
	case wsOrderbookChannel:
		return c.wsProcessOrderbook(respRaw)
	case wsTradesChannel:
		return c.wsProcessTrades(respRaw)
	case wsUserChannel:
		return c.wsProcessUserData(respRaw)
	default:
		if wsResp.Type == "error" {
			return fmt.Errorf("websocket error: %s", string(respRaw))
		}
		log.Warnf(log.ExchangeSys, "%s unhandled websocket message: %s", c.Name, string(respRaw))
	}

	return nil
}

// wsProcessTicker processes ticker data
func (c *CoinbaseAdvanced) wsProcessTicker(data []byte) error {
	var tickerData WsTickerData
	err := json.Unmarshal(data, &tickerData)
	if err != nil {
		return err
	}

	for _, event := range tickerData.Events {
		if len(event.Tickers) == 0 {
			continue
		}

		for _, t := range event.Tickers {
			pair, err := currency.NewPairFromString(t.ProductID)
			if err != nil {
				return err
			}

			price, err := strconv.ParseFloat(t.Price, 64)
			if err != nil {
				return err
			}

			c.Websocket.DataHandler <- &ticker.Price{
				Last:         price,
				ExchangeName: c.Name,
				AssetType:    asset.Spot,
				Pair:         pair,
			}
		}
	}

	return nil
}

// wsProcessOrderbook processes orderbook data
func (c *CoinbaseAdvanced) wsProcessOrderbook(data []byte) error {
	var orderbookData WsOrderbookData
	err := json.Unmarshal(data, &orderbookData)
	if err != nil {
		return err
	}

	for _, event := range orderbookData.Events {
		pair, err := currency.NewPairFromString(event.ProductID)
		if err != nil {
			return err
		}

		for _, update := range event.Updates {
			price, err := strconv.ParseFloat(update.PriceLevel, 64)
			if err != nil {
				return err
			}

			amount, err := strconv.ParseFloat(update.NewQuantity, 64)
			if err != nil {
				return err
			}

			// Simple orderbook update - in a full implementation, this would
			// properly manage the orderbook state
			log.Debugf(log.ExchangeSys, "%s orderbook update: %s (%s) %s %f@%f",
				c.Name, pair.String(), event.ProductID, update.Side, amount, price)
		}
	}

	return nil
}

// wsProcessTrades processes trade data
func (c *CoinbaseAdvanced) wsProcessTrades(data []byte) error {
	var tradeData WsTradeData
	err := json.Unmarshal(data, &tradeData)
	if err != nil {
		return err
	}

	for _, event := range tradeData.Events {
		for _, t := range event.Trades {
			pair, err := currency.NewPairFromString(t.ProductID)
			if err != nil {
				return err
			}

			price, err := strconv.ParseFloat(t.Price, 64)
			if err != nil {
				return err
			}

			amount, err := strconv.ParseFloat(t.Size, 64)
			if err != nil {
				return err
			}

			timestamp, err := time.Parse(time.RFC3339, t.Time)
			if err != nil {
				return err
			}

			var side order.Side
			if t.Side == "BUY" {
				side = order.Buy
			} else {
				side = order.Sell
			}

			trade := &trade.Data{
				Exchange:     c.Name,
				TID:          t.TradeID,
				CurrencyPair: pair,
				AssetType:    asset.Spot,
				Side:         side,
				Amount:       amount,
				Price:        price,
				Timestamp:    timestamp,
			}

			c.Websocket.DataHandler <- trade
		}
	}

	return nil
}

// wsProcessUserData processes user-specific data
func (c *CoinbaseAdvanced) wsProcessUserData(data []byte) error {
	// TODO: Implement user data processing (orders, fills, etc.)
	log.Debugf(log.ExchangeSys, "%s received user data: %s", c.Name, string(data))
	return nil
}

// GenerateDefaultSubscriptions generates default subscription channels
func (c *CoinbaseAdvanced) GenerateDefaultSubscriptions() (subscription.List, error) {
	var subscriptions subscription.List
	var channels = []string{wsTickerChannel, wsOrderbookChannel, wsTradesChannel}

	enabledPairs, err := c.GetEnabledPairs(asset.Spot)
	if err != nil {
		return nil, err
	}

	for _, channel := range channels {
		for j := range enabledPairs {
			subscriptions = append(subscriptions, &subscription.Subscription{
				Channel: channel,
				Pairs:   currency.Pairs{enabledPairs[j]},
				Asset:   asset.Spot,
			})
		}
	}

	return subscriptions, nil
}

// Subscribe subscribes to websocket channels
func (c *CoinbaseAdvanced) Subscribe(channelsToSubscribe subscription.List) error {
	ctx := context.TODO()
	var errs error
	for _, s := range channelsToSubscribe {
		req := WsSubscribeRequest{
			Type:       "subscribe",
			ProductIDs: s.Pairs.Strings(),
			Channel:    s.Channel,
		}

		if c.Websocket.CanUseAuthenticatedEndpoints() {
			var sig, timestamp string
			var err error
			sig, timestamp, err = c.generateWsSignature()
			if err != nil {
				errs = fmt.Errorf("%v %w", errs, err)
				continue
			}
			req.Signature = sig
			req.Timestamp = timestamp
		}

		err := c.Websocket.Conn.SendJSONMessage(ctx, request.Unset, req)
		if err != nil {
			errs = fmt.Errorf("%v %w", errs, err)
			continue
		}
	}
	if errs == nil {
		errs = c.Websocket.AddSuccessfulSubscriptions(c.Websocket.Conn, channelsToSubscribe...)
	}
	return errs
}

// Unsubscribe unsubscribes from websocket channels
func (c *CoinbaseAdvanced) Unsubscribe(channelsToUnsubscribe subscription.List) error {
	ctx := context.TODO()
	var errs error
	for _, s := range channelsToUnsubscribe {
		req := WsSubscribeRequest{
			Type:       "unsubscribe",
			ProductIDs: s.Pairs.Strings(),
			Channel:    s.Channel,
		}

		err := c.Websocket.Conn.SendJSONMessage(ctx, request.Unset, req)
		if err != nil {
			errs = fmt.Errorf("%v %w", errs, err)
			continue
		}
	}
	return errs
}

// generateWsSignature generates signature for authenticated websocket requests
func (c *CoinbaseAdvanced) generateWsSignature() (string, string, error) {
	credentials, err := c.GetCredentials(context.TODO())
	if err != nil {
		return "", "", err
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	message := timestamp + "GET" + "/users/self/verify"

	h, err := crypto.GetHMAC(crypto.HashSHA256,
		[]byte(message),
		[]byte(credentials.Secret))
	if err != nil {
		return "", "", err
	}

	return base64.StdEncoding.EncodeToString(h), timestamp, nil
}
