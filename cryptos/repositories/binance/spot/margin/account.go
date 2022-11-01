package margin

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"
	"log"
	"strconv"
	config "taoniu.local/cryptos/config/binance/spot"
)

type AccountRepository struct {
	Rdb *redis.Client
	Ctx context.Context
}

func (r *AccountRepository) Flush() error {
	client := binance.NewClient(config.ACCOUNT_API_KEY, config.ACCOUNT_SECRET_KEY)
	account, err := client.NewGetMarginAccountService().Do(r.Ctx)
	if err != nil {
		return err
	}
	for _, coin := range account.UserAssets {
		log.Println("coin:", coin)
		free, _ := strconv.ParseFloat(coin.Free, 64)
		if free <= 0.0 {
			r.Rdb.Del(r.Ctx, fmt.Sprintf("binance:spot:margin:balances:%s", coin.Asset))
			continue
		}
		r.Rdb.HMSet(
			r.Ctx,
			fmt.Sprintf("binance:spot:margin:balances:%s", coin.Asset),
			map[string]interface{}{
				"free":      coin.Free,
				"locked":    coin.Locked,
				"borrowed":  coin.Borrowed,
				"interrest": coin.Interest,
				"net_asset": coin.NetAsset,
			},
		)
	}

	return nil
}
