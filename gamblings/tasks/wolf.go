package tasks

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	tasks "taoniu.local/gamblings/tasks/wolf"
)

type WolfTask struct {
	Db       *gorm.DB
	Rdb      *redis.Client
	Ctx      context.Context
	DiceTask *tasks.DiceTask
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
