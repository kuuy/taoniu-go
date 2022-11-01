package strategies

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	models "taoniu.local/cryptos/models/binance/futures"
	repositories "taoniu.local/cryptos/repositories/binance/futures/strategies"
)

type DailyTask struct {
	Db         *gorm.DB
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.DailyRepository
}

func (t *DailyTask) Atr() error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Repository.Atr(symbol)
	}
	return nil
}

func (t *DailyTask) Zlema() error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Repository.Zlema(symbol)
	}
	return nil
}

func (t *DailyTask) HaZlema() error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Repository.HaZlema(symbol)
	}
	return nil
}

func (t *DailyTask) Kdj() error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Repository.Kdj(symbol)
	}
	return nil
}

func (t *DailyTask) BBands() error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		t.Repository.BBands(symbol)
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
