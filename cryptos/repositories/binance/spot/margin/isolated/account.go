package isolated

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"
	"strconv"
	config "taoniu.local/cryptos/config/binance"
)

type AccountRepository struct {
	Rdb *redis.Client
	Ctx context.Context
}

func (r *AccountRepository) Flush() error {
	client := binance.NewClient(config.ACCOUNT_API_KEY, config.ACCOUNT_SECRET_KEY)
	account, err := client.NewGetIsolatedMarginAccountService().Do(r.Ctx)
	if err != nil {
		return err
	}
	for _, coin := range account.Assets {
		baseBorrowed, _ := strconv.ParseFloat(coin.BaseAsset.Borrowed, 64)
		quoteBorrowed, _ := strconv.ParseFloat(coin.QuoteAsset.Borrowed, 64)
		if baseBorrowed <= 0.0 && quoteBorrowed <= 0.0 {
			r.Rdb.Del(r.Ctx, fmt.Sprintf("binance:spot:margin:isolated:balances:%s", coin.Symbol))
			continue
		}
		r.Rdb.HMSet(
			r.Ctx,
			fmt.Sprintf("binance:spot:margin:isolated:balances:%s", coin.Symbol),
			map[string]interface{}{
				"margin_ratio":      coin.MarginRatio,
				"liquidate_price":   coin.LiquidatePrice,
				"base_free":         coin.BaseAsset.Free,
				"base_locked":       coin.BaseAsset.Locked,
				"base_borrowed":     coin.BaseAsset.Borrowed,
				"base_interest":     coin.BaseAsset.Interest,
				"base_net_asset":    coin.BaseAsset.NetAsset,
				"base_total_asset":  coin.BaseAsset.TotalAsset,
				"quote_free":        coin.QuoteAsset.Free,
				"quote_locked":      coin.QuoteAsset.Locked,
				"quote_borrowed":    coin.QuoteAsset.Borrowed,
				"quote_interest":    coin.QuoteAsset.Interest,
				"quote_net_asset":   coin.QuoteAsset.NetAsset,
				"quote_total_asset": coin.QuoteAsset.TotalAsset,
			},
		)
	}

	return nil
}
