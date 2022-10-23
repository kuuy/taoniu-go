package strategies

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/xid"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/go-redis/redis/v8"

	models "taoniu.local/cryptos/models/binance/spot"
)

type DailyRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *DailyRepository) Atr(symbol string) error {
	day := time.Now().Format("0102")
	atrVal, err := r.Rdb.HGet(
		r.Ctx,
		fmt.Sprintf(
			"binance:spot:indicators:%s:%s",
			symbol,
			day,
		),
		"atr",
	).Result()
	if err != nil {
		return err
	}
	atr, _ := strconv.ParseFloat(atrVal, 64)
	priceVal, err := r.Rdb.HGet(
		r.Ctx,
		fmt.Sprintf(
			"binance:spot:realtime:%s",
			symbol,
		),
		"price",
	).Result()
	if err != nil {
		return err
	}
	price, _ := strconv.ParseFloat(priceVal, 64)

	profitTarget := 2*price - 1.5*atr
	stopLossPoint := price - atr
	riskRewardRatio := (price - stopLossPoint) / (profitTarget - stopLossPoint)
	takeProfitPrice := stopLossPoint + (profitTarget-stopLossPoint)/2
	takeProfitRatio := price / takeProfitPrice

	r.Rdb.HMSet(
		r.Ctx,
		fmt.Sprintf(
			"binance:spot:indicators:%s:%s",
			symbol,
			day,
		),
		map[string]interface{}{
			"profit_target":     profitTarget,
			"stop_loss_point":   stopLossPoint,
			"risk_reward_ratio": riskRewardRatio,
			"take_profit_price": takeProfitPrice,
			"take_profit_ratio": takeProfitRatio,
		},
	)

	return nil
}

func (r *DailyRepository) Zlema(symbol string) error {
	indicator := "zlema"
	duration := "1d"
	val, err := r.Rdb.HGet(
		r.Ctx,
		fmt.Sprintf(
			"binance:spot:indicators:%s:%s",
			symbol,
			time.Now().Format("0102"),
		),
		indicator,
	).Result()
	if err != nil {
		return err
	}
	data := strings.Split(val, ",")

	price, _ := strconv.ParseFloat(data[2], 64)
	zlema1, _ := strconv.ParseFloat(data[0], 64)
	zlema2, _ := strconv.ParseFloat(data[1], 64)
	timestamp, _ := strconv.ParseInt(data[3], 10, 64)
	if zlema1*zlema2 >= 0.0 {
		return nil
	}
	var signal int64
	if zlema2 > 0 {
		signal = 1
	} else {
		signal = 2
	}
	var entity models.Strategy
	result := r.Db.Where(
		"symbol=? AND indicator=? AND duration=?",
		symbol,
		indicator,
		duration,
	).Order(
		"timestamp DESC",
	).Take(&entity)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		if entity.Signal == signal {
			return nil
		}
		if entity.Timestamp >= timestamp {
			return nil
		}
	}
	entity = models.Strategy{
		ID:        xid.New().String(),
		Symbol:    symbol,
		Indicator: indicator,
		Duration:  duration,
		Price:     price,
		Signal:    signal,
		Timestamp: timestamp,
		Remark:    "",
	}
	r.Db.Create(&entity)

	return nil
}

func (r *DailyRepository) HaZlema(symbol string) error {
	indicator := "ha_zlema"
	duration := "1d"
	val, err := r.Rdb.HGet(
		r.Ctx,
		fmt.Sprintf(
			"binance:spot:indicators:%s:%s",
			symbol,
			time.Now().Format("0102"),
		),
		"ha_zlema",
	).Result()
	if err != nil {
		return err
	}
	data := strings.Split(val, ",")

	price, _ := strconv.ParseFloat(data[2], 64)
	zlema1, _ := strconv.ParseFloat(data[0], 64)
	zlema2, _ := strconv.ParseFloat(data[1], 64)
	timestamp, _ := strconv.ParseInt(data[3], 10, 64)
	if zlema1*zlema2 >= 0.0 {
		return nil
	}
	var signal int64
	if zlema2 > 0 {
		signal = 1
	} else {
		signal = 2
	}
	var entity models.Strategy
	result := r.Db.Where(
		"symbol=? AND indicator=? AND duration=?",
		symbol,
		indicator,
		duration,
	).Order(
		"timestamp DESC",
	).Take(&entity)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		if entity.Signal == signal {
			return nil
		}
		if entity.Timestamp >= timestamp {
			return nil
		}
	}
	entity = models.Strategy{
		ID:        xid.New().String(),
		Symbol:    symbol,
		Indicator: indicator,
		Duration:  duration,
		Price:     price,
		Signal:    signal,
		Timestamp: timestamp,
		Remark:    "",
	}
	r.Db.Create(&entity)

	return nil
}

