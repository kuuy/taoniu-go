package binance

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/socket/binance/futures"
)

type HubInterface interface {
  Subscribe(client interface{}, topic string)
  Unsubscribe(client interface{}, topic string)
  Broadcast(topic string, message []byte)
}

type Futures struct {
  SocketContext *common.SocketContext
  Hub           interface{}
}

func NewFutures(socketContext *common.SocketContext, hub interface{}) *Futures {
  return &Futures{
    SocketContext: socketContext,
    Hub:           hub,
  }
}

func (h *Futures) Subscribe(client interface{}, req interface{}) error {
  tickers := futures.NewTickers(h.SocketContext, h.Hub)
  return tickers.Subscribe(client, req)
}

func (h *Futures) Unsubscribe(client interface{}, req interface{}) error {
  tickers := futures.NewTickers(h.SocketContext, h.Hub)
  return tickers.Unsubscribe(client, req)
}
