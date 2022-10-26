package isolated

import (
	"context"
	"errors"
	"github.com/rs/xid"
	"gorm.io/gorm"
	"math"
	"strconv"

	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"

	config "taoniu.local/cryptos/config/binance"
	spotModels "taoniu.local/cryptos/models/binance/spot"
	marginModels "taoniu.local/cryptos/models/binance/spot/margin"
	models "taoniu.local/cryptos/models/binance/spot/margin/isolated"
	binanceRepositories "taoniu.local/cryptos/repositories/binance"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	marginRepositories "taoniu.local/cryptos/repositories/binance/spot/margin"
	tradingviewRepositories "taoniu.local/cryptos/repositories/tradingview"
)

type TradingsRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

type TradingsError struct {
	Message string
}

func (m *TradingsError) Error() string {
	return m.Message
}

func (r *TradingsRepository) TradingviewRepository() *tradingviewRepositories.AnalysisRepository {
	return &tradingviewRepositories.AnalysisRepository{
		Db: r.Db,
	}
}

func (r *TradingsRepository) SymbolsRepository() *binanceRepositories.SymbolsRepository {
	return &binanceRepositories.SymbolsRepository{
		Db:  r.Db,
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
}

func (r *TradingsRepository) GridsRepository() *spotRepositories.GridsRepository {
	return &spotRepositories.GridsRepository{
		Db:  r.Db,
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
}

func (r *TradingsRepository) OrdersRepository() *marginRepositories.OrdersRepository {
	return &marginRepositories.OrdersRepository{
		Db:  r.Db,
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
}

func (r *TradingsRepository) AccountRepository() *AccountRepository {
	return &AccountRepository{
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
}

func (r *TradingsRepository) Grids(symbol string) error {
	signal, err := r.TradingviewRepository().Signal(symbol)
	if err != nil {
		return err
	}
	if signal == 0 {
		return &TradingsError{"tradingview no trading signal"}
	}

	price, err := r.SymbolsRepository().Price(symbol)
	if err != nil {
		return err
	}

	grid, err := r.GridsRepository().Filter(symbol, price)
	if err != nil {
		return err
	}
	sellItems, err := r.FilterGrid(grid, price, signal)
	if err != nil {
		return err
	}

	if signal == 1 {
		amount := 10 * math.Pow(2, float64(grid.Step-1))
		return r.BuyGrid(grid, price, amount)
	} else {
		return r.SellGrid(grid, sellItems)
	}

	return nil
}

func (r *TradingsRepository) BuyGrid(grid *spotModels.Grids, price float64, amount float64) error {
	balance, _, err := r.AccountRepository().Balance(grid.Symbol)
	if err != nil {
		return err
	}

	buyPrice, buyQuantity := r.SymbolsRepository().Filter(grid.Symbol, price, amount)
	sellPrice := buyPrice * (1 + grid.TakeProfitPercent)
	sellQuantity := buyQuantity * grid.TriggerPercent
	sellPrice, sellQuantity = r.SymbolsRepository().Filter(grid.Symbol, sellPrice, sellPrice*sellQuantity)

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

func (r *TradingsRepository) SellGrid(grid *spotModels.Grids, entities []*models.TradingGrid) error {
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

func (r *TradingsRepository) UpdateGrids() error {
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

		var order *marginModels.Order
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

func (r *TradingsRepository) FilterGrid(grid *spotModels.Grids, price float64, signal int64) ([]*models.TradingGrid, error) {
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
		return nil, &TradingsError{"buy price too high"}
	}
	if signal == 2 && (takePrice == 0 || price < takePrice) {
		return nil, &TradingsError{"sell price too low"}
	}
	if signal == 2 && len(sellItems) == 0 {
		return nil, &TradingsError{"nothing sell"}
	}

	return sellItems, nil
}

func (r *TradingsRepository) Order(symbol string, side binance.SideType, price float64, quantity float64) (int64, error) {
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
	r.OrdersRepository().Flush(symbol, result.OrderID, true)

	return result.OrderID, nil
}
