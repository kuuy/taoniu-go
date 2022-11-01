package indicators

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	models "taoniu.local/cryptos/models/binance/spot"
	repositories "taoniu.local/cryptos/repositories/binance/spot/indicators"
)

type DailyTask struct {
	Db         *gorm.DB
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.DailyRepository
}

func (t *DailyTask) Pivot() error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Repository.Pivot(symbol)
	}
	return nil
}

func (t *DailyTask) Atr(period int, limit int) error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Repository.Atr(symbol, period, limit)
	}
	return nil
}

func (t *DailyTask) Zlema(period int, limit int) error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Repository.Zlema(symbol, period, limit)
	}
	return nil
}

func (t *DailyTask) HaZlema(period int, limit int) error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Repository.HaZlema(symbol, period, limit)
	}
	return nil
}

func (t *DailyTask) Kdj(longPeriod int, shortPeriod int, limit int) error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Repository.Kdj(symbol, longPeriod, shortPeriod, limit)
	}
	return nil
}

func (t *DailyTask) BBands(period int, limit int) error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Repository.BBands(symbol, period, limit)
	}
	return nil
}

func (t *DailyTask) Flush() error {
	t.Pivot()
	t.Atr(14, 100)
	t.Zlema(14, 100)
	t.HaZlema(14, 100)
	t.Kdj(9, 3, 100)
	t.BBands(14, 100)
	return nil
}
