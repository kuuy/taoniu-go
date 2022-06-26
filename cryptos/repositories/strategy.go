package repositories

import (
  "gorm.io/gorm"

  . "taoniu.local/cryptos/models"
)

type StrategyRepository struct {
  db *gorm.DB
}

func NewStrategyRepository(db *gorm.DB) *StrategyRepository {
  return &StrategyRepository{
    db: db,
  }
}

func (r *StrategyRepository) Listings() ([]Strategy, error) {
  offset := 0
  limit := 25

  var strategies []Strategy
  r.db.Select(
    "id",
    "symbol",
    "indicator",
    "price",
    "signal",
  ).Order(
    "created_at desc",
  ).Offset(
    offset,
  ).Limit(
    limit,
  ).Find(
    &strategies,
  )

  return strategies, nil
}

func (r *StrategyRepository) Get(id string) (interface{}, error) {
  return nil, nil
}

