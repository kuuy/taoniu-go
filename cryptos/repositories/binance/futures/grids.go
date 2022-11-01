package futures

import (
	"context"
	"errors"
	"math"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
	"gorm.io/gorm"

	models "taoniu.local/cryptos/models/binance/futures"
)

type GridsRepository struct {
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	SymbolsRepository *SymbolsRepository
}

type GridsError struct {
	Message string
}

func (m *GridsError) Error() string {
	return m.Message
}

func (r *GridsRepository) Symbols() *SymbolsRepository {
	if r.SymbolsRepository == nil {
		r.SymbolsRepository = &SymbolsRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.SymbolsRepository
}

func (r *GridsRepository) Flush(symbol string) error {
	var entity models.Grids
	result := r.Db.Where(
		"symbol=? AND status=1",
		symbol,
	).Order("step desc").Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	context := r.Symbols().Context(symbol)
	profitTarget, _ := strconv.ParseFloat(context["profit_target"].(string), 64)
	takeProfitPrice, _ := strconv.ParseFloat(context["take_profit_price"].(string), 64)
	stopLossPoint, _ := strconv.ParseFloat(context["stop_loss_point"].(string), 64)
	if profitTarget > entity.StopLossPoint {
		price, err := r.Symbols().Price(symbol)
		if err != nil {
			return err
		}
		if price < entity.ProfitTarget {
			return nil
		}

		if entity.Step == 1 {
			return r.Close(symbol)
		}

		r.Db.Model(&entity).Update("status", 2)

		return nil
	}
	amount := entity.Amount * math.Pow(2, float64(entity.Step-1))
	entity = models.Grids{
		ID:                xid.New().String(),
		Symbol:            symbol,
		Step:              entity.Step + 1,
		Balance:           amount,
		Amount:            amount,
		ProfitTarget:      profitTarget,
		StopLossPoint:     stopLossPoint,
		TakeProfitPrice:   takeProfitPrice,
		TriggerPercent:    entity.TriggerPercent * 0.8,
		TakeProfitPercent: 0.05,
		Status:            1,
	}
	r.Db.Create(&entity)

	return nil
}

func (r *GridsRepository) Open(symbol string, amount float64) error {
	var entity models.Grids
	result := r.Db.Where(
		"symbol=? AND status=1",
		symbol,
	).Take(&entity)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return &GridsError{"grid already opened"}
	}
	context := r.Symbols().Context(symbol)
	profitTarget, _ := strconv.ParseFloat(context["profit_target"].(string), 64)
	takeProfitPrice, _ := strconv.ParseFloat(context["take_profit_price"].(string), 64)
	stopLossPoint, _ := strconv.ParseFloat(context["stop_loss_point"].(string), 64)
	entity = models.Grids{
		ID:                xid.New().String(),
		Symbol:            symbol,
		Step:              1,
		Balance:           amount,
		Amount:            amount,
		ProfitTarget:      profitTarget,
		StopLossPoint:     stopLossPoint,
		TakeProfitPrice:   takeProfitPrice,
		TriggerPercent:    1,
		TakeProfitPercent: 0.05,
		Status:            1,
	}
	r.Db.Create(&entity)

	r.Rdb.SAdd(r.Ctx, "binance:futures:grids:symbols", symbol)

	return nil
}

func (r *GridsRepository) Close(symbol string) error {
	r.Db.Model(&models.Grids{}).Where(
		"symbol=? AND status=1",
		symbol,
	).Update("status", 2)

	r.Rdb.SRem(r.Ctx, "binance:futures:grids:symbols", symbol)

	return nil
}

func (r *GridsRepository) Filter(symbol string, price float64) (*models.Grids, error) {
	var entities []*models.Grids
	r.Db.Where(
		"symbol=? AND status=1",
		symbol,
	).Order(
		"step asc",
	).Find(&entities)
	for _, entity := range entities {
		if price > entity.TakeProfitPrice {
			continue
		}
		if price < entity.StopLossPoint {
			continue
		}
		return entity, nil
	}

	return nil, &GridsError{"no valid grid"}
}