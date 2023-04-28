package spot

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"

	tradingsJobs "taoniu.local/cryptos/queue/jobs/binance/spot/tradings"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
	plansRepositories "taoniu.local/cryptos/repositories/binance/spot/plans"
	tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
	fishersRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings/fishers"
	tasks "taoniu.local/cryptos/tasks/binance/spot/tradings"
)

type TradingsTask struct {
	Db           *gorm.DB
	Rdb          *redis.Client
	Ctx          context.Context
	Asynq        *asynq.Client
	FishersTask  *tasks.FishersTask
	ScalpingTask *tasks.ScalpingTask
}

func (t *TradingsTask) Fishers() *tasks.FishersTask {
	if t.FishersTask == nil {
		t.FishersTask = &tasks.FishersTask{
			Db:    t.Db,
			Rdb:   t.Rdb,
			Ctx:   t.Ctx,
			Asynq: t.Asynq,
		}
		t.FishersTask.Job = &tradingsJobs.Fishers{}
		t.FishersTask.Repository = &fishersRepositories.FishersRepository{
			Db: t.Db,
		}
	}
	return t.FishersTask
}

func (t *TradingsTask) Scalping() *tasks.ScalpingTask {
	if t.ScalpingTask == nil {
		t.ScalpingTask = &tasks.ScalpingTask{
			Asynq: t.Asynq,
		}
		t.ScalpingTask.Repository = &tradingsRepositories.ScalpingRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.ScalpingTask.Job = &tradingsJobs.Scalping{}
		t.ScalpingTask.Repository.SymbolsRepository = &repositories.SymbolsRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.ScalpingTask.PlansRepository = &plansRepositories.DailyRepository{
			Db: t.Db,
		}
	}
	return t.ScalpingTask
}
