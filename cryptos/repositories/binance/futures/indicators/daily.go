package indicators

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/go-redis/redis/v8"
	"github.com/markcheno/go-talib"

	models "taoniu.local/cryptos/models/binance/futures"
)

type DailyError struct {
	Message string
}

func (m *DailyError) Error() string {
	return m.Message
}

type DailyRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *DailyRepository) Pivot(symbol string) error {
	var kline models.Kline
	result := r.Db.Select(
		[]string{"close", "high", "low", "timestamp"},
	).Where(
		"symbol=? AND interval=?", symbol, "1d",
	).Order(
		"timestamp desc",
	).Take(&kline)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	p := (kline.Close + kline.High + kline.Low) / 3
	s1 := 2*p - kline.High
	r1 := 2*p - kline.Low
	s2 := p - (r1 - s1)
	r2 := p + (r1 - s1)
	s3 := kline.Low - 2*(kline.High-p)
	r3 := kline.High + 2*(p-kline.Low)

	day, err := r.Day(kline.Timestamp / 1000)
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf(
		"binance:futures:indicators:%s:%s",
		symbol,
		day,
	)
	exists, _ := r.Rdb.Exists(r.Ctx, redisKey).Result()
	r.Rdb.HMSet(
		r.Ctx,
		redisKey,
		map[string]interface{}{
			"r3": r3,
			"r2": r2,
			"r1": r1,
			"s1": s1,
			"s2": s2,
			"s3": s3,
		},
	)
	if exists != 1 {
		r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
	}

	return nil
}

func (r *DailyRepository) Atr(symbol string, period int, limit int) error {
	var klines []*models.Kline
	r.Db.Select(
		[]string{"close", "high", "low", "timestamp"},
	).Where(
		"symbol=? AND interval=?", symbol, "1d",
	).Order(
		"timestamp desc",
	).Limit(
		limit,
	).Find(
		&klines,
	)
	var highs []float64
	var lows []float64
	var prices []float64
	var timestamp int64
	for _, item := range klines {
		if timestamp > 0 && (timestamp-item.Timestamp) != 86400000 {
			return nil
		}
		prices = append([]float64{item.Close}, prices...)
		highs = append([]float64{item.High}, highs...)
		lows = append([]float64{item.Low}, lows...)
		timestamp = item.Timestamp
	}
	if len(prices) < limit {
		return nil
	}

	result := talib.Atr(
		highs,
		lows,
		prices,
		period,
	)

	day, err := r.Day(klines[0].Timestamp / 1000)
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf(
		"binance:futures:indicators:%s:%s",
		symbol,
		day,
	)
	exists, _ := r.Rdb.Exists(r.Ctx, redisKey).Result()
	r.Rdb.HSet(
		r.Ctx,
		redisKey,
		"atr",
		strconv.FormatFloat(result[limit-1], 'f', -1, 64),
	)
	if exists != 1 {
		r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
	}

	return nil
}

func (r *DailyRepository) Zlema(symbol string, period int, limit int) error {
	var klines []*models.Kline
	r.Db.Select(
		[]string{"close", "timestamp"},
	).Where(
		"symbol=? AND interval=?", symbol, "1d",
	).Order(
		"timestamp desc",
	).Limit(
		limit,
	).Find(
		&klines,
	)
	var data []float64
	var temp []float64
	var timestamp int64
	var lag = int((period - 1) / 2)
	for _, item := range klines {
		if len(temp) < lag {
			temp = append([]float64{item.Close}, temp...)
			continue
		}
		if timestamp > 0 && (timestamp-item.Timestamp) != 86400000 {
			return nil
		}
		data = append([]float64{item.Close - temp[lag-1]}, data...)
		temp = append([]float64{item.Close}, temp[:lag-1]...)
		timestamp = item.Timestamp
	}
	if len(data) < limit-lag {
		return nil
	}

	result := talib.Ema(data, period)

	day, err := r.Day(klines[0].Timestamp / 1000)
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf(
		"binance:futures:indicators:%s:%s",
		symbol,
		day,
	)
	exists, _ := r.Rdb.Exists(r.Ctx, redisKey).Result()
	r.Rdb.HSet(
		r.Ctx,
		redisKey,
		"zlema",
		fmt.Sprintf(
			"%s,%s,%s,%d",
			strconv.FormatFloat(result[limit-lag-2], 'f', -1, 64),
			strconv.FormatFloat(result[limit-lag-1], 'f', -1, 64),
			strconv.FormatFloat(klines[0].Close, 'f', -1, 64),
			time.Now().Unix(),
		),
	)
	if exists != 1 {
		r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
	}

	return nil
}

