package fishers

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "math"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "gorm.io/datatypes"
  "gorm.io/gorm"

  spotModels "taoniu.local/cryptos/models/binance/spot"
  models "taoniu.local/cryptos/models/binance/spot/tradings/fishers"
)

type AnalysisRepository interface {
  Summary(exchange string, symbol string, interval string) (map[string]interface{}, error)
}

type SymbolsRepository interface {
  Price(symbol string) (float64, error)
  Adjust(symbol string, price float64, amount float64) (float64, float64, error)
}

type AccountRepository interface {
  Balance(symbol string) (float64, float64, error)
  Flush() error
}

type OrdersRepository interface {
  Status(symbol string, orderID int64) string
  Create(symbol string, side string, price float64, quantity float64) (int64, error)
  Lost(symbol string, price float64, timestamp int64) int64
  Flush(symbol string, orderID int64) error
}

type FishersRepository struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  AnalysisRepository AnalysisRepository
  SymbolsRepository  SymbolsRepository
  AccountRepository  AccountRepository
  OrdersRepository   OrdersRepository
}

func (r *FishersRepository) Scan() []string {
  var symbols []string
  r.Db.Model(&spotModels.Fisher{}).Where("status", []int{1, 3}).Distinct().Pluck("symbol", &symbols)
  return symbols
}

func (r *FishersRepository) Apply(
    symbol string,
    amount float64,
    balance float64,
    targetBalance float64,
    stopBalance float64,
    tickers [][]float64,
) error {
  var fisher spotModels.Fisher
  result := r.Db.Where("symbol=? AND status IN ?", symbol, []int{1, 3, 4}).Take(&fisher)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    fisher = spotModels.Fisher{
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

func (r *FishersRepository) Flush(symbol string) error {
  var fisher spotModels.Fisher
  result := r.Db.Where("symbol=? AND status IN ?", symbol, []int{1, 3}).Take(&fisher)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("fishers empty")
  }

  if fisher.Status == 3 {
    var grid models.Grid
    result := r.Db.Where("symbol=? AND price=? AND status=0", symbol, fisher.Price).Take(&grid)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      return errors.New("grid not exists")
    }
    timestamp := grid.CreatedAt.Unix()
    orderID := r.OrdersRepository.Lost(symbol, fisher.Price, timestamp-30)
    if orderID > 0 {
      r.Db.Transaction(func(tx *gorm.DB) error {
        fisher.Status = 1
        if err := tx.Model(&spotModels.Fisher{ID: fisher.ID}).Updates(fisher).Error; err != nil {
          return err
        }
        grid.BuyOrderId = orderID
        if err := tx.Model(&models.Grid{ID: grid.ID}).Updates(grid).Error; err != nil {
          return err
        }
        return nil
      })
    }
    return errors.New("error occurs")
  }

  price, err := r.SymbolsRepository.Price(symbol)
  if err != nil {
    return err
  }
  r.Take(&fisher, price)

  var grids []*models.Grid
  r.Db.Where("symbol=? AND status IN ?", fisher.Symbol, []int{0, 2}).Find(&grids)
  for _, grid := range grids {
    if grid.Status == 0 {
      status := r.OrdersRepository.Status(symbol, grid.BuyOrderId)
      if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
        r.OrdersRepository.Flush(symbol, grid.BuyOrderId)
        continue
      }
      r.Db.Transaction(func(tx *gorm.DB) error {
        if status == "FILLED" {
          grid.Status = 1
        } else {
          fisher.Balance += grid.BuyPrice * grid.BuyQuantity
          if err := tx.Model(&spotModels.Fisher{ID: fisher.ID}).Updates(fisher).Error; err != nil {
            return err
          }
          grid.Status = 4
        }
        if err := tx.Model(&models.Grid{ID: grid.ID}).Updates(grid).Error; err != nil {
          return err
        }
        return nil
      })
    } else if grid.Status == 2 {
      status := r.OrdersRepository.Status(symbol, grid.SellOrderId)
      if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
        r.OrdersRepository.Flush(symbol, grid.BuyOrderId)
        continue
      }
      r.Db.Transaction(func(tx *gorm.DB) error {
        if status == "FILLED" {
          grid.Status = 3
        } else {
          fisher.Balance -= grid.SellPrice * grid.SellQuantity
          if err := tx.Model(&spotModels.Fisher{ID: fisher.ID}).Updates(fisher).Error; err != nil {
            return err
          }
          grid.Status = 5
        }
        if err := tx.Model(&models.Grid{ID: grid.ID}).Updates(grid).Error; err != nil {
          return err
        }
        return nil
      })
    }
  }

  return nil
}

