package socket

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/socket/binance"
)

type Binance struct {
  SocketContext *common.SocketContext
  FuturesSocket *binance.Futures
}

func NewBinance(socketContext *common.SocketContext) *Binance {
  return &Binance{
    SocketContext: socketContext,
  }
}

func (h *Binance) Futures() *binance.Futures {
  if h.FuturesSocket == nil {
    h.FuturesSocket = binance.NewFutures(h.SocketContext)
  }
  return h.FuturesSocket
}