func (r *DailyRepository) HaZlema(symbol string, period int, limit int) error {
	var klines []*models.Kline
	r.Db.Select(
		[]string{"open", "close", "high", "low", "timestamp"},
	).Where(
		"symbol=? AND interval=?", symbol, "1d",
	).Order(
		"timestamp desc",
	).Limit(
		limit,
	).Find(
		&klines,
	)
	var data []float64
	var temp []float64
	var timestamp int64
	var lag = int((period - 1) / 2)
	for _, item := range klines {
		var avgPrice = (item.Open + item.Close + item.High + item.Low) / 4
		if len(temp) < lag {
			temp = append([]float64{avgPrice}, temp...)
			continue
		}
		if timestamp > 0 && (timestamp-item.Timestamp) != 86400000 {
			return nil
		}
		data = append([]float64{avgPrice - temp[lag-1]}, data...)
		temp = append([]float64{avgPrice}, temp[:lag-1]...)
		timestamp = item.Timestamp
	}
	if len(data) < limit-lag {
		return nil
	}

	result := talib.Ema(data, period)

	day, err := r.Day(klines[0].Timestamp / 1000)
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf(
		"binance:futures:indicators:%s:%s",
		symbol,
		day,
	)
	exists, _ := r.Rdb.Exists(r.Ctx, redisKey).Result()
	r.Rdb.HSet(
		r.Ctx,
		redisKey,
		"ha_zlema",
		fmt.Sprintf(
			"%s,%s,%s,%d",
			strconv.FormatFloat(result[limit-lag-2], 'f', -1, 64),
			strconv.FormatFloat(result[limit-lag-1], 'f', -1, 64),
			strconv.FormatFloat(klines[0].Close, 'f', -1, 64),
			time.Now().Unix(),
		),
	)
	if exists != 1 {
		r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
	}

	return nil
}

func (r *DailyRepository) Kdj(symbol string, longPeriod int, shortPeriod int, limit int) error {
	var klines []*models.Kline
	r.Db.Select(
		[]string{"open", "close", "high", "low", "timestamp"},
	).Where(
		"symbol=? AND interval=?", symbol, "1d",
	).Order(
		"timestamp desc",
	).Limit(
		limit,
	).Find(
		&klines,
	)
	var highs []float64
	var lows []float64
	var prices []float64
	var timestamp int64
	for _, item := range klines {
		if timestamp > 0 && (timestamp-item.Timestamp) != 86400000 {
			return nil
		}
		var avgPrice = (item.Close + item.High + item.Low) / 3
		highs = append([]float64{item.High}, highs...)
		lows = append([]float64{item.Low}, lows...)
		prices = append([]float64{avgPrice}, prices...)
		timestamp = item.Timestamp
	}
	if len(prices) < limit {
		return nil
	}

	slowk, slowd := talib.Stoch(highs, lows, prices, longPeriod, shortPeriod, 0, shortPeriod, 0)
	var slowj []float64
	for i := 0; i < limit; i++ {
		slowj = append(slowj, 3*slowk[i]-2*slowd[i])
	}

	day, err := r.Day(klines[0].Timestamp / 1000)
	if err != nil {
		return err
	}

	r.Rdb.HSet(
		r.Ctx,
		fmt.Sprintf(
			"binance:futures:indicators:%s:%s",
			symbol,
			day,
		),
		"kdj",
		fmt.Sprintf(
			"%s,%s,%s,%s,%d",
			strconv.FormatFloat(slowk[limit-1], 'f', -1, 64),
			strconv.FormatFloat(slowd[limit-1], 'f', -1, 64),
			strconv.FormatFloat(slowj[limit-1], 'f', -1, 64),
			strconv.FormatFloat(klines[0].Close, 'f', -1, 64),
			time.Now().Unix(),
		),
	)

	return nil
}

