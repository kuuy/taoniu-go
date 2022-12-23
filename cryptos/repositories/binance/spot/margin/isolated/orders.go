package isolated

import (
	"context"
	"gorm.io/gorm"

	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"

	config "taoniu.local/cryptos/config/binance/spot"
)

type OrdersInterface interface {
	Sync(symbol string, isIsolated bool, limit int) error
	Save(order *binance.Order) error
}

type OrdersRepository struct {
	Db     *gorm.DB
	Rdb    *redis.Client
	Ctx    context.Context
	Parent OrdersInterface
}

func (r *OrdersRepository) Open(symbol string) error {
	client := binance.NewClient(config.ACCOUNT_API_KEY, config.ACCOUNT_SECRET_KEY)
	orders, err := client.NewListMarginOpenOrdersService().Symbol(symbol).IsIsolated(true).Do(r.Ctx)
	if err != nil {
		return err
	}
	for _, order := range orders {
		r.Parent.Save(order)
	}
	return nil
}

func (r *OrdersRepository) Sync(symbol string, limit int) error {
	return r.Parent.Sync(symbol, true, limit)
}
