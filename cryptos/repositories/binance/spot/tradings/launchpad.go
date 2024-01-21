package tradings

import (
  "errors"
  "fmt"
  "github.com/adshao/go-binance/v2/common"
  "github.com/rs/xid"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"
  "log"
  "math"
  spotModels "taoniu.local/cryptos/models/binance/spot"
  models "taoniu.local/cryptos/models/binance/spot/tradings"
  "time"
)

type LaunchpadRepository struct {
  Db                 *gorm.DB
  SymbolsRepository  SymbolsRepository
  OrdersRepository   OrdersRepository
  AccountRepository  AccountRepository
  PositionRepository PositionRepository
}

type LaunchpadBuyInfo struct {
  BuyPrice    float64
  BuyQuantity float64
  BuyAmount   float64
}

type LaunchpadSellInfo struct {
  SellPrice    float64
  SellQuantity float64
  SellAmount   float64
}

func (r *LaunchpadRepository) Scan() []string {
  var symbols []string
  r.Db.Model(&spotModels.Launchpad{}).Where("status", []int{1, 3}).Distinct().Pluck("symbol", &symbols)
  return symbols
}

func (r *LaunchpadRepository) Ids() []string {
  var ids []string
  r.Db.Model(&spotModels.Launchpad{}).Where("status", 1).Pluck("id", &ids)
  return ids
}

func (r *LaunchpadRepository) LaunchpadIds() []string {
  var ids []string
  r.Db.Model(&models.Launchpad{}).Select("launchpad_id").Where("status", []int{0, 1, 2}).Distinct().Pluck("launchpad_id", &ids)
  return ids
}

func (r *LaunchpadRepository) Place(id string) error {
  var launchpad *spotModels.Launchpad
  result := r.Db.First(&launchpad, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("launchpads empty")
  }

  if launchpad.IssuedAt.Unix() > time.Now().Unix() {
    return errors.New(
      fmt.Sprintf(
        "[%s] launchpad issued at %s",
        launchpad.Symbol,
        launchpad.IssuedAt.Format("2006-01-02 15:04"),
      ),
    )
  }

  if launchpad.ExpiredAt.Unix() < time.Now().Unix() {
    launchpad.Status = 4
    r.Db.Model(&launchpad).Updates(launchpad)
    return errors.New("launchpad expired")
  }

  entity, err := r.SymbolsRepository.Get(launchpad.Symbol)
  if err != nil {
    r.SymbolsRepository.Flush()
    return err
  }

  tickSize, stepSize, _, err := r.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return nil
  }

  price, err := r.SymbolsRepository.Price(launchpad.Symbol)
  if err != nil {
    return err
  }
  if price > launchpad.CorePrice {
    return errors.New(fmt.Sprintf("[%s] price must reach %v", launchpad.Symbol, launchpad.CorePrice))
  }

  balance, err := r.AccountRepository.Balance(entity.QuoteAsset)
  if err != nil {
    return err
  }

  buys := r.Buys(launchpad.Capital, launchpad.CorePrice, tickSize, stepSize)
  sells := r.Sells(launchpad.Capital, launchpad.CorePrice, tickSize, stepSize)
  for i := 0; i < len(buys); i++ {
    if balance["free"] < buys[i].BuyAmount {
      log.Println(fmt.Sprintf("[%s] balance not enough buy at %v", launchpad.Symbol, buys[i].BuyPrice))
      break
    }

    var basePrice float64
    if i < len(buys)-1 {
      basePrice = buys[i+1].BuyPrice
    }
    if !r.CanBuy(launchpad, buys[i].BuyPrice, basePrice) {
      log.Println(fmt.Sprintf("[%s] can not buy at %v", launchpad.Symbol, buys[i].BuyPrice))
      balance["free"], _ = decimal.NewFromFloat(balance["free"]).Sub(decimal.NewFromFloat(buys[i].BuyAmount)).Float64()
      continue
    }

    err := r.Db.Transaction(func(tx *gorm.DB) (err error) {
      orderID, err := r.OrdersRepository.Create(launchpad.Symbol, "BUY", buys[i].BuyPrice, buys[i].BuyQuantity)
      if err != nil {
        apiError, ok := err.(common.APIError)
        if ok {
          err := r.ApiError(apiError, launchpad)
          if err != nil {
            return err
          }
        }
        tx.Model(&launchpad).Where("version", launchpad.Version).Updates(map[string]interface{}{
          "remark":  err.Error(),
          "version": gorm.Expr("version + ?", 1),
        })
      }

      trading := models.Launchpad{
        ID:           xid.New().String(),
        Symbol:       launchpad.Symbol,
        LaunchpadID:  launchpad.ID,
        BuyOrderId:   orderID,
        BuyPrice:     buys[i].BuyPrice,
        BuyQuantity:  buys[i].BuyQuantity,
        SellPrice:    sells[i].SellPrice,
        SellQuantity: sells[i].SellQuantity,
      }
      return tx.Create(&trading).Error
    })
    if err != nil {
      log.Println(fmt.Sprintf("[%s] can not buy at %v", launchpad.Symbol, buys[i].BuyPrice), err.Error())
      break
    }
    balance["free"], _ = decimal.NewFromFloat(balance["free"]).Sub(decimal.NewFromFloat(buys[i].BuyAmount)).Float64()
  }

  return nil
}

