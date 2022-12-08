package margin

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"

	config "taoniu.local/cryptos/config/binance/spot"
	models "taoniu.local/cryptos/models/binance/spot/margin"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type OrdersRepository struct {
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	SymbolsRepository *repositories.SymbolsRepository
}

func (r *OrdersRepository) Symbols() *repositories.SymbolsRepository {
	if r.SymbolsRepository == nil {
		r.SymbolsRepository = &repositories.SymbolsRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.SymbolsRepository
}

func (r *OrdersRepository) Count(conditions map[string]interface{}) int64 {
	var total int64
	query := r.Db.Model(&models.Order{})
	if _, ok := conditions["symbols"]; ok {
		query.Where("symbol IN ?", conditions["symbols"].([]string))
	}
	query.Where("status IN ?", []string{"NEW"})
	query.Count(&total)
	return total
}

func (r *OrdersRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Order {
	offset := (current - 1) * pageSize

	var orders []*models.Order
	query := r.Db.Select([]string{
		"id",
		"symbol",
		"side",
		"price",
		"quantity",
		"status",
		"created_at",
		"updated_at",
	})
	if _, ok := conditions["symbols"]; ok {
		query.Where("symbol IN ?", conditions["symbols"].([]string))
	}
	query.Where("status IN ?", []string{"NEW"})
	query.Order("created_at desc")
	query.Offset(offset).Limit(pageSize).Find(&orders)
	return orders
}

func (r *OrdersRepository) Create(
	symbol string,
	side string,
	price float64,
	amount float64,
) (int64, error) {
	price, quantity, err := r.Symbols().Adjust(symbol, price, amount)
	if err != nil {
		return 0, err
	}
	client := binance.NewClient(config.TRADE_API_KEY, config.TRADE_SECRET_KEY)
	result, err := client.NewCreateMarginOrderService().Symbol(
		symbol,
	).Side(
		binance.SideType(side),
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
	r.Flush(symbol, result.OrderID, true)

	return result.OrderID, nil
}

func (r *OrdersRepository) Cancel(id string) error {
	var order models.Order
	result := r.Db.Where("id", id).Find(&order)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	client := binance.NewClient(config.TRADE_API_KEY, config.TRADE_SECRET_KEY)
	response, err := client.NewCancelMarginOrderService().Symbol(order.Symbol).OrderID(order.OrderID).IsIsolated(order.IsIsolated).Do(r.Ctx)
	if err != nil {
		return err
	}
	order.Status = string(response.Status)
	r.Db.Model(&models.Order{ID: order.ID}).Updates(order)
	return nil
}

func (r *OrdersRepository) Flush(symbol string, orderId int64, isIsolated bool) error {
	client := binance.NewClient(config.ACCOUNT_API_KEY, config.ACCOUNT_SECRET_KEY)
	order, err := client.NewGetMarginOrderService().Symbol(symbol).OrderID(orderId).IsIsolated(isIsolated).Do(r.Ctx)
	if err != nil {
		return err
	}
	r.Save(order)

	var isolated int8
	if isIsolated {
		isolated = 1
	} else {
		isolated = 0
	}
	r.Rdb.SRem(
		r.Ctx,
		"binance:spot:margin:orders:flush",
		fmt.Sprintf("%s,%d,%d", symbol, orderId, isolated),
	).Result()

	return nil
}

func (r *OrdersRepository) Sync(symbol string, isIsolated bool, limit int) error {
	yestoday := time.Now().Unix() - 86400
	client := binance.NewClient(config.ACCOUNT_API_KEY, config.ACCOUNT_SECRET_KEY)
	orders, err := client.NewListMarginOrdersService().Symbol(
		symbol,
	).IsIsolated(
		isIsolated,
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
		"is_isolated",
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
		r.Flush(order.Symbol, order.OrderID, order.IsIsolated)
	}
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
			Type:             string(order.Type),
			Side:             string(order.Side),
			Price:            price,
			StopPrice:        stopPrice,
			Quantity:         quantity,
			ExecutedQuantity: executedQuantity,
			OpenTime:         order.Time,
			UpdateTime:       order.UpdateTime,
			IsIsolated:       order.IsIsolated,
			Status:           string(order.Status),
			Remark:           "",
		}
		r.Db.Create(&entity)
	} else {
		entity.ExecutedQuantity = executedQuantity
		entity.UpdateTime = order.UpdateTime
		entity.Status = string(order.Status)
		r.Db.Model(&models.Order{ID: entity.ID}).Updates(entity)
	}

	return nil
}
