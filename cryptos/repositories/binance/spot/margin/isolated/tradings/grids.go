package tradings

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"strconv"
	models "taoniu.local/cryptos/models/binance/spot/margin/isolated/tradings"

	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"

	config "taoniu.local/cryptos/config/binance/spot"
	spotModels "taoniu.local/cryptos/models/binance/spot"
	marginModels "taoniu.local/cryptos/models/binance/spot/margin"
	//spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	//marginRepositories "taoniu.local/cryptos/repositories/binance/spot/margin"
	//isolatedRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
	//tradingviewRepositories "taoniu.local/cryptos/repositories/tradingview"
)

type MarginInterface interface {
	Isolated() IsolatedInterface
}

type IsolatedInterface interface {
	Account() AccountInterface
}

type AccountInterface interface {
	Balance(symbol string) (float64, float64, error)
}

type GridsRepository struct {
	Db               *gorm.DB
	Rdb              *redis.Client
	Ctx              context.Context
	MarginRepository MarginInterface
	//AccountRepository     *isolatedRepositories.AccountRepository
	//OrdersRepository      *marginRepositories.OrdersRepository
	//SymbolsRepository     *spotRepositories.SymbolsRepository
	//GridsRepository       *spotRepositories.GridsRepository
	//TradingviewRepository *tradingviewRepositories.AnalysisRepository
}

func (r *GridsRepository) Margin() MarginInterface {
	return r.MarginRepository
}

//func (r *GridsRepository) Account() *isolatedRepositories.AccountRepository {
//	if r.AccountRepository == nil {
//		r.AccountRepository = &isolatedRepositories.AccountRepository{
//			Rdb: r.Rdb,
//			Ctx: r.Ctx,
//		}
//	}
//	return r.AccountRepository
//}

//func (r *GridsRepository) Orders() *marginRepositories.OrdersRepository {
//	if r.OrdersRepository == nil {
//		r.OrdersRepository = &marginRepositories.OrdersRepository{
//			Db:  r.Db,
//			Rdb: r.Rdb,
//			Ctx: r.Ctx,
//		}
//	}
//	return r.OrdersRepository
//}

//func (r *GridsRepository) Symbols() *spotRepositories.SymbolsRepository {
//	if r.SymbolsRepository == nil {
//		r.SymbolsRepository = &spotRepositories.SymbolsRepository{
//			Db:  r.Db,
//			Rdb: r.Rdb,
//			Ctx: r.Ctx,
//		}
//	}
//	return r.SymbolsRepository
//}

//func (r *GridsRepository) Grids() *spotRepositories.GridsRepository {
//	if r.GridsRepository == nil {
//		r.GridsRepository = &spotRepositories.GridsRepository{
//			Db:  r.Db,
//			Rdb: r.Rdb,
//			Ctx: r.Ctx,
//		}
//	}
//	return r.GridsRepository
//}

//func (r *GridsRepository) Tradingview() *tradingviewRepositories.AnalysisRepository {
//	if r.TradingviewRepository == nil {
//		r.TradingviewRepository = &tradingviewRepositories.AnalysisRepository{
//			Db:  r.Db,
//			Rdb: r.Rdb,
//			Ctx: r.Ctx,
//		}
//	}
//	return r.TradingviewRepository
//}

func (r *GridsRepository) Count(conditions map[string]interface{}) int64 {
	var total int64
	query := r.Db.Model(&models.Grid{})
	if _, ok := conditions["symbols"]; ok {
		query.Where("symbol IN ?", conditions["symbols"].([]string))
	}
	query.Count(&total)

	return total
}

