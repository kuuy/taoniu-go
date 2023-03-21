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

	models "taoniu.local/cryptos/models/binance/spot/analysis/tradings/fishers"
	fishersModels "taoniu.local/cryptos/models/binance/spot/tradings/fishers"
)

type GridsRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *GridsRepository) Flush() error {
	now := time.Now().Add(time.Minute * -6)
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
	analysis.Profit = 0
	analysis.Data = make(datatypes.JSONMap)
	analysis.Data["today"] = []string{}
	analysis.Data["history"] = []string{}
	analysis.Data["quantity"] = map[string]float64{}

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
			analysis.Profit += entity.SellPrice*entity.SellQuantity - entity.BuyPrice*entity.BuyQuantity
			if _, ok := analysis.Data["quantity"].(map[string]float64)[entity.Symbol]; ok {
				analysis.Data["quantity"].(map[string]float64)[entity.Symbol] += entity.BuyQuantity - entity.SellQuantity
			} else {
				analysis.Data["quantity"].(map[string]float64)[entity.Symbol] = entity.BuyQuantity - entity.SellQuantity
			}
		}
	}

	r.Db.Where(
		"status=2 AND created_at<?",
		datetime,
	).Find(&entities)
	for _, entity := range entities {
		analysis.Data["history"] = append(analysis.Data["history"].([]string), entity.ID)
	}

	r.Db.Where(
		"status=3 AND created_at<? AND updated_at>=?",
		datetime,
		datetime,
	).Find(&entities)
	for _, entity := range entities {
		analysis.SellsCount += 1
		analysis.SellsAmount += entity.SellPrice * entity.SellQuantity
		analysis.Profit += entity.SellPrice*entity.SellQuantity - entity.BuyPrice*entity.BuyQuantity
		if _, ok := analysis.Data["quantity"].(map[string]float64)[entity.Symbol]; ok {
			analysis.Data["quantity"].(map[string]float64)[entity.Symbol] += entity.BuyQuantity - entity.SellQuantity
		} else {
			analysis.Data["quantity"].(map[string]float64)[entity.Symbol] = entity.BuyQuantity - entity.SellQuantity
		}
	}

	r.Db.Save(&analysis)

	return nil
}

func (r *GridsRepository) Count() int64 {
	var total int64
	r.Db.Model(&models.Grid{}).Count(&total)
	return total
}

func (r *GridsRepository) Listings(current int, pageSize int) []*models.Grid {
	var plans []*models.Grid
	r.Db.Select([]string{
		"id",
		"day",
		"buys_count",
		"sells_count",
		"buys_amount",
		"sells_amount",
		"profit",
		"data",
	}).Order(
		"day desc",
	).Offset(
		(current - 1) * pageSize,
	).Limit(
		pageSize,
	).Find(&plans)
	return plans
}

func (r *GridsRepository) Series(limit int) []interface{} {
	var grids []*models.Grid
	r.Db.Order("day desc").Limit(limit).Find(&grids)

	series := make([]interface{}, len(grids))
	for i, grid := range grids {
		series[i] = []interface{}{
			grid.BuysCount,
			grid.SellsCount,
			time.Time(grid.Day).Format("01/02"),
		}
	}
	return series
}

func (r *GridsRepository) JSONMap(in interface{}) datatypes.JSONMap {
	buf, _ := json.Marshal(in)

	var out datatypes.JSONMap
	json.Unmarshal(buf, &out)
	return out
}
