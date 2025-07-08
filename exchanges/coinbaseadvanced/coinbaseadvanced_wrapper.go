package coinbaseadvanced

import (
	"context"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/deposit"
	"github.com/thrasher-corp/gocryptotrader/exchanges/fundingrate"
	"github.com/thrasher-corp/gocryptotrader/exchanges/futures"
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
				OrderbookFetching: true,
			},
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
		},
	}

	c.Requester, err = request.New(c.Name, common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}
	c.API.Endpoints = c.NewEndpoints()
	err = c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: coinbaseAdvancedAPIURL,
	})
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}
}

// Setup initializes the exchange parameters with default settings
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

// GetAccountInfo returns account information
func (c *CoinbaseAdvanced) GetAccountInfo(ctx context.Context, assetType asset.Item) (account.Holdings, error) {
	var info account.Holdings
	info.Exchange = c.Name
	return info, common.ErrFunctionNotSupported
}

// UpdateAccountInfo updates account info
func (c *CoinbaseAdvanced) UpdateAccountInfo(ctx context.Context, assetType asset.Item) (account.Holdings, error) {
	return c.GetAccountInfo(ctx, assetType)
}

// GetAccountFundingHistory returns funding history
func (c *CoinbaseAdvanced) GetAccountFundingHistory(ctx context.Context) ([]exchange.FundingHistory, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetWithdrawalsHistory returns withdrawal history
func (c *CoinbaseAdvanced) GetWithdrawalsHistory(ctx context.Context, code currency.Code, assetType asset.Item) ([]exchange.WithdrawalHistory, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetRecentTrades returns recent trades
func (c *CoinbaseAdvanced) GetRecentTrades(ctx context.Context, p currency.Pair, assetType asset.Item) ([]trade.Data, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetHistoricTrades returns historic trades
func (c *CoinbaseAdvanced) GetHistoricTrades(ctx context.Context, p currency.Pair, assetType asset.Item, timestampStart, timestampEnd time.Time) ([]trade.Data, error) {
	return nil, common.ErrFunctionNotSupported
}

// SubmitOrder submits an order
func (c *CoinbaseAdvanced) SubmitOrder(ctx context.Context, s *order.Submit) (*order.SubmitResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// ModifyOrder modifies an order
func (c *CoinbaseAdvanced) ModifyOrder(ctx context.Context, action *order.Modify) (*order.ModifyResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// CancelOrder cancels an order
func (c *CoinbaseAdvanced) CancelOrder(ctx context.Context, ord *order.Cancel) error {
	return common.ErrFunctionNotSupported
}

// CancelBatchOrders cancels multiple orders
func (c *CoinbaseAdvanced) CancelBatchOrders(ctx context.Context, o []order.Cancel) (*order.CancelBatchResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// CancelAllOrders cancels all orders
func (c *CoinbaseAdvanced) CancelAllOrders(ctx context.Context, req *order.Cancel) (order.CancelAllResponse, error) {
	return order.CancelAllResponse{}, common.ErrFunctionNotSupported
}

// GetOrderInfo gets order information
func (c *CoinbaseAdvanced) GetOrderInfo(ctx context.Context, orderID string, pair currency.Pair, assetType asset.Item) (*order.Detail, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetDepositAddress gets deposit address
func (c *CoinbaseAdvanced) GetDepositAddress(ctx context.Context, code currency.Code, _, chain string) (*deposit.Address, error) {
	return nil, common.ErrFunctionNotSupported
}

// WithdrawCryptocurrencyFunds withdraws cryptocurrency
func (c *CoinbaseAdvanced) WithdrawCryptocurrencyFunds(ctx context.Context, withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// WithdrawFiatFunds withdraws fiat
func (c *CoinbaseAdvanced) WithdrawFiatFunds(ctx context.Context, withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// WithdrawFiatFundsToInternationalBank withdraws fiat to international bank
func (c *CoinbaseAdvanced) WithdrawFiatFundsToInternationalBank(ctx context.Context, withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetActiveOrders gets active orders
func (c *CoinbaseAdvanced) GetActiveOrders(ctx context.Context, req *order.MultiOrderRequest) (order.FilteredOrders, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetOrderHistory gets order history
func (c *CoinbaseAdvanced) GetOrderHistory(ctx context.Context, req *order.MultiOrderRequest) (order.FilteredOrders, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetHistoricCandles gets historic candles
func (c *CoinbaseAdvanced) GetHistoricCandles(ctx context.Context, p currency.Pair, a asset.Item, interval kline.Interval, start, end time.Time) (*kline.Item, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetHistoricCandlesExtended gets extended historic candles
func (c *CoinbaseAdvanced) GetHistoricCandlesExtended(ctx context.Context, p currency.Pair, a asset.Item, interval kline.Interval, start, end time.Time) (*kline.Item, error) {
	return nil, common.ErrFunctionNotSupported
}

// UpdateTicker updates ticker
func (c *CoinbaseAdvanced) UpdateTicker(ctx context.Context, p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	return nil, common.ErrFunctionNotSupported
}

// UpdateTickers updates all tickers
func (c *CoinbaseAdvanced) UpdateTickers(ctx context.Context, assetType asset.Item) error {
	return common.ErrFunctionNotSupported
}

// UpdateOrderbook updates orderbook
func (c *CoinbaseAdvanced) UpdateOrderbook(ctx context.Context, pair currency.Pair, assetType asset.Item) (*orderbook.Book, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetFeeByType gets fee by type
func (c *CoinbaseAdvanced) GetFeeByType(ctx context.Context, feeBuilder *exchange.FeeBuilder) (float64, error) {
	return 0, common.ErrFunctionNotSupported
}

// GetFuturesContractDetails returns all contracts from the exchange by asset type
func (c *CoinbaseAdvanced) GetFuturesContractDetails(ctx context.Context, item asset.Item) ([]futures.Contract, error) {
	if !item.IsFutures() {
		return nil, futures.ErrNotFuturesAsset
	}
	return nil, common.ErrFunctionNotSupported
}

// FetchTradablePairs returns tradable pairs
func (c *CoinbaseAdvanced) FetchTradablePairs(ctx context.Context, assetType asset.Item) (currency.Pairs, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetLatestFundingRates gets latest funding rates
func (c *CoinbaseAdvanced) GetLatestFundingRates(ctx context.Context, req *fundingrate.LatestRateRequest) ([]fundingrate.LatestRateResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetServerTime gets server time
func (c *CoinbaseAdvanced) GetServerTime(ctx context.Context, assetType asset.Item) (time.Time, error) {
	return time.Time{}, common.ErrFunctionNotSupported
}

// UpdateOrderExecutionLimits updates order execution limits
func (c *CoinbaseAdvanced) UpdateOrderExecutionLimits(ctx context.Context, assetType asset.Item) error {
	return common.ErrFunctionNotSupported
}

// UpdateTradablePairs updates tradable pairs
func (c *CoinbaseAdvanced) UpdateTradablePairs(ctx context.Context, forceUpdate bool) error {
	return common.ErrFunctionNotSupported
}

// ValidateAPICredentials validates API credentials
func (c *CoinbaseAdvanced) ValidateAPICredentials(ctx context.Context, assetType asset.Item) error {
	return common.ErrFunctionNotSupported
}
