package futures

import (
	"context"
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"

	models "taoniu.local/cryptos/models/binance/futures"
)

type KlinesRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *KlinesRepository) Flush(symbol string, interval string, limit int) error {
	client := binance.NewFuturesClient("", "")
	klines, err := client.NewKlinesService().Symbol(
		symbol,
	).Interval(
		interval,
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
		var entity models.Kline
		result := r.Db.Where(
			"symbol=? AND interval=? AND timestamp=?",
			symbol,
			interval,
			timestamp,
		).Take(&entity)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			entity = models.Kline{
				ID:        xid.New().String(),
				Symbol:    symbol,
				Interval:  interval,
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
			r.Db.Model(&models.Kline{ID: entity.ID}).Updates(entity)
		}
	}

	return nil
}

func (r *KlinesRepository) Clean() error {
	timestamp := time.Now().AddDate(0, 0, -100).Unix()
	r.Db.Where("timestamp < ?", timestamp).Delete(&models.Kline{})

	return nil
}