func (r *LaunchpadRepository) Buys(
  capital float64,
  entryPrice float64,
  tickSize float64,
  stepSize float64,
) (result []*LaunchpadBuyInfo) {
  var buyPrice float64
  var buyQuantity float64
  var buyAmount float64
  var investAmount float64

  maxCapital := capital * 100

  ipart, _ := math.Modf(maxCapital)
  places := 1
  for ; ipart >= 10; ipart = ipart / 10 {
    places++
  }

  entryQuantity := 10 / entryPrice
  entryQuantity, _ = decimal.NewFromFloat(entryQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
  entryAmount, _ := decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()

  for {
    var err error
    capital, err := r.PositionRepository.Capital(maxCapital, entryAmount, places)
    if err != nil {
      break
    }
    ratio := r.PositionRepository.Ratio(capital, entryAmount)
    buyAmount, _ = decimal.NewFromFloat(capital).Mul(decimal.NewFromFloat(ratio)).Float64()
    if buyAmount < 10 {
      buyAmount = 10
    }
    buyQuantity = r.PositionRepository.BuyQuantity(buyAmount, entryPrice, entryAmount)
    buyPrice, _ = decimal.NewFromFloat(buyAmount).Div(decimal.NewFromFloat(buyQuantity)).Float64()
    buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
    buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
    buyAmount, _ = decimal.NewFromFloat(buyPrice).Mul(decimal.NewFromFloat(buyQuantity)).Float64()
    entryQuantity, _ = decimal.NewFromFloat(entryQuantity).Add(decimal.NewFromFloat(buyQuantity)).Float64()
    entryAmount, _ = decimal.NewFromFloat(entryAmount).Add(decimal.NewFromFloat(buyAmount)).Float64()
    entryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryQuantity)).Float64()

    investAmount, _ = decimal.NewFromFloat(investAmount).Add(decimal.NewFromFloat(buyAmount)).Float64()
    if investAmount > capital {
      break
    }

    result = append(result, &LaunchpadBuyInfo{
      BuyPrice:    buyPrice,
      BuyQuantity: buyQuantity,
      BuyAmount:   buyAmount,
    })
  }

  return
}

