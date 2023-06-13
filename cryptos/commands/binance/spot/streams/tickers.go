package streams

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "log"
  "strconv"
  "strings"
  config "taoniu.local/cryptos/config/binance/spot"
  "time"

  "gorm.io/gorm"
  "nhooyr.io/websocket"

  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures/streams"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
  crossRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross"
  crossTradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross/tradings"
  isolatedRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
  isolatedTradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type TickersHandler struct {
  Db                         *gorm.DB
  Ctx                        context.Context
  Socket                     *websocket.Conn
  Nats                       *nats.Conn
  TickersJob                 *jobs.Tickers
  TradingsRepository         *repositories.TradingsRepository
  CrossTradingsRepository    *crossRepositories.TradingsRepository
  IsolatedTradingsRepository *isolatedRepositories.TradingsRepository
}

func NewTickersCommand() *cli.Command {
  var h TickersHandler
  return &cli.Command{
    Name:  "tickers",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TickersHandler{
        Db:         common.NewDB(),
        Ctx:        context.Background(),
        Nats:       common.NewNats(),
        TickersJob: &jobs.Tickers{},
      }
      h.TradingsRepository = &repositories.TradingsRepository{
        Db: h.Db,
      }
      h.TradingsRepository.FishersRepository = &tradingsRepositories.FishersRepository{
        Db: h.Db,
      }
      h.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
        Db: h.Db,
      }
      h.TradingsRepository.TriggersRepository = &tradingsRepositories.TriggersRepository{
        Db: h.Db,
      }
      h.CrossTradingsRepository = &crossRepositories.TradingsRepository{
        Db: h.Db,
      }
      h.CrossTradingsRepository.TriggersRepository = &crossTradingsRepositories.TriggersRepository{
        Db: h.Db,
      }
      h.IsolatedTradingsRepository = &isolatedRepositories.TradingsRepository{
        Db: h.Db,
      }
      h.IsolatedTradingsRepository.FishersRepository = &isolatedTradingsRepositories.FishersRepository{
        Db: h.Db,
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
    data, _ := json.Marshal(map[string]interface{}{
      "symbol":    data["s"].(string),
      "price":     price,
      "open":      open,
      "high":      high,
      "low":       low,
      "volume":    volume,
      "quota":     quota,
      "timestamp": time.Now().Unix(),
    })
    h.Nats.Publish(config.NATS_TICKERS_UPDATE, data)
    h.Nats.Flush()
  }
}

func (h *TickersHandler) start() (err error) {
  log.Println("stream start")

  var streams []string
  for _, symbol := range h.Scan() {
    streams = append(
      streams,
      fmt.Sprintf("%s@miniTicker", strings.ToLower(symbol)),
    )
  }

  if len(streams) < 1 {
    return errors.New("streams empty")
  }

  endpoint := fmt.Sprintf("wss://stream.binance.com/stream?streams=%s", strings.Join(streams, "/"))

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
      if len(streams) != len(h.Scan()) {
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

func (h *TickersHandler) Scan() []string {
  var symbols []string
  for _, symbol := range h.TradingsRepository.Scan() {
    if !h.contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  for _, symbol := range h.CrossTradingsRepository.Scan() {
    if !h.contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  for _, symbol := range h.IsolatedTradingsRepository.Scan() {
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