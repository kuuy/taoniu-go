package socket

import (
  "errors"
  socketio "github.com/vchitai/go-socket.io/v4"
  repositories "taoniu.local/cryptos/repositories/account"
)

type JwtHandler struct{}

func (h *JwtHandler) Authenticator(next func(socketio.Conn, map[string]interface{}) error) func(socketio.Conn, map[string]interface{}) error {
  return func(conn socketio.Conn, req map[string]interface{}) error {
    if _, ok := req["access_token"]; ok {
      accessToken := req["access_token"].(string)
      repository := &repositories.TokenRepository{}
      uid, err := repository.Uid(accessToken)
      if err != nil {
        conn.Close()
        if uid != "" {
          return err
        } else {
          return errors.New("access not allowed")
        }
      }
    } else {
      conn.Close()
    }
    return next(conn, req)
  }
}
