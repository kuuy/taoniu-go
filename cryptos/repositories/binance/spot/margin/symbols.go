package margin

import (
  "gorm.io/gorm"
  models "taoniu.local/cryptos/models/binance/spot"
)

type SymbolsRepository struct {
  Db *gorm.DB
}

func (r *SymbolsRepository) Assets() []string {
  var assets []string
  r.Db.Model(models.Symbol{}).Where("status=? AND is_margin=True", "TRADING").Distinct().Pluck("base_asset", &assets)
  return assets
}
