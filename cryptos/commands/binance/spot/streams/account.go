package streams

import (
  "context"
  "crypto/ed25519"
  "crypto/x509"
  "encoding/base64"
  "encoding/json"
  "encoding/pem"
  "errors"
  "fmt"
  "log"
  "net/http"
  "os"
  "os/signal"
  "strconv"
  "syscall"
  "time"

  "github.com/coder/websocket"
  "github.com/coder/websocket/wsjson"
  "github.com/go-redis/redis/v8"
  "github.com/google/uuid"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot/streams"
)

type AccountHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Nats       *nats.Conn
  Ctx        context.Context
  cancel     context.CancelFunc
  Socket     *websocket.Conn
  AccountJob *jobs.Account
  OrdersJob  *jobs.Orders
  workerChan chan map[string]interface{}
  pingChan   chan map[string]interface{}
}

func NewAccountCommand() *cli.Command {
  var h AccountHandler
  return &cli.Command{
    Name:  "account",
    Usage: "",
    Before: func(c *cli.Context) error {
      ctx, cancel := context.WithCancel(context.Background())
      h = AccountHandler{
        Db:         common.NewDB(1),
        Rdb:        common.NewRedis(1),
        Nats:       common.NewNats(1),
        Ctx:        ctx,
        cancel:     cancel,
        AccountJob: &jobs.Account{},
        OrdersJob:  &jobs.Orders{},
        workerChan: make(chan map[string]interface{}, 1024),
        pingChan:   make(chan map[string]interface{}, 16),
      }
      return nil
    },
    After: func(c *cli.Context) error {
      h.Rdb.Close()
      h.Nats.Close()
      return nil
    },
    Action: func(c *cli.Context) error {
      sigChan := make(chan os.Signal, 1)
      signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
      go func() {
        sig := <-sigChan
        log.Printf("received signal %v, shutting down...", sig)
        h.cancel()
      }()

      for i := 0; i < 4; i++ {
        go h.worker()
      }

      for {
        select {
        case <-h.Ctx.Done():
          return nil
        default:
          if err := h.Start(); err != nil {
            log.Printf("account stream error: %v, reconnecting in 5s...", err)
            time.Sleep(5 * time.Second)
          }
        }
      }
    },
  }
}

func (h *AccountHandler) worker() {
  for {
    select {
    case <-h.Ctx.Done():
      return
    case message := <-h.workerChan:
      h.processMessage(message)
    }
  }
}

func (h *AccountHandler) processMessage(message map[string]interface{}) {
  log.Println("message", message)

  if status, ok := message["status"].(float64); ok {
    if status == 200 {
      log.Printf("userDataStream response status: 200 (id: %v)", message["id"])
    } else {
      log.Printf("userDataStream response error status: %v (id: %v, error: %v)", status, message["id"], message["error"])
    }
  }

  if errResp, ok := message["error"].(map[string]interface{}); ok {
    log.Printf("binance websocket error: code=%v msg=%v", errResp["code"], errResp["msg"])
    return
  }

  var data map[string]interface{}
  if d, ok := message["data"].(map[string]interface{}); ok {
    data = d
  } else if d, ok := message["event"].(map[string]interface{}); ok {
    data = d
  } else {
    data = message
  }

  event, ok := data["e"].(string)
  if !ok {
    return
  }

  if event == "outboundAccountPosition" {
    balances, ok := data["B"].([]interface{})
    if !ok {
      return
    }
    for _, item := range balances {
      account, ok := item.(map[string]interface{})
      if !ok {
        continue
      }
      asset, ok := account["a"].(string)
      if !ok {
        continue
      }
      free, _ := strconv.ParseFloat(fmt.Sprintf("%v", account["f"]), 64)
      locked, _ := strconv.ParseFloat(fmt.Sprintf("%v", account["l"]), 64)

      if free <= 0 && locked <= 0 {
        h.Rdb.SRem(h.Ctx, config.REDIS_KEY_CURRENCIES, asset)
        h.Rdb.Del(h.Ctx, fmt.Sprintf(config.REDIS_KEY_BALANCE, asset))
      } else {
        h.Rdb.SAdd(h.Ctx, config.REDIS_KEY_CURRENCIES, asset)
        h.Rdb.HMSet(
          h.Ctx,
          fmt.Sprintf(config.REDIS_KEY_BALANCE, asset),
          map[string]interface{}{
            "free":   free,
            "locked": locked,
          },
        )
      }

      if h.Nats != nil {
        payload, _ := json.Marshal(map[string]interface{}{
          "asset":  asset,
          "free":   free,
          "locked": locked,
        })
        h.Nats.Publish(config.NATS_ACCOUNT_UPDATE, payload)
        h.Nats.Flush()
      }
    }
  }

  if event == "balanceUpdate" {
    asset, _ := data["a"].(string)
    delta, _ := strconv.ParseFloat(fmt.Sprintf("%v", data["d"]), 64)
    log.Printf("balance update event for asset %s, delta: %f", asset, delta)
  }

  if event == "executionReport" {
    symbol, ok := data["s"].(string)
    if !ok {
      return
    }
    orderId, _ := strconv.ParseInt(fmt.Sprintf("%.0f", data["i"]), 10, 64)
    status, _ := data["X"].(string)

    if h.Nats != nil {
      payload, _ := json.Marshal(map[string]interface{}{
        "symbol":   symbol,
        "order_id": orderId,
        "status":   status,
      })
      h.Nats.Publish(config.NATS_ORDERS_UPDATE, payload)
      h.Nats.Flush()
    }
  }

  //if event == "listStatus" {
  //  symbol, _ := data["s"].(string)
  //  orderListId, _ := strconv.ParseInt(fmt.Sprintf("%v", data["g"]), 10, 64)
  //  listStatusType, _ := data["l"].(string)
  //  log.Printf("list status event for symbol %s (listId %d): %s", symbol, orderListId, listStatusType)
  //}
}

