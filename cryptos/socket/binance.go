package socket

import (
	"strings"

	"taoniu.local/cryptos/common"
	"taoniu.local/cryptos/socket/binance"
)

type Binance struct {
	SocketContext *common.SocketContext
	Hub           *Hub
	SpotSocket    *binance.Spot
	FuturesSocket *binance.Futures
}

func NewBinance(socketContext *common.SocketContext, hub *Hub) *Binance {
	return &Binance{
		SocketContext: socketContext,
		Hub:           hub,
	}
}

func (h *Binance) Spot() *binance.Spot {
	if h.SpotSocket == nil {
		h.SpotSocket = binance.NewSpot(h.SocketContext, h.Hub)
	}
	return h.SpotSocket
}

func (h *Binance) Futures() *binance.Futures {
	if h.FuturesSocket == nil {
		h.FuturesSocket = binance.NewFutures(h.SocketContext, h.Hub)
	}
	return h.FuturesSocket
}

func (h *Binance) Subscribe(client *Client, req *RequestMessage) error {
	if strings.HasPrefix(req.Topic, "binance:spot") {
		return h.Spot().Subscribe(client, req)
	}
	return h.Futures().Subscribe(client, req)
}

func (h *Binance) Unsubscribe(client *Client, req *RequestMessage) error {
	if strings.HasPrefix(req.Topic, "binance:spot") {
		return h.Spot().Unsubscribe(client, req)
	}
	return h.Futures().Unsubscribe(client, req)
}
