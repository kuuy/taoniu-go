package streams

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "os"
  "strconv"
  "time"

  "github.com/coder/websocket"
  "github.com/coder/websocket/wsjson"
  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
)

type KlinesHandler struct {
  Db     *gorm.DB
  Rdb    *redis.Client
  Ctx    context.Context
  Socket *websocket.Conn
  Nats   *nats.Conn
}

func NewKlinesCommand() *cli.Command {
  var h KlinesHandler
  return &cli.Command{
    Name:  "klines",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = KlinesHandler{
        Rdb:  common.NewRedis(1),
        Ctx:  context.Background(),
        Nats: common.NewNats(),
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      if err := h.Run(); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *KlinesHandler) read() (message map[string]interface{}, err error) {
  var data []byte
  _, data, err = h.Socket.Read(h.Ctx)
  json.Unmarshal(data, &message)
  return
}

func (h *KlinesHandler) ping() error {
  ctx, cancel := context.WithTimeout(h.Ctx, time.Second*1)
  defer cancel()
  return h.Socket.Ping(ctx)
}

func (h *KlinesHandler) handler(message map[string]interface{}) {
  if message["type"] == "connected" {
    var symbols []string
    symbols = append(symbols, "BTC-USD")
    symbols = append(symbols, "DOT-USD")
    for _, symbol := range symbols {
      params := map[string]interface{}{}
      params["type"] = "subscribe"
      params["channel"] = "v4_candles"
      params["id"] = fmt.Sprintf("%v/1MIN", symbol)
      wsjson.Write(h.Ctx, h.Socket, params)
    }
  } else {
    log.Println("message", message["type"])
  }

  if message["type"] == "channel_data" && message["channel"] == "v4_candles" {
    contents := message["contents"].(map[string]interface{})
    log.Println("contents", contents)
  }
}

func (h *KlinesHandler) Run() (err error) {
  log.Println("stream start")

  endpoint := os.Getenv("DYDX_STREAMS_ENDPOINT")
  log.Println("endpoint", endpoint)

  h.Socket, _, err = websocket.Dial(h.Ctx, endpoint, &websocket.DialOptions{
    CompressionMode: websocket.CompressionDisabled,
  })
  if err != nil {
    return
  }
  defer h.Socket.Close(websocket.StatusInternalError, "the socket was closed abruptly")

  go func() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
      select {
      case <-ticker.C:
        err = h.ping()
        if err != nil {
          h.Socket.Close(websocket.StatusNormalClosure, "")
          return
        }
      case <-h.Ctx.Done():
        return
      }
    }
  }()

  for {
    select {
    case <-h.Ctx.Done():
      return
    default:
      message, err := h.read()
      if err != nil {
        return err
      }
      h.handler(message)
    }
  }
}

func (h *KlinesHandler) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}

func (h *KlinesHandler) Timestamp() int64 {
  timestamp := time.Now().UnixMicro()
  value, err := h.Rdb.HGet(h.Ctx, "dydx:server", "timediff").Result()
  if err != nil {
    return timestamp
  }
  timediff, _ := strconv.ParseInt(value, 10, 64)
  return timestamp - timediff
}
