package spot

import (
	"context"
	"gorm.io/gorm"
	"log"

	"github.com/go-redis/redis/v8"

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
		Rdb: r.Rdb,
		Ctx: r.Ctx,
	}
}

func (r *TradingsRepository) Scalping() error {
	plan, err := r.PlansRepository().Filter()
	if err != nil {
		return err
	}
	log.Println("plan", plan)

	return nil
}