func (r *LaunchpadRepository) Sells(
  capital float64,
  entryPrice float64,
  tickSize float64,
  stepSize float64,
) (result []*LaunchpadSellInfo) {
  var sellPrice float64
  var sellQuantity float64
  var sellAmount float64

  maxCapital := capital * 100

  ipart, _ := math.Modf(maxCapital)
  places := 1
  for ; ipart >= 10; ipart = ipart / 10 {
    places++
  }

  entryQuantity := 10 / entryPrice
  entryQuantity, _ = decimal.NewFromFloat(entryQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
  entryAmount, _ := decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()

  for {
    var err error
    capital, err := r.PositionRepository.Capital(maxCapital, entryAmount, places)
    if err != nil {
      break
    }
    ratio := r.PositionRepository.Ratio(capital, entryAmount)
    sellAmount, _ = decimal.NewFromFloat(capital).Mul(decimal.NewFromFloat(ratio)).Float64()
    if sellAmount < 10 {
      sellAmount = 10
    }
    sellQuantity = r.PositionRepository.SellQuantity(sellAmount, entryPrice, entryAmount)
    sellPrice, _ = decimal.NewFromFloat(sellAmount).Div(decimal.NewFromFloat(sellQuantity)).Float64()
    sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
    sellQuantity, _ = decimal.NewFromFloat(sellQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
    sellAmount, _ = decimal.NewFromFloat(sellPrice).Mul(decimal.NewFromFloat(sellQuantity)).Float64()
    entryQuantity, _ = decimal.NewFromFloat(entryQuantity).Add(decimal.NewFromFloat(sellQuantity)).Float64()
    entryAmount, _ = decimal.NewFromFloat(entryAmount).Add(decimal.NewFromFloat(sellAmount)).Float64()
    entryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryQuantity)).Float64()

    result = append(result, &LaunchpadSellInfo{
      SellPrice:    sellPrice,
      SellQuantity: sellQuantity,
      SellAmount:   sellAmount,
    })
  }

  return
}

func (r *LaunchpadRepository) Flush(id string) error {
  var launchpad *spotModels.Launchpad
  result := r.Db.First(&launchpad, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("launchpad empty")
  }

  price, err := r.SymbolsRepository.Price(launchpad.Symbol)
  if err != nil {
    return err
  }
  err = r.Take(launchpad, price)
  if err != nil {
    log.Println("take error", err)
  }

  var tradings []*models.Launchpad
  r.Db.Where("launchpad_id=? AND status IN ?", launchpad.ID, []int{0, 2}).Find(&tradings)

  for _, trading := range tradings {
    if trading.Status == 0 {
      timestamp := trading.CreatedAt.Unix()
      if trading.BuyOrderId == 0 {
        orderID := r.OrdersRepository.Lost(trading.Symbol, "BUY", trading.BuyQuantity, timestamp-30)
        if orderID > 0 {
          trading.BuyOrderId = orderID
          result := r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
            "buy_order_id": trading.BuyOrderId,
            "version":      gorm.Expr("version + ?", 1),
          })
          if result.Error != nil {
            return result.Error
          }
          if result.RowsAffected == 0 {
            return errors.New("order update failed")
          }
        }
      }

      status := r.OrdersRepository.Status(trading.Symbol, trading.BuyOrderId)
      if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
        r.OrdersRepository.Flush(trading.Symbol, trading.BuyOrderId)
        continue
      }

      if status == "FILLED" {
        trading.Status = 1
      } else {
        trading.Status = 4
      }

      result := r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
        "buy_order_id": trading.BuyOrderId,
        "status":       trading.Status,
        "version":      gorm.Expr("version + ?", 1),
      })
      if result.Error != nil {
        return result.Error
      }
      if result.RowsAffected == 0 {
        return errors.New("order update failed")
      }
    }

    if trading.Status == 2 {
      timestamp := trading.UpdatedAt.Unix()
      if trading.SellOrderId == 0 {
        orderID := r.OrdersRepository.Lost(trading.Symbol, "SELL", trading.SellQuantity, timestamp-30)
        if orderID > 0 {
          trading.SellOrderId = orderID
          result := r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
            "sell_order_id": trading.SellOrderId,
            "version":       gorm.Expr("version + ?", 1),
          })
          if result.Error != nil {
            return result.Error
          }
          if result.RowsAffected == 0 {
            return errors.New("order update failed")
          }
        }
      }

      status := r.OrdersRepository.Status(trading.Symbol, trading.SellOrderId)
      if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
        r.OrdersRepository.Flush(trading.Symbol, trading.SellOrderId)
        continue
      }

      if status == "FILLED" {
        trading.Status = 3
      } else if status == "CANCELED" {
        trading.SellOrderId = 0
        trading.Status = 1
      } else {
        trading.Status = 5
      }

      result := r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
        "sell_order_id": trading.SellOrderId,
        "status":        trading.Status,
        "version":       gorm.Expr("version + ?", 1),
      })
      if result.Error != nil {
        return result.Error
      }
      if result.RowsAffected == 0 {
        return errors.New("order update failed")
      }
    }
  }

  return nil
}

