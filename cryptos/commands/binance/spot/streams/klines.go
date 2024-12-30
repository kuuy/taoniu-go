package streams

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "log"
  "os"
  "slices"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "nhooyr.io/websocket"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot/streams"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type KlinesHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Socket             *websocket.Conn
  Asynq              *asynq.Client
  Ctx                context.Context
  AnsqContext        *common.AnsqClientContext
  Job                *jobs.Klines
  ScalpingRepository *repositories.ScalpingRepository
}

func NewKlinesCommand() *cli.Command {
  var h KlinesHandler
  return &cli.Command{
    Name:  "klines",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = KlinesHandler{
        Db:    common.NewDB(1),
        Rdb:   common.NewRedis(1),
        Asynq: common.NewAsynqClient("BINANCE_SPOT"),
        Ctx:   context.Background(),
      }
      h.AnsqContext = &common.AnsqClientContext{
        Db:   h.Db,
        Rdb:  h.Rdb,
        Ctx:  h.Ctx,
        Conn: h.Asynq,
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
      interval := c.Args().Get(1)
      if !slices.Contains([]string{"1m", "15m", "4h", "1d"}, interval) {
        log.Fatal("interval not valid")
        return nil
      }
      if err := h.Start(current, interval); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *KlinesHandler) read() (message map[string]interface{}, err error) {
  var data []byte
  _, data, err = h.Socket.Read(h.Ctx)
  json.Unmarshal(data, &message)
  return
}

func (h *KlinesHandler) ping() error {
  ctx, cancel := context.WithTimeout(h.Ctx, time.Second*1)
  defer cancel()
  return h.Socket.Ping(ctx)
}

func (h *KlinesHandler) handler(message map[string]interface{}) {
  data := message["data"].(map[string]interface{})
  kline := data["k"].(map[string]interface{})
  event := data["e"].(string)

  log.Println("kline", data)
  if event == "kline" {
    symbol := data["s"].(string)
    interval := kline["i"].(string)
    open, _ := strconv.ParseFloat(kline["o"].(string), 64)
    close, _ := strconv.ParseFloat(kline["c"].(string), 64)
    high, _ := strconv.ParseFloat(kline["h"].(string), 64)
    low, _ := strconv.ParseFloat(kline["l"].(string), 64)
    volume, _ := strconv.ParseFloat(kline["v"].(string), 64)
    quota, _ := strconv.ParseFloat(kline["q"].(string), 64)
    timestamp := int64(kline["t"].(float64))

    mutex := common.NewMutex(
      h.AnsqContext.Rdb,
      h.AnsqContext.Ctx,
      fmt.Sprintf(config.LOCKS_KLINES_STREAM, symbol, interval),
    )
    if !mutex.Lock(10 * time.Second) {
      return
    }

    task, err := h.Job.Update(
      symbol,
      interval,
      open,
      close,
      high,
      low,
      volume,
      quota,
      timestamp,
    )
    if err != nil {
      log.Println("err", err)
      return
    }
    _, err = h.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_KLINES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
    if err != nil {
      log.Println("err", err)
      return
    }
  }
}

func (h *KlinesHandler) Start(current int, interval string) (err error) {
  log.Println("stream start")

  symbols := h.ScalpingRepository.Scan()

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
