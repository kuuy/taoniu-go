package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
	tasks "taoniu.local/cryptos/tasks/binance/spot/tradings"
)

type TradingsTask struct {
	Db           *gorm.DB
	Rdb          *redis.Client
	Ctx          context.Context
	ScalpingTask *tasks.ScalpingTask
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
