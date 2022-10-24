package plans

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"

	models "taoniu.local/cryptos/models/binance/spot"
	binanceRepositories "taoniu.local/cryptos/repositories/binance"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
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

func (r *DailyRepository) SymbolsRepository() *binanceRepositories.SymbolsRepository {
	return &binanceRepositories.SymbolsRepository{
		Db:  r.Db,
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
}

func (r *DailyRepository) GridsRepository() *spotRepositories.GridsRepository {
	return &spotRepositories.GridsRepository{
		Db:  r.Db,
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
}

func (r *DailyRepository) Plans(signals map[string]interface{}, side int64) error {
	if _, ok := signals["kdj"]; !ok {
		return nil
	}
	now := time.Now()
	duration := time.Hour*time.Duration(8-now.Hour()) - time.Minute*time.Duration(now.Minute()) - time.Second*time.Duration(now.Second())
	timestamp := now.Add(duration).Unix()
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
		price, quantity := r.SymbolsRepository().Filter(symbol, price, amount)
		var entity models.Plans
		result := r.Db.Where(
			"symbol=? AND timestamp=?",
			symbol,
			timestamp,
		).Take(&entity)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context := r.SymbolsRepository().Context(symbol)
			isUpdate := false
			for key, val := range entity.Context {
				if val == nil {
					entity.Context[key] = context[key]
					isUpdate = true
				}
			}
			if isUpdate {
				r.Db.Model(&models.Plans{ID: entity.ID}).Updates(entity)
			}
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
			Context:   r.SymbolsRepository().Context(symbol),
		}
		r.Db.Create(&entity)

		r.GridsRepository().Flush(symbol)
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
