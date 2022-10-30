package spot

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/go-redis/redis/v8"

	models "taoniu.local/cryptos/models/binance"
	binanceRepositories "taoniu.local/cryptos/repositories/binance"
)

type SymbolsError struct {
	Message string
}

func (m *SymbolsError) Error() string {
	return m.Message
}

type SymbolsRepository struct {
	Db               *gorm.DB
	Rdb              *redis.Client
	Ctx              context.Context
	ParentRepository *binanceRepositories.SymbolsRepository
}

func (r *SymbolsRepository) Parent() *binanceRepositories.SymbolsRepository {
	if r.ParentRepository == nil {
		r.ParentRepository = &binanceRepositories.SymbolsRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.ParentRepository
}

func (r *SymbolsRepository) Symbols() []string {
	var symbols []string
	r.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	return symbols
}

func (r *SymbolsRepository) Price(symbol string) (float64, error) {
	fields := []string{
		"price",
		"timestamp",
	}
	data, _ := r.Rdb.HMGet(
		r.Ctx,
		fmt.Sprintf(
			"binance:spot:realtime:%s",
			symbol,
		),
		fields...,
	).Result()
	for i := 0; i < len(fields); i++ {
		if data[i] == nil {
			return 0, &SymbolsError{"price not exists"}
		}
	}

	timestamp := time.Now().Unix()
	price, _ := strconv.ParseFloat(data[0].(string), 64)
	lasttime, _ := strconv.ParseInt(data[1].(string), 10, 64)
	if timestamp-lasttime > 60 {
		return 0, &SymbolsError{"price long time not freshed"}
	}

	return price, nil
}

func (r *SymbolsRepository) Filter(symbol string, price float64, amount float64) (float64, float64) {
	return r.Parent().Filter(symbol, price, amount)
}

func (r *SymbolsRepository) Context(symbol string) map[string]interface{} {
	day := time.Now().Format("0102")
	fields := []string{
		"r3",
		"r2",
		"r1",
		"s1",
		"s2",
		"s3",
		"profit_target",
		"stop_loss_point",
		"take_profit_price",
	}
	data, _ := r.Rdb.HMGet(
		r.Ctx,
		fmt.Sprintf(
			"binance:spot:indicators:%s:%s",
			symbol,
			day,
		),
		fields...,
	).Result()
	var context = make(map[string]interface{})
	for i := 0; i < len(fields); i++ {
		context[fields[i]] = data[i]
	}

	return context
}
