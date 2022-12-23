package futures

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/rs/xid"
	"github.com/urfave/cli/v2"

	pool "taoniu.local/cryptos/common"
	models "taoniu.local/cryptos/models/binance/futures"
	repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type OrdersHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.OrdersRepository
}

func NewOrdersCommand() *cli.Command {
	var h OrdersHandler
	return &cli.Command{
		Name:  "orders",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = OrdersHandler{
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.OrdersRepository{
				Db:  pool.NewDB(),
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "open",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.open(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "flush",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.flush(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *OrdersHandler) saveOrder(db *gorm.DB, order *futures.Order) {
	symbol := order.Symbol
	orderID := order.OrderID

	price, _ := strconv.ParseFloat(order.Price, 64)
	avgPrice, _ := strconv.ParseFloat(order.AvgPrice, 64)
	activatePrice, _ := strconv.ParseFloat(order.ActivatePrice, 64)
	stopPrice, _ := strconv.ParseFloat(order.StopPrice, 64)
	quantity, _ := strconv.ParseFloat(order.OrigQuantity, 64)
	executedQuantity, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)

	var entity models.Order
	result := db.Where(
		"symbol=? AND order_id=?",
		symbol,
		orderID,
	).First(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		entity = models.Order{
			ID:               xid.New().String(),
			Symbol:           symbol,
			OrderID:          orderID,
			Type:             fmt.Sprint(order.Type),
			PositionSide:     fmt.Sprint(order.PositionSide),
			Side:             fmt.Sprint(order.Side),
			Price:            price,
			AvgPrice:         avgPrice,
			ActivatePrice:    activatePrice,
			StopPrice:        stopPrice,
			Quantity:         quantity,
			ExecutedQuantity: executedQuantity,
			OpenTime:         order.Time,
			UpdateTime:       order.UpdateTime,
			WorkingType:      fmt.Sprint(order.WorkingType),
			PriceProtect:     order.PriceProtect,
			ReduceOnly:       order.ReduceOnly,
			ClosePosition:    order.ClosePosition,
			Status:           fmt.Sprint(order.Status),
			Remark:           "",
		}
		db.Create(&entity)
	} else {
		entity.AvgPrice = avgPrice
		entity.ExecutedQuantity = executedQuantity
		entity.UpdateTime = order.UpdateTime
		entity.Status = fmt.Sprint(order.Status)
		db.Model(&models.Order{ID: entity.ID}).Updates(entity)
	}
}

func (h *OrdersHandler) open() error {
	ctx := context.Background()
	rdb := pool.NewRedis()
	defer rdb.Close()
	db := pool.NewDB()

	mutex := pool.NewMutex(
		rdb,
		ctx,
		"lock:binance:futures:orders:opened",
	)
	if mutex.Lock(5 * time.Second) {
		return nil
	}
	defer mutex.Unlock()

	apiKey := "HWRmhMZsd1JYEDgmrXfRYlMIgli6jA2g40Kk3kRAzmb35oPO713IcRFxSTvFwJ2u"
	secretKey := "jbeKzWMT66ovrD14BzZGH48Q9vER4gi7c5Hg4iPzgl78GpgVgIVV19abin9XNj6b"

	client := binance.NewFuturesClient(apiKey, secretKey)

	symbols, _ := rdb.SMembers(ctx, "binance:futures:websocket:symbols").Result()
	for _, symbol := range symbols {
		list, err := client.NewListOpenOrdersService().Symbol(symbol).Do(ctx)
		if err != nil {
			log.Fatalln("api failed", err)
			return nil
		}
		orderIDs := []int64{}
		for _, order := range list {
			orderID := order.OrderID
			orderIDs = append(orderIDs, orderID)
			h.saveOrder(db, order)
		}

		if len(orderIDs) == 0 {
			db.Model(&models.Order{}).Where(
				"symbol = ? AND status IN ?",
				symbol,
				[]string{"NEW", "PARTIALLY_FILLED"},
			).Update("status", "UNKNOW")
		} else {
			db.Model(&models.Order{}).Where(
				"symbol = ? AND status IN ? AND order_id NOT IN ?",
				symbol,
				[]string{"NEW", "PARTIALLY_FILLED"},
				orderIDs,
			).Update("status", "UNKNOW")
		}
	}

	db.Model(&models.Order{}).Where(
		"symbol NOT IN ? AND status IN ?",
		symbols,
		[]string{"NEW", "PARTIALLY_FILLED"},
	).Update("status", "UNKNOW")

	var entity models.Order
	result := db.Model(&models.Order{}).Where("status", "UNKNOW").First(&entity)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		order, err := client.NewGetOrderService().Symbol(entity.Symbol).OrderID(entity.OrderID).Do(ctx)
		if err == nil {
			h.saveOrder(db, order)
		}
	}

	list, _ := rdb.SMembers(ctx, "binance:futures:websocket:orders").Result()
	log.Println(list)
	for _, item := range list {
		data := strings.Split(item, ",")
		symbol := data[0]
		orderID, _ := strconv.ParseInt(data[1], 10, 64)
		result := db.Model(&models.Order{}).Where(
			"symbol=? AND order_id=?",
			symbol,
			orderID,
		).First(&entity)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			order, err := client.NewGetOrderService().Symbol(symbol).OrderID(orderID).Do(ctx)
			if err == nil {
				h.saveOrder(db, order)
				rdb.SRem(ctx, "binance:futures:websocket:orders", item)
			}
		} else {
			rdb.SRem(ctx, "binance:futures:websocket:orders", item)
		}
	}

	return nil
}

func (h *OrdersHandler) flush() error {
	log.Println("margin orders flush...")
	orders, err := h.Rdb.SMembers(h.Ctx, "binance:futures:orders:flush").Result()
	if err != nil {
		return nil
	}
	for _, order := range orders {
		data := strings.Split(order, ",")
		symbol := data[0]
		orderID, _ := strconv.ParseInt(data[1], 10, 64)
		h.Repository.Flush(symbol, orderID)
	}

	return nil
}
