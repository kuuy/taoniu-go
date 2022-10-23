package spot

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
	"gorm.io/gorm"
	"log"
	"strconv"
	models "taoniu.local/cryptos/models/binance/spot"
	"taoniu.local/cryptos/repositories/binance/spot/plans"
)

type GridsRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

type GridsError struct {
	Message string
}

func (m *GridsError) Error() string {
	return m.Message
}

func (r *GridsRepository) Open(symbol string, balance float64) error {
	var entity models.Grids
	result := r.Db.Where(
		"symbol=? AND status=1",
		symbol,
	).Take(&entity)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return &GridsError{"grid already opened"}
	}
	repository := plans.DailyRepository{
		Db:  r.Db,
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
	context := repository.Context(symbol)
	profitTarget, _ := strconv.ParseFloat(context["profit_target"].(string), 64)
	takeProfitPrice, _ := strconv.ParseFloat(context["take_profit_price"].(string), 64)
	stopLossPoint, _ := strconv.ParseFloat(context["stop_loss_point"].(string), 64)
	entity = models.Grids{
		ID:                xid.New().String(),
		Symbol:            symbol,
		Step:              1,
		Balance:           balance,
		ProfitTarget:      profitTarget,
		StopLossPoint:     stopLossPoint,
		TakeProfitPrice:   takeProfitPrice,
		TriggerPercent:    1,
		TakeProfitPercent: 0.05,
		Status:            1,
	}
	r.Db.Create(&entity)

	return nil
}

func (r *GridsRepository) Close(symbol string) error {
	log.Println("open", symbol)
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
