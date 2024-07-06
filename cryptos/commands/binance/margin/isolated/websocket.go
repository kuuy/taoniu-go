package isolated

import (
  "context"
  "fmt"
  "log"
  "os"
  "strconv"

  "github.com/adshao/go-binance/v2"
  "github.com/bitly/go-simplejson"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "nhooyr.io/websocket"

  "taoniu.local/cryptos/common"
)

type WebsocketHandler struct {
  Rdb *redis.Client
  Ctx context.Context
}

func NewWebsocketCommand() *cli.Command {
  var h WebsocketHandler
  return &cli.Command{
    Name:  "websocket",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = WebsocketHandler{
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      if c.NArg() < 1 {
        return nil
      }
      symbol := c.Args().Get(0)
      if err := h.start(symbol); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *WebsocketHandler) newJSON(data []byte) (j *simplejson.Json, err error) {
  j, err = simplejson.NewJson(data)
  if err != nil {
    return nil, err
  }
  return j, nil
}

func (h *WebsocketHandler) handler(message []byte) {
  j, err := h.newJSON(message)
  if err != nil {
    panic(err)
  }

  event := j.Get("e").MustString()
  if event == "executionReport" {
    order := j.MustMap()
    symbol := fmt.Sprint(order["s"])
    orderId, _ := strconv.ParseInt(fmt.Sprint(order["i"]), 10, 64)
    status := order["X"]

    if status != "NEW" || status != "PARTIALLY_FILLED" {
      h.Rdb.SAdd(
        h.Ctx,
        "binance:spot:margin:orders:flush",
        fmt.Sprintf("%s,%d,1", symbol, orderId),
      )
    }
  }
}

func (h *WebsocketHandler) start(symbol string) error {
  client := binance.NewClient(
    os.Getenv("BINANCE_SPOT_STREAMS_API_KEY"),
    os.Getenv("BINANCE_SPOT_STREAMS_API_SECRET"),
  )
  client.BaseURL = os.Getenv("BINANCE_SPOT_API_ENDPOINT")

  listenKey, err := client.NewStartIsolatedMarginUserStreamService().Symbol(symbol).Do(h.Ctx)
  if err != nil {
    return err
  }
  log.Println("listenKey:", listenKey)
  defer client.NewCloseIsolatedMarginUserStreamService().ListenKey(listenKey).Symbol(symbol).Do(h.Ctx)

  endpoint := fmt.Sprintf("wss://stream.binance.com/ws/%s", listenKey)
  socket, _, err := websocket.Dial(h.Ctx, endpoint, nil)
  if err != nil {
    return err
  }
  socket.SetReadLimit(655350)

  for {
    _, message, readErr := socket.Read(h.Ctx)
    if readErr != nil {
      return readErr
    }
    h.handler(message)
  }

  return nil
}
