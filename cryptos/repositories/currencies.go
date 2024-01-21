package repositories

import (
  "encoding/json"
  "errors"
  "github.com/rs/xid"
  "gorm.io/datatypes"
  "gorm.io/gorm"
  "taoniu.local/cryptos/models"
)

type CurrenciesRepository struct {
  Db *gorm.DB
}

func (r *CurrenciesRepository) Add(
  symbol string,
  selectorID string,
  totalSupply float64,
  circulatingSupply float64,
  price float64,
  volume float64,
) error {
  var marketCap float64
  if price > 0 && volume > 0 {
    marketCap = circulatingSupply * price
  }
  var entity *models.Currency
  result := r.Db.Where("symbol", symbol).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity = &models.Currency{
      ID:                xid.New().String(),
      Symbol:            symbol,
      SectorID:          selectorID,
      TotalSupply:       totalSupply,
      CirculatingSupply: circulatingSupply,
      Price:             price,
      Volume:            volume,
      MarketCap:         marketCap,
      Exchanges:         r.JSON([]string{}),
    }
    r.Db.Create(&entity)
  }
  return nil
}

func (r *CurrenciesRepository) About(symbol string) (string, error) {
  var about string
  result := r.Db.Model(&models.Currency{}).Select("about").Where("symbol", symbol).Take(&about)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return "", result.Error
  }
  return about, nil
}

func (r *CurrenciesRepository) SetAbout(symbol string, about string) error {
  r.Db.Model(&models.Currency{}).Where("symbol", symbol).Update("about", about)
  return nil
}

func (r *CurrenciesRepository) JSON(in interface{}) datatypes.JSON {
  buf, _ := json.Marshal(in)

  var out datatypes.JSON
  json.Unmarshal(buf, &out)
  return out
}
