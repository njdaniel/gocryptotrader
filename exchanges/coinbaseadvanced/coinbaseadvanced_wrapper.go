package coinbaseadvanced

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/deposit"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/protocol"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/trade"
	"github.com/thrasher-corp/gocryptotrader/log"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
)

// SetDefaults sets default values for the exchange
func (c *CoinbaseAdvanced) SetDefaults() {
	c.Name = "CoinbaseAdvanced"
	c.Enabled = true
	c.Verbose = true
	c.API.CredentialsValidator.RequiresKey = true
	c.API.CredentialsValidator.RequiresSecret = true

	requestFmt := &currency.PairFormat{Delimiter: currency.DashDelimiter, Uppercase: true}
	configFmt := &currency.PairFormat{Delimiter: currency.DashDelimiter, Uppercase: true}
	err := c.SetGlobalPairsManager(requestFmt, configFmt, asset.Spot)
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	c.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: true,
			RESTCapabilities: protocol.Features{
				TickerFetching:    true,
				KlineFetching:     true,
				TradeFetching:     true,
				OrderbookFetching: true,
				AutoPairUpdates:   true,
				AccountInfo:       true,
				GetOrder:          true,
				GetOrders:         true,
				CancelOrders:      true,
				CancelOrder:       true,
				SubmitOrder:       true,
				DepositHistory:    true,
				WithdrawalHistory: true,
				UserTradeHistory:  true,
				CryptoDeposit:     true,
				CryptoWithdrawal:  true,
				TradeFee:          true,
				CandleHistory:     true,
			},
			WebsocketCapabilities: protocol.Features{
				TickerFetching:         true,
				OrderbookFetching:      true,
				Subscribe:              true,
				Unsubscribe:            true,
				AuthenticatedEndpoints: true,
				MessageSequenceNumbers: true,
				GetOrders:              true,
				GetOrder:               true,
			},
			WithdrawPermissions: exchange.AutoWithdrawCryptoWithAPIPermission,
			Kline: kline.ExchangeCapabilitiesSupported{
				DateRanges: true,
				Intervals:  true,
			},
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
			Kline: kline.ExchangeCapabilitiesEnabled{
				Intervals: kline.DeployExchangeIntervals(
					kline.IntervalCapacity{Interval: kline.OneMin},
					kline.IntervalCapacity{Interval: kline.FiveMin},
					kline.IntervalCapacity{Interval: kline.FifteenMin},
					kline.IntervalCapacity{Interval: kline.ThirtyMin},
					kline.IntervalCapacity{Interval: kline.OneHour},
					kline.IntervalCapacity{Interval: kline.TwoHour},
					kline.IntervalCapacity{Interval: kline.SixHour},
					kline.IntervalCapacity{Interval: kline.OneDay},
				),
			},
		},
	}

	c.Requester, err = request.New(c.Name,
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout),
		request.WithLimiter(SetRateLimit()))
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}
	c.API.Endpoints = c.NewEndpoints()
	err = c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot:      coinbaseAdvancedAPIURL,
		exchange.WebsocketSpot: coinbaseAdvancedWSURL,
	})
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}
}

// Setup sets exchange configuration profile
func (c *CoinbaseAdvanced) Setup(exch *config.Exchange) error {
	err := exch.Validate()
	if err != nil {
		return err
	}
	if !exch.Enabled {
		c.SetEnabled(false)
		return nil
	}
	err = c.SetupDefaults(exch)
	if err != nil {
		return err
	}

	return nil
}

// GetDefaultConfig returns a default exchange config
func (c *CoinbaseAdvanced) GetDefaultConfig(ctx context.Context) (*config.Exchange, error) {
	c.SetDefaults()
	exchCfg := new(config.Exchange)
	exchCfg.Name = c.Name
	exchCfg.HTTPTimeout = exchange.DefaultHTTPTimeout
	exchCfg.BaseCurrencies = currency.Currencies{currency.USD, currency.USDC}

	err := c.SetupDefaults(exchCfg)
	if err != nil {
		return nil, err
	}

	if c.Features.Supports.RESTCapabilities.AutoPairUpdates {
		err := c.UpdateTradablePairs(ctx, true)
		if err != nil {
			return nil, err
		}
	}

	return exchCfg, nil
}

