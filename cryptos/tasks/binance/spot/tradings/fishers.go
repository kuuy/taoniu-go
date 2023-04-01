package tradings

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"

	config "taoniu.local/cryptos/config/queue"
	jobs "taoniu.local/cryptos/queue/jobs/binance/spot/tradings"
	savingsRepositories "taoniu.local/cryptos/repositories/binance/savings"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	repositories "taoniu.local/cryptos/repositories/binance/spot/tradings/fishers"
	tasks "taoniu.local/cryptos/tasks/binance/spot/tradings/fishers"
)

type FishersTask struct {
	Db         *gorm.DB
	Rdb        *redis.Client
	Ctx        context.Context
	Asynq      *asynq.Client
	Job        *jobs.Fishers
	Repository *repositories.FishersRepository
	GridsTask  *tasks.GridsTask
}

func (t *FishersTask) Grids() *tasks.GridsTask {
	if t.GridsTask == nil {
		t.GridsTask = &tasks.GridsTask{}
		t.GridsTask.Repository = &repositories.GridsRepository{
			Db: t.Db,
		}
		t.GridsTask.Repository.AccountRepository = &spotRepositories.AccountRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.GridsTask.Repository.ProductsRepository = &savingsRepositories.ProductsRepository{
			Db: t.Db,
		}
	}
	return t.GridsTask
}

func (t *FishersTask) Flush() error {
	symbols := t.Repository.Scan()
	for _, symbol := range symbols {
		task, err := t.Job.Flush(symbol)
		if err != nil {
			return err
		}
		t.Asynq.Enqueue(
			task,
			asynq.Queue(config.BINANCE_SPOT_TRADINGS_FISHERS),
			asynq.MaxRetry(0),
			asynq.Timeout(5*time.Minute),
		)
	}
	return nil
}

func (t *FishersTask) Place() error {
	symbols := t.Repository.Scan()
	for _, symbol := range symbols {
		task, err := t.Job.Place(symbol)
		if err != nil {
			return err
		}
		t.Asynq.Enqueue(
			task,
			asynq.Queue(config.BINANCE_SPOT_TRADINGS_FISHERS),
			asynq.MaxRetry(0),
			asynq.Timeout(5*time.Minute),
		)
	}
	return nil
}
