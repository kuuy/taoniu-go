package tasks

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	repositories "taoniu.local/gamblings/repositories/wolf"
	tasks "taoniu.local/gamblings/tasks/wolf"
)

type WolfTask struct {
	Db          *gorm.DB
	Rdb         *redis.Client
	Ctx         context.Context
	DiceTask    *tasks.DiceTask
	AccountTask *tasks.AccountTask
}

func (t *WolfTask) Account() *tasks.AccountTask {
	if t.AccountTask == nil {
		t.AccountTask = &tasks.AccountTask{}
		t.AccountTask.Repository = &repositories.AccountRepository{
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.AccountTask
}

func (t *WolfTask) Dice() *tasks.DiceTask {
	if t.DiceTask == nil {
		t.DiceTask = &tasks.DiceTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.DiceTask
}
