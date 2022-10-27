package margin

import (
	"context"
	"errors"
	"github.com/rs/xid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	isolatedModels "taoniu.local/cryptos/models/binance/spot/margin/isolated"
	"time"

	"github.com/go-redis/redis/v8"

	models "taoniu.local/cryptos/models/binance/spot/analysis/daily/margin"
)

type IsolatedRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *IsolatedRepository) Grids() error {
	now := time.Now().Add(time.Minute * -5)
	duration := time.Hour*time.Duration(-now.Hour()) + time.Minute*time.Duration(-now.Minute()) + time.Second*time.Duration(-now.Second())
	datetime := now.Add(duration)

	var analysis *models.Isolated
	result := r.Db.Where("day", datatypes.Date(datetime)).Take(&analysis)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		analysis = &models.Isolated{
			ID:  xid.New().String(),
			Day: datatypes.Date(datetime),
		}
	}
	analysis.GridsBuysCount = 0
	analysis.GridsSellsCount = 0
	analysis.GridsBuysAmount = 0
	analysis.GridsSellsAmount = 0

	var entities []*isolatedModels.TradingGrid
	r.Db.Where(
		"updated_at>?",
		datetime,
	).Find(&entities)
	for _, entity := range entities {
		if entity.Status == 1 || entity.Status == 3 {
			analysis.GridsBuysCount += 1
			analysis.GridsBuysAmount += entity.BuyPrice * entity.BuyQuantity
		}
		if entity.Status == 3 {
			analysis.GridsSellsCount += 1
			analysis.GridsSellsAmount += entity.SellPrice * entity.SellQuantity
			analysis.GridsProfit += (entity.SellPrice - entity.BuyPrice) * entity.SellQuantity
		}
	}
	r.Db.Save(&analysis)

	return nil
}
