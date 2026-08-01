package streams

import (
  "context"
  "crypto/ed25519"
  "crypto/hmac"
  "crypto/sha256"
  "crypto/x509"
  "encoding/base64"
  "encoding/hex"
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
  config "taoniu.local/cryptos/config/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures/streams"
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
        Db:         common.NewDB(2),
        Rdb:        common.NewRedis(2),
        Nats:       common.NewNats(2),
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
      if h.Rdb != nil {
        h.Rdb.Close()
      }
      if h.Nats != nil {
        h.Nats.Close()
      }
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
            log.Printf("futures account stream error: %v, reconnecting in 5s...", err)
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
  if status, ok := message["status"].(float64); ok {
    if status == 200 {
      log.Printf("futures userDataStream response status: 200 (id: %v)", message["id"])
    } else {
      log.Printf("futures userDataStream response error status: %v (id: %v, error: %v)", status, message["id"], message["error"])
    }
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

  if event == "ACCOUNT_UPDATE" {
    info, ok := data["a"].(map[string]interface{})
    if !ok {
      return
    }
    balances, ok := info["B"].([]interface{})
    if ok {
      for _, item := range balances {
        account, ok := item.(map[string]interface{})
        if !ok {
          continue
        }
        asset, ok := account["a"].(string)
        if !ok {
          continue
        }
        balance, _ := strconv.ParseFloat(fmt.Sprintf("%v", account["wb"]), 64)
        free, _ := strconv.ParseFloat(fmt.Sprintf("%v", account["cw"]), 64)
        up, _ := strconv.ParseFloat(fmt.Sprintf("%v", account["up"]), 64)
        margin, _ := strconv.ParseFloat(fmt.Sprintf("%v", account["iw"]), 64)

        if balance <= 0 {
          h.Rdb.Del(h.Ctx, fmt.Sprintf(config.REDIS_KEY_BALANCE, asset))
        } else {
          h.Rdb.HMSet(
            h.Ctx,
            fmt.Sprintf(config.REDIS_KEY_BALANCE, asset),
            map[string]interface{}{
              "balance":           balance,
              "free":              free,
              "unrealized_profit": up,
              "margin":            margin,
            },
          )
        }

        if h.Nats != nil {
          payload, _ := json.Marshal(map[string]interface{}{
            "asset":             asset,
            "balance":           balance,
            "free":              free,
            "unrealized_profit": up,
            "margin":            margin,
          })
          h.Nats.Publish(config.NATS_ACCOUNT_UPDATE, payload)
          h.Nats.Flush()
        }
      }
    }
  }

  if event == "ORDER_TRADE_UPDATE" {
    order, ok := data["o"].(map[string]interface{})
    if !ok {
      return
    }
    symbol, ok := order["s"].(string)
    if !ok {
      return
    }
    orderId, _ := strconv.ParseInt(fmt.Sprintf("%.0f", order["i"]), 10, 64)
    status, _ := order["X"].(string)

    if status != "NEW" && status != "PARTIALLY_FILLED" {
      h.Rdb.SAdd(
        h.Ctx,
        "binance:futures:orders:flush",
        fmt.Sprintf("%s,%d", symbol, orderId),
      )
    }

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

  if event == "MARGIN_CALL" {
    log.Println("MARGIN_CALL event received:", data)
  }

  if event == "ACCOUNT_CONFIG_UPDATE" {
    log.Println("ACCOUNT_CONFIG_UPDATE event received:", data)
  }
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
    return ed25519.PrivateKey(keyData), nil
  }

  return nil, errors.New("invalid ed25519 key format")
}

func (h *AccountHandler) sessionLogon(apiKey string, apiSecret string) error {
  reqID := uuid.NewString()
  timestamp := time.Now().UnixMilli()

  payload := fmt.Sprintf("apiKey=%s&timestamp=%d", apiKey, timestamp)

  var signature string
  privateKey, err := parseEd25519PrivateKey(apiSecret)
  if err == nil && privateKey != nil {
    signatureBytes := ed25519.Sign(privateKey, []byte(payload))
    signature = base64.StdEncoding.EncodeToString(signatureBytes)
    log.Println("ed25519 signature generated:", signature)
  } else {
    mac := hmac.New(sha256.New, []byte(apiSecret))
    mac.Write([]byte(payload))
    signature = hex.EncodeToString(mac.Sum(nil))
    log.Println("hmac signature generated:", signature)
  }

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

  if err := wsjson.Write(ctx, h.Socket, req); err != nil {
    return err
  }

  var resp map[string]interface{}
  if err := wsjson.Read(ctx, h.Socket, &resp); err != nil {
    return err
  }

  log.Println("session.logon response:", resp)
  return nil
}

func (h *AccountHandler) start() (listenKey string, err error) {
  reqID := uuid.NewString()
  req := map[string]interface{}{
    "id":     reqID,
    "method": "userDataStream.start",
  }

  ctx, cancel := context.WithTimeout(h.Ctx, 10*time.Second)
  defer cancel()

  if err := wsjson.Write(ctx, h.Socket, req); err != nil {
    return "", err
  }

  var resp map[string]interface{}
  if err := wsjson.Read(ctx, h.Socket, &resp); err != nil {
    return "", err
  }

  if status, ok := resp["status"].(float64); ok && status == 200 {
    if result, ok := resp["result"].(map[string]interface{}); ok {
      if listenKey, ok := result["listenKey"].(string); ok && listenKey != "" {
        log.Println("userDataStream.start success, listenKey:", listenKey)
        return listenKey, nil
      }
    }
    return "", nil
  }

  return "", fmt.Errorf("userDataStream.start failed: %v", resp)
}

func (h *AccountHandler) subscribe() (err error) {
  reqID := uuid.NewString()
  req := map[string]interface{}{
    "id":     reqID,
    "method": "userDataStream.subscribe",
  }

  ctx, cancel := context.WithTimeout(h.Ctx, 10*time.Second)
  defer cancel()

  if err := wsjson.Write(ctx, h.Socket, req); err != nil {
    return err
  }

  var resp map[string]interface{}
  if err := wsjson.Read(ctx, h.Socket, &resp); err != nil {
    return err
  }

  if status, ok := resp["status"].(float64); ok && status == 200 {
    log.Println("userDataStream.subscribe success:", resp)
    return nil
  }

  return fmt.Errorf("userDataStream.subscribe failed: %v", resp)
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
      params := map[string]interface{}{
        "listenKey": listenKey,
      }
      req := map[string]interface{}{
        "id":     reqID,
        "method": "userDataStream.ping",
        "params": params,
      }
      pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
      err := wsjson.Write(pingCtx, h.Socket, req)
      cancel()
      if err != nil {
        log.Println("futures userDataStream.ping write error:", err)
        continue
      }

      select {
      case <-ctx.Done():
        return
      case resp := <-h.pingChan:
        if status, ok := resp["status"].(float64); ok && status == 200 {
          log.Printf("futures userDataStream.ping status 200 OK (id: %v)", resp["id"])
        } else {
          log.Printf("futures userDataStream.ping response error status: %v (resp: %v)", resp["status"], resp)
        }
      case <-time.After(10 * time.Second):
        log.Println("futures userDataStream.ping response timeout")
      }
    }
  }
}

