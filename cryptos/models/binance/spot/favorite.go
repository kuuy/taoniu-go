package spot

type FavoriteSymbol struct {
  UID    string `gorm:"size:20;not null;uniqueIndex:unq_binance_spot_favorite"`
  Symbol string `gorm:"size:20;not null;uniqueIndex:unq_binance_spot_favorite"`
}

func (m *FavoriteSymbol) TableName() string {
  return "binance_spot_favorite_symbols"
}
