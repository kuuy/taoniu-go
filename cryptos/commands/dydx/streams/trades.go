package streams

import (
  "context"
  "encoding/json"
  "errors"
  "log"
  "nhooyr.io/websocket/wsjson"
  "os"
  "strconv"
  config "taoniu.local/cryptos/config/dydx"
  "time"

  "gorm.io/gorm"
  "nhooyr.io/websocket"

  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
  tradingsRepositories "taoniu.local/cryptos/repositories/dydx/tradings"
)

type TradesHandler struct {
  Db      *gorm.DB
  Ctx     context.Context
  Socket  *websocket.Conn
  Nats    *nats.Conn
  Symbols []string
  //TradesJob       *jobs.Trades
  TradingsRepository *repositories.TradingsRepository
}

func NewTradesCommand() *cli.Command {
  var h TradesHandler
  return &cli.Command{
    Name:  "trades",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TradesHandler{
        Db:   common.NewDB(1),
        Ctx:  context.Background(),
        Nats: common.NewNats(),
        //TradesJob: &jobs.Trades{},
      }
      h.TradingsRepository = &repositories.TradingsRepository{
        Db: h.Db,
      }
      h.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
        Db: h.Db,
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      no, _ := strconv.Atoi(c.Args().Get(0))
      if no < 1 {
        log.Fatal("no is less than 1")
        return nil
      }
      if err := h.Run(no); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *TradesHandler) read() (message map[string]interface{}, err error) {
  var data []byte
  _, data, err = h.Socket.Read(h.Ctx)
  json.Unmarshal(data, &message)
  return
}

func (h *TradesHandler) ping() error {
  ctx, cancel := context.WithTimeout(h.Ctx, time.Second*1)
  defer cancel()
  return h.Socket.Ping(ctx)
}

func (h *TradesHandler) handler(message map[string]interface{}) {
  if message["type"] == "connected" {
    for _, symbol := range h.Symbols {
      wsjson.Write(h.Ctx, h.Socket, map[string]interface{}{
        "type":           "subscribe",
        "channel":        "v3_trades",
        "id":             symbol,
        "includeOffsets": false,
      })
    }
  }
  if message["type"] == "channel_data" && message["channel"] == "v3_trades" {
    contents := message["contents"].(map[string]interface{})
    trades := contents["trades"].([]interface{})
    trade := trades[0].(map[string]interface{})
    price, _ := strconv.ParseFloat(trade["price"].(string), 64)
    data, _ := json.Marshal(map[string]interface{}{
      "symbol": message["id"].(string),
      "price":  price,
      "side":   trade["side"],
    })
    h.Nats.Publish(config.NATS_TRADES_UPDATE, data)
    h.Nats.Flush()
  }
}

func (h *TradesHandler) Run(current int) (err error) {
  log.Println("stream start")

  symbols := h.Scan()

  offset := (current - 1) * 25
  if offset >= len(symbols) {
    err = errors.New("symbols out of range")
    return
  }
  endPos := offset + 25
  if endPos > len(symbols) {
    endPos = len(symbols)
  }

  if len(symbols[offset:endPos]) < 1 {
    return errors.New("streams empty")
  }

  h.Symbols = symbols[offset:endPos]

  endpoint := os.Getenv("DYDX_STREAMS_ENDPOINT")
  log.Println("endpoint", endpoint)

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
      if len(symbols) != len(h.Scan()) {
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

func (h *TradesHandler) Scan() []string {
  var symbols []string
  for _, symbol := range h.TradingsRepository.Scan() {
    if !h.contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  return symbols
}

func (h *TradesHandler) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}
