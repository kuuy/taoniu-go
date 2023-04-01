package isolated

import (
	"context"
	"github.com/hibiken/asynq"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	tradingsJobs "taoniu.local/cryptos/queue/jobs/binance/spot/margin/isolated/tradings"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
	fishersRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings/fishers"
	tasks "taoniu.local/cryptos/tasks/binance/spot/margin/isolated/tradings"
)

type TradingsTask struct {
	Db          *gorm.DB
	Rdb         *redis.Client
	Ctx         context.Context
	Asynq       *asynq.Client
	FishersTask *tasks.FishersTask
	GridsTask   *tasks.GridsTask
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