func (r *DailyRepository) Kdj(symbol string) error {
	indicator := "kdj"
	duration := "1d"
	val, err := r.Rdb.HGet(
		r.Ctx,
		fmt.Sprintf(
			"binance:spot:indicators:%s:%s",
			symbol,
			time.Now().Format("0102"),
		),
		indicator,
	).Result()
	if err != nil {
		return err
	}
	data := strings.Split(val, ",")

	k, _ := strconv.ParseFloat(data[0], 64)
	d, _ := strconv.ParseFloat(data[1], 64)
	j, _ := strconv.ParseFloat(data[2], 64)
	price, _ := strconv.ParseFloat(data[3], 64)
	timestamp, _ := strconv.ParseInt(data[4], 10, 64)
	var signal int64
	if k < 20 && d < 30 && j < 60 {
		signal = 1
	}
	if k > 80 && d > 70 && j > 90 {
		signal = 2
	}
	if signal == 0 {
		return nil
	}
	var entity models.Strategy
	result := r.Db.Where(
		"symbol=? AND indicator=? AND duration=?",
		symbol,
		indicator,
		duration,
	).Order(
		"timestamp DESC",
	).Take(&entity)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		if entity.Signal == signal {
			return nil
		}
		if entity.Timestamp >= timestamp {
			return nil
		}
	}
	entity = models.Strategy{
		ID:        xid.New().String(),
		Symbol:    symbol,
		Indicator: indicator,
		Duration:  duration,
		Price:     price,
		Signal:    signal,
		Timestamp: timestamp,
		Remark:    "",
	}
	r.Db.Create(&entity)

	return nil
}

func (r *DailyRepository) BBands(symbol string) error {
	indicator := "bbands"
	duration := "1d"
	val, err := r.Rdb.HGet(
		r.Ctx,
		fmt.Sprintf(
			"binance:spot:indicators:%s:%s",
			symbol,
			time.Now().Format("0102"),
		),
		indicator,
	).Result()
	if err != nil {
		return err
	}
	data := strings.Split(val, ",")

	b1, _ := strconv.ParseFloat(data[0], 64)
	b2, _ := strconv.ParseFloat(data[1], 64)
	b3, _ := strconv.ParseFloat(data[2], 64)
	w1, _ := strconv.ParseFloat(data[3], 64)
	w2, _ := strconv.ParseFloat(data[4], 64)
	w3, _ := strconv.ParseFloat(data[5], 64)
	price, _ := strconv.ParseFloat(data[6], 64)
	timestamp, _ := strconv.ParseInt(data[7], 10, 64)
	var signal int64
	if b1 < 0.5 && b2 < 0.5 && b3 > 0.5 {
		signal = 1
	}
	if b1 > 0.5 && b2 < 0.5 && b3 < 0.5 {
		signal = 2
	}
	if b1 > 0.8 && b2 > 0.8 && b3 > 0.8 {
		signal = 1
	}
	if b1 > 0.8 && b2 > 0.8 && b3 < 0.8 {
		signal = 2
	}
	if w1 < 0.1 && w2 < 0.1 && w3 < 0.1 {
		if w1 < 0.03 || w2 < 0.03 || w3 > 0.03 {
			return nil
		}
	}
	if signal == 0 {
		return nil
	}
	var entity models.Strategy
	result := r.Db.Where(
		"symbol=? AND indicator=? AND duration=?",
		symbol,
		indicator,
		duration,
	).Order(
		"timestamp DESC",
	).Take(&entity)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		if entity.Signal == signal {
			return nil
		}
		if entity.Timestamp >= timestamp {
			return nil
		}
	}
	entity = models.Strategy{
		ID:        xid.New().String(),
		Symbol:    symbol,
		Indicator: indicator,
		Duration:  duration,
		Price:     price,
		Signal:    signal,
		Timestamp: timestamp,
		Remark:    "",
	}
	r.Db.Create(&entity)

	return nil
}
