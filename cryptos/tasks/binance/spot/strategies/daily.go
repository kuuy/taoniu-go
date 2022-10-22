package strategies

import (
	"context"
	"github.com/gammazero/workerpool"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	models "taoniu.local/cryptos/models/binance"
	repositories "taoniu.local/cryptos/repositories/binance/spot/strategies"
)

type DailyTask struct {
	Db         *gorm.DB
	Rdb        *redis.Client
	Ctx        context.Context
	Wp         *workerpool.WorkerPool
	Repository *repositories.DailyRepository
}

func (t *DailyTask) Atr() error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Wp.Submit(func() {
			t.Repository.Atr(symbol)
		})
	}
	return nil
}

func (t *DailyTask) Zlema() error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Wp.Submit(func() {
			t.Repository.Zlema(symbol)
		})
	}
	return nil
}

func (t *DailyTask) HaZlema() error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Wp.Submit(func() {
			t.Repository.HaZlema(symbol)
		})
	}
	return nil
}

func (t *DailyTask) Kdj() error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Wp.Submit(func() {
			t.Repository.Kdj(symbol)
		})
	}
	return nil
}

func (t *DailyTask) BBands() error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Wp.Submit(func() {
			t.Repository.BBands(symbol)
		})
	}
	return nil
}

func (t *DailyTask) Flush() error {
	t.Atr()
	t.Zlema()
	t.HaZlema()
	t.Kdj()
	t.BBands()
	return nil
}
