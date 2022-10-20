package indicators

import (
	"context"
	"github.com/gammazero/workerpool"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	models "taoniu.local/cryptos/models/binance"
	repositories "taoniu.local/cryptos/repositories/binance/spot/indicators"
)

type DailyTask struct {
	Db         *gorm.DB
	Rdb        *redis.Client
	Ctx        context.Context
	Wp         *workerpool.WorkerPool
	Repository *repositories.DailyRepository
}

func (t *DailyTask) Pivot() error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Wp.Submit(func() {
			t.Repository.Pivot(symbol)
		})
	}
	return nil
}

func (t *DailyTask) Atr(period int, limit int) error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Wp.Submit(func() {
			t.Repository.Atr(symbol, period, limit)
		})
	}
	return nil
}

func (t *DailyTask) Zlema(period int, limit int) error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Wp.Submit(func() {
			t.Repository.Zlema(symbol, period, limit)
		})
	}
	return nil
}

func (t *DailyTask) HaZlema(period int, limit int) error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Wp.Submit(func() {
			t.Repository.HaZlema(symbol, period, limit)
		})
	}
	return nil
}

func (t *DailyTask) Kdj(longPeriod int, shortPeriod int, limit int) error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Wp.Submit(func() {
			t.Repository.Kdj(symbol, longPeriod, shortPeriod, limit)
		})
	}
	return nil
}

func (t *DailyTask) BBands(period int, limit int) error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Wp.Submit(func() {
			t.Repository.BBands(symbol, period, limit)
		})
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
