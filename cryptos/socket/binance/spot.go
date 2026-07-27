package binance

import (
	"taoniu.local/cryptos/common"
	"taoniu.local/cryptos/socket/binance/spot"
)

type Spot struct {
	SocketContext *common.SocketContext
	Hub           interface{}
}

func NewSpot(socketContext *common.SocketContext, hub interface{}) *Spot {
	return &Spot{
		SocketContext: socketContext,
		Hub:           hub,
	}
}

func (h *Spot) Subscribe(client interface{}, req interface{}) error {
	tickers := spot.NewTickers(h.SocketContext, h.Hub)
	return tickers.Subscribe(client, req)
}

func (h *Spot) Unsubscribe(client interface{}, req interface{}) error {
	tickers := spot.NewTickers(h.SocketContext, h.Hub)
	return tickers.Unsubscribe(client, req)
}
