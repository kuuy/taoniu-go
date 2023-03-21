package tradings

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	savingsRepositories "taoniu.local/cryptos/repositories/binance/savings"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	repositories "taoniu.local/cryptos/repositories/binance/spot/tradings/fishers"
	tasks "taoniu.local/cryptos/tasks/binance/spot/tradings/fishers"
)

type FishersTask struct {
	Db         *gorm.DB
	Rdb        *redis.Client
	Ctx        context.Context
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
		err := t.Repository.Flush(symbol)
		if err != nil {
			log.Println("fishers flush error", err)
		}
	}
	return nil
}

func (t *FishersTask) Place() error {
	symbols := t.Repository.Scan()
	for _, symbol := range symbols {
		err := t.Repository.Place(symbol)
		if err != nil {
			log.Println("fishers Place error", err)
		}
	}
	return nil
}
