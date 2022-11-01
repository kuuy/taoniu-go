package repositories

import (
	"gorm.io/gorm"
	. "taoniu.local/cryptos/models/binance/futures"
)

type StrategyRepository struct {
	db *gorm.DB
}

func NewStrategyRepository(db *gorm.DB) *StrategyRepository {
	return &StrategyRepository{
		db: db,
	}
}

func (r *StrategyRepository) Count() (int64, error) {
	var count int64
	r.db.Model(&Strategy{}).Count(&count)

	return count, nil
}

func (r *StrategyRepository) Listings(current int, pageSize int) ([]Strategy, error) {
	offset := (current - 1) * pageSize

	var strategies []Strategy
	r.db.Select(
		"id",
		"symbol",
		"indicator",
		"price",
		"signal",
		"created_at",
	).Order(
		"created_at desc",
	).Offset(
		offset,
	).Limit(
		pageSize,
	).Find(
		&strategies,
	)

	return strategies, nil
}

func (r *StrategyRepository) Get(id string) (interface{}, error) {
	return nil, nil
}
