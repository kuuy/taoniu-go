package spot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"
)

type TickersRepository struct {
	Rdb *redis.Client
	Ctx context.Context
}

func (r *TickersRepository) Flush(symbols []string) error {
	client := binance.NewClient("", "")
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
				"timestamp": timestamp,
			},
		)

		r.Rdb.ZRem(r.Ctx, "binance:spot:tickers:flush", item.Symbol)
	}

	return nil
}

func (r *TickersRepository) Gets(symbols []string, fields []string) []string {
	var script = redis.NewScript(`
	local hmget = function (key)
		local hash = {}
		local data = redis.call('HMGET', key, unpack(ARGV))
		for i = 1, #ARGV do
			hash[i] = data[i]
		end
		return hash
	end
	local data = {}
	for i = 1, #KEYS do
		local key = 'binance:spot:realtime:' .. KEYS[i]
		if redis.call('EXISTS', key) == 0 then
			data[i] = false
		else
			data[i] = hmget(key)
		end
	end
	return data
  `)
	args := make([]interface{}, len(fields))
	for i := 0; i < len(fields); i++ {
		args[i] = fields[i]
	}
	result, _ := script.Run(r.Ctx, r.Rdb, symbols, args...).Result()

	tickers := make([]string, len(symbols))
	for i := 0; i < len(symbols); i++ {
		item := result.([]interface{})[i]
		if item == nil {
			continue
		}
		data := make([]string, len(fields))
		for j := 0; j < len(fields); j++ {
			if item.([]interface{})[j] == nil {
				continue
			}
			data[j] = fmt.Sprintf("%v", item.([]interface{})[j])
		}
		tickers[i] = strings.Join(data, ",")
	}

	return tickers
}
