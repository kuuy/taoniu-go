package streams

import (
  "context"
  "errors"
  "fmt"
  "log"
  "net/http"
  "os"
  "slices"
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

type KlinesHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Socket             *websocket.Conn
  Ctx                context.Context
  ScalpingRepository *repositories.ScalpingRepository
}

func NewKlinesCommand() *cli.Command {
  var h KlinesHandler
  return &cli.Command{
    Name:  "klines",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = KlinesHandler{
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
      interval := c.Args().Get(0)
      if !slices.Contains([]string{"1m", "15m", "4h", "1d"}, interval) {
        log.Fatal("interval not valid")
        return nil
      }
      current, _ := strconv.Atoi(c.Args().Get(1))
      if current < 1 {
        log.Fatal("current is less than 1")
        return nil
      }
      if err := h.Start(interval, current); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *KlinesHandler) read() (message map[string]interface{}, err error) {
  ctx, cancel := context.WithTimeout(h.Ctx, 10*time.Second)
  defer cancel()
  err = wsjson.Read(ctx, h.Socket, &message)
  return
}

func (h *KlinesHandler) ping() (err error) {
  ctx, cancel := context.WithTimeout(h.Ctx, 25*time.Second)
  defer cancel()
  err = h.Socket.Ping(ctx)
  if err != nil {
    log.Println("ping error", err.Error())
  }
  return
}

func (h *KlinesHandler) handler(message map[string]interface{}) {
  data := message["data"].(map[string]interface{})
  kline := data["k"].(map[string]interface{})
  event := data["e"].(string)

  if event == "kline" {
    symbol := data["s"].(string)
    interval := kline["i"].(string)
    open, _ := strconv.ParseFloat(kline["o"].(string), 64)
    close, _ := strconv.ParseFloat(kline["c"].(string), 64)
    high, _ := strconv.ParseFloat(kline["h"].(string), 64)
    low, _ := strconv.ParseFloat(kline["l"].(string), 64)
    volume, _ := strconv.ParseFloat(kline["v"].(string), 64)
    quota, _ := strconv.ParseFloat(kline["q"].(string), 64)
    change, _ := decimal.NewFromFloat(close).Sub(decimal.NewFromFloat(open)).Div(decimal.NewFromFloat(open)).Round(4).Float64()
    timestamp := int64(kline["t"].(float64))

    duration := time.Second * 0
    if interval == "1m" {
      duration = time.Second * (30 + 60)
    } else if interval == "15m" {
      duration = time.Second * (30 + 900)
    } else if interval == "4h" {
      duration = time.Second * (30 + 14400)
    } else if interval == "1d" {
      duration = time.Second * (30 + 86400)
    }

    redisKey := fmt.Sprintf(config.REDIS_KEY_KLINES, interval, symbol, timestamp)
    h.Rdb.HMSet(
      h.Ctx,
      redisKey,
      map[string]interface{}{
        "symbol":    symbol,
        "open":      open,
        "close":     close,
        "change":    change,
        "high":      high,
        "low":       low,
        "volume":    volume,
        "quota":     quota,
        "timestamp": timestamp,
        "lasttime":  time.Now().UnixMilli(),
      },
    )

    ttl, _ := h.Rdb.TTL(h.Ctx, redisKey).Result()
    if -1 == ttl.Nanoseconds() {
      h.Rdb.Expire(h.Ctx, redisKey, duration)
    }
  }
}

func (h *KlinesHandler) Start(interval string, current int) (err error) {
  log.Println("stream start")

  symbols := h.ScalpingRepository.Scan()

  pageSize := common.GetEnvInt("BINANCE_SPOT_SYMBOLS_SIZE")
  startPos := (current - 1) * pageSize
  if startPos >= len(symbols) {
    err = errors.New("symbols out of range")
    return
  }
  endPos := startPos + pageSize
  if endPos > len(symbols) {
    endPos = len(symbols)
  }

  var streams []string
  for _, symbol := range symbols[startPos:endPos] {
    streams = append(
      streams,
      fmt.Sprintf("%s@kline_%s", strings.ToLower(symbol), interval),
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

  var httpClient *http.Client

  proxy := common.GetEnvString(fmt.Sprintf("BINANCE_PROXY_%v", current))
  if proxy != "" {
    tr := &http.Transport{}
    tr.DialContext = (&common.ProxySession{
      Proxy: proxy,
    }).DialContext
    httpClient = &http.Client{
      Transport: tr,
    }
  }

  h.Socket, _, err = websocket.Dial(h.Ctx, endpoint, &websocket.DialOptions{
    HTTPClient:      httpClient,
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
      go h.handler(message)
    }
  }
}
