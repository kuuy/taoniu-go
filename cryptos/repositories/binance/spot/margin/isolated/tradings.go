package isolated

import (
	"context"
	"gorm.io/gorm"
	"strconv"

	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"

	config "taoniu.local/cryptos/config/binance"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin"
)

type TradingsRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *TradingsRepository) Scalping() error {
	return nil
}

func (r *TradingsRepository) Trade(symbol string, side binance.SideType, price float64, quantity float64) (int64, error) {
	if quantity == 0 {
		return 0, nil
	} else {
		return 0, nil
	}
	client := binance.NewClient(config.TRADE_API_KEY, config.TRADE_SECRET_KEY)
	result, err := client.NewCreateMarginOrderService().Symbol(
		symbol,
	).Side(
		side,
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
