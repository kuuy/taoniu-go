package tradings

import "time"

type Triggers struct {
	ID           string    `gorm:"size:20;primaryKey"`
	Symbol       string    `gorm:"size:20;not null;index:idx_binance_spot_tradings_triggers_symbol_status"`
	BuyOrderId   int64     `gorm:"not null"`
	SellOrderId  int64     `gorm:"not null"`
	BuyPrice     float64   `gorm:"not null"`
	SellPrice    float64   `gorm:"not null"`
	BuyQuantity  float64   `gorm:"not null"`
	SellQuantity float64   `gorm:"not null"`
	Status       int       `gorm:"size:30;not null;index;index:idx_binance_spot_tradings_triggers_symbol_status"`
	Remark       string    `gorm:"size:5000;not null"`
	ExpiredAt    time.Time `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

func (m *Triggers) TableName() string {
	return "binance_spot_tradings_triggers"
}
