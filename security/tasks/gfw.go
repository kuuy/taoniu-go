package tasks

import (
  "context"

  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"

  "taoniu.local/security/grpc/services"
  repositories "taoniu.local/security/repositories/gfw"
  tasks "taoniu.local/security/tasks/gfw"
)

type GfwTask struct {
  Db      *gorm.DB
  Rdb     *redis.Client
  Ctx     context.Context
  DnsTask *tasks.DnsTask
}

func (t *GfwTask) Dns() *tasks.DnsTask {
  if t.DnsTask == nil {
    t.DnsTask = &tasks.DnsTask{}
    t.DnsTask.Repository = &repositories.DnsRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
    t.DnsTask.Repository.Service = &services.Aes{
      Ctx: t.Ctx,
    }
  }
  return t.DnsTask
}
