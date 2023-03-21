package spot

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	config "taoniu.local/cryptos/config/binance/spot"
	binanceModels "taoniu.local/cryptos/models/binance/spot"
)

type AccountRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *AccountRepository) Flush() error {
	client := binance.NewClient(config.ACCOUNT_API_KEY, config.ACCOUNT_SECRET_KEY)
	account, err := client.NewGetAccountService().Do(r.Ctx)
	if err != nil {
		return err
	}
	oldCurrencies, _ := r.Rdb.SMembers(r.Ctx, "binance:spot:currencies").Result()
	var currencies []string
	for _, coin := range account.Balances {
		free, _ := strconv.ParseFloat(coin.Free, 64)
		if free <= 0.0 {
			r.Rdb.SRem(r.Ctx, "binance:spot:currencies", coin.Asset)
			r.Rdb.Del(r.Ctx, fmt.Sprintf("binance:spot:balances:%s", coin.Asset))
			continue
		}
		r.Rdb.SAdd(r.Ctx, "binance:spot:currencies", coin.Asset)
		r.Rdb.HMSet(
			r.Ctx,
			fmt.Sprintf("binance:spot:balances:%s", coin.Asset),
			map[string]interface{}{
				"free":   coin.Free,
				"locked": coin.Locked,
			},
		)
		currencies = append(currencies, coin.Asset)
	}
	for _, currency := range oldCurrencies {
		if !r.contains(currencies, currency) {
			r.Rdb.SRem(r.Ctx, "binance:spot:currencies", currency)
			r.Rdb.Del(r.Ctx, fmt.Sprintf("binance:spot:balances:%s", currency))
		}
	}

	return nil
}

func (r *AccountRepository) Balance(symbol string) (float64, float64, error) {
	var entity *binanceModels.Symbol
	result := r.Db.Select([]string{"base_asset", "quote_asset"}).Where("symbol", symbol).Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, 0, errors.New("no symbol data")
	}

	var balance float64 = 0
	var quantity float64 = 0

	var val string
	var err error
	val, err = r.Rdb.HGet(
		r.Ctx,
		fmt.Sprintf("binance:spot:balances:%s", entity.QuoteAsset),
		"free",
	).Result()
	if err == nil {
		balance, _ = strconv.ParseFloat(val, 64)
	}

	val, err = r.Rdb.HGet(
		r.Ctx,
		fmt.Sprintf("binance:spot:balances:%s", entity.BaseAsset),
		"free",
	).Result()
	if err == nil {
		quantity, _ = strconv.ParseFloat(val, 64)
	}

	return balance, quantity, nil
}

func (r *AccountRepository) contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
