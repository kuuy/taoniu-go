package spot

import (
	"context"
	"errors"
	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
	"gorm.io/gorm"
	"strconv"
	config "taoniu.local/cryptos/config/binance"
	models "taoniu.local/cryptos/models/binance/spot"
	marginModels "taoniu.local/cryptos/models/binance/spot/margin"

	binanceRepositories "taoniu.local/cryptos/repositories/binance"
	plansRepository "taoniu.local/cryptos/repositories/binance/spot/plans"
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

func (r *TradingsRepository) TradingviewRepository() *tradingviewRepositories.AnalysisRepository {
	return &tradingviewRepositories.AnalysisRepository{
		Db: r.Db,
	}
}

func (r *TradingsRepository) PlansRepository() *plansRepository.DailyRepository {
	return &plansRepository.DailyRepository{
		Db:  r.Db,
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
}

func (r *TradingsRepository) OrdersRepository() *OrdersRepository {
	return &OrdersRepository{
		Db:  r.Db,
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
}

func (r *TradingsRepository) AccountRepository() *AccountRepository {
	return &AccountRepository{
		Db:  r.Db,
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
}

func (r *TradingsRepository) Scalping() error {
	plan, err := r.PlansRepository().Filter()
	if err != nil {
		return err
	}
	var entity *models.TradingScalping
	result := r.Db.Model(&models.TradingScalping{}).Where("symbol=? AND status=0", plan.Symbol).Take(&entity)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	balance, _, err := r.AccountRepository().Balance(plan.Symbol)
	if err != nil {
		return err
	}

	if plan.Side != 1 {
		if plan.Price < entity.SellPrice {
			return nil
		}

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
			}
		}

		entity.SellOrderId = sellOrderId
		entity.Status = status
		entity.Remark = remark

		r.Db.Model(&models.TradingScalping{ID: entity.ID}).Updates(entity)

		return nil
	}

	buyPrice, buyQuantity := r.SymbolsRepository().Filter(plan.Symbol, plan.Price, 10)
	sellPrice := buyPrice * (1 + 0.05)
	sellQuantity := buyQuantity
	sellPrice, sellQuantity = r.SymbolsRepository().Filter(plan.Symbol, sellPrice, sellPrice*sellQuantity)

	buyAmount := buyPrice * buyQuantity

	var buyOrderId int64 = 0
	var status int64 = 0
	var remark = ""
	if balance < buyAmount {
		status = 1
	} else {
		buyOrderId, err = r.Order(plan.Symbol, binance.SideTypeBuy, plan.Price, buyQuantity)
		if err != nil {
			remark = err.Error()
		}
	}

	entity = &models.TradingScalping{
		ID:           xid.New().String(),
		Symbol:       plan.Symbol,
		BuyOrderId:   buyOrderId,
		BuyPrice:     buyPrice,
		BuyQuantity:  buyQuantity,
		SellPrice:    sellPrice,
		SellQuantity: sellQuantity,
		Status:       status,
		Remark:       remark,
	}
	r.Db.Create(&entity)

	plan.Status = 1
	r.Db.Model(&models.Plans{ID: plan.ID}).Updates(plan)

	return nil
}

func (r *TradingsRepository) UpdateScalping() error {
	var entities []*models.TradingScalping
	r.Db.Where(
		"status IN ?",
		[]int64{0, 2},
	).Find(&entities)
	for _, entity := range entities {
		orderID := entity.BuyOrderId
		if entity.Status == 2 {
			orderID = entity.SellOrderId
		}

		var order *models.Order
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

func (r *TradingsRepository) Order(symbol string, side binance.SideType, price float64, quantity float64) (int64, error) {
	client := binance.NewClient(config.TRADE_API_KEY, config.TRADE_SECRET_KEY)
	result, err := client.NewCreateOrderService().Symbol(
		symbol,
	).Side(
		side,
	).Type(
		binance.OrderTypeLimit,
	).Price(
		strconv.FormatFloat(price, 'f', -1, 64),
	).Quantity(
		strconv.FormatFloat(quantity, 'f', -1, 64),
	).TimeInForce(
		binance.TimeInForceTypeGTC,
	).NewOrderRespType(
		binance.NewOrderRespTypeRESULT,
	).Do(r.Ctx)
	if err != nil {
		return 0, err
	}

	r.OrdersRepository().Flush(symbol, result.OrderID)

	return result.OrderID, nil
}
