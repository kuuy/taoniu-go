package futures

import (
  "fmt"
  "math"
	"context"
	"time"
  "strconv"
  "errors"
  "strings"
  "log"

  "github.com/rs/xid"
  "github.com/adshao/go-binance/v2"
  "github.com/adshao/go-binance/v2/futures"

  "gorm.io/gorm"

	future "taoniu.local/cryptos/models"
	pool "taoniu.local/cryptos/common"
)

func saveOrderResponse(db *gorm.DB, order *futures.CreateOrderResponse, remark string) {
  symbol := order.Symbol
  orderID := order.OrderID

  price,_ := strconv.ParseFloat(order.Price, 64)
  avgPrice,_ := strconv.ParseFloat(order.AvgPrice, 64)
  activatePrice,_ := strconv.ParseFloat(order.ActivatePrice, 64)
  stopPrice,_ := strconv.ParseFloat(order.StopPrice, 64)
  quantity,_ := strconv.ParseFloat(order.OrigQuantity, 64)
  executedQuantity,_ := strconv.ParseFloat(order.ExecutedQuantity, 64)

  var entity future.Order
  result := db.Where(
    "symbol=? AND order_id=?",
    symbol,
    orderID,
  ).First(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity = future.Order{
      ID:xid.New().String(),
      Symbol:symbol,
      OrderID:orderID,
      Type:fmt.Sprint(order.Type),
      PositionSide:fmt.Sprint(order.PositionSide),
      Side:fmt.Sprint(order.Side),
      Price:price,
      AvgPrice:avgPrice,
      ActivatePrice:activatePrice,
      StopPrice:stopPrice,
      Quantity:quantity,
      ExecutedQuantity:executedQuantity,
      OpenTime:order.UpdateTime,
      UpdateTime:order.UpdateTime,
      WorkingType:fmt.Sprint(order.WorkingType),
      PriceProtect:order.PriceProtect,
      ReduceOnly:order.ReduceOnly,
      ClosePosition:order.ClosePosition,
      Status:fmt.Sprint(order.Status),
      Remark:remark,
    }
    db.Create(&entity)
  } else {
    entity.AvgPrice = avgPrice
    entity.ExecutedQuantity = executedQuantity
    entity.UpdateTime = order.UpdateTime
    entity.Status = fmt.Sprint(order.Status)
    db.Model(&future.Order{ID:entity.ID}).Updates(entity)
  }
}

type Trade struct {
  Signal int64
  Price float64
  Quantity float64
  StopPrice float64
  StrategyID string
  StrategyRemark string
  Skip bool
}

func TakeProfit() error {
  ctx := context.Background()
  rdb := pool.NewRedis()
  defer rdb.Close()
  db := pool.NewDB()

  mutex := pool.NewMutex(
    rdb,
    ctx,
    "lock:binance:futures:orders:take_profit",
  )
  if mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  items,err := rdb.HGetAll(ctx, "binance:futures:orders:take_profit").Result()
  if err != nil {
    return err
  }

  apiKey := "xxx"
  secretKey := "xxx"
  client := binance.NewFuturesClient(apiKey, secretKey)

  for field,value := range items {
    x := strings.Split(field, ",")
    y := strings.Split(value, ",")
    symbol := x[0]
    orderID,_ := strconv.ParseInt(x[1], 10, 64)
    signal,_ := strconv.ParseInt(y[0], 10, 64)
    stopPrice := y[2]
    var order future.Order
    result := db.First(&order, "symbol=? AND order_id=?", symbol, orderID)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      continue
    }
    if order.Status == "CANCELED" || order.Status == "EXPIRED" {
      rdb.HDel(ctx, "binance:futures:orders:take_profit", field)
    }
    if order.Status != "FILLED" {
      continue
    }

    positionSide := futures.PositionSideTypeLong
    side := futures.SideTypeSell
    if signal == 2 {
      positionSide = futures.PositionSideTypeShort
      side = futures.SideTypeBuy
    }

    _,apiErr := client.NewCreateOrderService().Symbol(
      symbol,
    ).PositionSide(
      positionSide,
    ).Side(
      side,
    ).Type(
      futures.OrderTypeLimit,
    ).Price(
      stopPrice,
    ).Quantity(
      fmt.Sprint(order.Quantity),
    ).PriceProtect(
      true,
    ).NewClientOrderID(
      fmt.Sprintf("profit-%d", orderID),
    ).WorkingType(
      futures.WorkingTypeContractPrice,
    ).TimeInForce(
      futures.TimeInForceTypeGTC,
    ).Do(ctx)
    if apiErr == nil {
      rdb.HDel(
        ctx,
        "binance:futures:orders:take_profit",
        fmt.Sprintf("%s,%d", symbol, orderID),
      )
    }
    if apiErr != nil {
      return apiErr
    }
  }

  return nil
}

