package streams

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "os"
  "strconv"
  "time"

  "github.com/adshao/go-binance/v2"
  "github.com/coder/websocket"
  "github.com/coder/websocket/wsjson"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"

  "github.com/go-redis/redis/v8"
  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures/streams"
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
        Rdb:        common.NewRedis(2),
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
  err = wsjson.Read(h.Ctx, h.Socket, &message)
  return
}

func (h *AccountHandler) ping() (err error) {
  ctx, cancel := context.WithTimeout(h.Ctx, 25*time.Second)
  defer cancel()
  err = h.Socket.Ping(ctx)
  if err != nil {
    log.Println("ping error", err.Error())
  }
  return
}

func (h *AccountHandler) handler(message map[string]interface{}) {
  event := message["e"].(string)

  if event == "ACCOUNT_UPDATE" {
    //info := message["a"].(map[string]interface{})
    //for _, item := range info["B"].([]interface{}) {
    //  account := item.(map[string]interface{})
    //  asset := account["a"].(string)
    //  if asset != "USDT" {
    //    continue
    //  }
    //  balance, _ := strconv.ParseFloat(fmt.Sprintf("%v", account["wb"]), 64)
    //
    //  h.Rdb.HMSet(
    //    h.Ctx,
    //    fmt.Sprintf(config.REDIS_KEY_BALANCE, asset),
    //    map[string]interface{}{
    //      "balance": balance,
    //    },
    //  )
    //}
  }

  if event == "ORDER_TRADE_UPDATE" {
    order := message["o"].(map[string]interface{})
    orderId, _ := strconv.ParseInt(fmt.Sprintf("%.0f", order["i"]), 10, 64)
    data, _ := json.Marshal(map[string]interface{}{
      "symbol":   order["s"].(string),
      "order_id": orderId,
      "status":   order["X"].(string),
    })
    h.Nats.Publish(config.NATS_ORDERS_UPDATE, data)
    h.Nats.Flush()
  }
}

func (h *AccountHandler) start() (err error) {
  log.Println("stream start")

  var apiKey, apiSecret string
  var isTestNet bool
  if common.GetEnvInt("BINANCE_FUTURES_TESTNET_ENABLE") == 1 {
    apiKey = common.GetEnvString("BINANCE_FUTURES_TESTNET_API_KEY")
    apiSecret = common.GetEnvString("BINANCE_FUTURES_TESTNET_API_SECRET")
    isTestNet = true
  } else {
    apiKey = common.GetEnvString("BINANCE_FUTURES_STREAMS_API_KEY")
    apiSecret = common.GetEnvString("BINANCE_FUTURES_STREAMS_API_SECRET")
  }

  client := binance.NewFuturesClient(
    apiKey,
    apiSecret,
  )
  if isTestNet {
    client.BaseURL = common.GetEnvString("BINANCE_FUTURES_TESTNET_API_ENDPOINT")
  } else {
    client.BaseURL = common.GetEnvString("BINANCE_FUTURES_API_ENDPOINT")
  }
  listenKey, err := client.NewStartUserStreamService().Do(h.Ctx)
  if err != nil {
    return err
  }
  defer client.NewCloseUserStreamService().ListenKey(listenKey).Do(h.Ctx)

  var endpoint string
  if isTestNet {
    endpoint = fmt.Sprintf(
      "%s/ws/%s",
      os.Getenv("BINANCE_FUTURES_TESTNET_STREAMS_ENDPOINT"),
      listenKey,
    )
  } else {
    endpoint = fmt.Sprintf(
      "%s/ws/%s",
      os.Getenv("BINANCE_FUTURES_STREAMS_ENDPOINT"),
      listenKey,
    )
  }

  log.Println("endpoint", endpoint, client.BaseURL, isTestNet, listenKey)

  h.Socket, _, err = websocket.Dial(h.Ctx, endpoint, &websocket.DialOptions{
    CompressionMode: websocket.CompressionDisabled,
  })
  if err != nil {
    return
  }
  defer h.Socket.Close(websocket.StatusInternalError, "the socket was closed abruptly")

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
