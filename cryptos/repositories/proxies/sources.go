package proxies

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type SourcesRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

type SourcesError struct {
	Message string
}

func (m *SourcesError) Error() string {
	return m.Message
}

func (r *SourcesRepository) Add() error {
	return nil
}