// Start starts the CoinbaseAdvanced go routine
func (c *CoinbaseAdvanced) Start(ctx context.Context, wg *sync.WaitGroup) error {
	if wg == nil {
		return fmt.Errorf("%T %w", wg, common.ErrNilPointer)
	}
	wg.Add(1)
	go func() {
		c.Run(ctx)
		wg.Done()
	}()
	return nil
}

// Run implements the CoinbaseAdvanced wrapper
func (c *CoinbaseAdvanced) Run(ctx context.Context) {
	if c.Verbose {
		log.Debugf(log.ExchangeSys,
			"%s Websocket: %s.",
			c.Name,
			common.IsEnabled(c.Websocket.IsEnabled()))
	}

	if !c.GetEnabledFeatures().AutoPairUpdates && !c.BypassConfigFormatUpgrades {
		return
	}

	err := c.UpdateTradablePairs(ctx, false)
	if err != nil {
		log.Errorf(log.ExchangeSys, "%s failed to update tradable pairs. Err: %s", c.Name, err)
	}
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (c *CoinbaseAdvanced) FetchTradablePairs(ctx context.Context, a asset.Item) (currency.Pairs, error) {
	if !c.SupportsAsset(a) {
		return nil, fmt.Errorf("%w %v", asset.ErrNotSupported, a)
	}

	products, err := c.GetProducts(ctx)
	if err != nil {
		return nil, err
	}

	pairs := make(currency.Pairs, 0, len(products))
	for _, product := range products {
		if product.Status != "online" {
			continue
		}

		pair, err := currency.NewPairFromString(product.ProductID)
		if err != nil {
			return nil, err
		}
		pairs = pairs.Add(pair)
	}

	return pairs, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores them in the exchanges config
func (c *CoinbaseAdvanced) UpdateTradablePairs(ctx context.Context, forceUpdate bool) error {
	pairs, err := c.FetchTradablePairs(ctx, asset.Spot)
	if err != nil {
		return err
	}
	return c.UpdatePairs(pairs, asset.Spot, false, forceUpdate)
}

// UpdateAccountInfo retrieves balances for all enabled currencies
func (c *CoinbaseAdvanced) UpdateAccountInfo(ctx context.Context, assetType asset.Item) (account.Holdings, error) {
	var info account.Holdings
	var acc account.SubAccount
	info.Exchange = c.Name

	if !c.AllowAuthenticatedRequest() {
		return info, fmt.Errorf("%s %w", c.Name, exchange.ErrAuthenticatedRequestWithoutCredentialsSet)
	}

	accounts, err := c.GetAccounts(ctx)
	if err != nil {
		return info, err
	}

	currencyBalance := make(map[currency.Code]account.Balance)
	for _, acc := range accounts {
		currCode := currency.NewCode(acc.Currency)
		available, err := strconv.ParseFloat(acc.AvailableBalance.Value, 64)
		if err != nil {
			return info, err
		}
		hold, err := strconv.ParseFloat(acc.Hold.Value, 64)
		if err != nil {
			return info, err
		}

		currencyBalance[currCode] = account.Balance{
			Currency: currCode,
			Total:    available + hold,
			Hold:     hold,
			Free:     available,
		}
	}

	acc.Currencies = make([]account.Balance, 0, len(currencyBalance))
	for _, balance := range currencyBalance {
		acc.Currencies = append(acc.Currencies, balance)
	}

	acc.AssetType = assetType
	info.Accounts = append(info.Accounts, acc)
	return info, nil
}

// FetchAccountInfo retrieves balances for all enabled currencies
func (c *CoinbaseAdvanced) FetchAccountInfo(ctx context.Context, assetType asset.Item) (account.Holdings, error) {
	creds, err := c.GetCredentials(ctx)
	if err != nil {
		return account.Holdings{}, err
	}
	acc, err := c.UpdateAccountInfo(ctx, assetType)
	if err != nil {
		return acc, err
	}
	acc.Exchange = c.Name
	if creds.ClientID != "" {
		acc.ClientID = creds.ClientID
	}
	return acc, nil
}

// GetAccountFundingHistory returns funding history, deposits and withdrawals
func (c *CoinbaseAdvanced) GetAccountFundingHistory(ctx context.Context) ([]exchange.FundingHistory, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetWithdrawalsHistory retrieves the account's withdrawal history
func (c *CoinbaseAdvanced) GetWithdrawalsHistory(ctx context.Context, code currency.Code, assetType asset.Item) ([]exchange.WithdrawalHistory, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetRecentTrades returns the most recent trades for a currency and asset
func (c *CoinbaseAdvanced) GetRecentTrades(ctx context.Context, p currency.Pair, assetType asset.Item) ([]trade.Data, error) {
	return c.GetHistoricTrades(ctx, p, assetType, time.Now().Add(-time.Minute*15), time.Now())
}

// GetHistoricTrades retrieves historic trade data within the timeframe provided
func (c *CoinbaseAdvanced) GetHistoricTrades(ctx context.Context, p currency.Pair, assetType asset.Item, timestampStart, timestampEnd time.Time) ([]trade.Data, error) {
	if !c.SupportsAsset(assetType) {
		return nil, fmt.Errorf("%w %v", asset.ErrNotSupported, assetType)
	}

	err := common.StartEndTimeCheck(timestampStart, timestampEnd)
	if err != nil {
		return nil, fmt.Errorf("invalid time range supplied. Start: %v End %v %w", timestampStart, timestampEnd, err)
	}
	var resp []trade.Data
	ts := timestampStart
	limit := 1000
allTrades:
	for {
		var trades []TradeData
		trades, err = c.GetMarketTrades(ctx,
			p.String(),
			limit,
			ts,
			timestampEnd)
		if err != nil {
			return nil, err
		}
		for i := range trades {
			var side order.Side
			side, err = order.StringToOrderSide(trades[i].Side)
			if err != nil {
				log.Errorf(log.ExchangeSys, "%s %v", c.Name, err)
			}
			var price, amount float64
			price, err = strconv.ParseFloat(trades[i].Price, 64)
			if err != nil {
				return nil, err
			}
			amount, err = strconv.ParseFloat(trades[i].Size, 64)
			if err != nil {
				return nil, err
			}
			resp = append(resp, trade.Data{
				Exchange:     c.Name,
				TID:          trades[i].TradeID,
				CurrencyPair: p,
				AssetType:    assetType,
				Side:         side,
				Amount:       amount,
				Price:        price,
				Timestamp:    trades[i].Time,
			})
			if trades[i].Time.Equal(timestampEnd) || trades[i].Time.After(timestampEnd) {
				break allTrades
			}
		}
		if len(trades) != limit {
			break allTrades
		}
		ts = resp[len(resp)-1].Timestamp
	}

	err = trade.AddTradesToBuffer(c.Name, resp...)
	if err != nil {
		return nil, err
	}

	sort.Sort(trade.ByDate(resp))
	return trade.FilterTradesByTime(resp, timestampStart, timestampEnd), nil
}

// SubmitOrder submits a new order
func (c *CoinbaseAdvanced) SubmitOrder(ctx context.Context, s *order.Submit) (*order.SubmitResponse, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}

	if !c.SupportsAsset(s.AssetType) {
		return nil, fmt.Errorf("%w %v", asset.ErrNotSupported, s.AssetType)
	}

	fpair, err := c.FormatExchangeCurrency(s.Pair, s.AssetType)
	if err != nil {
		return nil, err
	}

	var orderResponse *OrderData
	switch s.Type {
	case order.Market:
		if s.Side == order.Buy {
			orderResponse, err = c.PlaceMarketOrder(ctx,
				fpair.String(),
				s.Side.String(),
				"",
				s.QuoteAmount,
			)
		} else {
			orderResponse, err = c.PlaceMarketOrder(ctx,
				fpair.String(),
				s.Side.String(),
				s.Amount,
				0,
			)
		}
	case order.Limit:
		orderResponse, err = c.PlaceLimitOrder(ctx,
			fpair.String(),
			s.Side.String(),
			s.Amount,
			s.Price,
			"GTC",
		)
	default:
		return nil, fmt.Errorf("%w %v", order.ErrTypeIsInvalid, s.Type)
	}

	if err != nil {
		return nil, err
	}

	return &order.SubmitResponse{
		OrderID: orderResponse.OrderID,
	}, nil
}

// ModifyOrder will allow of changing orderbook placement and limit to market conversion
func (c *CoinbaseAdvanced) ModifyOrder(ctx context.Context, action *order.Modify) (*order.ModifyResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// CancelOrder cancels an order by its corresponding ID number
func (c *CoinbaseAdvanced) CancelOrder(ctx context.Context, ord *order.Cancel) error {
	if err := ord.Validate(ord.StandardCancel()); err != nil {
		return err
	}

	_, err := c.CancelOrderByID(ctx, ord.OrderID)
	return err
}

// CancelBatchOrders cancels orders by their corresponding ID numbers
func (c *CoinbaseAdvanced) CancelBatchOrders(ctx context.Context, o []order.Cancel) (*order.CancelBatchResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// CancelAllOrders cancels all orders associated with a currency pair
func (c *CoinbaseAdvanced) CancelAllOrders(ctx context.Context, orderCancellation *order.Cancel) (order.CancelAllResponse, error) {
	if err := orderCancellation.Validate(); err != nil {
		return order.CancelAllResponse{}, err
	}

	var resp order.CancelAllResponse
	orders, err := c.CancelAllOrdersByProductID(ctx, orderCancellation.Pair.String())
	if err != nil {
		return resp, err
	}

	resp.Status = make(map[string]string)
	for _, ord := range orders {
		resp.Status[ord.OrderID] = "cancelled"
	}

	return resp, nil
}

// GetOrderInfo returns information on a current open order. Order must be executed by Coinbase Advanced.
func (c *CoinbaseAdvanced) GetOrderInfo(ctx context.Context, orderID string, pair currency.Pair, assetType asset.Item) (*order.Detail, error) {
	if pair.IsEmpty() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	if !c.SupportsAsset(assetType) {
		return nil, fmt.Errorf("%w %v", asset.ErrNotSupported, assetType)
	}

	resp, err := c.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return c.parseOrderData(resp, assetType)
}

// GetDepositAddress returns a deposit address for a specified currency
func (c *CoinbaseAdvanced) GetDepositAddress(ctx context.Context, cryptocurrency currency.Code, _, chain string) (*deposit.Address, error) {
	return nil, common.ErrFunctionNotSupported
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is submitted
func (c *CoinbaseAdvanced) WithdrawCryptocurrencyFunds(ctx context.Context, withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// WithdrawFiatFunds returns a withdrawal ID when a withdrawal is submitted
func (c *CoinbaseAdvanced) WithdrawFiatFunds(ctx context.Context, withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a withdrawal is submitted
func (c *CoinbaseAdvanced) WithdrawFiatFundsToInternationalBank(ctx context.Context, withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetFeeByType returns an estimate of fee based on type of transaction
func (c *CoinbaseAdvanced) GetFeeByType(ctx context.Context, feeBuilder *exchange.FeeBuilder) (float64, error) {
	if feeBuilder == nil {
		return 0, fmt.Errorf("%T %w", feeBuilder, common.ErrNilPointer)
	}
	if !c.AreCredentialsValid(ctx) && // Todo check connection status
		feeBuilder.FeeType == exchange.CryptocurrencyTradeFee {
		feeBuilder.FeeType = exchange.OfflineTradeFee
	}
	return c.GetFee(feeBuilder)
}

// GetActiveOrders retrieves any orders that are active/open
func (c *CoinbaseAdvanced) GetActiveOrders(ctx context.Context, req *order.MultiOrderRequest) (order.FilteredOrders, error) {
	err := req.Validate()
	if err != nil {
		return nil, err
	}

	var resp order.FilteredOrders
	for _, a := range req.AssetType {
		if !c.SupportsAsset(a) {
			return nil, fmt.Errorf("%w %v", asset.ErrNotSupported, a)
		}

		for _, p := range req.Pairs {
			fpair, err := c.FormatExchangeCurrency(p, a)
			if err != nil {
				return nil, err
			}

			orders, err := c.ListOrders(ctx, fpair.String(), "OPEN", "", "", 0)
			if err != nil {
				return nil, err
			}

			for _, orderData := range orders {
				orderDetail, err := c.parseOrderData(&orderData, a)
				if err != nil {
					return nil, err
				}
				resp = append(resp, *orderDetail)
			}
		}
	}

	return resp, nil
}

// GetOrderHistory retrieves account order information. Can Limit the number of entries returned.
func (c *CoinbaseAdvanced) GetOrderHistory(ctx context.Context, req *order.MultiOrderRequest) (order.FilteredOrders, error) {
	err := req.Validate()
	if err != nil {
		return nil, err
	}

	var resp order.FilteredOrders
	for _, a := range req.AssetType {
		if !c.SupportsAsset(a) {
			return nil, fmt.Errorf("%w %v", asset.ErrNotSupported, a)
		}

		for _, p := range req.Pairs {
			fpair, err := c.FormatExchangeCurrency(p, a)
			if err != nil {
				return nil, err
			}

			orders, err := c.ListOrders(ctx, fpair.String(), "FILLED", "", "", 0)
			if err != nil {
				return nil, err
			}

			for _, orderData := range orders {
				orderDetail, err := c.parseOrderData(&orderData, a)
				if err != nil {
					return nil, err
				}
				resp = append(resp, *orderDetail)
			}
		}
	}

	return resp, nil
}

// GetFeeByType returns an estimate of fee based on the type of transaction
func (c *CoinbaseAdvanced) ValidateAPICredentials(ctx context.Context, assetType asset.Item) error {
	_, err := c.UpdateAccountInfo(ctx, assetType)
	return c.CheckTransientError(err)
}

// GetHistoricCandles returns candles between a time period for a set time interval
func (c *CoinbaseAdvanced) GetHistoricCandles(ctx context.Context, pair currency.Pair, a asset.Item, interval kline.Interval, start, end time.Time) (*kline.Item, error) {
	req, err := c.GetKlineRequest(pair, a, interval, start, end, false)
	if err != nil {
		return nil, err
	}

	candles, err := c.GetCandles(ctx, req.RequestFormatted.String(), start, end, interval)
	if err != nil {
		return nil, err
	}

	ret := kline.Item{
		Exchange: c.Name,
		Pair:     pair,
		Asset:    a,
		Interval: interval,
	}

	for i := range candles {
		ret.Candles = append(ret.Candles, kline.Candle{
			Time:   candles[i].Start,
			Open:   candles[i].Low,
			High:   candles[i].High,
			Low:    candles[i].Low,
			Close:  candles[i].Close,
			Volume: candles[i].Volume,
		})
	}

	ret.SortCandlesByTimestamp(false)
	return req.ProcessResponse(&ret)
}

// GetHistoricCandlesExtended returns candles between a time period for a set time interval
func (c *CoinbaseAdvanced) GetHistoricCandlesExtended(ctx context.Context, pair currency.Pair, a asset.Item, interval kline.Interval, start, end time.Time) (*kline.Item, error) {
	return c.GetHistoricCandles(ctx, pair, a, interval, start, end)
}

// UpdateTicker updates and returns the ticker for a currency pair
func (c *CoinbaseAdvanced) UpdateTicker(ctx context.Context, p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	if err := c.CurrencyPairs.IsAssetEnabled(assetType); err != nil {
		return nil, err
	}

	if !c.SupportsAsset(assetType) {
		return nil, fmt.Errorf("%w %v", asset.ErrNotSupported, assetType)
	}

	fPair, err := c.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return nil, err
	}

	tick, err := c.GetTicker(ctx, fPair.String())
	if err != nil {
		return nil, err
	}

	last, _ := strconv.ParseFloat(tick.Price, 64)
	bid, _ := strconv.ParseFloat(tick.Bid, 64)
	ask, _ := strconv.ParseFloat(tick.Ask, 64)
	volume, _ := strconv.ParseFloat(tick.Volume, 64)
	err = ticker.ProcessTicker(&ticker.Price{
		Last:         last,
		Bid:          bid,
		Ask:          ask,
		Volume:       volume,
		Pair:         p,
		ExchangeName: c.Name,
		AssetType:    assetType,
	})
	if err != nil {
		return nil, err
	}

	return ticker.GetTicker(c.Name, p, assetType)
}

// UpdateTickers updates all currency pair tickers and returns the data
func (c *CoinbaseAdvanced) UpdateTickers(ctx context.Context, assetType asset.Item) error {
	pairs, err := c.GetEnabledPairs(assetType)
	if err != nil {
		return err
	}

	for i := range pairs {
		_, err := c.UpdateTicker(ctx, pairs[i], assetType)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetOrderbook retrieves the orderbook for a currency pair
func (c *CoinbaseAdvanced) UpdateOrderbook(ctx context.Context, pair currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	if err := c.CurrencyPairs.IsAssetEnabled(assetType); err != nil {
		return nil, err
	}

	if !c.SupportsAsset(assetType) {
		return nil, fmt.Errorf("%w %v", asset.ErrNotSupported, assetType)
	}

	fPair, err := c.FormatExchangeCurrency(pair, assetType)
	if err != nil {
		return nil, err
	}

	orderbookNew, err := c.GetOrderbook(ctx, fPair.String(), 2)
	if err != nil {
		return nil, err
	}

	book := &orderbook.Book{
		Exchange:          c.Name,
		Pair:              pair,
		Asset:             assetType,
		LastUpdated:       orderbookNew.Time,
		ValidateOrderbook: c.ValidateOrderbook,
	}

	book.Bids = make(orderbook.Levels, len(orderbookNew.Bids))
	for i := range orderbookNew.Bids {
		amount, _ := strconv.ParseFloat(orderbookNew.Bids[i][1], 64)
		price, _ := strconv.ParseFloat(orderbookNew.Bids[i][0], 64)
		book.Bids[i] = orderbook.Level{
			Amount: amount,
			Price:  price,
		}
	}

	book.Asks = make(orderbook.Levels, len(orderbookNew.Asks))
	for i := range orderbookNew.Asks {
		amount, _ := strconv.ParseFloat(orderbookNew.Asks[i][1], 64)
		price, _ := strconv.ParseFloat(orderbookNew.Asks[i][0], 64)
		book.Asks[i] = orderbook.Level{
			Amount: amount,
			Price:  price,
		}
	}

	err = book.Process()
	if err != nil {
		return book, err
	}

	return orderbook.Get(c.Name, pair, assetType)
}

// GetFee returns an estimate of fee based on type of transaction
func (c *CoinbaseAdvanced) GetFee(feeBuilder *exchange.FeeBuilder) (float64, error) {
	var fee float64
	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyTradeFee:
		fee = c.calculateTradingFee(feeBuilder.PurchasePrice, feeBuilder.Amount)
	case exchange.OfflineTradeFee:
		fee = c.getOfflineTradeFee(feeBuilder.PurchasePrice, feeBuilder.Amount)
	}

	if fee < 0 {
		fee = 0
	}

	return fee, nil
}

// calculateTradingFee calculates trading fee
func (c *CoinbaseAdvanced) calculateTradingFee(price, amount float64) float64 {
	return 0.006 * price * amount // 0.6% default fee
}

// getOfflineTradeFee calculates offline trading fee
func (c *CoinbaseAdvanced) getOfflineTradeFee(price, amount float64) float64 {
	return 0.006 * price * amount // 0.6% default fee
}
