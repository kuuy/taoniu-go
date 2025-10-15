package streams

import (
  "context"
  "errors"
  "fmt"
  "log"
  "os"
  "strconv"
  "strings"
  "time"

  "github.com/coder/websocket"
  "github.com/coder/websocket/wsjson"
  "github.com/go-redis/redis/v8"
  "github.com/shopspring/decimal"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type TickersHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  Socket             *websocket.Conn
  ScalpingRepository *repositories.ScalpingRepository
}

func NewTickersCommand() *cli.Command {
  var h TickersHandler
  return &cli.Command{
    Name:  "tickers",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TickersHandler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
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
  err = wsjson.Read(h.Ctx, h.Socket, &message)
  return
}

func (h *TickersHandler) ping() (err error) {
  ctx, cancel := context.WithTimeout(h.Ctx, 25*time.Second)
  defer cancel()
  err = h.Socket.Ping(ctx)
  if err != nil {
    log.Println("ping error", err.Error())
  }
  return
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

  symbols := h.ScalpingRepository.Scan()

  pageSize := common.GetEnvInt("BINANCE_SPOT_SYMBOLS_SIZE")
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
    os.Getenv("BINANCE_SPOT_STREAMS_ENDPOINT"),
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

  go func() {
    ticker := time.NewTicker(30 * time.Second)
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
