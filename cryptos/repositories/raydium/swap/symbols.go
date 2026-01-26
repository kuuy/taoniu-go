package swap

import (
  "github.com/rs/xid"
  "gorm.io/gorm"
)

import (
  "errors"
  models "taoniu.local/cryptos/models/raydium/swap"
)

type SymbolsRepository struct {
  Db *gorm.DB
}

func (r *SymbolsRepository) Symbols() []string {
  var symbols []string
  r.Db.Model(models.Symbol{}).Select("symbol").Where("status", 1).Find(&symbols)
  return symbols
}

func (r *SymbolsRepository) Get(symbol string) (entity *models.Symbol, err error) {
  err = r.Db.Where("symbol", symbol).Take(&entity).Error
  return
}

func (r *SymbolsRepository) Apply(
  symbol string,
  baseAddress string,
  quoteAddress string,
) error {
  var scalping *models.Symbol
  result := r.Db.Where("symbol = ? AND status = 1", symbol).Take(&scalping)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity := &models.Symbol{
      ID:           xid.New().String(),
      Symbol:       symbol,
      BaseAddress:  baseAddress,
      QuoteAddress: quoteAddress,
      Status:       1,
    }
    r.Db.Create(&entity)
  } else {
    return errors.New("symbol exists")
  }
  return nil
}
