package commands

import (
  "context"
  "fmt"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"
  socketio "github.com/vchitai/go-socket.io/v4"
  "log"
  "net/http"
  "os"
  "sync"

  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/socket"
)

type SocketHandler struct {
  Jwt  socket.JwtHandler
  Ctx  context.Context
  Nats *nats.Conn
}

func NewSocketCommand() *cli.Command {
  var h SocketHandler
  return &cli.Command{
    Name:  "socket",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = SocketHandler{
        Ctx:  context.Background(),
        Nats: common.NewNats(),
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      if err := h.Run(); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *SocketHandler) Run() error {
  log.Println("sockets running...")

  server := common.NewSocketServer(nil)
  go server.Serve()
  defer server.Close()

  server.OnConnect("/", h.Jwt.Authenticator(func(conn socketio.Conn, req map[string]interface{}) error {
    conn.SetContext(h.Ctx)
    return nil
  }))

  server.OnEvent("/binance/futures", "subscribe", func(conn socketio.Conn, req map[string]interface{}) error {
    wg := &sync.WaitGroup{}
    wg.Add(1)

    socketContext := &common.SocketContext{
      Socket: server,
      Conn:   conn,
      Nats:   h.Nats,
    }
    err := socket.NewBinance(socketContext).Futures().Subscribe(req)
    if err != nil {
      log.Println("error", err.Error())
    }

    <-h.wait(wg)

    return nil
  })

  server.OnEvent("/binance/futures", "unsubscribe", func(conn socketio.Conn, req map[string]interface{}) error {
    socketContext := &common.SocketContext{
      Socket: server,
      Conn:   conn,
      Nats:   h.Nats,
    }
    err := socket.NewBinance(socketContext).Futures().UnSubscribe(req)
    if err != nil {
      log.Println("error", err.Error())
    }
    return nil
  })

  http.Handle("/socket.io/cryptos/", server)
  http.ListenAndServe(
    fmt.Sprintf("127.0.0.1:%v", os.Getenv("CRYPTOS_SOCKET_PORT")),
    nil,
  )

  return nil
}

func (h *SocketHandler) wait(wg *sync.WaitGroup) chan bool {
  ch := make(chan bool)
  go func() {
    wg.Wait()
    ch <- true
  }()
  return ch
}
