package tasks

import (
	"gorm.io/gorm"
	repositories "taoniu.local/security/repositories/gfw"
	tasks "taoniu.local/security/tasks/gfw"
)

type GfwTask struct {
	Db      *gorm.DB
	DnsTask *tasks.DnsTask
}

func (t *GfwTask) Dns() *tasks.DnsTask {
	if t.DnsTask == nil {
		t.DnsTask = &tasks.DnsTask{}
		t.DnsTask.Repository = &repositories.DnsRepository{
			Db: t.Db,
		}
	}
	return t.DnsTask
}
