package socket

import (
  "context"
  "encoding/json"
  "log"
  "strings"
  "sync"
  "time"

  "github.com/coder/websocket"

  "taoniu.local/cryptos/common"
)

type RequestMessage struct {
  Action      string      `json:"action"`
  Topic       string      `json:"topic"`
  Symbols     []string    `json:"symbols"`
  AccessToken string      `json:"access_token"`
  Payload     interface{} `json:"payload"`
}

type PushMessage struct {
  Event string      `json:"event"`
  Topic string      `json:"topic"`
  Data  interface{} `json:"data"`
}

type Client struct {
  Hub    *Hub
  Conn   *websocket.Conn
  Send   chan []byte
  Topics map[string]bool
  Uid    string
  mu     sync.Mutex
  Jwe    *common.Jwe
}

func NewClient(hub *Hub, conn *websocket.Conn, uid string) *Client {
	return &Client{
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan []byte, 2048),
		Topics: make(map[string]bool),
		Uid:    uid,
		Jwe:    &common.Jwe{},
	}
}

func (c *Client) SendBytes(data []byte) {
	jwe := &common.Jwe{}
	jweCompact, err := jwe.Encrypt(data)
	if err != nil {
		return
	}
	msgBytes := []byte(jweCompact)
	select {
	case c.Send <- msgBytes:
	default:
		select {
		case <-c.Send:
		default:
		}
		select {
		case c.Send <- msgBytes:
		default:
		}
	}
}

func (c *Client) ReadLoop(ctx context.Context, onMessage func(c *Client, req *RequestMessage)) {
  defer func() {
    c.Hub.Unregister(c)
    c.Conn.Close(websocket.StatusNormalClosure, "")
  }()

  for {
    _, data, err := c.Conn.Read(ctx)
    if err != nil {
      break
    }

    payload, err := c.Jwe.Decrypt(string(data))
    if err != nil {
      var rawReq RequestMessage
      if jsonErr := json.Unmarshal(data, &rawReq); jsonErr == nil && rawReq.Action == "ping" {
        pong, _ := json.Marshal(map[string]string{"event": "pong"})
        c.SendBytes(pong)
        continue
      }
      if strings.TrimSpace(string(data)) == "ping" {
        pong, _ := json.Marshal(map[string]string{"event": "pong"})
        c.SendBytes(pong)
        continue
      }
      continue
    }

    var req RequestMessage
    if err := json.Unmarshal(payload, &req); err != nil {
      continue
    }

    if req.Action == "ping" {
      pong, _ := json.Marshal(map[string]string{"event": "pong"})
      c.SendBytes(pong)
      continue
    }

    if onMessage != nil {
      onMessage(c, &req)
    }
  }
}

func (c *Client) WriteLoop(ctx context.Context) {
  ticker := time.NewTicker(25 * time.Second)
  defer func() {
    ticker.Stop()
    c.Conn.Close(websocket.StatusNormalClosure, "")
  }()

  for {
    select {
    case message, ok := <-c.Send:
      if !ok {
        c.Conn.Close(websocket.StatusNormalClosure, "")
        return
      }
      writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
      err := c.Conn.Write(writeCtx, websocket.MessageText, message)
      cancel()
      if err != nil {
        return
      }
    case <-ticker.C:
      writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
      err := c.Conn.Ping(writeCtx)
      cancel()
      if err != nil {
        return
      }
    case <-ctx.Done():
      return
    }
  }
}

type Hub struct {
  clients map[*Client]bool
  topics  map[string]map[*Client]bool
  mu      sync.RWMutex
  Jwe     *common.Jwe
}

func NewHub() *Hub {
  return &Hub{
    clients: make(map[*Client]bool),
    topics:  make(map[string]map[*Client]bool),
    Jwe:     &common.Jwe{},
  }
}

func (h *Hub) Register(client *Client) {
  h.mu.Lock()
  defer h.mu.Unlock()
  h.clients[client] = true
}

func (h *Hub) Unregister(client *Client) {
  h.mu.Lock()
  defer h.mu.Unlock()
  if _, ok := h.clients[client]; ok {
    delete(h.clients, client)
    close(client.Send)
    for topic, clientSet := range h.topics {
      delete(clientSet, client)
      if len(clientSet) == 0 {
        delete(h.topics, topic)
      }
    }
  }
}

func (h *Hub) Subscribe(client interface{}, topic string) {
  c, ok := client.(*Client)
  if !ok {
    return
  }

  h.mu.Lock()
  defer h.mu.Unlock()

  c.mu.Lock()
  c.Topics[topic] = true
  c.mu.Unlock()

  if h.topics[topic] == nil {
    h.topics[topic] = make(map[*Client]bool)
  }
  h.topics[topic][c] = true
}

func (h *Hub) Unsubscribe(client interface{}, topic string) {
  c, ok := client.(*Client)
  if !ok {
    return
  }

  h.mu.Lock()
  defer h.mu.Unlock()

  c.mu.Lock()
  delete(c.Topics, topic)
  c.mu.Unlock()

  if clientSet, ok := h.topics[topic]; ok {
    delete(clientSet, c)
    if len(clientSet) == 0 {
      delete(h.topics, topic)
    }
  }
}

func (h *Hub) Broadcast(topic string, message []byte) {
  h.mu.RLock()
  defer h.mu.RUnlock()

  clientSet, ok := h.topics[topic]
  if !ok || len(clientSet) == 0 {
    return
  }

  jweCompact, err := h.Jwe.Encrypt(message)
  if err != nil {
    return
  }
  encryptedMessage := []byte(jweCompact)

  for client := range clientSet {
    select {
    case client.Send <- encryptedMessage:
    default:
      select {
      case <-client.Send:
      default:
      }
      select {
      case client.Send <- encryptedMessage:
      default:
        log.Printf("Client send queue full, dropping message for topic: %s", topic)
      }
    }
  }
}
