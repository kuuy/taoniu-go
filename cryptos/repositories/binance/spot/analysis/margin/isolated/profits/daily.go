package profits

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"log"
	"strconv"
	"time"
)

type DailyRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *DailyRepository) Flush(symbol string) error {
	log.Println("symbol:", symbol)
	balance, _ := r.Rdb.HGetAll(r.Ctx, fmt.Sprintf("binance:spot:margin:isolated:balances:%s", symbol)).Result()
	marginRatio, _ := strconv.ParseFloat(balance["margin_ratio"], 64)
	baseTotal, _ := strconv.ParseFloat(balance["base_total_asset"], 64)
	baseBorrowed, _ := strconv.ParseFloat(balance["base_borrowed"], 64)
	baseInterest, _ := strconv.ParseFloat(balance["base_interest"], 64)
	quoteTotal, _ := strconv.ParseFloat(balance["quote_total_asset"], 64)
	quoteBorrowed, _ := strconv.ParseFloat(balance["quote_borrowed"], 64)
	quoteInterest, _ := strconv.ParseFloat(balance["quote_interest"], 64)

	timestamp := time.Now().Unix()
	data, err := r.Rdb.HMGet(r.Ctx, fmt.Sprintf("binance:spot:realtime:%s", symbol), "price", "timestamp").Result()
	if err != nil {
		return err
	}
	price, _ := strconv.ParseFloat(fmt.Sprint(data[0]), 64)
	lasttime, _ := strconv.ParseInt(fmt.Sprint(data[1]), 10, 64)
	if timestamp-lasttime > 60 {
		return nil
	}

	var totalProfit float64
	baseProfit := (baseTotal - baseBorrowed - baseInterest - baseBorrowed/marginRatio) * price
	quoteProfit := quoteTotal - quoteBorrowed - quoteBorrowed/marginRatio - quoteInterest
	totalProfit = baseProfit + quoteProfit
	log.Println("balance", symbol, price, totalProfit)

	return nil
}

func (r *DailyRepository) Summary() error {
	return nil
}
