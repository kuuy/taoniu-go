package isolated

import (
  "encoding/json"
  "errors"

  "github.com/rs/xid"
  "gorm.io/datatypes"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/spot/margin/isolated"
)

type FishersRepository struct {
  Db *gorm.DB
}

func (r *FishersRepository) Apply(
  symbol string,
  amount float64,
  balance float64,
  targetBalance float64,
  stopBalance float64,
  tickers [][]float64,
) error {
  var fisher models.Fisher
  result := r.Db.Where("symbol=? AND status IN ?", symbol, []int{1, 3, 4}).Take(&fisher)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    fisher = models.Fisher{
      ID:            xid.New().String(),
      Symbol:        symbol,
      Price:         0,
      Balance:       balance,
      Tickers:       r.JSON(tickers),
      StartAmount:   amount,
      StartBalance:  balance,
      TargetBalance: targetBalance,
      StopBalance:   stopBalance,
      Status:        1,
    }
    r.Db.Create(&fisher)
  } else {
    if fisher.Status == 4 {
      return errors.New("stop loss occured")
    }
    if fisher.Status == 3 {
      return errors.New("fisher error waiting")
    }
    return errors.New("fisher not finished")
  }
  return nil
}

func (r *FishersRepository) JSON(in interface{}) datatypes.JSON {
  var out datatypes.JSON
  buf, _ := json.Marshal(in)
  json.Unmarshal(buf, &out)
  return out
}
