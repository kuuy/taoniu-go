package dice

import "time"

type Multiple struct {
	ID         string    `gorm:"size:20;primaryKey"`
	Currency   string    `gorm:"size:20;not null"`
	Amount     float64   `gorm:"not null"`
	Balance    float64   `gorm:"not null"`
	Invest     float64   `gorm:"not null;index"`
	Profit     float64   `gorm:"not null"`
	WinAmount  float64   `gorm:"not null"`
	LossAmount float64   `gorm:"not null"`
	WinCount   int       `gorm:"not null"`
	LossCount  int       `gorm:"not null"`
	Remark     string    `gorm:"size:5000;not null"`
	Status     uint8     `gorm:"not null;index;index:idx_wolf_dice_multiple_updated_status,priority:2"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null;index;index:idx_wolf_dice_multiple_updated_status,priority:1"`
}

func (m *Multiple) TableName() string {
	return "wolf_dice_multiple"
}
