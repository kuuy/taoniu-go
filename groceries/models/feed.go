package models

type Feed struct {
  ID        string `gorm:"size:20;primaryKey"`
  Type      int    `gorm:"not null"`
  RelatedID string `gorm:"size:20"`
  score     int64  `gorm:"not null;index"`
}

func (m *Feed) TableName() string {
  return "groceries_feeds"
}
