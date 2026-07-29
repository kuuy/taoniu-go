package streams

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "log"
  "net/http"
  "os"
  "os/signal"
  "strconv"
  "strings"
  "syscall"
  "time"

  "github.com/coder/websocket"
  "github.com/coder/websocket/wsjson"
  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type TickersHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Nats               *nats.Conn
  Ctx                context.Context
  cancel             context.CancelFunc
  Socket             *websocket.Conn
  ScalpingRepository *repositories.ScalpingRepository
  workerChan         chan map[string]interface{}
  Symbols            []string
}

func NewTickersCommand() *cli.Command {
  var h TickersHandler
  return &cli.Command{
    Name:  "tickers",
    Usage: "",
    Before: func(c *cli.Context) error {
      ctx, cancel := context.WithCancel(context.Background())
      h = TickersHandler{
        Db:         common.NewDB(2),
        Rdb:        common.NewRedis(2),
        Nats:       common.NewNats(2),
        Ctx:        ctx,
        cancel:     cancel,
        workerChan: make(chan map[string]interface{}, 1024),
      }
      h.ScalpingRepository = &repositories.ScalpingRepository{
        Db: h.Db,
      }
      return nil
    },
    After: func(c *cli.Context) error {
      h.Rdb.Close()
      h.Nats.Close()
      return nil
    },
    Action: func(c *cli.Context) error {
      current, _ := strconv.Atoi(c.Args().Get(0))
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

      for i := 0; i < 8; i++ {
        go h.worker()
      }

      for {
        select {
        case <-h.Ctx.Done():
          return nil
        default:
          if err := h.Start(current); err != nil {
            log.Printf("stream error: %v, reconnecting in 5s...", err)
            time.Sleep(5 * time.Second)
          }
        }
      }
    },
  }
}

func (h *TickersHandler) Start(current int) error {
  if len(h.Symbols) == 0 {
    symbols := h.ScalpingRepository.Scan(2)
    if len(symbols) == 0 {
      return errors.New("no symbols found")
    }
    h.Symbols = symbols
    if sqlDB, err := h.Db.DB(); err == nil {
      sqlDB.Close()
    }
  }

  pageSize := common.GetEnvInt("BINANCE_FUTURES_SYMBOLS_SIZE")
  if pageSize <= 0 {
    pageSize = 100
  }

  offset := (current - 1) * pageSize
  if offset >= len(h.Symbols) {
    return errors.New("symbols out of range")
  }

  endPos := offset + pageSize
  if endPos > len(h.Symbols) {
    endPos = len(h.Symbols)
  }

  streams := make([]string, 0, endPos-offset)
  for _, symbol := range h.Symbols[offset:endPos] {
    streams = append(streams, fmt.Sprintf("%s@miniTicker", strings.ToLower(symbol)))
  }

  endpoint := fmt.Sprintf(
    "%s/market/stream?streams=%s",
    os.Getenv("BINANCE_FUTURES_STREAMS_ENDPOINT"),
    strings.Join(streams, "/"),
  )
  log.Printf("connecting to endpoint: %s", endpoint)

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

  connCtx, connCancel := context.WithCancel(h.Ctx)
  defer connCancel()

  go h.pingLoop(connCtx)

  log.Printf("stream started for index %d (%d symbols)", current, len(streams))

  for {
    var message map[string]interface{}
    err = wsjson.Read(connCtx, h.Socket, &message)

    if err != nil {
      if errors.Is(err, context.Canceled) {
        return nil
      }
      return fmt.Errorf("read error: %w", err)
    }

    select {
    case h.workerChan <- message:
    default:
      log.Println("worker channel full, dropping ticker message")
    }
  }
}

func (h *TickersHandler) worker() {
  for {
    select {
    case <-h.Ctx.Done():
      return
    case message := <-h.workerChan:
      h.processMessage(message)
    }
  }
}

func (h *TickersHandler) processMessage(message map[string]interface{}) {
  data, ok := message["data"].(map[string]interface{})
  if !ok {
    return
  }

  event, _ := data["e"].(string)
  if event != "24hrMiniTicker" {
    return
  }

  symbol, _ := data["s"].(string)
  priceStr, _ := data["c"].(string)
  openStr, _ := data["o"].(string)
  highStr, _ := data["h"].(string)
  lowStr, _ := data["l"].(string)
  volumeStr, _ := data["v"].(string)
  quotaStr, _ := data["q"].(string)

  price, _ := strconv.ParseFloat(priceStr, 64)
  open, _ := strconv.ParseFloat(openStr, 64)
  high, _ := strconv.ParseFloat(highStr, 64)
  low, _ := strconv.ParseFloat(lowStr, 64)
  volume, _ := strconv.ParseFloat(volumeStr, 64)
  quota, _ := strconv.ParseFloat(quotaStr, 64)

  var change float64
  if open > 0 {
    change = (price - open) / open
    change = float64(int(change*10000)) / 10000
  }

  timestamp := time.Now().UnixMilli()

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
      "timestamp": timestamp,
    },
  )

  if h.Nats != nil {
    payload, _ := json.Marshal(map[string]interface{}{
      "symbol":    symbol,
      "open":      open,
      "price":     price,
      "high":      high,
      "low":       low,
      "volume":    volume,
      "quota":     quota,
      "timestamp": timestamp,
    })
    h.Nats.Publish(config.NATS_TICKERS_UPDATE, payload)
    h.Nats.Flush()
  }
}

func (h *TickersHandler) pingLoop(ctx context.Context) {
  ticker := time.NewTicker(20 * time.Second)
  defer ticker.Stop()

  for {
    select {
    case <-ctx.Done():
      return
    case <-ticker.C:
      if h.Socket == nil {
        return
      }
      pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
      err := h.Socket.Ping(pingCtx)
      pingCancel()
      if err != nil {
        log.Printf("ping error: %v", err)
        return
      }
    }
  }
}
