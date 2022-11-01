package tradings

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/adshao/go-binance/v2"
	service "github.com/adshao/go-binance/v2/futures"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
	"github.com/urfave/cli/v2"

	pool "taoniu.local/cryptos/common"
	models "taoniu.local/cryptos/models/binance/futures"
	repositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type GridsHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.GridsRepository
}

func NewGridsCommand() *cli.Command {
	var h GridsHandler
	return &cli.Command{
		Name:  "grids",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = GridsHandler{
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.GridsRepository{
				Db:  pool.NewDB(),
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			return nil
		},
		Subcommands: []*cli.Command{
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

func (h *GridsHandler) flush() error {
	log.Println("futures tradings grids flush...")
	return nil
}

func (h *GridsHandler) saveTrade(db *gorm.DB, order *service.CreateOrderResponse, remark string) {
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
			OpenTime:         order.UpdateTime,
			UpdateTime:       order.UpdateTime,
			WorkingType:      fmt.Sprint(order.WorkingType),
			PriceProtect:     order.PriceProtect,
			ReduceOnly:       order.ReduceOnly,
			ClosePosition:    order.ClosePosition,
			Status:           fmt.Sprint(order.Status),
			Remark:           remark,
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

type Trade struct {
	Signal         int64
	Price          float64
	Quantity       float64
	StopPrice      float64
	StrategyID     string
	StrategyRemark string
	Skip           bool
}

func (h *GridsHandler) takeProfit() error {
	ctx := context.Background()
	rdb := pool.NewRedis()
	defer rdb.Close()
	db := pool.NewDB()

	log.Println("take profit")

	items, err := rdb.HGetAll(ctx, "binance:futures:orders:take_profit").Result()
	if err != nil {
		log.Println("take profit", err)
		return err
	}

	apiKey := "1ezcGDyXqV6fHPqockPILt5KMiXzUr4feoPMNmmqsmWakKJyK32GOvnL9LNoBg8n"
	secretKey := "AXHKOh04ndgWkQlwc8Ro4m6ZSBFudNno8b2zlLKtSwzy9B6cZbvsTyyWynzNMvCw"
	client := binance.NewFuturesClient(apiKey, secretKey)

	log.Println("take profit", items)
	for field, value := range items {
		x := strings.Split(field, ",")
		y := strings.Split(value, ",")
		symbol := x[0]
		orderID, _ := strconv.ParseInt(x[1], 10, 64)
		signal, _ := strconv.ParseInt(y[0], 10, 64)
		stopPrice, _ := strconv.ParseFloat(y[2], 64)
		var order models.Order
		result := db.First(&order, "symbol=? AND order_id=?", symbol, orderID)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			continue
		}
		if order.Status == "CANCELED" || order.Status == "EXPIRED" {
			rdb.HDel(ctx, "binance:futures:orders:take_profit", field)
		}
		if order.Status != "FILLED" {
			continue
		}

		positionSide := service.PositionSideTypeLong
		side := service.SideTypeSell
		if signal == 2 {
			positionSide = service.PositionSideTypeShort
			side = service.SideTypeBuy
		}

		_, err := client.NewCreateOrderService().Symbol(
			symbol,
		).PositionSide(
			positionSide,
		).Side(
			side,
		).Type(
			service.OrderTypeTakeProfitMarket,
		).Quantity(
			fmt.Sprint(order.Quantity),
		).NewClientOrderID(
			fmt.Sprintf("profit-%d", orderID),
		).WorkingType(
			service.WorkingTypeMarkPrice,
		).StopPrice(
			fmt.Sprint(stopPrice),
		).TimeInForce(
			service.TimeInForceTypeGTC,
		).NewOrderResponseType(
			service.NewOrderRespTypeACK,
		).Do(ctx)
		if err == nil {
			rdb.HDel(
				ctx,
				"binance:futures:orders:take_profit",
				fmt.Sprintf("%s,%d", symbol, orderID),
			)
		}
		if err != nil {
			log.Println("order submit failed", err)
		}

		log.Println(signal, order.Quantity, stopPrice)
	}

	return nil
}

func (h *GridsHandler) submit() error {

	ctx := context.Background()
	rdb := pool.NewRedis()
	defer rdb.Close()
	db := pool.NewDB()

	mutex := pool.NewMutex(
		rdb,
		ctx,
		"lock:binance:futures:trades:submit",
	)
	if mutex.Lock(30 * time.Second) {
		return nil
	}
	defer mutex.Unlock()

	location, _ := time.LoadLocation("Asia/Shanghai")
	expireTime := time.Now().In(location).Add(-20 * time.Minute)

	var strategies []models.Strategy
	db.Where("created_at > ?", expireTime).Order("created_at desc").Find(&strategies)

	trades := make(map[string]Trade)
	for _, strategy := range strategies {
		symbol := strategy.Symbol
		signal := strategy.Signal
		price := strategy.Price
		if item, ok := trades[symbol]; ok {
			if item.Skip {
				continue
			}
		}
		//if strategy.Volume > 0 {
		//	trades[symbol] = Trade{
		//		Skip: true,
		//	}
		//	continue
		//}

		timestamp := time.Now().Unix()
		realtime, _ := rdb.HMGet(
			ctx,
			fmt.Sprintf("binance:futures:realtime:%s", symbol),
			"r3",
			"r2",
			"r1",
			"s1",
			"s2",
			"s3",
			"quantity_step",
			"ticket_step",
			"timestamp",
		).Result()
		if realtime[0] == nil || realtime[6] == nil {
			trades[symbol] = Trade{
				Skip: true,
			}
			continue
		}
		r3, _ := strconv.ParseFloat(fmt.Sprint(realtime[0]), 64)
		r2, _ := strconv.ParseFloat(fmt.Sprint(realtime[1]), 64)
		r1, _ := strconv.ParseFloat(fmt.Sprint(realtime[2]), 64)
		s1, _ := strconv.ParseFloat(fmt.Sprint(realtime[3]), 64)
		s2, _ := strconv.ParseFloat(fmt.Sprint(realtime[4]), 64)
		s3, _ := strconv.ParseFloat(fmt.Sprint(realtime[5]), 64)
		quantityStep, _ := strconv.ParseFloat(fmt.Sprint(realtime[6]), 64)
		ticketStep, _ := strconv.ParseFloat(fmt.Sprint(realtime[7]), 64)
		lasttime, _ := strconv.ParseInt(fmt.Sprint(realtime[8]), 10, 64)

		p1 := r1 / s1
		p2 := r2 / s2
		p3 := r3 / s3
		if p1 < 1.01 || p2 < 1.02 || p3 < 1.03 {
			log.Println("submit invalid", symbol, p1, p2, p3)
			trades[symbol] = Trade{
				Skip: true,
			}
			continue
		}

		if timestamp-lasttime > 5 {
			trades[symbol] = Trade{
				Skip: true,
			}
			continue
		}
		position, _ := rdb.HGetAll(
			ctx,
			fmt.Sprintf("binance:futures:positions:%s", symbol),
		).Result()
		if _, ok := position["symbol"]; !ok {
			trades[symbol] = Trade{
				Skip: true,
			}
			continue
		}
		tradePrice := price
		stopPrice := price
		if signal == 1 && price > s1 && s1 > 0 {
			tradePrice = s1
		}
		if signal == 2 && price < r1 {
			tradePrice = r1
		}
		if strategy.Signal == 1 {
			entryPrice, _ := strconv.ParseFloat(
				fmt.Sprint(position["long_entry_price"]),
				64,
			)
			if tradePrice > entryPrice || entryPrice == 0.0 {
				continue
			}
			stopPrice = r1

			tradePrice = math.Floor(tradePrice/ticketStep) / math.Ceil(1/ticketStep)
			stopPrice = math.Ceil(stopPrice/ticketStep) / math.Ceil(1/ticketStep)
			log.Println("long price:", tradePrice, stopPrice)
		}
		if strategy.Signal == 2 {
			entryPrice, _ := strconv.ParseFloat(
				fmt.Sprint(position["short_entry_price"]),
				64,
			)
			if tradePrice < entryPrice || entryPrice == 0.0 {
				continue
			}
			stopPrice = s1

			tradePrice = math.Ceil(tradePrice/ticketStep) / math.Ceil(1/ticketStep)
			stopPrice = math.Floor(stopPrice/ticketStep) / math.Ceil(1/ticketStep)
			log.Println("short price:", tradePrice, stopPrice)
		}

		quantity := math.Ceil(20/(tradePrice*quantityStep)) / math.Ceil(1/quantityStep)
		if quantity > 10 {
			quantity = math.Floor(quantity)
		}

		trades[symbol] = Trade{
			Signal:         signal,
			Price:          tradePrice,
			Quantity:       quantity,
			StopPrice:      stopPrice,
			StrategyID:     strategy.ID,
			StrategyRemark: strategy.Remark,
			Skip:           false,
		}
	}

	apiKey := "1ezcGDyXqV6fHPqockPILt5KMiXzUr4feoPMNmmqsmWakKJyK32GOvnL9LNoBg8n"
	secretKey := "AXHKOh04ndgWkQlwc8Ro4m6ZSBFudNno8b2zlLKtSwzy9B6cZbvsTyyWynzNMvCw"
	client := binance.NewFuturesClient(apiKey, secretKey)

	for symbol, trade := range trades {
		if trade.Skip {
			continue
		}

		signal := trade.Signal
		quantity := trade.Quantity
		tradePrice := trade.Price
		stopPrice := trade.StopPrice
		strategyID := trade.StrategyID
		strategyRemark := trade.StrategyRemark

		log.Println("price:", tradePrice, stopPrice)

		positionSide := service.PositionSideTypeLong
		side := service.SideTypeBuy
		if signal == 2 {
			positionSide = service.PositionSideTypeShort
			side = service.SideTypeSell
		}
		result, err := client.NewCreateOrderService().Symbol(
			symbol,
		).PositionSide(
			positionSide,
		).Side(
			side,
		).Type(
			service.OrderTypeLimit,
		).Quantity(
			fmt.Sprint(quantity),
		).Price(
			fmt.Sprint(tradePrice),
		).NewClientOrderID(
			fmt.Sprintf("taoniu-%s", strategyID),
		).WorkingType(
			service.WorkingTypeContractPrice,
		).PriceProtect(
			true,
		).TimeInForce(
			service.TimeInForceTypeGTC,
		).NewOrderResponseType(
			service.NewOrderRespTypeRESULT,
		).Do(ctx)
		if err != nil {
			log.Println("order submit failed", err)
			return err
		}
		orderID := result.OrderID

		remark := fmt.Sprintf(
			"%s o:%d p:%f q:%f a:%f",
			strategyRemark,
			orderID,
			tradePrice,
			quantity,
			stopPrice,
		)
		db.Model(&models.Strategy{ID: strategyID}).Updates(
			map[string]interface{}{
				"volume": 20,
				"remark": remark,
			},
		)

		rdb.HSet(
			ctx,
			"binance:futures:orders:take_profit",
			strings.Join([]string{
				symbol,
				fmt.Sprint(orderID),
			}, ","),
			strings.Join([]string{
				fmt.Sprint(signal),
				fmt.Sprint(quantity),
				fmt.Sprint(stopPrice),
			}, ","),
		)

		h.saveTrade(
			db,
			result,
			fmt.Sprintf(
				"s:%s p:%f q:%f s:%f",
				strategyID,
				tradePrice,
				quantity,
				stopPrice,
			),
		)
	}

	return nil
}