func (r *DailyRepository) BBands(symbol string, period int, limit int) error {
	var klines []*models.Kline
	r.Db.Select(
		[]string{"open", "close", "high", "low", "timestamp"},
	).Where(
		"symbol=? AND interval=?", symbol, "1d",
	).Order(
		"timestamp desc",
	).Limit(
		limit,
	).Find(
		&klines,
	)
	var prices []float64
	var timestamp int64
	for _, item := range klines {
		if timestamp > 0 && (timestamp-item.Timestamp) != 86400000 {
			return nil
		}
		var avgPrice = (item.Close + item.High + item.Low) / 3
		prices = append([]float64{avgPrice}, prices...)
		timestamp = item.Timestamp
	}
	if len(prices) < limit {
		return nil
	}

	uBands, mBands, lBands := talib.BBands(prices, period, 2, 2, 0)
	p1 := (klines[2].Close + klines[2].High + klines[2].Low) / 3
	p2 := (klines[1].Close + klines[1].High + klines[1].Low) / 3
	p3 := (klines[0].Close + klines[0].High + klines[0].Low) / 3
	b1 := (p1 - lBands[limit-3]) / (uBands[limit-3] - lBands[limit-3])
	b2 := (p2 - lBands[limit-2]) / (uBands[limit-2] - lBands[limit-2])
	b3 := (p3 - lBands[limit-1]) / (uBands[limit-1] - lBands[limit-1])
	w1 := (uBands[limit-3] - lBands[limit-3]) / mBands[limit-3]
	w2 := (uBands[limit-2] - lBands[limit-2]) / mBands[limit-2]
	w3 := (uBands[limit-1] - lBands[limit-1]) / mBands[limit-1]

	day, err := r.Day(klines[0].Timestamp / 1000)
	if err != nil {
		return err
	}

	r.Rdb.HSet(
		r.Ctx,
		fmt.Sprintf(
			"binance:futures:indicators:%s:%s",
			symbol,
			day,
		),
		"bbands",
		fmt.Sprintf(
			"%s,%s,%s,%s,%s,%s,%s,%d",
			strconv.FormatFloat(b1, 'f', -1, 64),
			strconv.FormatFloat(b2, 'f', -1, 64),
			strconv.FormatFloat(b3, 'f', -1, 64),
			strconv.FormatFloat(w1, 'f', -1, 64),
			strconv.FormatFloat(w2, 'f', -1, 64),
			strconv.FormatFloat(w3, 'f', -1, 64),
			strconv.FormatFloat(klines[0].Close, 'f', -1, 64),
			time.Now().Unix(),
		),
	)

	return nil
}

func (r *DailyRepository) Day(timestamp int64) (string, error) {
	now := time.Now()
	_, offset := now.Zone()

	utc := now.Add(time.Second * time.Duration(-offset))
	duration := time.Hour*time.Duration(8-utc.Hour()) - time.Minute*time.Duration(utc.Minute()) - time.Second*time.Duration(utc.Second())
	open := utc.Add(duration)
	last := time.Unix(timestamp, 0)
	if open.Format("0102") != last.Format("0102") {
		return "", &DailyError{"timestamp is not today"}
	}

	return now.Format("0102"), nil
}