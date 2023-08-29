package dydx

import (
  "context"
  "github.com/go-redis/redis/v8"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type ServerTask struct {
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.ServerRepository
}

func (t *ServerTask) Time() error {
  t.Repository.Time()
  return nil
}
