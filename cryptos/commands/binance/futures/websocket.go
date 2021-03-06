package main

import (
	"os"
  "fmt"
  "log"
  "context"
  "strconv"
  "strings"

  "nhooyr.io/websocket"

  "github.com/urfave/cli/v2"
  "github.com/go-redis/redis/v8"
  "github.com/adshao/go-binance/v2"
  "github.com/adshao/go-binance/v2/futures"
  "github.com/bitly/go-simplejson"

  pool "taoniu.local/cryptos/common"
)

var (
  rdb *redis.Client
  ctx context.Context
)

func main() {
  app := &cli.App{
    Name: "tasks",
    Usage: "taoniu cryptos tasks",
    Action: func(c *cli.Context) error {
      fmt.Println("binance cli error", c.Err)
      return nil
    },
    Commands: []*cli.Command{
      {
        Name: "start",
        Usage: "start stream",
        Action: func(c *cli.Context) error {
          if err := start(); err != nil {
            return cli.NewExitError(err.Error(), 1)
          }
          return nil
        },
      },
    },
    Version: "0.0.0",
  }

  rdb = pool.NewRedis()
  ctx = context.Background()

  err := app.Run(os.Args)
  if err != nil {
    log.Fatalln("app start fatal", err)
  }
}

func newJSON(data []byte) (j *simplejson.Json, err error) {
  j, err = simplejson.NewJson(data)
  if err != nil {
      return nil, err
  }
  return j, nil
}

func handler(message []byte) {
  j, err := newJSON(message)
  if err != nil {
    panic(err)
  }

  event := j.Get("e").MustString()
  timestamp := j.Get("T").MustInt64()

  if event == "ACCOUNT_UPDATE" {
    redisKey := "binance:futures:balance:USDT"
    value, err := rdb.HGet(ctx, redisKey, "timestamp").Result()
    if err != redis.Nil {
      lasttime,_ := strconv.ParseInt(value, 10, 64)
      if lasttime > timestamp {
        return
      }
    }
    accounts := j.Get("a").MustMap()
    for _,item := range accounts["B"].([]interface{}) {
      account := item.(map[string]interface{})
      if account["a"] != "USDT" {
        continue
      }
      rdb.HMSet(ctx, redisKey, map[string]interface{}{
        "balance":account["wb"],
        "timestamp":timestamp,
      })
    }
  }

  if event == "ORDER_TRADE_UPDATE" {
    order := j.Get("o").MustMap()
    symbol := fmt.Sprint(order["s"])
    orderID,_ := strconv.ParseInt(fmt.Sprint(order["i"]), 10, 64)
    status := order["X"]

    log.Println("order", order)
    if status != "NEW" || status != "PARTIALLY_FILLED" {
      rdb.SAdd(
        ctx,
        "binance:futures:websocket:orders",
        fmt.Sprintf("%s,%d", symbol, orderID),
      )
    }
    if status == "CANCELED" || status == "EXPIRED" {
      rdb.HDel(
        ctx,
        "binance:futures:orders:take_profit",
        fmt.Sprintf("%s,%d", symbol, orderID),
      )
    }
    if status == "FILLED" {
      item,err := rdb.HGet(
        ctx,
        "binance:futures:orders:take_profit",
        fmt.Sprintf("%s,%d", symbol, orderID),
      ).Result()
      if err != redis.Nil {
        data := strings.Split(item, ",")
        signal,_ := strconv.ParseInt(data[0], 10, 64)
        quantity := data[1]
        stopPrice := data[2]

        apiKey := "1ezcGDyXqV6fHPqockPILt5KMiXzUr4feoPMNmmqsmWakKJyK32GOvnL9LNoBg8n"
        secretKey := "AXHKOh04ndgWkQlwc8Ro4m6ZSBFudNno8b2zlLKtSwzy9B6cZbvsTyyWynzNMvCw"
        client := binance.NewFuturesClient(apiKey, secretKey)

        positionSide := futures.PositionSideTypeLong
        side := futures.SideTypeSell
        if signal == 2 {
          positionSide = futures.PositionSideTypeShort
          side = futures.SideTypeBuy
        }

        _,err := client.NewCreateOrderService().Symbol(
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
          fmt.Sprint(quantity),
        ).PriceProtect(
          true,
        ).NewClientOrderID(
          fmt.Sprintf("profit-%d", orderID),
        ).WorkingType(
          futures.WorkingTypeContractPrice,
        ).TimeInForce(
          futures.TimeInForceTypeGTC,
        ).Do(ctx)
        if err == nil {
          rdb.HDel(
            ctx,
            "binance:futures:orders:take_profit",
            fmt.Sprintf("%s,%d", symbol, orderID),
          )
        }
        if err != nil {
          log.Println("order submit failed", err)
        }

        log.Println("data:", positionSide, quantity, stopPrice)
      }
    }
  }
}

func start() error {
  log.Println("stream start")
    
  apiKey := "UE05rlVBaVjKpIUwFdTBQ7OOnM7E8YIz25ybQsg88odc2a8P9DvkUVJwmASx0ujB"
  secretKey := "eYIAk46ytB1PltICG1tZ3KuDPBHDcC12XkbFFiTURxwFBSUJIch7Kfufmzyh6mr7"

  client := binance.NewFuturesClient(apiKey, secretKey)

  listenKey,err := client.NewStartUserStreamService().Do(ctx)
  if err != nil {
    return err
  }
  log.Println("listenKey:", listenKey)
  defer client.NewCloseUserStreamService().ListenKey(listenKey).Do(ctx)

  endpoint := fmt.Sprintf("wss://fstream.binance.com/ws/%s", listenKey)

  socket, _, err := websocket.Dial(ctx, endpoint, nil)
  if err != nil {
    return err
  }
  socket.SetReadLimit(655350)

  for {
    _, message, readErr := socket.Read(ctx)
    if readErr != nil {
      return readErr
    }
    handler(message)
  }

  return nil
}

