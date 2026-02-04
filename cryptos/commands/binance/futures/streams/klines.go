package streams

import (
  "context"
  "errors"
  "fmt"
  "log"
  "net/http"
  "os"
  "os/signal"
  "slices"
  "strconv"
  "strings"
  "syscall"
  "time"

  "github.com/coder/websocket"
  "github.com/coder/websocket/wsjson"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type KlinesHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  cancel             context.CancelFunc
  Socket             *websocket.Conn
  ScalpingRepository *repositories.ScalpingRepository
  workerChan         chan map[string]interface{}
}

func NewKlinesCommand() *cli.Command {
  var h KlinesHandler
  return &cli.Command{
    Name:  "klines",
    Usage: "",
    Before: func(c *cli.Context) error {
      ctx, cancel := context.WithCancel(context.Background())
      h = KlinesHandler{
        Db:         common.NewDB(2),
        Rdb:        common.NewRedis(2),
        Ctx:        ctx,
        cancel:     cancel,
        workerChan: make(chan map[string]interface{}, 2048),
      }
      h.ScalpingRepository = &repositories.ScalpingRepository{
        Db: h.Db,
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      interval := c.Args().Get(0)
      if !slices.Contains([]string{"1m", "15m", "4h", "1d"}, interval) {
        return errors.New("invalid interval")
      }
      current, _ := strconv.Atoi(c.Args().Get(1))
      if current < 1 {
        return errors.New("current index must be >= 1")
      }

      sigChan := make(chan os.Signal, 1)
      signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
      go func() {
        sig := <-sigChan
        log.Printf("received signal %v, shutting down...", sig)
        h.cancel()
      }()

      for i := 0; i < 16; i++ {
        go h.worker()
      }

      for {
        select {
        case <-h.Ctx.Done():
          return nil
        default:
          if err := h.Start(interval, current); err != nil {
            log.Printf("klines stream error: %v, reconnecting in 5s...", err)
            time.Sleep(5 * time.Second)
          }
        }
      }
    },
  }
}

func (h *KlinesHandler) Start(interval string, current int) error {
  symbols := h.ScalpingRepository.Scan(2)
  pageSize := common.GetEnvInt("BINANCE_FUTURES_SYMBOLS_SIZE")
  if pageSize <= 0 {
    pageSize = 50
  }

  offset := (current - 1) * pageSize
  if offset >= len(symbols) {
    return errors.New("symbols out of range")
  }

  endPos := offset + pageSize
  if endPos > len(symbols) {
    endPos = len(symbols)
  }

  var streams []string
  for _, symbol := range symbols[offset:endPos] {
    streams = append(streams, fmt.Sprintf("%s@kline_%s", strings.ToLower(symbol), interval))
  }

  endpoint := fmt.Sprintf(
    "%s/stream?streams=%s",
    os.Getenv("BINANCE_FUTURES_STREAMS_ENDPOINT"),
    strings.Join(streams, "/"),
  )

  var httpClient *http.Client
  proxy := common.GetEnvString(fmt.Sprintf("BINANCE_PROXY_%v", current))
  if proxy != "" {
    tr := &http.Transport{}
    tr.DialContext = (&common.ProxySession{Proxy: proxy}).DialContext
    httpClient = &http.Client{Transport: tr}
  }

  dialCtx, dialCancel := context.WithTimeout(h.Ctx, 30*time.Second)
  defer dialCancel()

  var err error
  h.Socket, _, err = websocket.Dial(dialCtx, endpoint, &websocket.DialOptions{
    HTTPClient:      httpClient,
    CompressionMode: websocket.CompressionDisabled,
  })
  if err != nil {
    return fmt.Errorf("dial error: %w", err)
  }
  defer h.Socket.Close(websocket.StatusNormalClosure, "")

  go h.pingLoop()

  log.Printf("klines stream [%s] started for index %d", interval, current)

  for {
    var message map[string]interface{}
    readCtx, readCancel := context.WithTimeout(h.Ctx, 10*time.Second)
    err = wsjson.Read(readCtx, h.Socket, &message)
    readCancel()

    if err != nil {
      if errors.Is(err, context.Canceled) {
        return nil
      }
      return fmt.Errorf("read error: %w", err)
    }

    select {
    case h.workerChan <- message:
    default:
      log.Println("klines worker channel full, dropping message")
    }
  }
}

func (h *KlinesHandler) worker() {
  for {
    select {
    case <-h.Ctx.Done():
      return
    case message := <-h.workerChan:
      h.processMessage(message)
    }
  }
}

func (h *KlinesHandler) processMessage(message map[string]interface{}) {
  data, ok := message["data"].(map[string]interface{})
  if !ok {
    return
  }
  kline, ok := data["k"].(map[string]interface{})
  if !ok {
    return
  }
  if data["e"].(string) != "kline" {
    return
  }

  symbol := data["s"].(string)
  interval := kline["i"].(string)
  open, _ := strconv.ParseFloat(kline["o"].(string), 64)
  close, _ := strconv.ParseFloat(kline["c"].(string), 64)
  high, _ := strconv.ParseFloat(kline["h"].(string), 64)
  low, _ := strconv.ParseFloat(kline["l"].(string), 64)
  volume, _ := strconv.ParseFloat(kline["v"].(string), 64)
  quota, _ := strconv.ParseFloat(kline["q"].(string), 64)
  timestamp := int64(kline["t"].(float64))

  var change float64
  if open > 0 {
    change = (close - open) / open
    change = float64(int(change*10000)) / 10000
  }

  var expiration time.Duration
  switch interval {
  case "1m":
    expiration = 1*time.Minute + 30*time.Second
  case "15m":
    expiration = 15*time.Minute + 30*time.Second
  case "4h":
    expiration = 4*time.Hour + 30*time.Second
  case "1d":
    expiration = 24*time.Hour + 30*time.Second
  default:
    expiration = 30 * time.Second
  }

  redisKey := fmt.Sprintf(config.REDIS_KEY_KLINES, interval, symbol, timestamp)

  h.Rdb.HMSet(h.Ctx, redisKey, map[string]interface{}{
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
  })

  ttl, _ := h.Rdb.TTL(h.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    h.Rdb.Expire(h.Ctx, redisKey, expiration)
  }
}

func (h *KlinesHandler) pingLoop() {
  ticker := time.NewTicker(20 * time.Second)
  defer ticker.Stop()
  for {
    select {
    case <-h.Ctx.Done():
      return
    case <-ticker.C:
      if h.Socket == nil {
        return
      }
      ctx, cancel := context.WithTimeout(h.Ctx, 5*time.Second)
      _ = h.Socket.Ping(ctx)
      cancel()
    }
  }
}