func (r *GridsRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Grid {
	var trades []*models.Grid
	query := r.Db.Select([]string{
		"id",
		"symbol",
		"buy_price",
		"sell_price",
		"status",
		"created_at",
	})
	if _, ok := conditions["symbols"]; ok {
		query.Where("symbol IN ?", conditions["symbols"].([]string))
	}
	query.Order("created_at desc")
	offset := (current - 1) * pageSize
	query.Offset(offset).Limit(pageSize).Find(&trades)

	return trades
}

func (r *GridsRepository) Flush(symbol string) error {
	//signal, _, err := r.Tradingview().Signal(symbol)
	//if err != nil {
	//	return err
	//}
	signal := 0
	if signal == 0 {
		return errors.New("tradingview no trading signal")
	}

	//price, err := r.Symbols().Price(symbol)
	//if err != nil {
	//	return err
	//}

	//grid, err := r.Grids().Filter(symbol, price)
	//if err != nil {
	//	return err
	//}
	//sellItems, err := r.FilterGrid(grid, price, signal)
	//if err != nil {
	//	return err
	//}

	//if signal == 1 {
	//	amount := 10 * math.Pow(2, float64(grid.Step-1))
	//	return r.Buy(grid, price, amount)
	//} else {
	//	return r.Sell(grid, sellItems)
	//}

	return nil
}

//func (r *GridsRepository) Buy(grid *spotModels.Grid, price float64, amount float64) error {
//	//balance, _, err := r.Account().Balance(grid.Symbol)
//	//if err != nil {
//	//	return err
//	//}
//	var err error
//	balance := 0.0
//
//	buyPrice, buyQuantity, _ := r.Symbols().Adjust(grid.Symbol, price, amount)
//	sellPrice := buyPrice * (1 + grid.TakeProfitPercent)
//	sellQuantity := buyQuantity * grid.TriggerPercent
//	sellPrice, sellQuantity, _ = r.Symbols().Adjust(grid.Symbol, sellPrice, sellPrice*sellQuantity)
//
//	buyAmount := buyPrice * buyQuantity
//
//	var buyOrderId int64 = 0
//	var status = 0
//	var remark = ""
//	if balance < buyAmount || grid.Balance < buyAmount {
//		status = 1
//	} else {
//		grid.Balance = grid.Balance - buyAmount
//		r.Db.Model(&spotModels.Grid{ID: grid.ID}).Updates(grid)
//	}
//
//	var entity *models.Grid
//	entity = &models.Grid{
//		ID:           xid.New().String(),
//		Symbol:       grid.Symbol,
//		GridID:       grid.ID,
//		BuyOrderId:   buyOrderId,
//		BuyPrice:     buyPrice,
//		BuyQuantity:  buyQuantity,
//		SellPrice:    sellPrice,
//		SellQuantity: sellQuantity,
//		Status:       status,
//		Remark:       remark,
//	}
//	r.Db.Create(entity)
//
//	if entity.Status == 0 {
//		buyOrderId, err = r.Order(grid.Symbol, binance.SideTypeBuy, price, buyQuantity)
//		if err != nil {
//			entity.Remark = err.Error()
//		} else {
//			entity.BuyOrderId = buyOrderId
//			entity.Status = 1
//		}
//		r.Db.Model(&models.Grid{ID: grid.ID}).Updates(entity)
//	}
//
//	return nil
//}
//
//func (r *GridsRepository) Sell(grid *spotModels.Grid, entities []*models.Grid) error {
//	for _, entity := range entities {
//		sellAmount := entity.SellPrice * entity.SellQuantity
//
//		var sellOrderId int64 = 0
//		var err error
//		var status = 2
//		var remark = entity.Remark
//		if entity.BuyOrderId == 0 {
//			status = 3
//		} else {
//			grid.Balance = grid.Balance + sellAmount
//			r.Db.Model(&spotModels.Grid{ID: grid.ID}).Updates(grid)
//		}
//		entity.SellOrderId = sellOrderId
//		entity.Status = status
//		entity.Remark = remark
//		r.Db.Model(&models.Grid{ID: entity.ID}).Updates(entity)
//
//		if entity.Status == 2 {
//			sellOrderId, err = r.Order(entity.Symbol, binance.SideTypeSell, entity.SellPrice, entity.SellQuantity)
//			if err != nil {
//				remark = err.Error()
//			} else {
//				entity.SellOrderId = sellOrderId
//				entity.Status = 3
//			}
//			r.Db.Model(&models.Grid{ID: entity.ID}).Updates(entity)
//		}
//	}
//
//	return nil
//}

func (r *GridsRepository) Update() error {
	var entities []*models.Grid
	r.Db.Where(
		"status IN ?",
		[]int64{0, 2},
	).Find(&entities)
	for _, entity := range entities {
		orderID := entity.BuyOrderId
		if entity.Status == 2 {
			orderID = entity.SellOrderId
		}

		var order marginModels.Order
		result := r.Db.Where(
			"symbol=? AND order_id=?",
			entity.Symbol,
			orderID,
		).Take(&order)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			continue
		}
		if order.Status == "NEW" || order.Status == "PARTIALLY_FILLED" {
			continue
		}

		status := 0
		if entity.Status == 0 {
			if order.Status != "FILLED" {
				status = 4
			} else {
				status = 1
			}
		}
		if entity.Status == 2 {
			if order.Status != "FILLED" {
				status = 5
			} else {
				status = 3
			}
		}
		entity.Status = status

		r.Db.Model(&marginModels.Order{ID: entity.ID}).Updates(entity)
	}

	return nil
}

func (r *GridsRepository) FilterGrid(grid *spotModels.Grid, price float64, signal int64) ([]*models.Grid, error) {
	var entryPrice float64
	var takePrice float64
	var entities []*models.Grid
	r.Db.Where(
		"grid_id=? AND status IN ?",
		grid.ID,
		[]int64{0, 1},
	).Find(&entities)
	var sellItems []*models.Grid
	for _, entity := range entities {
		if entryPrice == 0 || entryPrice > entity.BuyPrice*(1-grid.TakeProfitPercent) {
			entryPrice = entity.BuyPrice / (1 + grid.TakeProfitPercent)
		}
		if takePrice == 0 || (entity.Status == 1 && takePrice < entity.SellPrice) {
			takePrice = entity.SellPrice
		}
		if entity.Status == 1 && price > entity.SellPrice {
			sellItems = append(sellItems, entity)
		}
	}
	if signal == 1 && entryPrice > 0 && price > entryPrice {
		return nil, errors.New("buy price too high")
	}
	if signal == 2 && (takePrice == 0 || price < takePrice) {
		return nil, errors.New("sell price too low")
	}
	if signal == 2 && len(sellItems) == 0 {
		return nil, errors.New("nothing sell")
	}

	return sellItems, nil
}

func (r *GridsRepository) Order(symbol string, side binance.SideType, price float64, quantity float64) (int64, error) {
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
	//r.Orders().Flush(symbol, result.OrderID, true)

	return result.OrderID, nil
}
