package isolated

import (
	"context"
	"taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings/fishers"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
	tvRepositories "taoniu.local/cryptos/repositories/tradingview"
	tasks "taoniu.local/cryptos/tasks/binance/spot/margin/isolated/tradings"
)

type TradingsTask struct {
	Db          *gorm.DB
	Rdb         *redis.Client
	Ctx         context.Context
	FishersTask *tasks.FishersTask
	GridsTask   *tasks.GridsTask
}

func (t *TradingsTask) Fishers() *tasks.FishersTask {
	if t.FishersTask == nil {
		t.FishersTask = &tasks.FishersTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.FishersTask.Repository = &fishers.FishersRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.FishersTask.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.FishersTask.Repository.AnalysisRepository = &tvRepositories.AnalysisRepository{
			Db: t.Db,
		}
		marginRepository := &spotRepositories.MarginRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.FishersTask.Repository.AccountRepository = marginRepository.Isolated().Account()
		t.FishersTask.Repository.OrdersRepository = marginRepository.Orders()
	}
	return t.FishersTask
}

func (t *TradingsTask) Grids() *tasks.GridsTask {
	if t.GridsTask == nil {
		t.GridsTask = &tasks.GridsTask{
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.GridsTask.Repository = &repositories.GridsRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.GridsTask
}
