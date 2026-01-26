package swap

import (
	"time"
)

type Transaction struct {
	ID          string    `gorm:"size:20;primaryKey"`
	Signature   string    `gorm:"size:88;not null;uniqueIndex"`
	PoolAddress string    `gorm:"size:44;not null;index"`
	MintIn      string    `gorm:"size:44;not null"`
	AmountIn    float64   `gorm:"type:double precision;not null"`
	MintOut     string    `gorm:"size:44;not null"`
	AmountOut   float64   `gorm:"type:double precision;not null"`
	Timestamp   time.Time `gorm:"not null;index"`
	Status      int       `gorm:"type:integer;not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (m *Transaction) TableName() string {
	return "raydium_swap_transactions"
}
