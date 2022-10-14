package isolated

import (
	"context"
	"gorm.io/gorm"

	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"

	config "taoniu.local/cryptos/config/binance"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin"
)

type OrdersRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *OrdersRepository) Open(symbol string) error {
	client := binance.NewClient(config.ACCOUNT_API_KEY, config.ACCOUNT_SECRET_KEY)
	orders, err := client.NewListMarginOpenOrdersService().Symbol(symbol).IsIsolated(true).Do(r.Ctx)
	if err != nil {
		return err
	}
	repository := repositories.OrdersRepository{
		Db:  r.Db,
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
	for _, order := range orders {
		repository.Save(order)
	}

	return nil
}

func (r *OrdersRepository) Sync(symbol string, limit int) error {
	repository := repositories.OrdersRepository{
		Db:  r.Db,
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
	repository.Sync(symbol, true, limit)

	return nil
}
