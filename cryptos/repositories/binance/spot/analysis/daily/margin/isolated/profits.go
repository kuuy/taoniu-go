package isolated

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/xid"
	"gorm.io/datatypes"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/go-redis/redis/v8"

	models "taoniu.local/cryptos/models/binance/spot/analysis/margin/isolated"
)

type ProfitsRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *ProfitsRepository) Flush(symbol string) error {
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

	now := time.Now()
	today := datatypes.Date(now)

	var totalProfit float64
	baseProfit := (baseTotal - baseBorrowed - baseInterest - baseBorrowed/marginRatio) * price
	quoteProfit := quoteTotal - quoteBorrowed - quoteBorrowed/marginRatio - quoteInterest
	totalProfit = baseProfit + quoteProfit

	var entity models.Daily
	var tx *gorm.DB
	tx = r.Db.Where(
		"symbol=? AND day=?",
		symbol,
		today,
	).First(&entity)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		entity = models.Daily{
			ID:          xid.New().String(),
			Symbol:      symbol,
			Day:         today,
			TotalProfit: totalProfit,
		}
		r.Db.Create(&entity)
	} else {
		entity.TotalProfit = totalProfit
		r.Db.Model(&models.Daily{ID: entity.ID}).Updates(entity)
	}

	return nil
}
