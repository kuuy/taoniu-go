package klines

import (
	"context"
	"errors"
	"strconv"

	"gorm.io/gorm"

	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"

	config "taoniu.local/cryptos/config/binance"
	models "taoniu.local/cryptos/models/binance/spot"
)

type DailyRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *DailyRepository) Flush(symbol string, limit int) error {
	client := binance.NewClient(config.REST_API_KEY, config.REST_SECRET_KEY)
	klines, err := client.NewKlinesService().Symbol(
		symbol,
	).Interval(
		"1d",
	).Limit(
		limit,
	).Do(r.Ctx)
	if err != nil {
		return err
	}
	for _, kline := range klines {
		open, _ := strconv.ParseFloat(kline.Open, 64)
		close, _ := strconv.ParseFloat(kline.Close, 64)
		high, _ := strconv.ParseFloat(kline.High, 64)
		low, _ := strconv.ParseFloat(kline.Low, 64)
		volume, _ := strconv.ParseFloat(kline.Volume, 64)
		quota, _ := strconv.ParseFloat(kline.QuoteAssetVolume, 64)
		timestamp := kline.OpenTime
		var entity models.Kline1d
		result := r.Db.Where(
			"symbol=? AND timestamp=?",
			symbol,
			timestamp,
		).First(&entity)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			entity = models.Kline1d{
				ID:        xid.New().String(),
				Symbol:    symbol,
				Open:      open,
				Close:     close,
				High:      high,
				Low:       low,
				Volume:    volume,
				Quota:     quota,
				Timestamp: timestamp,
			}
			r.Db.Create(&entity)
		} else {
			entity.Open = open
			entity.Close = close
			entity.High = high
			entity.Low = low
			entity.Volume = volume
			entity.Quota = quota
			entity.Timestamp = timestamp
			r.Db.Model(&models.Kline1d{ID: entity.ID}).Updates(entity)
		}
	}

	return nil
}
