package tradings

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
	"gorm.io/gorm"
	"time"

	spotModels "taoniu.local/cryptos/models/binance/spot"
	models "taoniu.local/cryptos/models/binance/spot/tradings"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	plansRepositories "taoniu.local/cryptos/repositories/binance/spot/plans"
)

type ScalpingRepository struct {
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	SymbolsRepository *spotRepositories.SymbolsRepository
	OrdersRepository  *spotRepositories.OrdersRepository
	AccountRepository *spotRepositories.AccountRepository
	PlansRepository   *plansRepositories.DailyRepository
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

func (r *ScalpingRepository) Flush() error {
	var entities []*models.Scalping
	r.Db.Where(
		"status IN ?",
		[]int64{0, 2},
	).Find(&entities)
	for _, entity := range entities {
		if entity.Status == 0 {
			timestamp := entity.CreatedAt.Unix()
			if entity.BuyOrderId == 0 {
				orderID := r.Orders().Lost(entity.Symbol, "BUY", entity.BuyPrice, timestamp-30)
				if orderID > 0 {
					entity.BuyOrderId = orderID
					if err := r.Db.Model(&models.Scalping{ID: entity.ID}).Updates(entity).Error; err != nil {
						return err
					}
				} else {
					if timestamp > time.Now().Unix()-300 {
						r.Db.Model(&models.Scalping{ID: entity.ID}).Update("status", 1)
					}
					return nil
				}
			}
			if entity.BuyOrderId > 0 {
				status := r.Orders().Status(entity.Symbol, entity.BuyOrderId)
				if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
					r.Orders().Flush(entity.Symbol, entity.BuyOrderId)
					continue
				}
				if status == "FILLED" {
					entity.Status = 1
				} else {
					entity.Status = 4
				}
			}
			r.Db.Model(&models.Scalping{ID: entity.ID}).Updates(entity)
		} else if entity.Status == 2 {
			timestamp := entity.UpdatedAt.Unix()
			if entity.SellOrderId == 0 {
				orderID := r.Orders().Lost(entity.Symbol, "SELL", entity.BuyPrice, timestamp-30)
				if orderID > 0 {
					entity.SellOrderId = orderID
					if err := r.Db.Model(&models.Scalping{ID: entity.ID}).Updates(entity).Error; err != nil {
						return err
					}
				} else {
					if timestamp > time.Now().Unix()-300 {
						r.Db.Model(&models.Scalping{ID: entity.ID}).Update("status", 1)
					}
					return nil
				}
			}
			if entity.SellOrderId > 0 {
				status := r.Orders().Status(entity.Symbol, entity.SellOrderId)
				if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
					r.Orders().Flush(entity.Symbol, entity.SellOrderId)
					continue
				}
				if status == "FILLED" {
					entity.Status = 3
				} else {
					entity.Status = 5
				}
			}
			r.Db.Model(&models.Scalping{ID: entity.ID}).Updates(entity)
		}
	}

	return nil
}

func (r *ScalpingRepository) Place() error {
	plan, err := r.Plans().Filter()
	if err != nil {
		return err
	}

	buyPrice, buyQuantity, err := r.Symbols().Adjust(plan.Symbol, plan.Price, plan.Amount)
	if err != nil {
		return err
	}
	balance, _, err := r.Account().Balance(plan.Symbol)
	if err != nil {
		return err
	}
	if balance < buyPrice*buyQuantity {
		return errors.New("balance not enough")
	}
	var sellPrice float64
	if plan.Amount > 15 {
		sellPrice = buyPrice * 1.02
	} else {
		sellPrice = buyPrice * 1.015
	}
	sellPrice, sellQuantity, err := r.Symbols().Adjust(plan.Symbol, sellPrice, plan.Amount)
	if err != nil {
		return err
	}

	r.Db.Transaction(func(tx *gorm.DB) error {
		var remark string
		orderID, err := r.Orders().Create(plan.Symbol, "BUY", buyPrice, buyQuantity)
		if err != nil {
			remark = err.Error()
		}
		entity := &models.Scalping{
			ID:           xid.New().String(),
			Symbol:       plan.Symbol,
			BuyOrderId:   orderID,
			BuyPrice:     buyPrice,
			BuyQuantity:  buyQuantity,
			SellPrice:    sellPrice,
			SellQuantity: sellQuantity,
			Status:       0,
			Remark:       remark,
		}
		if err := tx.Create(&entity).Error; err != nil {
			return err
		}

		if err := tx.Model(&spotModels.Plan{ID: plan.ID}).Update("status", 1).Error; err != nil {
			return err
		}

		return nil
	})

	r.Account().Flush()

	return nil
}
