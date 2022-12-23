package tor

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type ProxiesRepository struct {
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	BridgesRepository *BridgesRepository
}

func (r *ProxiesRepository) Bridges() *BridgesRepository {
	if r.BridgesRepository == nil {
		r.BridgesRepository = &BridgesRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.BridgesRepository
}

func (r *ProxiesRepository) Start(id int) error {
	limit := 20
	bridges, err := r.Bridges().Random(id, limit)
	if err != nil {
		return err
	}
	return r.Bridges().Monitor(id, bridges, false)
}
