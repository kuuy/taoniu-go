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
	BetTask      *tasks.BetTask
	HellsTask    *tasks.HellsTask
	PlansTask    *tasks.PlansTask
	MultipleTask *tasks.MultipleTask
}

func (t *DiceTask) Bet() *tasks.BetTask {
	if t.BetTask == nil {
		t.BetTask = &tasks.BetTask{}
		t.BetTask.Repository = &repositories.BetRepository{
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.BetTask
}

func (t *DiceTask) Hells() *tasks.HellsTask {
	if t.PlansTask == nil {
		t.HellsTask = &tasks.HellsTask{}
		t.HellsTask.Repository = &repositories.HellsRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.HellsTask
}

func (t *DiceTask) Plans() *tasks.PlansTask {
	if t.PlansTask == nil {
		t.PlansTask = &tasks.PlansTask{}
		t.PlansTask.Repository = &repositories.PlansRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.PlansTask
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
