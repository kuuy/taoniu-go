package fishers

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	models "taoniu.local/cryptos/models/binance/spot/analysis/margin/isolated/tradings/fishers"
	fishersModels "taoniu.local/cryptos/models/binance/spot/margin/isolated/tradings/fishers"
)

type GridsRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *GridsRepository) Flush() error {
	now := time.Now().Add(time.Minute * -5)
	duration := time.Hour*time.Duration(-now.Hour()) + time.Minute*time.Duration(-now.Minute()) + time.Second*time.Duration(-now.Second())
	datetime := now.Add(duration)

	var analysis *models.Grid
	result := r.Db.Where("day", datatypes.Date(datetime)).Take(&analysis)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		analysis = &models.Grid{
			ID:  xid.New().String(),
			Day: datatypes.Date(datetime),
		}
	}

	analysis.BuysCount = 0
	analysis.BuysAmount = 0
	analysis.SellsCount = 0
	analysis.SellsAmount = 0
	analysis.ProfitAmount = 0
	analysis.ProfitQuantity = 0
	analysis.Data = make(datatypes.JSONMap)
	analysis.Data["today"] = []string{}
	analysis.Data["history"] = []string{}

	var entities []*fishersModels.Grid
	r.Db.Where(
		"created_at>=?",
		datetime,
	).Find(&entities)
	for _, entity := range entities {
		if entity.Status == 1 || entity.Status == 2 || entity.Status == 3 {
			analysis.BuysCount += 1
			analysis.BuysAmount += entity.BuyPrice * entity.BuyQuantity
			if entity.Status != 3 {
				analysis.Data["today"] = append(analysis.Data["today"].([]string), entity.ID)
			}
		}
		if entity.Status == 3 {
			analysis.SellsCount += 1
			analysis.SellsAmount += entity.SellPrice * entity.SellQuantity
			analysis.ProfitAmount += entity.SellPrice*entity.SellQuantity - entity.BuyPrice*entity.BuyQuantity
			analysis.ProfitQuantity += entity.BuyQuantity - entity.SellQuantity
		}
	}

	r.Db.Where(
		"status=2 AND created_at<?",
		datetime,
	).Find(&entities)
	for _, entity := range entities {
		analysis.Data["history"] = append(analysis.Data["history"].([]string), entity.ID)
	}

	r.Db.Save(&analysis)

	return nil
}

func (r *GridsRepository) JSONMap(in interface{}) datatypes.JSONMap {
	buf, _ := json.Marshal(in)

	var out datatypes.JSONMap
	json.Unmarshal(buf, &out)
	return out
}
