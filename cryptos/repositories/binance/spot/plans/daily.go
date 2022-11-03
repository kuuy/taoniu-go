package plans

import (
	"context"
	"errors"
	"github.com/rs/xid"
	"time"

	"gorm.io/gorm"

	"github.com/go-redis/redis/v8"

	models "taoniu.local/cryptos/models/binance/spot"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	tradingviewRepositories "taoniu.local/cryptos/repositories/tradingview"
)

type DailyError struct {
	Message string
}

func (m *DailyError) Error() string {
	return m.Message
}

type DailyRepository struct {
	Db                    *gorm.DB
	Rdb                   *redis.Client
	Ctx                   context.Context
	SymbolsRepository     *spotRepositories.SymbolsRepository
	TradingviewRepository *tradingviewRepositories.AnalysisRepository
}

func (r *DailyRepository) Tradingview() *tradingviewRepositories.AnalysisRepository {
	if r.TradingviewRepository == nil {
		r.TradingviewRepository = &tradingviewRepositories.AnalysisRepository{
			Db: r.Db,
		}
	}
	return r.TradingviewRepository
}

func (r *DailyRepository) Symbols() *spotRepositories.SymbolsRepository {
	if r.SymbolsRepository == nil {
		r.SymbolsRepository = &spotRepositories.SymbolsRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.SymbolsRepository
}

func (r *DailyRepository) Flush() error {
	buys, sells := r.Signals()
	r.Create(buys, 1)
	r.Create(sells, 2)
	return nil
}

func (r *DailyRepository) Create(signals map[string]interface{}, side int64) error {
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
		price, quantity := r.Symbols().Filter(symbol, price, amount)
		if price == 0 || quantity == 0 {
			continue
		}
		var entity models.Plan
		result := r.Db.Where(
			"symbol=? AND timestamp=?",
			symbol,
			timestamp,
		).Take(&entity)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			continue
		}
		entity = models.Plan{
			ID:        xid.New().String(),
			Symbol:    symbol,
			Side:      side,
			Price:     price,
			Quantity:  quantity,
			Amount:    amount,
			Timestamp: timestamp,
			Context:   r.Symbols().Context(symbol),
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
		"indicator in ? AND interval = ? AND timestamp > ?",
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

func (r *DailyRepository) Filter() (*models.Plan, error) {
	now := time.Now()
	duration := time.Hour*time.Duration(8-now.Hour()) - time.Minute*time.Duration(now.Minute()) - time.Second*time.Duration(now.Second())
	timestamp := now.Add(duration).Unix()

	var entities []*models.Plan
	r.Db.Where(
		"timestamp=? AND status=0",
		timestamp,
	).Find(&entities)
	for _, entity := range entities {
		if entity.Side == 1 && entity.Amount < 20 {
			continue
		}
		signal, _, err := r.Tradingview().Signal(entity.Symbol)
		if err != nil {
			continue
		}
		if entity.Side == 1 && signal != 1 {
			continue
		}
		if entity.Side == 2 && signal != 2 {
			continue
		}

		price, err := r.Symbols().Price(entity.Symbol)
		if err != nil {
			continue
		}
		if signal == 1 && price > entity.Price {
			continue
		}
		if signal == 2 && price < entity.Price {
			continue
		}

		return entity, nil
	}

	return nil, &DailyError{"no valid plan"}
}
