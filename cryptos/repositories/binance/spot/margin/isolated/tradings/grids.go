package tradings

import (
	"context"
	"errors"
	"math"
	"strconv"
	config "taoniu.local/cryptos/config/binance/spot"

	"gorm.io/gorm"

	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"

	spotModels "taoniu.local/cryptos/models/binance/spot"
	marginModels "taoniu.local/cryptos/models/binance/spot/margin"
	models "taoniu.local/cryptos/models/binance/spot/margin/isolated"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	marginRepositories "taoniu.local/cryptos/repositories/binance/spot/margin"
	isolatedRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
	tradingviewRepositories "taoniu.local/cryptos/repositories/tradingview"
)

type GridsRepository struct {
	Db                    *gorm.DB
	Rdb                   *redis.Client
	Ctx                   context.Context
	AccountRepository     *isolatedRepositories.AccountRepository
	OrdersRepository      *marginRepositories.OrdersRepository
	SymbolsRepository     *spotRepositories.SymbolsRepository
	GridsRepository       *spotRepositories.GridsRepository
	TradingviewRepository *tradingviewRepositories.AnalysisRepository
}

type GridsError struct {
	Message string
}

func (m *GridsError) Error() string {
	return m.Message
}

func (r *GridsRepository) Account() *isolatedRepositories.AccountRepository {
	if r.AccountRepository == nil {
		r.AccountRepository = &isolatedRepositories.AccountRepository{
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.AccountRepository
}

func (r *GridsRepository) Orders() *marginRepositories.OrdersRepository {
	if r.OrdersRepository == nil {
		r.OrdersRepository = &marginRepositories.OrdersRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.OrdersRepository
}

func (r *GridsRepository) Symbols() *spotRepositories.SymbolsRepository {
	if r.SymbolsRepository == nil {
		r.SymbolsRepository = &spotRepositories.SymbolsRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.SymbolsRepository
}

func (r *GridsRepository) Grids() *spotRepositories.GridsRepository {
	if r.GridsRepository == nil {
		r.GridsRepository = &spotRepositories.GridsRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.GridsRepository
}

func (r *GridsRepository) Tradingview() *tradingviewRepositories.AnalysisRepository {
	if r.TradingviewRepository == nil {
		r.TradingviewRepository = &tradingviewRepositories.AnalysisRepository{
			Db: r.Db,
		}
	}
	return r.TradingviewRepository
}

func (r *GridsRepository) Flush(symbol string) error {
	signal, _, err := r.Tradingview().Signal(symbol)
	if err != nil {
		return err
	}
	if signal == 0 {
		return &GridsError{"tradingview no trading signal"}
	}

	price, err := r.Symbols().Price(symbol)
	if err != nil {
		return err
	}

	grid, err := r.Grids().Filter(symbol, price)
	if err != nil {
		return err
	}
	sellItems, err := r.FilterGrid(grid, price, signal)
	if err != nil {
		return err
	}

	if signal == 1 {
		amount := 10 * math.Pow(2, float64(grid.Step-1))
		return r.Buy(grid, price, amount)
	} else {
		return r.Sell(grid, sellItems)
	}

	return nil
}

func (r *GridsRepository) Buy(grid *spotModels.Grids, price float64, amount float64) error {
	balance, _, err := r.Account().Balance(grid.Symbol)
	if err != nil {
		return err
	}

	buyPrice, buyQuantity := r.Symbols().Filter(grid.Symbol, price, amount)
	sellPrice := buyPrice * (1 + grid.TakeProfitPercent)
	sellQuantity := buyQuantity * grid.TriggerPercent
	sellPrice, sellQuantity = r.Symbols().Filter(grid.Symbol, sellPrice, sellPrice*sellQuantity)

	buyAmount := buyPrice * buyQuantity

	var buyOrderId int64 = 0
	var status int64 = 0
	var remark = ""
	if balance < buyAmount || grid.Balance < buyAmount {
		status = 1
	} else {
		buyOrderId, err = r.Order(grid.Symbol, binance.SideTypeBuy, price, buyQuantity)
		if err != nil {
			remark = err.Error()
		} else {
			grid.Balance = grid.Balance - buyAmount
			r.Db.Model(&models.TradingGrid{ID: grid.ID}).Updates(grid)
		}
	}

	var entity *models.TradingGrid
	entity = &models.TradingGrid{
		ID:           xid.New().String(),
		Symbol:       grid.Symbol,
		GridID:       grid.ID,
		BuyOrderId:   buyOrderId,
		BuyPrice:     buyPrice,
		BuyQuantity:  buyQuantity,
		SellPrice:    sellPrice,
		SellQuantity: sellQuantity,
		Status:       status,
		Remark:       remark,
	}

	r.Db.Create(entity)

	return nil
}

func (r *GridsRepository) Sell(grid *spotModels.Grids, entities []*models.TradingGrid) error {
	for _, entity := range entities {
		var sellOrderId int64 = 0
		var err error
		var status int64 = 2
		var remark = entity.Remark
		if entity.BuyOrderId == 0 {
			status = 3
		} else {
			sellOrderId, err = r.Order(entity.Symbol, binance.SideTypeSell, entity.SellPrice, entity.SellQuantity)
			if err != nil {
				remark = err.Error()
			} else {
				grid.Balance = grid.Balance + entity.SellPrice*entity.SellQuantity
				r.Db.Model(&models.TradingGrid{ID: grid.ID}).Updates(grid)
			}
		}

		entity.SellOrderId = sellOrderId
		entity.Status = status
		entity.Remark = remark

		r.Db.Model(&models.TradingGrid{ID: entity.ID}).Updates(entity)
	}

	return nil
}

func (r *GridsRepository) Update() error {
	var entities []*models.TradingGrid
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

		var status int64
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

func (r *GridsRepository) FilterGrid(grid *spotModels.Grids, price float64, signal int64) ([]*models.TradingGrid, error) {
	var entryPrice float64
	var takePrice float64
	var entities []*models.TradingGrid
	r.Db.Where(
		"grid_id=? AND status IN ?",
		grid.ID,
		[]int64{0, 1},
	).Find(&entities)
	var sellItems []*models.TradingGrid
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
		return nil, &GridsError{"buy price too high"}
	}
	if signal == 2 && (takePrice == 0 || price < takePrice) {
		return nil, &GridsError{"sell price too low"}
	}
	if signal == 2 && len(sellItems) == 0 {
		return nil, &GridsError{"nothing sell"}
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
	r.Orders().Flush(symbol, result.OrderID, true)

	return result.OrderID, nil
}
