package indicators

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/markcheno/go-talib"

	config "taoniu.local/cryptos/config/binance/futures"
)

type IchimokuCloudRepository struct {
	BaseRepository
}

func (r *IchimokuCloudRepository) Get(symbol, interval string) (
	signal int,
	conversionLine,
	baseLine,
	senkouSpanA,
	senkouSpanB,
	chikouSpan,
	price float64,
	timestamp int64,
	err error,
) {
	day := time.Now().Format("0102")
	redisKey := fmt.Sprintf(
		config.REDIS_KEY_INDICATORS,
		interval,
		symbol,
		day,
	)
	val, err := r.Rdb.HGet(
		r.Ctx,
		redisKey,
		"ichimoku_cloud",
	).Result()
	if err != nil {
		return
	}
	data := strings.Split(val, ",")
	if len(data) < 8 {
		err = fmt.Errorf("invalid data in redis")
		return
	}
	signal, _ = strconv.Atoi(data[0])
	conversionLine, _ = strconv.ParseFloat(data[1], 64)
	baseLine, _ = strconv.ParseFloat(data[2], 64)
	senkouSpanA, _ = strconv.ParseFloat(data[3], 64)
	senkouSpanB, _ = strconv.ParseFloat(data[4], 64)
	chikouSpan, _ = strconv.ParseFloat(data[5], 64)
	price, _ = strconv.ParseFloat(data[6], 64)
	timestamp, _ = strconv.ParseInt(data[7], 10, 64)
	return
}

func (r *IchimokuCloudRepository) Flush(
	symbol string,
	interval string,
	tenkanPeriod int,
	kijunPeriod int,
	senkouPeriod int,
	limit int,
) error {
	data, timestamps, err := r.Klines(symbol, interval, limit, "high", "low", "close")
	if err != nil {
		return err
	}

	highs := data[0]
	lows := data[1]
	closes := data[2]
	lastIdx := len(timestamps) - 1

	highsTenkan := talib.Max(highs, tenkanPeriod)
	lowsTenkan := talib.Min(lows, tenkanPeriod)
	highsKijun := talib.Max(highs, kijunPeriod)
	lowsKijun := talib.Min(lows, kijunPeriod)
	highsSenkou := talib.Max(highs, senkouPeriod)
	lowsSenkou := talib.Min(lows, senkouPeriod)

	currConversionLine := (highsTenkan[lastIdx] + lowsTenkan[lastIdx]) / 2
	currBaseLine := (highsKijun[lastIdx] + lowsKijun[lastIdx]) / 2

	prevConversionLine := (highsTenkan[lastIdx-1] + lowsTenkan[lastIdx-1]) / 2
	prevBaseLine := (highsKijun[lastIdx-1] + lowsKijun[lastIdx-1]) / 2

	senkouSpanA := (currConversionLine + currBaseLine) / 2
	senkouSpanB := (highsSenkou[lastIdx] + lowsSenkou[lastIdx]) / 2
	chikouSpan := closes[lastIdx-kijunPeriod]

	var signal int
	if currConversionLine > currBaseLine && prevConversionLine < prevBaseLine {
		signal = 1
	}
	if currConversionLine < currBaseLine && prevConversionLine > prevBaseLine {
		signal = 2
	}

	day, err := r.Day(timestamps[lastIdx] / 1000)
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf(
		config.REDIS_KEY_INDICATORS,
		interval,
		symbol,
		day,
	)

	if signal == 0 {
		val, _ := r.Rdb.HGet(
			r.Ctx,
			redisKey,
			"ichimoku_cloud",
		).Result()
		values := strings.Split(val, ",")
		if len(values) == 8 {
			prevConversionLine, _ = strconv.ParseFloat(values[1], 64)
			prevBaseLine, _ = strconv.ParseFloat(values[2], 64)
			if currConversionLine > currBaseLine && prevConversionLine < prevBaseLine {
				signal = 1
			}
			if currConversionLine < currBaseLine && prevConversionLine > prevBaseLine {
				signal = 2
			}
		}
	}

	r.Rdb.HSet(
		r.Ctx,
		redisKey,
		"ichimoku_cloud",
		fmt.Sprintf(
			"%d,%s,%s,%s,%s,%s,%s,%d",
			signal,
			strconv.FormatFloat(currConversionLine, 'f', -1, 64),
			strconv.FormatFloat(currBaseLine, 'f', -1, 64),
			strconv.FormatFloat(senkouSpanA, 'f', -1, 64),
			strconv.FormatFloat(senkouSpanB, 'f', -1, 64),
			strconv.FormatFloat(chikouSpan, 'f', -1, 64),
			strconv.FormatFloat(closes[lastIdx], 'f', -1, 64),
			timestamps[lastIdx],
		),
	)
	ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
	if -1 == ttl.Nanoseconds() {
		r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
	}

	return nil
}