func (h *AccountHandler) Start() (err error) {
  log.Println("futures account stream start")

  var apiKey string
  var apiSecret string
  var endpoint string
  if common.GetEnvInt("BINANCE_FUTURES_TESTNET_ENABLE") == 1 {
    apiKey = common.GetEnvString("BINANCE_FUTURES_TESTNET_API_KEY")
    apiSecret = common.GetEnvString("BINANCE_FUTURES_TESTNET_API_SECRET")
    endpoint = common.GetEnvString("BINANCE_FUTURES_TESTNET_STREAMS_ENDPOINT") + "/ws"
  } else {
    apiKey = common.GetEnvString("BINANCE_FUTURES_STREAMS_API_KEY")
    apiSecret = common.GetEnvString("BINANCE_FUTURES_STREAMS_API_SECRET")
    endpoint = common.GetEnvString("BINANCE_FUTURES_STREAMS_ENDPOINT") + "/ws"
  }

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

  log.Println("futures endpoint:", endpoint)

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

  log.Println("futures account stream connected")

  if err = h.sessionLogon(apiKey, apiSecret); err != nil {
    log.Println("futures session logon warning:", err.Error())
  }

  listenKey, err := h.start()
  if err != nil {
    log.Println("failed to start userDataStream", err.Error())
    return err
  }
  defer h.stop(listenKey)

  if err = h.subscribe(); err != nil {
    log.Println("failed to subscribe userDataStream", err.Error())
    return err
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
      log.Println("futures account worker channel full, dropping message")
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