func parseEd25519PrivateKey(secret string) (ed25519.PrivateKey, error) {
  block, _ := pem.Decode([]byte(secret))
  var keyData []byte
  if block != nil {
    keyData = block.Bytes
  } else {
    decoded, err := base64.StdEncoding.DecodeString(secret)
    if err == nil {
      keyData = decoded
    } else {
      keyData = []byte(secret)
    }
  }

  if key, err := x509.ParsePKCS8PrivateKey(keyData); err == nil {
    if edKey, ok := key.(ed25519.PrivateKey); ok {
      return edKey, nil
    }
  }

  if len(keyData) == ed25519.SeedSize {
    return ed25519.NewKeyFromSeed(keyData), nil
  }
  if len(keyData) == ed25519.PrivateKeySize {
    return keyData, nil
  }

  return nil, errors.New("invalid ed25519 key format")
}

func (h *AccountHandler) sessionLogon(apiKey string, apiSecret string) (err error) {
  reqID := uuid.NewString()
  timestamp := time.Now().UnixMilli()

  payload := fmt.Sprintf("apiKey=%s&timestamp=%d", apiKey, timestamp)

  privateKey, err := parseEd25519PrivateKey(apiSecret)
  if err != nil {
    return
  }
  signatureBytes := ed25519.Sign(privateKey, []byte(payload))
  signature := base64.StdEncoding.EncodeToString(signatureBytes)

  params := map[string]interface{}{
    "apiKey":    apiKey,
    "timestamp": timestamp,
    "signature": signature,
  }

  req := map[string]interface{}{
    "id":     reqID,
    "method": "session.logon",
    "params": params,
  }

  ctx, cancel := context.WithTimeout(h.Ctx, 10*time.Second)
  defer cancel()

  if err = wsjson.Write(ctx, h.Socket, req); err != nil {
    return
  }

  var resp map[string]interface{}
  if err = wsjson.Read(ctx, h.Socket, &resp); err != nil {
    return
  }

  if status, ok := resp["status"].(float64); ok && status == 200 {
    return
  }

  err = fmt.Errorf("session.logon failed: %v", resp)
  return
}

func (h *AccountHandler) start() (listenKey string, err error) {
  reqID := uuid.NewString()
  req := map[string]interface{}{
    "id":     reqID,
    "method": "userDataStream.start",
  }

  ctx, cancel := context.WithTimeout(h.Ctx, 10*time.Second)
  defer cancel()

  if err = wsjson.Write(ctx, h.Socket, req); err != nil {
    return
  }

  var resp map[string]interface{}
  if err = wsjson.Read(ctx, h.Socket, &resp); err != nil {
    return
  }

  if status, ok := resp["status"].(float64); ok && status == 200 {
    if result, ok := resp["result"].(map[string]interface{}); ok {
      if lk, ok := result["listenKey"].(string); ok && lk != "" {
        listenKey = lk
        log.Println("userDataStream.start success, listenKey:", listenKey)
        return
      }
    }
    return
  }

  err = fmt.Errorf("userDataStream.start failed: %v", resp)
  return
}