func SubmitOrder() error {
  ctx := context.Background()
  rdb := pool.NewRedis()
  defer rdb.Close()
  db := pool.NewDB()

  mutex := pool.NewMutex(
    rdb,
    ctx,
    "lock:binance:futures:orders:submit",
  )
  if mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  location,_ := time.LoadLocation("Asia/Shanghai")
  expireTime := time.Now().In(location).Add(-20 * time.Minute)

  var strategies []future.Strategy
  db.Where("updated_at > ?", expireTime).Order("updated_at desc").Find(&strategies)

  trades := make(map[string]Trade)
  for _,strategy := range strategies {
    symbol := strategy.Symbol
    signal := strategy.Signal
    price := strategy.Price
    if item,ok := trades[symbol]; ok {
      if item.Skip {
        continue
      }
    }
    if strategy.Volume > 0 {
      trades[symbol] = Trade{
        Skip:true,
      }
      continue
    }

    timestamp := time.Now().Unix()
    realtime,err := rdb.HMGet(
      ctx,
      fmt.Sprintf("binance:futures:realtime:%s", symbol),
      "r3",
      "r2",
      "r1",
      "s1",
      "s2",
      "s3",
      "quantity_step",
      "ticket_step",
      "timestamp",
    ).Result()
    if err != nil {
      trades[symbol] = Trade{
        Skip: true,
      }
      continue
    }
    r3,_ := strconv.ParseFloat(fmt.Sprint(realtime[0]), 64)
    r2,_ := strconv.ParseFloat(fmt.Sprint(realtime[1]), 64)
    r1,_ := strconv.ParseFloat(fmt.Sprint(realtime[2]), 64)
    s1,_ := strconv.ParseFloat(fmt.Sprint(realtime[3]), 64)
    s2,_ := strconv.ParseFloat(fmt.Sprint(realtime[4]), 64)
    s3,_ := strconv.ParseFloat(fmt.Sprint(realtime[5]), 64)
    quantityStep,_ := strconv.ParseFloat(fmt.Sprint(realtime[6]), 64)
    ticketStep,_ := strconv.ParseFloat(fmt.Sprint(realtime[7]), 64)
    lasttime,_ := strconv.ParseInt(fmt.Sprint(realtime[8]), 10, 64)
    
    p1 := r1 / s1
    p2 := r2 / s2
    p3 := r3 / s3
    if p1 < 1.01 || p2 < 1.02 ||  p3 < 1.03 {
      log.Println("less profit", p1, p2, p3)
      trades[symbol] = Trade{
        Skip : true,
      }
      continue
    }

    if timestamp - lasttime > 5 {
      trades[symbol] = Trade{
        Skip : true,
      }
      continue
    }
    position,rdbErr := rdb.HGetAll(
      ctx,
      fmt.Sprintf("binance:futures:positions:%s", symbol),
    ).Result()
    if rdbErr != nil {
      trades[symbol] = Trade{
        Skip : true,
      }
      continue
    }
    tradePrice := price
    stopPrice := price
    if signal == 1 && price > s1 && s1 > 0 {
      tradePrice = s1
    }
    if signal == 2 && price < r1 {
      tradePrice = r1
    }
    if strategy.Signal == 1 {
      entryPrice,_ := strconv.ParseFloat(
        fmt.Sprint(position["long_entry_price"]),
        64,
      )
      if tradePrice > entryPrice || entryPrice == 0.0 {
        continue
      }
      stopPrice = r1

      tradePrice = math.Floor(tradePrice / ticketStep) / math.Ceil(1 / ticketStep)
      stopPrice = math.Ceil(stopPrice / ticketStep) / math.Ceil(1 / ticketStep)
    }
    if strategy.Signal == 2 {
      entryPrice,_ := strconv.ParseFloat(
        fmt.Sprint(position["short_entry_price"]),
        64,
      )
      if tradePrice < entryPrice || entryPrice == 0.0 {
        continue
      }
      stopPrice = s1

      tradePrice = math.Ceil(tradePrice / ticketStep) / math.Ceil(1 / ticketStep)
      stopPrice = math.Floor(stopPrice / ticketStep) / math.Ceil(1 / ticketStep)
    }

    quantity := math.Ceil(100 / (tradePrice * quantityStep)) / math.Ceil(1 / quantityStep)
    if quantity > 10 {
      quantity = math.Floor(quantity)
    }

    trades[symbol] = Trade{
      Signal:signal,
      Price:tradePrice,
      Quantity:quantity,
      StopPrice:stopPrice,
      StrategyID:strategy.ID,
      StrategyRemark:strategy.Remark,
      Skip:false,
    }
  }

  apiKey := "1ezcGDyXqV6fHPqockPILt5KMiXzUr4feoPMNmmqsmWakKJyK32GOvnL9LNoBg8n"
  secretKey := "AXHKOh04ndgWkQlwc8Ro4m6ZSBFudNno8b2zlLKtSwzy9B6cZbvsTyyWynzNMvCw"
  client := binance.NewFuturesClient(apiKey, secretKey)

  for symbol,trade := range trades {
    if trade.Skip {
      continue
    }

    signal := trade.Signal
    quantity := trade.Quantity
    tradePrice := trade.Price
    stopPrice := trade.StopPrice
    strategyID := trade.StrategyID
    strategyRemark := trade.StrategyRemark

    positionSide := futures.PositionSideTypeLong
    side := futures.SideTypeBuy
    if signal == 2 {
      positionSide = futures.PositionSideTypeShort
      side = futures.SideTypeSell
    }
    result,apiErr := client.NewCreateOrderService().Symbol(
      symbol,
    ).PositionSide(
      positionSide,
    ).Side(
      side,
    ).Type(
      futures.OrderTypeLimit,
    ).Quantity(
      fmt.Sprint(quantity),
    ).Price(
      fmt.Sprint(tradePrice),
    ).NewClientOrderID(
      fmt.Sprintf("taoniu-%s", strategyID),
    ).WorkingType(
      futures.WorkingTypeContractPrice,
    ).PriceProtect(
      true,
    ).TimeInForce(
      futures.TimeInForceTypeGTC,
    ).NewOrderResponseType(
      futures.NewOrderRespTypeRESULT,
    ).Do(ctx)
    if apiErr != nil {
      log.Println("api error:", apiErr)
      continue
    }
    orderID := result.OrderID

    remark := fmt.Sprintf(
      "%s o:%d p:%f q:%f a:%f",
      strategyRemark,
      orderID,
      tradePrice,
      quantity,
      stopPrice,
    )
    db.Model(&future.Strategy{ID:strategyID}).Updates(
      map[string]interface{}{
        "volume":100,
        "remark":remark,
      },
    )

    rdb.HSet(
      ctx,
      "binance:futures:orders:take_profit",
      strings.Join([]string{
        symbol,
        fmt.Sprint(orderID),
      }, ","),
      strings.Join([]string{
        fmt.Sprint(signal),
        fmt.Sprint(quantity),
        fmt.Sprint(stopPrice),
      }, ","),
    )

    saveOrderResponse(
      db,
      result,
      fmt.Sprintf(
        "s:%s p:%f q:%f s:%f",
        strategyID,
        tradePrice,
        quantity,
        stopPrice,
      ),
    )
  }

  return nil
}

