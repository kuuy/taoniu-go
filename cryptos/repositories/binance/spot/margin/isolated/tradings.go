package isolated

import (
	"context"
	"github.com/rs/xid"
	"gorm.io/gorm"
	"strconv"
	spotModels "taoniu.local/cryptos/models/binance/spot"

	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"

	config "taoniu.local/cryptos/config/binance"
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

func (r *TradingsRepository) TradingviewRepository() *tradingviewRepositories.AnalysisRepository {
	return &tradingviewRepositories.AnalysisRepository{
		Db: r.Db,
	}
}

func (r *TradingsRepository) AccountRepository() *AccountRepository {
	return &AccountRepository{
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
}

func (r *TradingsRepository) Scalping() error {
	return nil
}

func (r *TradingsRepository) Grids(symbol string) error {
	price, err := r.SymbolsRepository().Price(symbol)
	if err != nil {
		return err
	}
	signal, err := r.TradingviewRepository().Signal(symbol)
	if err != nil {
		return err
	}
	if signal == 0 {
		return &TradingsError{"tradingview no trading signal"}
	}

	grid, err := r.GridsRepository().Filter(symbol, price)
	if err != nil {
		return err
	}
	sellItems, err := r.FilterGrid(grid.ID, price, signal)
	if err != nil {
		return err
	}

	if signal == 1 {
		return r.BuyGrid(grid, price, 10)
	} else {
		return r.SellGrid(sellItems)
	}

	return nil
}

func (r *TradingsRepository) BuyGrid(grid *spotModels.Grids, price float64, amount float64) error {
	balance, _, err := r.AccountRepository().Balance(grid.Symbol)
	if err != nil {
		return err
	}

	buyPrice, buyQuantity := r.SymbolsRepository().Filter(grid.Symbol, price, amount)
	if balance < price*buyQuantity*1.005 {
		//return &TradingsError{"balance not enough"}
	}
	sellPrice := buyPrice * (1 + grid.TakeProfitPercent)
	sellQuantity := buyQuantity * grid.TriggerPercent
	sellPrice, sellQuantity = r.SymbolsRepository().Filter(grid.Symbol, sellPrice, sellPrice*sellQuantity)

	var entity *models.TradingGrid
	var buyOrderId int64 = 100
	//buyOrderId, err := r.Trade(grid.Symbol, binance.SideTypeBuy, price, buyQuantity)
	//if err != nil {
	//	return err
	//}
	entity = &models.TradingGrid{
		ID:           xid.New().String(),
		Symbol:       grid.Symbol,
		GridID:       grid.ID,
		BuyOrderId:   buyOrderId,
		BuyPrice:     buyPrice,
		SellPrice:    sellPrice,
		SellQuantity: sellQuantity,
		Status:       1,
	}
	r.Db.Create(entity)

	return nil
}

func (r *TradingsRepository) SellGrid(entities []*models.TradingGrid) error {
	var sellOrderId int64
	for _, entity := range entities {
		sellOrderId = 200
		//sellOrderId, err := r.Trade(entity.Symbol, binance.SideTypeSell, entity.SellPrice, entity.SellQuantity)
		//if err != nil {
		//	continue
		//}
		entity.SellOrderId = sellOrderId
		entity.Status = 3
		r.Db.Model(&models.TradingGrid{ID: entity.ID}).Updates(entity)
	}

	return nil
}

func (r *TradingsRepository) FilterGrid(gid string, price float64, signal int64) ([]*models.TradingGrid, error) {
	var entryPrice float64
	var takePrice float64
	var entities []*models.TradingGrid
	var sellItems []*models.TradingGrid
	r.Db.Where(
		"grid_id=? AND status IN ?",
		gid,
		[]int64{0, 1},
	).Find(&entities)
	for _, entity := range entities {
		if entryPrice == 0 || entryPrice > entity.BuyPrice {
			entryPrice = entity.BuyPrice
		}
		if takePrice == 0 || (entity.Status == 1 && takePrice < entity.SellPrice) {
			takePrice = entity.SellPrice
		}
		if entity.Status == 1 && price > entity.SellPrice*1.005 {
			sellItems = append(sellItems, entity)
		}
	}
	if signal == 1 && entryPrice > 0 && price > entryPrice*0.995 {
		return nil, &TradingsError{"buy price too high"}
	}
	if signal == 2 && (takePrice == 0 || price < takePrice*1.005) {
		return nil, &TradingsError{"sell price too low"}
	}
	if signal == 2 && len(sellItems) == 0 {
		return nil, &TradingsError{"nothing sell"}
	}

	return sellItems, nil
}

func (r *TradingsRepository) Trade(symbol string, side binance.SideType, price float64, quantity float64) (int64, error) {
	if quantity == 0 {
		return 0, nil
	} else {
		return 0, nil
	}
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
	marginRepository := marginRepositories.OrdersRepository{
		Db:  r.Db,
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
	marginRepository.Flush(symbol, result.OrderID, true)

	return result.OrderID, nil
}
