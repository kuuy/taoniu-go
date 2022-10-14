package margin

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	repositories "taoniu.local/cryptos/repositories/binance/spot/margin"
)

type OrdersTask struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.OrdersRepository
}

func (t *OrdersTask) Flush() error {
	orders, err := t.Rdb.SMembers(t.Ctx, "binance:spot:margin:orders:flush").Result()
	if err != nil {
		return nil
	}
	for _, order := range orders {
		data := strings.Split(order, ",")
		symbol := data[0]
		orderID, _ := strconv.ParseInt(data[1], 10, 64)
		isIsolated, _ := strconv.ParseBool(data[2])
		t.Repository.Flush(symbol, orderID, isIsolated)
	}
	return nil
}

func (t *OrdersTask) Fix() error {
	t.Repository.Fix(time.Now().Add(-30*time.Minute), 20)
	return nil
}