func (r *FishersRepository) Place(symbol string) error {
  var fisher spotModels.Fisher
  result := r.Db.Where("symbol=? AND status IN ?", symbol, []int{1, 3}).Take(&fisher)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("fishers empty")
  }
  if fisher.Status == 3 {
    return errors.New("error occurs")
  }
  price, err := r.SymbolsRepository.Price(symbol)
  if err != nil {
    return err
  }
  minPrice := 0.0
  maxPrice := 0.0
  side := 0
  step := 0

  var tickers [][]float64
  var buf []byte

  buf, _ = fisher.Tickers.MarshalJSON()
  json.Unmarshal(buf, &tickers)
  for i, items := range tickers {
    for _, ticker := range items {
      if price >= ticker {
        minPrice = ticker
        step = i
        side = 1
        break
      }
      maxPrice = ticker
    }
    if side != 0 {
      break
    }
  }

  exchange := "BINANCE"
  interval := "1m"
  summary, err := r.AnalysisRepository.Summary(exchange, symbol, interval)
  if err != nil {
    return err
  }

  timestamp := time.Now().Unix()
  redisKey := fmt.Sprintf("binance:spot:tradings:fishers:recommendation")
  item, err := r.Rdb.HGet(r.Ctx, redisKey, symbol).Result()
  if err == nil {
    data := strings.Split(item, ",")
    if data[0] == "BUY" || data[0] == "STRONG_BUY" {
      return errors.New("recommendation not changed")
    }
    if summary["RECOMMENDATION"] == data[0] {
      return errors.New("recommendation not changed")
    }
  }

  if summary["RECOMMENDATION"] != "BUY" && summary["RECOMMENDATION"] != "STRONG_BUY" {
    r.Rdb.HSet(
      r.Ctx,
      redisKey,
      symbol,
      fmt.Sprintf("%s:%v", summary["RECOMMENDATION"], timestamp),
    )
    return errors.New("tradingview recommendation not for buy")
  }

  if side == 0 {
    return errors.New("fishers place waiting")
  }

  if side == 1 {
    if !r.CanBuy(symbol, price, minPrice, maxPrice) {
      return errors.New("can not buy now")
    }
  }

  if side == 1 {
    amount := fisher.StartAmount * math.Pow(2, float64(step))
    if amount > fisher.Balance {
      return errors.New(fmt.Sprintf("[%s] balance not enough", symbol))
    }
    buyPrice, buyQuantity, err := r.SymbolsRepository.Adjust(symbol, price, amount)
    if err != nil {
      return err
    }
    balance, _, err := r.AccountRepository.Balance(symbol)
    if err != nil {
      return err
    }
    if balance < buyPrice*buyQuantity {
      return errors.New(fmt.Sprintf("[%s] balance not enough", symbol))
    }
    sellPrice := buyPrice * 1.0035
    sellPrice, sellQuantity, err := r.SymbolsRepository.Adjust(symbol, sellPrice, amount)
    if err != nil {
      return err
    }
    if fisher.Balance <= fisher.StopBalance-buyPrice*buyQuantity {
      return errors.New("reached stop balance")
    }
    r.Db.Transaction(func(tx *gorm.DB) error {
      fisher.Price = buyPrice
      fisher.Balance -= buyPrice * buyQuantity
      orderID, err := r.OrdersRepository.Create(symbol, "BUY", buyPrice, buyQuantity)
      if err != nil {
        fisher.Status = 3
        fisher.Remark = err.Error()
      }
      if err := tx.Model(&spotModels.Fisher{ID: fisher.ID}).Updates(fisher).Error; err != nil {
        return err
      }
      grid := models.Grid{
        ID:           xid.New().String(),
        Symbol:       symbol,
        FisherID:     fisher.ID,
        BuyOrderId:   orderID,
        BuyPrice:     buyPrice,
        BuyQuantity:  buyQuantity,
        SellPrice:    sellPrice,
        SellQuantity: sellQuantity,
        Status:       0,
      }
      if err := tx.Create(&grid).Error; err != nil {
        return err
      }
      return nil
    })

    r.Rdb.HSet(
      r.Ctx,
      redisKey,
      symbol,
      fmt.Sprintf("%s:%v", summary["RECOMMENDATION"], timestamp),
    )

    r.AccountRepository.Flush()
  }

  return nil
}

func (r *FishersRepository) CanBuy(
    symbol string,
    price float64,
    minPrice float64,
    maxPrice float64,
) bool {
  if minPrice*maxPrice == 0 {
    return false
  }

  var grid models.Grid
  result := r.Db.Where("symbol=? AND status IN ?", symbol, []int{0, 1, 2}).Order("buy_price asc").Take(&grid)
  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    if maxPrice >= grid.BuyPrice {
      return false
    }
    if price > grid.BuyPrice*0.995 {
      return false
    }
  }
  return true
}

func (r *FishersRepository) Take(fisher *spotModels.Fisher, price float64) error {
  var grid models.Grid
  result := r.Db.Where("symbol=? AND status=?", fisher.Symbol, 1).Order("sell_price asc").Take(&grid)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("empty grid")
  }
  if price < grid.SellPrice {
    return errors.New("price too low")
  }
  r.Db.Transaction(func(tx *gorm.DB) error {
    fisher.Balance += grid.SellPrice * grid.SellQuantity
    orderID, err := r.OrdersRepository.Create(grid.Symbol, "SELL", grid.SellPrice, grid.SellQuantity)
    if err != nil {
      fisher.Status = 3
      fisher.Remark = err.Error()
    }
    if err := tx.Model(&spotModels.Fisher{ID: fisher.ID}).Updates(fisher).Error; err != nil {
      return err
    }
    grid.SellOrderId = orderID
    grid.Status = 2
    if err := tx.Model(&models.Grid{ID: grid.ID}).Updates(grid).Error; err != nil {
      return err
    }
    return nil
  })
  return nil
}

func (r *FishersRepository) JSON(in interface{}) datatypes.JSON {
  var out datatypes.JSON
  buf, _ := json.Marshal(in)
  json.Unmarshal(buf, &out)
  return out
}
