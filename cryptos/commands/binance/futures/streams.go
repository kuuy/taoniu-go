package main

import (
	"os"
  "fmt"
  "time"
  "context"
  "strings"
  "strconv"

  "nhooyr.io/websocket"
	
  "github.com/urfave/cli/v2"
  "github.com/go-redis/redis/v8"
  "github.com/RichardKnop/machinery/v2/log"
  "github.com/bitly/go-simplejson"

  pool "taoniu.local/cryptos/common"
)

var (
  rdb *redis.Client
  ctx context.Context
)

func main() {
  app := &cli.App{
    Name: "tasks queues",
    Usage: "taoniu cryptos tasks",
    Action: func(c *cli.Context) error {
      fmt.Println("binance cli error", c.Err)
      return nil
    },
    Commands: []*cli.Command{
      {
        Name: "start",
        Usage: "start websocket",
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
    log.FATAL.Fatalln("app start fatal", err)
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

  data := j.Get("data").MustMap()

  if data["e"] != "24hrMiniTicker" {
    return
  }

  timestamp := time.Now().Unix()
  redisKey := fmt.Sprintf("binance:futures:realtime:%s", data["s"])
  value, err := rdb.HGet(ctx, redisKey, "price").Result()
  if err != redis.Nil {
    lasttime,_ := strconv.ParseInt(value, 10, 64)
    if lasttime > timestamp {
      return
    }
  }
  rdb.HMSet(
    ctx,
    redisKey,
    map[string]interface{} {
      "symbol": data["s"],
      "price": data["c"],
      "open": data["o"],
      "high": data["h"],
      "low": data["l"],
      "volume": data["v"],
      "quota": data["q"],
      "timestamp": fmt.Sprint(timestamp),
    },
  )
}

func start() error {
  symbols, _ := rdb.SMembers(ctx, "binance:futures:websocket:symbols").Result()
  streams := []string{}
  for _,symbol := range symbols{
    streams = append(
      streams,
      fmt.Sprintf("%s@miniTicker", strings.ToLower(symbol)),
    )
  }
  endpoint := "wss://fstream.binance.com/stream?streams=" + strings.Join(streams, "/")

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

