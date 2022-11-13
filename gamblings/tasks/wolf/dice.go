package wolf

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	repositories "taoniu.local/gamblings/repositories/wolf/dice"
	tasks "taoniu.local/gamblings/tasks/wolf/dice"
)

type DiceTask struct {
	Db           *gorm.DB
	Rdb          *redis.Client
	Ctx          context.Context
	MultipleTask *tasks.MultipleTask
}

func (t *DiceTask) Multiple() *tasks.MultipleTask {
	if t.MultipleTask == nil {
		t.MultipleTask = &tasks.MultipleTask{}
		t.MultipleTask.Repository = &repositories.MultipleRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.MultipleTask
}

func (t *DiceTask) Clean() error {
	t.Multiple().Clean()
	return nil
}
