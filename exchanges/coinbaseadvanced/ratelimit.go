package coinbaseadvanced

import (
	"time"

	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
)

const (
	coinbaseAdvancedRateInterval = time.Second
	coinbaseAdvancedRequestRate  = 10
)

// SetRateLimit returns the rate limit for the exchange
func SetRateLimit() request.RateLimitDefinitions {
	return request.RateLimitDefinitions{
		request.Auth:   request.NewRateLimitWithWeight(coinbaseAdvancedRateInterval, coinbaseAdvancedRequestRate, 1),
		request.UnAuth: request.NewRateLimitWithWeight(coinbaseAdvancedRateInterval, coinbaseAdvancedRequestRate, 1),
	}
}
