package plans

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
	"gorm.io/gorm"
	models "taoniu.local/cryptos/models/binance/spot"
	"taoniu.local/cryptos/repositories/binance"
	"time"
)

type DailyRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *DailyRepository) Flush() error {
	buys, sells := r.Signals()
	r.Plans(buys, 1)
	r.Plans(sells, 2)

	return nil
}

func (r *DailyRepository) Plans(signals map[string]interface{}, side int64) error {
	if _, ok := signals["kdj"]; !ok {
		return nil
	}
	now := time.Now()
	duration := time.Hour*time.Duration(8-now.Hour()) - time.Minute*time.Duration(now.Minute()) - time.Second*time.Duration(now.Second())
	timestamp := now.Add(duration).Unix()
	repository := binance.SymbolsRepository{
		Db:  r.Db,
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
	for symbol, price := range signals["kdj"].(map[string]float64) {
		amount := 10.0
		if _, ok := signals["bbands"]; ok {
			if p, ok := signals["bbands"].(map[string]float64)[symbol]; ok {
				if p < price {
					price = p
				}
				amount += 10
			}
		}
		if _, ok := signals["ha_zlema"]; ok {
			if p, ok := signals["ha_zlema"].(map[string]float64)[symbol]; ok {
				if p < price {
					price = p
				}
				amount += 5
			}
		}
		price, quantity := repository.Filter(symbol, price, amount)
		var entity models.Plans
		result := r.Db.Where(
			"symbol=? AND timestamp=?",
			symbol,
			timestamp,
		).First(&entity)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			continue
		}
		entity = models.Plans{
			ID:        xid.New().String(),
			Symbol:    symbol,
			Side:      side,
			Price:     price,
			Quantity:  quantity,
			Amount:    amount,
			Timestamp: timestamp,
			Context:   r.Context(symbol),
		}
		r.Db.Create(&entity)
	}

	return nil
}

func (r *DailyRepository) Signals() (map[string]interface{}, map[string]interface{}) {
	timestamp := time.Now().Unix() - 86400
	var strategies []*models.Strategy
	r.Db.Select([]string{
		"symbol",
		"indicator",
		"price",
		"signal",
	}).Where(
		"indicator in ? AND duration = ? AND timestamp > ?",
		[]string{
			"kdj",
			"bbands",
			"ha_zlema",
		},
		"1d",
		timestamp,
	).Order(
		"timestamp desc",
	).Find(&strategies)
	var buys = make(map[string]interface{})
	var sells = make(map[string]interface{})
	for _, strategy := range strategies {
		if _, ok := buys[strategy.Indicator]; strategy.Signal == 1 && !ok {
			buys[strategy.Indicator] = make(map[string]float64)
		}
		if strategy.Signal == 1 {
			buys[strategy.Indicator].(map[string]float64)[strategy.Symbol] = strategy.Price
		}
		if _, ok := sells[strategy.Indicator]; strategy.Signal == 2 && !ok {
			sells[strategy.Indicator] = make(map[string]float64)
		}
		if strategy.Signal == 2 {
			sells[strategy.Indicator].(map[string]float64)[strategy.Symbol] = strategy.Price
		}
	}

	return buys, sells
}

func (r *DailyRepository) Context(symbol string) map[string]interface{} {
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
