package streams

import (
  "context"
  "crypto/hmac"
  "crypto/sha256"
  "encoding/hex"
  "errors"
  "fmt"
  "log"
  "net/http"
  "os"
  "strconv"
  "time"

  "github.com/coder/websocket"
  "github.com/coder/websocket/wsjson"
  "github.com/go-redis/redis/v8"
  "github.com/google/uuid"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot/streams"
)

type AccountHandler struct {
  Rdb        *redis.Client
  Ctx        context.Context
  Socket     *websocket.Conn
  AccountJob *jobs.Account
  OrdersJob  *jobs.Orders
  workerChan chan map[string]interface{}
}

func NewAccountCommand() *cli.Command {
  var h AccountHandler
  return &cli.Command{
    Name:  "account",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AccountHandler{
        Rdb:        common.NewRedis(1),
        Ctx:        context.Background(),
        AccountJob: &jobs.Account{},
        OrdersJob:  &jobs.Orders{},
        workerChan: make(chan map[string]interface{}, 1024),
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      if err := h.Start(); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *AccountHandler) processMessage(message map[string]interface{}) {
  data, ok := message["event"].(map[string]interface{})
  if !ok {
    return
  }

  event, ok := data["e"].(string)
  if !ok {
    return
  }

  if event == "outboundAccountPosition" {
    for _, item := range data["B"].([]interface{}) {
      account := item.(map[string]interface{})
      asset := account["a"].(string)
      if asset != "USDT" {
        continue
      }
      free, _ := strconv.ParseFloat(fmt.Sprintf("%v", account["f"]), 64)
      locked, _ := strconv.ParseFloat(fmt.Sprintf("%v", account["l"]), 64)

      h.Rdb.HMSet(
        h.Ctx,
        fmt.Sprintf(config.REDIS_KEY_BALANCE, asset),
        map[string]interface{}{
          "free":   free,
          "locked": locked,
        },
      )
    }
  }

  if event == "executionReport" {
    //orderId, _ := strconv.ParseInt(fmt.Sprintf("%.0f", data["i"]), 10, 64)
    //data, _ := json.Marshal(map[string]interface{}{
    //  "symbol":   data["s"].(string),
    //  "order_id": orderId,
    //  "status":   data["X"].(string),
    //})
    //h.Nats.Publish(config.NATS_ORDERS_UPDATE, data)
    //h.Nats.Flush()
  }
}

func (h *AccountHandler) read() (message map[string]interface{}, err error) {
  ctx, cancel := context.WithTimeout(h.Ctx, 10*time.Second)
  defer cancel()
  err = wsjson.Read(ctx, h.Socket, &message)
  return
}

func (h *AccountHandler) ping() (err error) {
  ctx, cancel := context.WithTimeout(h.Ctx, 25*time.Second)
  defer cancel()
  err = h.Socket.Ping(ctx)
  if err != nil {
    log.Println("ping error", err.Error())
  }
  return
}

func (h *AccountHandler) handler(message map[string]interface{}) {
  raw, ok := message["event"].(map[string]interface{})
  if !ok {
    return
  }
  event, ok := raw["e"].(string)
  if !ok {
    return
  }

  if event == "outboundAccountPosition" {
    for _, item := range raw["B"].([]interface{}) {
      account := item.(map[string]interface{})
      asset := account["a"].(string)
      if asset != "USDT" {
        continue
      }
      free, _ := strconv.ParseFloat(fmt.Sprintf("%v", account["f"]), 64)
      locked, _ := strconv.ParseFloat(fmt.Sprintf("%v", account["l"]), 64)

      h.Rdb.HMSet(
        h.Ctx,
        fmt.Sprintf(config.REDIS_KEY_BALANCE, asset),
        map[string]interface{}{
          "free":   free,
          "locked": locked,
        },
      )
    }
  }

  if event == "executionReport" {
    //orderId, _ := strconv.ParseInt(fmt.Sprintf("%.0f", raw["i"]), 10, 64)
    //data, _ := json.Marshal(map[string]interface{}{
    //  "symbol":   raw["s"].(string),
    //  "order_id": orderId,
    //  "status":   raw["X"].(string),
    //})
    //h.Nats.Publish(config.NATS_ORDERS_UPDATE, data)
    //h.Nats.Flush()
  }
}

func (h *AccountHandler) subscribe() error {
  if 1 > 0 {
    return nil
  }

  apiKey := os.Getenv("BINANCE_SPOT_STREAMS_API_KEY")
  apiSecret := os.Getenv("BINANCE_SPOT_STREAMS_API_SECRET")
  timestamp := time.Now().UnixMilli()

  payload := fmt.Sprintf("apiKey=%s&timestamp=%d", apiKey, timestamp)
  mac := hmac.New(sha256.New, []byte(apiSecret))
  mac.Write([]byte(payload))
  signature := hex.EncodeToString(mac.Sum(nil))

  req := map[string]interface{}{
    "id":     uuid.NewString(),
    "method": "userDataStream.subscribe",
    "params": map[string]interface{}{
      "apiKey":    apiKey,
      "timestamp": timestamp,
      "signature": signature,
    },
  }

  log.Println("req", req)

  ctx, cancel := context.WithTimeout(h.Ctx, 10*time.Second)
  defer cancel()
  return wsjson.Write(ctx, h.Socket, req)
}

func (h *AccountHandler) Start() (err error) {
  log.Println("stream start")

  endpoint := os.Getenv("BINANCE_SPOT_WS_API_ENDPOINT")

  var httpClient *http.Client

  proxy := common.GetEnvString("BINANCE_PROXY")
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

  go h.pingLoop()

  log.Println("account stream started")

  if err = h.subscribe(); err != nil {
    log.Println("failed to subscribe", err.Error())
    return
  }

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
      log.Println("account worker channel full, dropping message")
    }
  }
}

func (h *AccountHandler) pingLoop() {
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
