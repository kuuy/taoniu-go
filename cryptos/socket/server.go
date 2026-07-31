package socket

import (
  "context"
  "net/http"
  "strings"

  "github.com/coder/websocket"
  "taoniu.local/cryptos/common"
)

type Server struct {
  Hub           *Hub
  SocketContext *common.SocketContext
  JwtHandler    *JwtHandler
}

func NewServer(socketContext *common.SocketContext) *Server {
  return &Server{
    Hub:           NewHub(),
    SocketContext: socketContext,
    JwtHandler:    &JwtHandler{},
  }
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  token := ""
  bearer := r.Header.Get("Authorization")
  if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "TAONIU" {
    token = bearer[7:]
  }
  if token == "" {
    token = r.URL.Query().Get("access_token")
  }

  uid, authErr := s.JwtHandler.Authenticate(token)
  if authErr != nil {
    http.Error(w, authErr.Error(), http.StatusUnauthorized)
    return
  }

  opts := &websocket.AcceptOptions{
    InsecureSkipVerify: true,
  }

  conn, err := websocket.Accept(w, r, opts)
  if err != nil {
    return
  }

  client := NewClient(s.Hub, conn, uid)
  s.Hub.Register(client)

  clientCtx, clientCancel := context.WithCancel(context.Background())
  go func() {
    client.WriteLoop(clientCtx)
    clientCancel()
  }()
  go func() {
    client.ReadLoop(clientCtx, func(c *Client, req *RequestMessage) {
      s.handleRequest(c, req)
    })
    clientCancel()
  }()
}

func (s *Server) handleRequest(client *Client, req *RequestMessage) {
  binanceHandler := NewBinance(s.SocketContext, s.Hub)
  switch req.Action {
  case "subscribe":
    binanceHandler.Subscribe(client, req)
  case "unsubscribe":
    binanceHandler.Unsubscribe(client, req)
  }
}
