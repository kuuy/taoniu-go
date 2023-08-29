package binance

import (
  "errors"
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/socket/binance/futures"
)

type Futures struct {
  SocketContext *common.SocketContext
}

func NewFutures(socketContext *common.SocketContext) *Futures {
  return &Futures{socketContext}
}

func (h *Futures) Subscribe(req map[string]interface{}) error {
  if _, ok := req["topic"]; !ok {
    return errors.New("topic is empty")
  }
  switch req["topic"].(string) {
  case "tickers":
    return futures.NewTickers(h.SocketContext).Subscribe(req)
  default:
    return errors.New("topic not supported")
  }
}

func (h *Futures) UnSubscribe(req map[string]interface{}) error {
  if _, ok := req["topic"]; !ok {
    return errors.New("topic is empty")
  }
  switch req["topic"].(string) {
  case "tickers":
    return futures.NewTickers(h.SocketContext).UnSubscribe()
  default:
    return errors.New("topic not supported")
  }
}
