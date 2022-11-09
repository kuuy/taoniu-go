package isolated

import (
	"context"
	"errors"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"
	"strconv"
	config "taoniu.local/cryptos/config/binance/spot"
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
		baseTotalAsset, _ := strconv.ParseFloat(coin.BaseAsset.TotalAsset, 64)
		quoteTotalAsset, _ := strconv.ParseFloat(coin.QuoteAsset.TotalAsset, 64)
		if baseTotalAsset <= 0.0 && quoteTotalAsset <= 0.0 {
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

func (r *AccountRepository) Balance(symbol string) (float64, float64, error) {
	fields := []string{
		"quote_free",
		"base_free",
	}
	data, _ := r.Rdb.HMGet(
		r.Ctx,
		fmt.Sprintf(
			"binance:spot:margin:isolated:balances:%s",
			symbol,
		),
		fields...,
	).Result()
	for i := 0; i < len(fields); i++ {
		if data[i] == nil {
			return 0, 0, errors.New("price not exists")
		}
	}
	balance, _ := strconv.ParseFloat(data[0].(string), 64)
	quantity, _ := strconv.ParseFloat(data[1].(string), 64)

	return balance, quantity, nil
}
