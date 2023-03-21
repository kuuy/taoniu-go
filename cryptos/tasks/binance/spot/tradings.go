package spot

import (
	"context"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
	fishersRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings/fishers"
	tvRepositories "taoniu.local/cryptos/repositories/tradingview"
	tasks "taoniu.local/cryptos/tasks/binance/spot/tradings"
)

type TradingsTask struct {
	Db           *gorm.DB
	Rdb          *redis.Client
	Ctx          context.Context
	FishersTask  *tasks.FishersTask
	ScalpingTask *tasks.ScalpingTask
}

func (t *TradingsTask) Fishers() *tasks.FishersTask {
	if t.FishersTask == nil {
		t.FishersTask = &tasks.FishersTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.FishersTask.Repository = &fishersRepositories.FishersRepository{
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
		t.FishersTask.Repository.AccountRepository = &spotRepositories.AccountRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.FishersTask.Repository.OrdersRepository = &spotRepositories.OrdersRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.FishersTask
}

func (t *TradingsTask) Scalping() *tasks.ScalpingTask {
	if t.ScalpingTask == nil {
		t.ScalpingTask = &tasks.ScalpingTask{}
		t.ScalpingTask.Repository = &repositories.ScalpingRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.ScalpingTask
}