func (r *LaunchpadRepository) Take(launchpad *spotModels.Launchpad, price float64) error {
  var sellPrice float64
  var trading *models.Launchpad

  entity, err := r.SymbolsRepository.Get(launchpad.Symbol)
  if err != nil {
    return err
  }

  tickSize, _, _, err := r.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return nil
  }

  result := r.Db.Where("launchpad_id=? AND status=?", launchpad.ID, 1).Order("sell_price asc").Take(&trading)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New(fmt.Sprintf("[%s] empty trading", launchpad.Symbol))
  }

  if price < trading.SellPrice {
    return errors.New("price too low")
  }

  if sellPrice < price*0.9985 {
    sellPrice = price * 0.9985
  } else {
    sellPrice = trading.SellPrice
  }
  sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()

  orderID, err := r.OrdersRepository.Create(trading.Symbol, "SELL", sellPrice, trading.SellQuantity)
  if err != nil {
    apiError, ok := err.(common.APIError)
    if ok {
      err := r.ApiError(apiError, launchpad)
      if err != nil {
        return err
      }
    }
    r.Db.Model(&launchpad).Where("version", launchpad.Version).Updates(map[string]interface{}{
      "remark":  err.Error(),
      "version": gorm.Expr("version + ?", 1),
    })
  }

  r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
    "sell_order_id": orderID,
    "status":        2,
    "version":       gorm.Expr("version + ?", 1),
  })

  return nil
}

func (r *LaunchpadRepository) Pending() map[string]float64 {
  var result []*PendingInfo
  r.Db.Model(&models.Scalping{}).Select(
    "symbol",
    "sum(sell_quantity) as quantity",
  ).Where("status", 1).Group("symbol").Find(&result)
  data := make(map[string]float64)
  for _, item := range result {
    data[item.Symbol] = item.Quantity
  }
  return data
}

func (r *LaunchpadRepository) CanBuy(
  launchpad *spotModels.Launchpad,
  buyPrice float64,
  basePrice float64,
) bool {
  var trading models.Launchpad
  query := r.Db.Where("launchpad_id=? AND status IN ?", launchpad.ID, []int{0, 1, 2})
  if basePrice > 0 {
    query.Where("buy_price > ?", basePrice)
  }
  result := query.Order("buy_price asc").Take(&trading)
  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    if buyPrice >= trading.BuyPrice {
      return false
    }
  }

  return true
}

func (r *LaunchpadRepository) ApiError(apiError common.APIError, launchpad *spotModels.Launchpad) error {
  if apiError.Code == -1013 || apiError.Code == -1111 || apiError.Code == -1121 || apiError.Code == -2010 || apiError.Code == -4016 {
    r.Db.Model(&spotModels.Launchpad{ID: launchpad.ID}).Where("version", launchpad.Version).Updates(map[string]interface{}{
      "remark":  apiError.Error(),
      "version": gorm.Expr("version + ?", 1),
    })
    return apiError
  }
  return nil
}