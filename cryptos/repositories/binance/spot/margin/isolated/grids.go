package isolated

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
	"gorm.io/gorm"
	"strconv"
	models "taoniu.local/cryptos/models/binance/spot/margin/isolated"
	"time"
)

type GridsRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *GridsRepository) Flush(symbol string) error {
	day := time.Now().Format("0102")
	data, err := r.Rdb.HMGet(
		r.Ctx,
		fmt.Sprintf(
			"binance:spot:indicators:%s:%s",
			symbol,
			day,
		),
		"r3",
		"r2",
		"r1",
		"s1",
		"s2",
		"s3",
		"profit_target",
		"stop_loss_point",
		"take_profit_price",
	).Result()
	if err != nil {
		return err
	}
	r3, _ := strconv.ParseFloat(data[0].(string), 64)
	r2, _ := strconv.ParseFloat(data[1].(string), 64)
	r1, _ := strconv.ParseFloat(data[2].(string), 64)
	s1, _ := strconv.ParseFloat(data[3].(string), 64)
	s2, _ := strconv.ParseFloat(data[4].(string), 64)
	s3, _ := strconv.ParseFloat(data[5].(string), 64)
	profitTarget, _ := strconv.ParseFloat(data[6].(string), 64)
	stopLossPoint, _ := strconv.ParseFloat(data[7].(string), 64)
	takeProfitPrice, _ := strconv.ParseFloat(data[8].(string), 64)
	var entity models.Grids
	var tx *gorm.DB
	tx = r.Db.Where(
		"symbol=?",
		symbol,
	).Order("status asc").First(&entity)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		entity = models.Grids{
			ID:              xid.New().String(),
			Symbol:          symbol,
			Step:            1,
			R3:              r3,
			R2:              r2,
			R1:              r1,
			S1:              s1,
			S2:              s2,
			S3:              s3,
			ProfitTarget:    profitTarget,
			StopLossPoint:   stopLossPoint,
			TakeProfitPrice: takeProfitPrice,
		}
		r.Db.Create(&entity)
	}
	if entity.Status == 0 {
		entity.R3 = r3
		entity.R2 = r2
		entity.R1 = r1
		entity.S1 = s1
		entity.S2 = s2
		entity.S3 = s3
		entity.ProfitTarget = profitTarget
		entity.StopLossPoint = stopLossPoint
		entity.TakeProfitPrice = takeProfitPrice
		r.Db.Model(&models.Grids{ID: entity.ID}).Updates(entity)
	}

	return nil
}
