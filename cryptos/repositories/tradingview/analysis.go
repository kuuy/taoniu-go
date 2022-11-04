package tradingview

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"strings"
	"time"

	models "taoniu.local/cryptos/models/tradingview"
)

type AnalysisRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *AnalysisRepository) Signal(symbol string) (int64, bool, error) {
	var entity models.Analysis
	result := r.Db.Where(
		"exchange=? AND symbol=? AND interval=?",
		"BINANCE",
		symbol,
		"1m",
	).Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, false, result.Error
	}

	var signal int64 = 0
	var isStrong bool = false
	if strings.Contains(entity.Summary["RECOMMENDATION"].(string), "BUY") {
		signal = 1
	}
	if strings.Contains(entity.Summary["RECOMMENDATION"].(string), "SELL") {
		signal = 2
	}
	if strings.Contains(entity.Summary["RECOMMENDATION"].(string), "STRONG") {
		isStrong = true
	}

	timestamp := time.Now().Unix()
	if entity.UpdatedAt.Unix() < timestamp-60 {
		r.Rdb.ZAdd(r.Ctx, "tradingview:analysis:flush:1m", &redis.Z{
			float64(timestamp),
			symbol,
		})
		return 0, false, errors.New("tradingview analysis long time not freshed")
	}

	return signal, isStrong, nil
}
