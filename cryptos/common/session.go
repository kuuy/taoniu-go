package common

import (
  "context"
  "h12.io/socks"
  "net"
)

type ProxySession struct {
  Proxy string
}

func (session *ProxySession) DialContext(ctx context.Context, net, addr string) (net.Conn, error) {
  dialer := socks.Dial(session.Proxy)
  return dialer(net, addr)
}
