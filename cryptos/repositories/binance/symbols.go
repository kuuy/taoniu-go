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
	"strconv"
	config "taoniu.local/cryptos/config/binance"
	models "taoniu.local/cryptos/models/binance"
)

type SymbolsRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
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
		result := r.Db.Where("symbol", item.Symbol).First(&entity)
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
