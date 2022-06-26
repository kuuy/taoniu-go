package repositories

import (
  "gorm.io/gorm"

  . "taoniu.local/cryptos/models"
)

type OrderRepository struct {
  db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
  return &OrderRepository{
    db: db,
  }
}

func (r *OrderRepository) Listings() ([]Order, error) {
  offset := 0
  limit := 25

  var orders []Order
  r.db.Select(
    "id",
    "symbol",
    "type",
    "position_side",
    "side",
    "price",
    "open_time",
    "update_time",
    "status",
  ).Order(
    "open_time desc",
  ).Offset(
    offset,
  ).Limit(
    limit,
  ).Find(
    &orders,
  )

  return orders, nil
}

func (r *OrderRepository) Get(id string) (interface{}, error) {
  return nil, nil
}

