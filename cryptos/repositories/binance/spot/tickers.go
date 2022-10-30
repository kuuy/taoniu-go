package spot

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"

	config "taoniu.local/cryptos/config/binance"
)

type TickersRepository struct {
	Rdb *redis.Client
	Ctx context.Context
}

func (r *TickersRepository) Flush(symbols []string) error {
	client := binance.NewClient(config.REST_API_KEY, config.REST_SECRET_KEY)
	tickers, err := client.NewListSymbolTickerService().Symbols(symbols).Do(r.Ctx)
	if err != nil {
		return err
	}
	timestamp := time.Now().Unix()
	for _, item := range tickers {
		redisKey := fmt.Sprintf("binance:spot:realtime:%s", item.Symbol)
		value, err := r.Rdb.HGet(r.Ctx, redisKey, "price").Result()
		if err != nil {
			lasttime, _ := strconv.ParseInt(value, 10, 64)
			if lasttime > timestamp {
				continue
			}
		}
		price, _ := strconv.ParseFloat(item.LastPrice, 64)
		open, _ := strconv.ParseFloat(item.OpenPrice, 64)
		high, _ := strconv.ParseFloat(item.HighPrice, 64)
		low, _ := strconv.ParseFloat(item.LowPrice, 64)
		volume, _ := strconv.ParseFloat(item.Volume, 64)
		quota, _ := strconv.ParseFloat(item.QuoteVolume, 64)
		r.Rdb.HMSet(
			r.Ctx,
			redisKey,
			map[string]interface{}{
				"symbol":    item.Symbol,
				"price":     price,
				"open":      open,
				"high":      high,
				"low":       low,
				"volume":    volume,
				"quota":     quota,
				"timestamp": fmt.Sprint(timestamp),
			},
		)
	}

	return nil
}
