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

  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/shopspring/decimal"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures/streams"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type TickersHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  Socket             *websocket.Conn
  Nats               *nats.Conn
  TickersJob         *jobs.Tickers
  ScalpingRepository *repositories.ScalpingRepository
}

func NewTickersCommand() *cli.Command {
  var h TickersHandler
  return &cli.Command{
    Name:  "tickers",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TickersHandler{
        Db:         common.NewDB(2),
        Rdb:        common.NewRedis(2),
        Ctx:        context.Background(),
        Nats:       common.NewNats(),
        TickersJob: &jobs.Tickers{},
      }
      h.ScalpingRepository = &repositories.ScalpingRepository{
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
    symbol := data["s"].(string)
    open, _ := strconv.ParseFloat(data["o"].(string), 64)
    price, _ := strconv.ParseFloat(data["c"].(string), 64)
    high, _ := strconv.ParseFloat(data["h"].(string), 64)
    low, _ := strconv.ParseFloat(data["l"].(string), 64)
    volume, _ := strconv.ParseFloat(data["v"].(string), 64)
    quota, _ := strconv.ParseFloat(data["q"].(string), 64)
    change, _ := decimal.NewFromFloat(price).Sub(decimal.NewFromFloat(open)).Div(decimal.NewFromFloat(open)).Round(4).Float64()

    h.Rdb.HMSet(
      h.Ctx,
      fmt.Sprintf(config.REDIS_KEY_TICKERS, symbol),
      map[string]interface{}{
        "symbol":    symbol,
        "open":      open,
        "price":     price,
        "change":    change,
        "high":      high,
        "low":       low,
        "volume":    volume,
        "quota":     quota,
        "timestamp": time.Now().UnixMilli(),
      },
    )
  }
}

func (h *TickersHandler) Start(current int) (err error) {
  log.Println("streams tickers current", current)

  symbols := h.ScalpingRepository.Scan(2)

  pageSize := common.GetEnvInt("BINANCE_FUTURES_SYMBOLS_SIZE")
  offset := (current - 1) * pageSize
  if offset >= len(symbols) {
    err = errors.New("symbols out of range")
    return
  }
  endPos := offset + pageSize
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
    case <-h.Ctx.Done():
      return
    case <-quit:
      return
    }
  }
}
