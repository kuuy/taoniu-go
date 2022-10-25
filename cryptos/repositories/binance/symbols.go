package binance

import (
	"context"
	"errors"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"math"
	"strconv"
	"strings"
	config "taoniu.local/cryptos/config/binance"
	models "taoniu.local/cryptos/models/binance"
	"time"
)

type SymbolsRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

type SymbolsError struct {
	Message string
}

func (m *SymbolsError) Error() string {
	return m.Message
}

func (r *SymbolsRepository) Flush() error {
	client := binance.NewClient(config.REST_API_KEY, config.REST_SECRET_KEY)
	result, err := client.NewExchangeInfoService().Do(r.Ctx)
	if err != nil {
		return err
	}
	for _, item := range result.Symbols {
		if item.QuoteAsset != "BUSD" {
			continue
		}
		var filters = make(datatypes.JSONMap)
		for _, filter := range item.Filters {
			if filter["filterType"].(string) == string(binance.SymbolFilterTypePriceFilter) {
				if _, ok := filter["maxPrice"]; !ok {
					continue
				}
				if _, ok := filter["minPrice"]; !ok {
					continue
				}
				if _, ok := filter["tickSize"]; !ok {
					continue
				}
				maxPrice, _ := strconv.ParseFloat(filter["maxPrice"].(string), 64)
				minPrice, _ := strconv.ParseFloat(filter["minPrice"].(string), 64)
				tickSize, _ := strconv.ParseFloat(filter["tickSize"].(string), 64)
				filters["price"] = fmt.Sprintf(
					"%s,%s,%s",
					strconv.FormatFloat(maxPrice, 'f', -1, 64),
					strconv.FormatFloat(minPrice, 'f', -1, 64),
					strconv.FormatFloat(tickSize, 'f', -1, 64),
				)
			}
			if filter["filterType"].(string) == string(binance.SymbolFilterTypeLotSize) {
				if _, ok := filter["maxQty"]; !ok {
					continue
				}
				if _, ok := filter["minQty"]; !ok {
					continue
				}
				if _, ok := filter["stepSize"]; !ok {
					continue
				}
				maxQty, _ := strconv.ParseFloat(filter["maxQty"].(string), 64)
				minQty, _ := strconv.ParseFloat(filter["minQty"].(string), 64)
				stepSize, _ := strconv.ParseFloat(filter["stepSize"].(string), 64)
				filters["quote"] = fmt.Sprintf(
					"%s,%s,%s",
					strconv.FormatFloat(maxQty, 'f', -1, 64),
					strconv.FormatFloat(minQty, 'f', -1, 64),
					strconv.FormatFloat(stepSize, 'f', -1, 64),
				)
			}
		}
		if _, ok := filters["price"]; !ok {
			continue
		}
		if _, ok := filters["quote"]; !ok {
			continue
		}
		var entity models.Symbol
		result := r.Db.Where("symbol", item.Symbol).Take(&entity)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			entity = models.Symbol{
				ID:         xid.New().String(),
				Symbol:     item.Symbol,
				BaseAsset:  item.BaseAsset,
				QuoteAsset: item.QuoteAsset,
				Filters:    filters,
				IsSpot:     item.IsSpotTradingAllowed,
				IsMargin:   item.IsMarginTradingAllowed,
				Status:     item.Status,
			}
			r.Db.Create(&entity)
		} else {
			entity.Filters = filters
			entity.IsSpot = item.IsSpotTradingAllowed
			entity.IsMargin = item.IsMarginTradingAllowed
			entity.Status = item.Status
			r.Db.Model(&models.Symbol{ID: entity.ID}).Updates(entity)
		}
	}

	return nil
}

func (r *SymbolsRepository) Count() error {
	var count int64
	r.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Count(&count)
	r.Rdb.HMSet(
		r.Ctx,
		fmt.Sprintf("binance:symbols:count"),
		map[string]interface{}{
			"spot": count,
		},
	)

	return nil
}

func (r *SymbolsRepository) Price(symbol string) (float64, error) {
	fields := []string{
		"price",
		"timestamp",
	}
	data, _ := r.Rdb.HMGet(
		r.Ctx,
		fmt.Sprintf(
			"binance:spot:realtime:%s",
			symbol,
		),
		fields...,
	).Result()
	for i := 0; i < len(fields); i++ {
		if data[i] == nil {
			return 0, &SymbolsError{"price not exists"}
		}
	}

	timestamp := time.Now().Unix()
	price, _ := strconv.ParseFloat(data[0].(string), 64)
	lasttime, _ := strconv.ParseInt(data[1].(string), 10, 64)
	if lasttime < timestamp-30 {
		return 0, &SymbolsError{"price long time not freshed"}
	}

	return price, nil
}

func (r *SymbolsRepository) Filter(symbol string, price float64, amount float64) (float64, float64) {
	var entity models.Symbol
	result := r.Db.Select("filters").Where("symbol", symbol).Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, 0
	}
	var data []string
	data = strings.Split(entity.Filters["price"].(string), ",")
	maxPrice, _ := strconv.ParseFloat(data[0], 64)
	minPrice, _ := strconv.ParseFloat(data[1], 64)
	tickSize, _ := strconv.ParseFloat(data[2], 64)

	if price > maxPrice {
		return 0, 0
	}
	if price < minPrice {
		price = minPrice
	}
	price = math.Ceil(price/tickSize) / math.Ceil(1/tickSize)

	data = strings.Split(entity.Filters["quote"].(string), ",")
	maxQty, _ := strconv.ParseFloat(data[0], 64)
	minQty, _ := strconv.ParseFloat(data[1], 64)
	stepSize, _ := strconv.ParseFloat(data[2], 64)

	quantity := math.Ceil(amount/(price*stepSize)) / math.Ceil(1/stepSize)
	if quantity > maxQty {
		return 0, 0
	}
	if quantity < minQty {
		quantity = minQty
	}

	return price, quantity
}

func (r *SymbolsRepository) Context(symbol string) map[string]interface{} {
	day := time.Now().Format("0102")
	fields := []string{
		"r3",
		"r2",
		"r1",
		"s1",
		"s2",
		"s3",
		"profit_target",
		"stop_loss_point",
		"take_profit_price",
	}
	data, _ := r.Rdb.HMGet(
		r.Ctx,
		fmt.Sprintf(
			"binance:spot:indicators:%s:%s",
			symbol,
			day,
		),
		fields...,
	).Result()
	var context = make(map[string]interface{})
	for i := 0; i < len(fields); i++ {
		context[fields[i]] = data[i]
	}

	return context
}
