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
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	ScannerRepository *ScannerRepository
}

func (r *AnalysisRepository) Scanner() *ScannerRepository {
	if r.ScannerRepository == nil {
		r.ScannerRepository = &ScannerRepository{}
	}
	return r.ScannerRepository
}

func (r *AnalysisRepository) Flush(exchange string, symbol string, interval string) error {
	analysis, err := r.Scanner().Scan(exchange, symbol, interval)
	if err != nil {
		return err
	}

	var entity models.Analysis
	result := r.Db.Where(
		"exchange=? AND symbol=? AND interval=?",
		exchange,
		symbol,
		interval,
	).Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	entity.Summary["BUY"] = analysis.BuyCount
	entity.Summary["SELL"] = analysis.SellCount
	entity.Summary["NEUTRAL"] = analysis.NeutralCount
	entity.Summary["RECOMMENDATION"] = analysis.Recommend.Summary
	r.Db.Model(&models.Analysis{ID: entity.ID}).Updates(entity)

	return nil
}

func (r *AnalysisRepository) Count(conditions map[string]interface{}) int64 {
	var total int64
	query := r.Db.Model(&models.Analysis{})
	if _, ok := conditions["exchange"]; ok {
		query.Where("exchange", conditions["exchange"])
	}
	if _, ok := conditions["interval"]; ok {
		query.Where("interval", conditions["interval"])
	}
	query.Count(&total)
	return total
}

func (r *AnalysisRepository) Listings(
	current int,
	pageSize int,
	conditions map[string]interface{},
) []*models.Analysis {
	offset := (current - 1) * pageSize

	var analysis []*models.Analysis
	query := r.Db.Select([]string{
		"id",
		"symbol",
		"summary",
		"updated_at",
	})
	if _, ok := conditions["exchange"]; ok {
		query.Where("exchange", conditions["exchange"])
	}
	if _, ok := conditions["interval"]; ok {
		query.Where("interval", conditions["interval"])
	}
	query.Order("created_at desc")
	query.Offset(offset).Limit(pageSize).Find(&analysis)
	return analysis
}

func (r *AnalysisRepository) Summary(
	exchange string,
	symbol string,
	interval string,
) (map[string]interface{}, error) {
	var entity models.Analysis
	result := r.Db.Model(&models.Analysis{}).Select("summary").Where(
		"exchange=? AND symbol=? AND interval=?",
		exchange,
		symbol,
		interval,
	).Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}
	return entity.Summary, nil
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
