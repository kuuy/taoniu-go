package spot

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
	account, err := client.NewGetAccountService().Do(r.Ctx)
	if err != nil {
		return err
	}
	for _, coin := range account.Balances {
		free, _ := strconv.ParseFloat(coin.Free, 64)
		if free <= 0.0 {
			r.Rdb.Del(r.Ctx, fmt.Sprintf("binance:spot:balances:%s", coin.Asset))
			continue
		}
		r.Rdb.HMSet(
			r.Ctx,
			fmt.Sprintf("binance:spot:balances:%s", coin.Asset),
			map[string]interface{}{
				"free":   coin.Free,
				"locked": coin.Locked,
			},
		)
	}

	return nil
}
