package spot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
	"gorm.io/gorm"

	config "taoniu.local/cryptos/config/binance"
	models "taoniu.local/cryptos/models/binance/spot"
)

type OrdersRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *OrdersRepository) Open(symbol string) error {
	client := binance.NewClient(config.ACCOUNT_API_KEY, config.ACCOUNT_SECRET_KEY)
	orders, err := client.NewListOpenOrdersService().Symbol(symbol).Do(r.Ctx)
	if err != nil {
		return err
	}
	for _, order := range orders {
		r.Save(order)
	}

	return nil
}

func (r *OrdersRepository) Sync(symbol string, limit int) error {
	yestoday := time.Now().Unix() - 86400
	client := binance.NewClient(config.ACCOUNT_API_KEY, config.ACCOUNT_SECRET_KEY)
	orders, err := client.NewListOrdersService().Symbol(
		symbol,
	).StartTime(
		yestoday * 1000,
	).Limit(
		limit,
	).Do(r.Ctx)
	if err != nil {
		return err
	}
	for _, order := range orders {
		r.Save(order)
	}

	return nil
}

func (r *OrdersRepository) Fix(time time.Time, limit int) error {
	var orders []*models.Order
	r.Db.Select([]string{
		"symbol",
		"order_id",
	}).Where(
		"updated_at < ? AND status IN ?",
		time,
		[]string{
			"NEW",
			"PARTIALLY_FILLED",
		},
	).Order(
		"updated_at asc",
	).Limit(
		limit,
	).Find(&orders)
	for _, order := range orders {
		r.Flush(order.Symbol, order.OrderID)
	}

	return nil
}

func (r *OrdersRepository) Flush(symbol string, orderId int64) error {
	client := binance.NewClient(config.ACCOUNT_API_KEY, config.ACCOUNT_SECRET_KEY)
	order, err := client.NewGetOrderService().Symbol(symbol).OrderID(orderId).Do(r.Ctx)
	if err != nil {
		return err
	}
	r.Save(order)

	r.Rdb.SRem(
		r.Ctx,
		"binance:spot:orders:flush",
		fmt.Sprintf("%s,%d,%d", symbol, orderId),
	).Result()

	return nil
}

func (r *OrdersRepository) Save(order *binance.Order) error {
	symbol := order.Symbol
	orderID := order.OrderID

	price, _ := strconv.ParseFloat(order.Price, 64)
	stopPrice, _ := strconv.ParseFloat(order.StopPrice, 64)
	quantity, _ := strconv.ParseFloat(order.OrigQuantity, 64)
	executedQuantity, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)

	var entity models.Order
	result := r.Db.Where(
		"symbol=? AND order_id=?",
		symbol,
		orderID,
	).Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		entity = models.Order{
			ID:               xid.New().String(),
			Symbol:           symbol,
			OrderID:          orderID,
			Type:             fmt.Sprint(order.Type),
			Side:             fmt.Sprint(order.Side),
			Price:            price,
			StopPrice:        stopPrice,
			Quantity:         quantity,
			ExecutedQuantity: executedQuantity,
			OpenTime:         order.Time,
			UpdateTime:       order.UpdateTime,
			Status:           fmt.Sprint(order.Status),
			Remark:           "",
		}
		r.Db.Create(&entity)
	} else {
		entity.ExecutedQuantity = executedQuantity
		entity.UpdateTime = order.UpdateTime
		entity.Status = fmt.Sprint(order.Status)
		r.Db.Model(&models.Order{ID: entity.ID}).Updates(entity)
	}

	return nil
}
