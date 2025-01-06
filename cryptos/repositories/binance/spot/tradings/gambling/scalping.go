package gambling

import (
  "context"
  "errors"
  "fmt"
  "log"
  "math"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  models "taoniu.local/cryptos/models/binance/spot"
  tradingsModels "taoniu.local/cryptos/models/binance/spot/tradings"
)

type ScalpingRepository struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  SymbolsRepository  SymbolsRepository
  AccountRepository  AccountRepository
  OrdersRepository   OrdersRepository
  PositionRepository PositionRepository
}

func (r *ScalpingRepository) Place(id string) (err error) {
  var scalping *models.Scalping
  result := r.Db.Take(&scalping, "id", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    err = errors.New("scalping not found")
    return
  }

  if scalping.ExpiredAt.Unix() < time.Now().Unix() {
    r.Db.Model(&scalping).Update("status", 4)
    return errors.New("scalping expired")
  }

  positionSide := "LONG"
  side := "BUY"

  entity, err := r.SymbolsRepository.Get(scalping.Symbol)
  if err != nil {
    return
  }

  tickSize, stepSize, notional, err := r.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return
  }

  position, err := r.PositionRepository.Get(scalping.Symbol)
  if err != nil {
    return
  }

  if position.EntryQuantity == 0 {
    return errors.New(fmt.Sprintf("gambling scalping [%s] empty position", scalping.Symbol))
  }

  entryPrice := position.EntryPrice
  entryQuantity := position.EntryQuantity
  entryAmount, _ := decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()

  price, err := r.SymbolsRepository.Price(scalping.Symbol)
  if err != nil {
    return
  }

  if price > entryPrice {
    return errors.New(fmt.Sprintf("gambling scalping [%s] long price big than entry price", scalping.Symbol))
  }

  var capital float64
  var buyPrice float64
  var buyQuantity float64
  var buyAmount float64

  var cachedEntryPrice float64
  var cachedEntryQuantity float64
  redisKey := fmt.Sprintf(config.REDIS_KEY_TRADINGS_GAMBLING_SCALPING_PLACE, scalping.Symbol)
  values, _ := r.Rdb.HMGet(r.Ctx, redisKey, []string{
    "entry_price",
    "entry_quantity",
    "buy_quantity",
  }...).Result()
  if len(values) == 4 && values[0] != nil && values[1] != nil {
    cachedEntryPrice, _ = strconv.ParseFloat(values[0].(string), 64)
    cachedEntryQuantity, _ = strconv.ParseFloat(values[1].(string), 64)
  }

  if cachedEntryPrice == entryPrice && cachedEntryQuantity == entryQuantity {
    buyQuantity, _ = strconv.ParseFloat(values[3].(string), 64)
    buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
    log.Println("load from cached price", buyQuantity)
  } else {
    cachedEntryPrice = entryPrice
    cachedEntryQuantity = entryQuantity

    ipart, _ := math.Modf(scalping.Capital)
    places := 1
    for ; ipart >= 10; ipart = ipart / 10 {
      places++
    }

    capital, err = r.PositionRepository.Capital(scalping.Capital, entryAmount, places)
    if err != nil {
      err = errors.New(fmt.Sprintf("scalping [%s] %s reach the max invest capital", scalping.Symbol, positionSide))
      return
    }
    ratio := r.PositionRepository.Ratio(capital, entryAmount)
    buyAmount, _ = decimal.NewFromFloat(capital).Mul(decimal.NewFromFloat(ratio)).Float64()
    if buyAmount < notional {
      buyAmount = notional
    }
    buyQuantity = r.PositionRepository.BuyQuantity(buyAmount, entryPrice, entryAmount)
    buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()

    r.Rdb.HMSet(
      r.Ctx,
      redisKey,
      map[string]interface{}{
        "entry_price":    cachedEntryPrice,
        "entry_quantity": cachedEntryQuantity,
        "buy_quantity":   buyQuantity,
      },
    )
  }

  buyPrice, _ = decimal.NewFromFloat(entryPrice).Mul(
    decimal.NewFromFloat(100).Sub(
      decimal.NewFromFloat(config.GAMBLING_SCALPING_PRICE_LOSE_PERCENT),
    ).Div(decimal.NewFromFloat(100)),
  ).Float64()
  if price < buyPrice {
    buyPrice = price
  }
  buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()

  sellPrice := buyPrice * 1.0385
  sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()

  if price > buyPrice {
    return errors.New(fmt.Sprintf("gambling scalping [%s] %s price must reach %v", scalping.Symbol, positionSide, buyPrice))
  }

  buyAmount, _ = decimal.NewFromFloat(buyQuantity).Mul(decimal.NewFromFloat(buyPrice)).Float64()
  if buyAmount < config.GAMBLING_SCALPING_MIN_AMOUNT {
    return errors.New(fmt.Sprintf("gambling scalping [%s] amount must reach %v", scalping.Symbol, config.GAMBLING_SCALPING_MIN_AMOUNT))
  }
  if buyAmount > config.GAMBLING_SCALPING_MAX_AMOUNT {
    return errors.New(fmt.Sprintf("gambling scalping [%s] amount can not exceed %v", scalping.Symbol, config.GAMBLING_SCALPING_MAX_AMOUNT))
  }

  if !r.CanBuy(scalping, buyPrice) {
    return errors.New(fmt.Sprintf("gambling scalping [%s] can not buy now", scalping.Symbol))
  }

  balance, err := r.AccountRepository.Balance(entity.QuoteAsset)
  if err != nil {
    return err
  }

  if balance["free"] < config.GAMBLING_SCALPING_MIN_BINANCE {
    return errors.New(fmt.Sprintf("gambling scalping free balance must reach %v", config.GAMBLING_SCALPING_MIN_BINANCE))
  }

  mutex := common.NewMutex(
    r.Rdb,
    r.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_PLACE, scalping.Symbol),
  )
  if !mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  orderId, err := r.OrdersRepository.Create(scalping.Symbol, side, buyPrice, buyQuantity)
  if err != nil {
    if common.IsBinanceAPIError(err) {
      return
    }
    r.Db.Model(&scalping).Where("version", scalping.Version).Updates(map[string]interface{}{
      "remark":  err.Error(),
      "version": gorm.Expr("version + ?", 1),
    })
  }

  trading := &tradingsModels.Scalping{
    ID:           xid.New().String(),
    Symbol:       scalping.Symbol,
    ScalpingId:   scalping.ID,
    PlanId:       "",
    BuyOrderId:   orderId,
    BuyPrice:     buyPrice,
    BuyQuantity:  buyQuantity,
    SellPrice:    sellPrice,
    SellQuantity: buyQuantity,
    Version:      1,
  }
  r.Db.Create(&trading)

  return
}

func (r *ScalpingRepository) CanBuy(
  scalping *models.Scalping,
  price float64,
) bool {
  var buyPrice float64

  val, _ := r.Rdb.Get(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_LAST_PRICE, scalping.Symbol)).Result()
  if val != "" {
    buyPrice, _ = strconv.ParseFloat(val, 64)
    if price >= buyPrice*0.9615 {
      return false
    }
  }

  isChange := false

  var tradings []*tradingsModels.Scalping
  r.Db.Select([]string{"status", "buy_price"}).Where("scalping_id=? AND status IN ?", scalping.ID, []int{0, 1, 2}).Find(&tradings)
  for _, trading := range tradings {
    if trading.Status == 0 {
      return false
    }
    if price >= trading.BuyPrice*0.9615 {
      return false
    }
    if buyPrice == 0 {
      buyPrice = trading.BuyPrice
      isChange = true
    } else {
      if buyPrice > trading.BuyPrice {
        buyPrice = trading.BuyPrice
        isChange = true
      }
    }
  }

  if isChange {
    r.Rdb.Set(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_LAST_PRICE, scalping.Symbol), buyPrice, -1)
  }

  return true
}
