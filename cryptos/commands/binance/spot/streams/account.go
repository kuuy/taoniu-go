package streams

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "os"
  "strconv"
  "time"

  "nhooyr.io/websocket"

  "github.com/adshao/go-binance/v2"
  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot/streams"
)

type AccountHandler struct {
  Rdb        *redis.Client
  Ctx        context.Context
  Socket     *websocket.Conn
  Nats       *nats.Conn
  AccountJob *jobs.Account
  OrdersJob  *jobs.Orders
}

func NewAccountCommand() *cli.Command {
  var h AccountHandler
  return &cli.Command{
    Name:  "account",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AccountHandler{
        Rdb:        common.NewRedis(1),
        Ctx:        context.Background(),
        Nats:       common.NewNats(),
        AccountJob: &jobs.Account{},
        OrdersJob:  &jobs.Orders{},
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      if err := h.start(); err != nil {
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
  event := message["e"].(string)

  if event == "outboundAccountPosition" {
    for _, item := range message["B"].([]interface{}) {
      account := item.(map[string]interface{})
      asset := account["a"].(string)
      if asset != "USDT" {
        continue
      }
      free, _ := strconv.ParseFloat(fmt.Sprintf("%v", account["f"]), 64)
      locked, _ := strconv.ParseFloat(fmt.Sprintf("%v", account["l"]), 64)

      h.Rdb.HMSet(
        h.Ctx,
        fmt.Sprintf(config.REDIS_KEY_BALANCE, asset),
        map[string]interface{}{
          "free":   free,
          "locked": locked,
        },
      )
    }
  }

  if event == "executionReport" {
    orderId, _ := strconv.ParseInt(fmt.Sprintf("%.0f", message["i"]), 10, 64)
    data, _ := json.Marshal(map[string]interface{}{
      "symbol":   message["s"].(string),
      "order_id": orderId,
      "status":   message["X"].(string),
    })
    h.Nats.Publish(config.NATS_ORDERS_UPDATE, data)
    h.Nats.Flush()
  }
}

func (h *AccountHandler) start() (err error) {
  log.Println("stream start")

  client := binance.NewClient(
    os.Getenv("BINANCE_SPOT_STREAMS_API_KEY"),
    os.Getenv("BINANCE_SPOT_STREAMS_API_SECRET"),
  )
  client.BaseURL = os.Getenv("BINANCE_SPOT_API_ENDPOINT")

  listenKey, err := client.NewStartUserStreamService().Do(h.Ctx)
  if err != nil {
    return err
  }
  defer client.NewCloseUserStreamService().ListenKey(listenKey).Do(h.Ctx)

  endpoint := fmt.Sprintf(
    "%s/ws/%s",
    os.Getenv("BINANCE_SPOT_STREAMS_ENDPOINT"),
    listenKey,
  )

  h.Socket, _, err = websocket.Dial(h.Ctx, endpoint, &websocket.DialOptions{
    CompressionMode: websocket.CompressionDisabled,
  })
  if err != nil {
    return
  }
  defer h.Socket.Close(websocket.StatusInternalError, "the socket was closed abruptly")
  h.Socket.SetReadLimit(655350)

  quit := make(chan struct{})
  go func() {
    defer close(quit)
    for {
      select {
      case <-h.Ctx.Done():
        return
      default:
        message, err := h.read()
        if err != nil {
          return
        }
        h.handler(message)
      }
    }
  }()

  ticker := time.NewTicker(time.Minute)
  defer ticker.Stop()

  for {
    select {
    case <-ticker.C:
      err = client.NewKeepaliveUserStreamService().ListenKey(listenKey).Do(h.Ctx)
      if err != nil {
        h.Socket.Close(websocket.StatusNormalClosure, "")
        return
      }
      err = h.ping()
      if err != nil {
        h.Socket.Close(websocket.StatusNormalClosure, "")
        return
      }
    case <-h.Ctx.Done():
      return
    case <-quit:
      return
    }
  }
}
