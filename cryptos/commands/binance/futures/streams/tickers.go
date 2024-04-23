package streams

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "log"
  "os"
  "strconv"
  "strings"
  "time"

  "gorm.io/gorm"
  "nhooyr.io/websocket"

  "github.com/nats-io/nats.go"
  "github.com/shopspring/decimal"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures/streams"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type TickersHandler struct {
  Db                 *gorm.DB
  Ctx                context.Context
  Socket             *websocket.Conn
  Nats               *nats.Conn
  TickersJob         *jobs.Tickers
  TradingsRepository *repositories.TradingsRepository
}

func NewTickersCommand() *cli.Command {
  var h TickersHandler
  return &cli.Command{
    Name:  "tickers",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TickersHandler{
        Db:         common.NewDB(2),
        Ctx:        context.Background(),
        Nats:       common.NewNats(),
        TickersJob: &jobs.Tickers{},
      }
      h.TradingsRepository = &repositories.TradingsRepository{
        Db: h.Db,
      }
      h.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
        Db: h.Db,
      }
      h.TradingsRepository.TriggersRepository = &tradingsRepositories.TriggersRepository{
        Db: h.Db,
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      current, _ := strconv.Atoi(c.Args().Get(0))
      if current < 1 {
        log.Fatal("current is less than 1")
        return nil
      }
      if err := h.Start(current); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *TickersHandler) read() (message map[string]interface{}, err error) {
  var data []byte
  _, data, err = h.Socket.Read(h.Ctx)
  json.Unmarshal(data, &message)
  return
}

func (h *TickersHandler) ping() error {
  ctx, cancel := context.WithTimeout(h.Ctx, time.Second*1)
  defer cancel()
  return h.Socket.Ping(ctx)
}

func (h *TickersHandler) handler(message map[string]interface{}) {
  data := message["data"].(map[string]interface{})
  event := data["e"].(string)

  if event == "24hrMiniTicker" {
    open, _ := strconv.ParseFloat(data["o"].(string), 64)
    price, _ := strconv.ParseFloat(data["c"].(string), 64)
    high, _ := strconv.ParseFloat(data["h"].(string), 64)
    low, _ := strconv.ParseFloat(data["l"].(string), 64)
    volume, _ := strconv.ParseFloat(data["v"].(string), 64)
    quota, _ := strconv.ParseFloat(data["q"].(string), 64)
    change, _ := decimal.NewFromFloat(price).Sub(decimal.NewFromFloat(open)).Div(decimal.NewFromFloat(open)).Round(4).Float64()
    data, _ := json.Marshal(map[string]interface{}{
      "symbol":    data["s"].(string),
      "price":     price,
      "open":      open,
      "high":      high,
      "low":       low,
      "volume":    volume,
      "quota":     quota,
      "change":    change,
      "timestamp": time.Now().Unix(),
    })
    h.Nats.Publish(config.NATS_TICKERS_UPDATE, data)
    h.Nats.Flush()
  }
}

func (h *TickersHandler) Start(current int) (err error) {
  log.Println("stream start")

  symbols := h.Scan()

  offset := (current - 1) * 33
  if offset >= len(symbols) {
    err = errors.New("symbols out of range")
    return
  }
  endPos := offset + 33
  if endPos > len(symbols) {
    endPos = len(symbols)
  }

  var streams []string
  for _, symbol := range symbols[offset:endPos] {
    streams = append(
      streams,
      fmt.Sprintf("%s@miniTicker", strings.ToLower(symbol)),
    )
  }

  if len(streams) < 1 {
    return errors.New("streams empty")
  }

  endpoint := fmt.Sprintf(
    "%s/stream?streams=%s",
    os.Getenv("BINANCE_FUTURES_STREAMS_ENDPOINT"),
    strings.Join(streams, "/"),
  )
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
      err = h.ping()
      if err != nil {
        h.Socket.Close(websocket.StatusNormalClosure, "")
        return
      }
      if len(symbols) != len(h.Scan()) {
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

func (h *TickersHandler) Scan() []string {
  var symbols []string
  for _, symbol := range h.TradingsRepository.Scan() {
    if !h.contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  return symbols
}

func (h *TickersHandler) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}
