package tradings

import (
	"context"
	"errors"
	"strconv"
	config "taoniu.local/cryptos/config/binance/spot"
	"taoniu.local/cryptos/models/binance/spot/tradings"

	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
	"gorm.io/gorm"

	models "taoniu.local/cryptos/models/binance/spot"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	plansRepositories "taoniu.local/cryptos/repositories/binance/spot/plans"
	tradingviewRepositories "taoniu.local/cryptos/repositories/tradingview"
)

type ScalpingRepository struct {
	Db                    *gorm.DB
	Rdb                   *redis.Client
	Ctx                   context.Context
	SymbolsRepository     *spotRepositories.SymbolsRepository
	OrdersRepository      *spotRepositories.OrdersRepository
	AccountRepository     *spotRepositories.AccountRepository
	PlansRepository       *plansRepositories.DailyRepository
	TradingviewRepository *tradingviewRepositories.AnalysisRepository
}

func (r *ScalpingRepository) Symbols() *spotRepositories.SymbolsRepository {
	if r.SymbolsRepository == nil {
		r.SymbolsRepository = &spotRepositories.SymbolsRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.SymbolsRepository
}

func (r *ScalpingRepository) Orders() *spotRepositories.OrdersRepository {
	if r.OrdersRepository == nil {
		r.OrdersRepository = &spotRepositories.OrdersRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.OrdersRepository
}

func (r *ScalpingRepository) Account() *spotRepositories.AccountRepository {
	if r.AccountRepository == nil {
		r.AccountRepository = &spotRepositories.AccountRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.AccountRepository
}

func (r *ScalpingRepository) Plans() *plansRepositories.DailyRepository {
	if r.PlansRepository == nil {
		r.PlansRepository = &plansRepositories.DailyRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.PlansRepository
}

func (r *ScalpingRepository) Tradingview() *tradingviewRepositories.AnalysisRepository {
	if r.TradingviewRepository == nil {
		r.TradingviewRepository = &tradingviewRepositories.AnalysisRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.TradingviewRepository
}

func (r *ScalpingRepository) Flush() error {
	plan, err := r.Plans().Filter()
	if err != nil {
		return err
	}
	var entity *tradings.Scalping
	result := r.Db.Model(
		&tradings.Scalping{},
	).Where(
		"symbol=? AND status=0",
		plan.Symbol,
	).Take(&entity)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	balance, _, err := r.Account().Balance(plan.Symbol)
	if err != nil {
		return err
	}

	if plan.Side != 1 {
		if plan.Price < entity.SellPrice {
			return nil
		}

		var err error
		sellOrderId := int64(0)
		status := 2
		remark := entity.Remark
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

		r.Db.Model(&tradings.Scalping{ID: entity.ID}).Updates(entity)

		return nil
	}

	buyPrice, buyQuantity, _ := r.Symbols().Adjust(plan.Symbol, plan.Price, 10)
	sellPrice := buyPrice * (1 + 0.05)
	sellQuantity := buyQuantity
	sellPrice, sellQuantity, _ = r.Symbols().Adjust(plan.Symbol, sellPrice, sellPrice*sellQuantity)

	buyAmount := buyPrice * buyQuantity

	buyOrderId := int64(0)
	status := 0
	remark := ""
	if balance < buyAmount {
		status = 1
	} else {
		buyOrderId, err = r.Order(plan.Symbol, binance.SideTypeBuy, plan.Price, buyQuantity)
		if err != nil {
			remark = err.Error()
		}
	}

	entity = &tradings.Scalping{
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
	r.Db.Model(&models.Plan{ID: plan.ID}).Updates(plan)

	return nil
}

func (r *ScalpingRepository) Update() error {
	var entities []*tradings.Scalping
	r.Db.Where(
		"status IN ?",
		[]int64{0, 2},
	).Find(&entities)
	for _, entity := range entities {
		orderID := entity.BuyOrderId
		if entity.Status == 2 {
			orderID = entity.SellOrderId
		}

		var order models.Order
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

		status := entity.Status
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

		r.Db.Model(&tradings.Scalping{ID: entity.ID}).Updates(entity)
	}

	return nil
}

func (r *ScalpingRepository) Order(symbol string, side binance.SideType, price float64, quantity float64) (int64, error) {
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

	r.Orders().Flush(symbol, result.OrderID)

	return result.OrderID, nil
}
