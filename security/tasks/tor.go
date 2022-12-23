package tasks

import (
	"gorm.io/gorm"
	repositories "taoniu.local/security/repositories/tor"
	tasks "taoniu.local/security/tasks/tor"
)

type TorTask struct {
	Db          *gorm.DB
	BridgesTask *tasks.BridgesTask
}

func (t *TorTask) Bridges() *tasks.BridgesTask {
	if t.BridgesTask == nil {
		t.BridgesTask = &tasks.BridgesTask{}
		t.BridgesTask.Repository = &repositories.BridgesRepository{
			Db: t.Db,
		}
	}
	return t.BridgesTask
}
