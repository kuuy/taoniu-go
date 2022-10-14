package isolated

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"log"
	"math"
	"strconv"
	"strings"
	config "taoniu.local/cryptos/config/binance"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin"
	"time"

	"gorm.io/gorm"

	"github.com/go-redis/redis/v8"

	models "taoniu.local/cryptos/models/binance/spot"
)

type TradingsRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *TradingsRepository) Scalping() error {
	buys, sells := r.Signals()
	r.Buys(buys)
	r.Sells(sells)
	return nil
}

func (r *TradingsRepository) Signals() (map[string]interface{}, map[string]interface{}) {
	timestamp := time.Now().Unix() - 86400
	var strategies []*models.Strategy
	r.Db.Select([]string{
		"symbol",
		"indicator",
		"price",
		"signal",
	}).Where(
		"indicator in ? AND duration = ? AND timestamp > ?",
		[]string{
			"kdj",
			"bbands",
			"ha_zlema",
		},
		"1d",
		timestamp,
	).Order(
		"timestamp desc",
	).Find(&strategies)
	var buys = make(map[string]interface{})
	var sells = make(map[string]interface{})
	for _, strategy := range strategies {
		if _, ok := buys[strategy.Indicator]; strategy.Signal == 1 && !ok {
			buys[strategy.Indicator] = make(map[string]float64)
		}
		if strategy.Signal == 1 {
			buys[strategy.Indicator].(map[string]float64)[strategy.Symbol] = strategy.Price
		}
		if _, ok := sells[strategy.Indicator]; strategy.Signal == 2 && !ok {
			sells[strategy.Indicator] = make(map[string]float64)
		}
		if strategy.Signal == 2 {
			sells[strategy.Indicator].(map[string]float64)[strategy.Symbol] = strategy.Price
		}
	}
	return buys, sells
}

func (r *TradingsRepository) Buy(symbol string, price float64, amount float64) (int64, error) {
	val, err := r.Rdb.HGet(
		r.Ctx,
		fmt.Sprintf(
			"binance:spot:margin:isolated:balances:%s",
			symbol,
		),
		"quote_free",
	).Result()
	if err != nil {
		return 0, err
	}
	balance, _ := strconv.ParseFloat(val, 64)
	if balance < amount {
		return 0, nil
	}
	price, quantity := r.Filter(symbol, price, amount)
	if quantity == 0 {
		return 0, nil
	}
	client := binance.NewClient(config.TRADE_API_KEY, config.TRADE_SECRET_KEY)
	result, err := client.NewCreateMarginOrderService().Symbol(
		symbol,
	).Side(
		binance.SideTypeBuy,
	).Type(
		binance.OrderTypeLimit,
	).Price(
		strconv.FormatFloat(price, 'f', -1, 64),
	).Quantity(
		strconv.FormatFloat(quantity, 'f', -1, 64),
	).IsIsolated(
		true,
	).TimeInForce(
		binance.TimeInForceTypeGTC,
	).NewOrderRespType(
		binance.NewOrderRespTypeRESULT,
	).Do(r.Ctx)
	if err != nil {
		return 0, err
	}
	repository := repositories.OrdersRepository{
		Db:  r.Db,
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
	repository.Flush(symbol, result.OrderID, true)

	return result.OrderID, nil
}

func (r *TradingsRepository) Filter(symbol string, price float64, amount float64) (float64, float64) {
	filters, err := r.Rdb.HGetAll(r.Ctx, fmt.Sprintf("binance:spot:symbols:filters:%s", symbol)).Result()
	if err != nil {
		return 0, 0
	}
	var data []string
	data = strings.Split(filters["price"], ",")
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

	data = strings.Split(filters["quote"], ",")
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

func (r *TradingsRepository) Sell(symbol string, price float64, amount float64) error {
	log.Println("sell:", symbol, price, amount)
	return nil
}

func (r *TradingsRepository) Buys(buys map[string]interface{}) error {
	if _, ok := buys["kdj"]; !ok {
		return nil
	}
	for symbol, price := range buys["kdj"].(map[string]float64) {
		amount := 10.0
		if _, ok := buys["bbands"]; ok {
			if p, ok := buys["bbands"].(map[string]float64)[symbol]; ok {
				if p < price {
					price = p
				}
				amount += 10
			}
		}
		if _, ok := buys["ha_zlema"]; ok {
			if p, ok := buys["ha_zlema"].(map[string]float64)[symbol]; ok {
				if p < price {
					price = p
				}
				amount += 5
			}
		}
		exists, _ := r.Rdb.Exists(
			r.Ctx, fmt.Sprintf("binance:spot:margin:isolated:balances:%s", symbol),
		).Result()
		if exists != 0 {
			r.Buy(symbol, price, amount)
		}
	}

	return nil
}

func (r *TradingsRepository) Sells(sells map[string]interface{}) error {
	if _, ok := sells["kdj"]; !ok {
		return nil
	}
	for symbol, price := range sells["kdj"].(map[string]float64) {
		amount := 10.0
		if _, ok := sells["bbands"]; ok {
			if p, ok := sells["bbands"].(map[string]float64)[symbol]; ok {
				if p > price {
					price = p
				}
				amount += 10
			}
		}
		if _, ok := sells["ha_zlema"]; ok {
			if p, ok := sells["ha_zlema"].(map[string]float64)[symbol]; ok {
				if p > price {
					price = p
				}
				amount += 5
			}
		}
		exists, _ := r.Rdb.Exists(
			r.Ctx, fmt.Sprintf("binance:spot:margin:isolated:balances:%s", symbol),
		).Result()
		if exists != 0 {
			r.Sell(symbol, price, amount)
		}
	}

	return nil
}
