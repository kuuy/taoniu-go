package streams

import (
  "context"
  "crypto/hmac"
  "crypto/sha256"
  "encoding/base64"
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
  config "taoniu.local/cryptos/config/dydx"
)

type AccountHandler struct {
  Db     *gorm.DB
  Rdb    *redis.Client
  Ctx    context.Context
  Socket *websocket.Conn
  Nats   *nats.Conn
}

func NewAccountCommand() *cli.Command {
  var h AccountHandler
  return &cli.Command{
    Name:  "account",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AccountHandler{
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

func (h *AccountHandler) read() (message map[string]interface{}, err error) {
  var data []byte
  _, data, err = h.Socket.Read(h.Ctx)
  json.Unmarshal(data, &message)
  return
}

func (h *AccountHandler) ping() error {
  ctx, cancel := context.WithTimeout(h.Ctx, time.Second*1)
  defer cancel()
  return h.Socket.Ping(ctx)
}

func (h *AccountHandler) handler(message map[string]interface{}) {
  if message["type"] == "connected" {
    isoTimestamp := time.Unix(0, h.Timestamp()).UTC().Format("2006-01-02T15:04:05.000Z")
    payload := fmt.Sprintf("%sGET/ws/accounts", isoTimestamp)
    secret, _ := base64.URLEncoding.DecodeString(os.Getenv("DYDX_TRADE_API_SECRET"))
    mac := hmac.New(sha256.New, secret)
    _, err := mac.Write([]byte(payload))
    if err != nil {
      log.Println("error", err.Error())
      return
    }
    signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))
    params := map[string]interface{}{}
    params["type"] = "subscribe"
    params["channel"] = "v3_accounts"
    params["accountNumber"] = "0"
    params["apiKey"] = os.Getenv("DYDX_TRADE_API_KEY")
    params["passphrase"] = os.Getenv("DYDX_TRADE_API_PASSPHRASE")
    params["timestamp"] = isoTimestamp
    params["signature"] = signature
    wsjson.Write(h.Ctx, h.Socket, params)
  }

  if message["type"] == "channel_data" && message["channel"] == "v3_accounts" {
    contents := message["contents"].(map[string]interface{})
    orders := contents["orders"].([]interface{})
    for _, order := range orders {
      order := order.(map[string]interface{})
      data, _ := json.Marshal(map[string]interface{}{
        "symbol":   order["market"].(string),
        "order_id": order["id"].(string),
        "status":   order["status"].(string),
      })
      h.Nats.Publish(config.NATS_ORDERS_UPDATE, data)
    }
    h.Nats.Flush()
  }
}

func (h *AccountHandler) Run() (err error) {
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

func (h *AccountHandler) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}

func (h *AccountHandler) Timestamp() int64 {
  timestamp := time.Now().UnixMicro()
  value, err := h.Rdb.HGet(h.Ctx, "dydx:server", "timediff").Result()
  if err != nil {
    return timestamp
  }
  timediff, _ := strconv.ParseInt(value, 10, 64)
  return timestamp - timediff
}
