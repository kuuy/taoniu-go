package tasks

import (
	"context"
	"github.com/go-redis/redis/v8"
	"taoniu.local/cryptos/tasks/proxies"

	repositories "taoniu.local/cryptos/repositories/proxies"
)

type ProxiesTask struct {
	Rdb *redis.Client
	Ctx context.Context
}

func (t *ProxiesTask) Tor() *proxies.TorTask {
	return &proxies.TorTask{
		Rdb: t.Rdb,
		Ctx: t.Ctx,
		Repository: &repositories.TorRepository{
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		},
	}
}
