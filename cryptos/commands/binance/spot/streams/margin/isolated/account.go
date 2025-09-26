package isolated

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "os"
  "time"

  "github.com/adshao/go-binance/v2"
  "github.com/coder/websocket"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot/streams"
)

type AccountHandler struct {
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
        Ctx:        context.Background(),
        Nats:       common.NewNats(),
        AccountJob: &jobs.Account{},
        OrdersJob:  &jobs.Orders{},
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      symbol := c.Args().Get(0)
      if symbol == "" {
        log.Fatal("symbol can not be empty")
        return nil
      }
      if err := h.start(symbol); err != nil {
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

  log.Println("event", event)
  //if event == "outboundAccountPosition" {
  //  for _, item := range message["B"].([]interface{}) {
  //    account := item.(map[string]interface{})
  //    if account["a"].(string) != "USDT" {
  //      continue
  //    }
  //    free, _ := strconv.ParseFloat(fmt.Sprintf("%v", account["f"]), 64)
  //    locked, _ := strconv.ParseFloat(fmt.Sprintf("%v", account["l"]), 64)
  //    data, _ := json.Marshal(map[string]interface{}{
  //      "asset":  account["a"].(string),
  //      "free":   free,
  //      "locked": locked,
  //    })
  //    h.Nats.Publish(config.NATS_ACCOUNT_UPDATE, data)
  //    h.Nats.Flush()
  //  }
  //}
}

func (h *AccountHandler) start(symbol string) (err error) {
  log.Println("stream start")

  client := binance.NewClient(
    os.Getenv("BINANCE_SPOT_STREAMS_API_KEY"),
    os.Getenv("BINANCE_SPOT_STREAMS_API_SECRET"),
  )
  client.BaseURL = os.Getenv("BINANCE_SPOT_API_ENDPOINT")

  listenKey, err := client.NewStartIsolatedMarginUserStreamService().Symbol(symbol).Do(h.Ctx)
  if err != nil {
    return err
  }
  defer client.NewCloseIsolatedMarginUserStreamService().ListenKey(listenKey).Symbol(symbol).Do(h.Ctx)

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