func (h *AccountHandler) subscribe() (err error) {
  reqID := uuid.NewString()
  req := map[string]interface{}{
    "id":     reqID,
    "method": "userDataStream.subscribe",
  }

  ctx, cancel := context.WithTimeout(h.Ctx, 10*time.Second)
  defer cancel()

  if err = wsjson.Write(ctx, h.Socket, req); err != nil {
    return
  }

  var resp map[string]interface{}
  if err = wsjson.Read(ctx, h.Socket, &resp); err != nil {
    return
  }

  if status, ok := resp["status"].(float64); ok && status == 200 {
    log.Println("userDataStream.subscribe success:", resp)
    return
  }

  err = fmt.Errorf("userDataStream.subscribe failed: %v", resp)
  return
}

func (h *AccountHandler) stop(listenKey string) {
  if h.Socket == nil || listenKey == "" {
    return
  }
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()
  req := map[string]interface{}{
    "id":     uuid.NewString(),
    "method": "userDataStream.stop",
    "params": map[string]interface{}{
      "listenKey": listenKey,
    },
  }
  _ = wsjson.Write(ctx, h.Socket, req)
}

func (h *AccountHandler) keepaliveLoop(ctx context.Context, listenKey string) {
  ticker := time.NewTicker(30 * time.Minute)
  defer ticker.Stop()
  for {
    select {
    case <-ctx.Done():
      return
    case <-ticker.C:
      if h.Socket == nil {
        return
      }
      reqID := uuid.NewString()
      params := map[string]interface{}{}
      params["listenKey"] = listenKey
      req := map[string]interface{}{
        "id":     reqID,
        "method": "userDataStream.ping",
        "params": params,
      }
      pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
      err := wsjson.Write(pingCtx, h.Socket, req)
      cancel()
      if err != nil {
        log.Println("userDataStream.ping write error:", err.Error())
        continue
      }

      select {
      case <-ctx.Done():
        return
      case resp := <-h.pingChan:
        if status, ok := resp["status"].(float64); ok && status == 200 {
          log.Printf("userDataStream.ping status 200 OK (id: %v)", resp["id"])
        } else {
          log.Printf("userDataStream.ping response error status: %v (resp: %v)", resp["status"], resp)
        }
      case <-time.After(10 * time.Second):
        log.Println("userDataStream.ping response timeout")
      }
    }
  }
}

func (h *AccountHandler) Start() (err error) {
  log.Println("account stream start")

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

  endpoint := common.GetEnvString("BINANCE_SPOT_WS_API_ENDPOINT")
  h.Socket, _, err = websocket.Dial(h.Ctx, endpoint, &websocket.DialOptions{
    HTTPClient:      httpClient,
    CompressionMode: websocket.CompressionDisabled,
  })
  if err != nil {
    return fmt.Errorf("dial error: %w", err)
  }
  defer h.Socket.Close(websocket.StatusInternalError, "the socket was closed abruptly")

  connCtx, connCancel := context.WithCancel(h.Ctx)
  defer connCancel()

  go h.pingLoop(connCtx)

  log.Println("account stream connected")

  apiKey := common.GetEnvString("BINANCE_SPOT_STREAMS_API_KEY")
  apiSecret := common.GetEnvString("BINANCE_SPOT_STREAMS_API_SECRET")

  if err = h.sessionLogon(apiKey, apiSecret); err != nil {
    log.Println("session logon warning:", err.Error())
  }

  listenKey, err := h.start()
  if err != nil {
    log.Println("failed to start userDataStream", err.Error())
    return
  }
  defer h.stop(listenKey)

  if err = h.subscribe(); err != nil {
    log.Println("failed to subscribe userDataStream", err.Error())
    return
  }

  go h.keepaliveLoop(connCtx, listenKey)

  for {
    var message map[string]interface{}
    err = wsjson.Read(connCtx, h.Socket, &message)

    if err != nil {
      if errors.Is(err, context.Canceled) {
        return nil
      }
      return fmt.Errorf("read error: %w", err)
    }

    if _, ok := message["status"]; ok {
      select {
      case h.pingChan <- message:
      default:
      }
    }

    select {
    case h.workerChan <- message:
    default:
      log.Println("account worker channel full, dropping message")
    }
  }
}

func (h *AccountHandler) pingLoop(ctx context.Context) {
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
      pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
      _ = h.Socket.Ping(pingCtx)
      cancel()
    }
  }
}
