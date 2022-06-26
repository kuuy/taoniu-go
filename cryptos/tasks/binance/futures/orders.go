package futures

import (
  "fmt"
	"context"
  "errors"
	"time"
  "strconv"
  "strings"

  "github.com/rs/xid"
  "github.com/adshao/go-binance/v2"
  "github.com/adshao/go-binance/v2/futures"

  "gorm.io/gorm"

	future "taoniu.local/cryptos/models"
	pool "taoniu.local/cryptos/common"
)

func saveOrder(db *gorm.DB, order *futures.Order) {
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
      OpenTime:order.Time,
      UpdateTime:order.UpdateTime,
      WorkingType:fmt.Sprint(order.WorkingType),
      PriceProtect:order.PriceProtect,
      ReduceOnly:order.ReduceOnly,
      ClosePosition:order.ClosePosition,
      Status:fmt.Sprint(order.Status),
      Remark:"",
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

func FlushOrders() error {
  
  ctx := context.Background()
  rdb := pool.NewRedis()
  defer rdb.Close()
  db := pool.NewDB()

  mutex := pool.NewMutex(
    rdb,
    ctx,
    "lock:binance:futures:orders:flush",
  )
  if mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  apiKey := "xxx"
  secretKey := "xxx"

  client := binance.NewFuturesClient(apiKey, secretKey)
  
  symbols, _ := rdb.SMembers(ctx, "binance:futures:websocket:symbols").Result()
  for _, symbol := range symbols {
    list,err := client.NewListOpenOrdersService().Symbol(symbol).Do(ctx) 
    if err != nil {
      continue
    }
    orderIds := []int64{}
    for _,order := range list {
      orderID := order.OrderID
      orderIds = append(orderIds, orderID)
      saveOrder(db, order)
    }

    if len(orderIds) == 0 {
      db.Model(&future.Order{}).Where(
        "symbol = ? AND status IN ?",
        symbol,
        []string{"NEW","PARTIALLY_FILLED"},
      ).Update("status", "UNKNOW")
    } else {
      db.Model(&future.Order{}).Where(
        "symbol = ? AND status IN ? AND order_id NOT IN ?",
        symbol,
        []string{"NEW","PARTIALLY_FILLED"},
        orderIds,
      ).Update("status", "UNKNOW")
    }
  }

  db.Model(&future.Order{}).Where(
    "symbol NOT IN ? AND status IN ?",
    symbols,
    []string{"NEW","PARTIALLY_FILLED"},
  ).Update("status", "UNKNOW")

  var entity future.Order
  result := db.Model(&future.Order{}).Where("status", "UNKNOW").First(&entity)
  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    order,err := client.NewGetOrderService().Symbol(entity.Symbol).OrderID(entity.OrderID).Do(ctx)
    if err == nil {
      saveOrder(db, order)
    }
  }
  
  list, _ := rdb.SMembers(ctx, "binance:futures:websocket:orders").Result()
  for _,item := range list {
    data := strings.Split(item, ",")
    symbol := data[0]
    orderID,_ := strconv.ParseInt(data[1], 10, 64)
    result := db.Model(&future.Order{}).Where(
      "symbol=? AND order_id=?",
      symbol,
      orderID,
    ).First(&entity)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {    
      order,err := client.NewGetOrderService().Symbol(symbol).OrderID(orderID).Do(ctx)
      if err == nil {
        saveOrder(db, order)
        rdb.SRem(ctx, "binance:futures:websocket:orders", item)
      }
    } else {
      rdb.SRem(ctx, "binance:futures:websocket:orders", item)
    }
  }

  return nil
}

