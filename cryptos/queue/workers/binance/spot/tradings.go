package spot

import (
	"github.com/hibiken/asynq"
	"taoniu.local/cryptos/queue/workers/binance/spot/tradings"
)

type Tradings struct{}

func NewTradings() *Tradings {
	return &Tradings{}
}

func (h *Tradings) Register(mux *asynq.ServeMux) error {
	mux.HandleFunc("binance:spot:tradings:fishers:flush", tradings.NewFishers().Flush)
	mux.HandleFunc("binance:spot:tradings:fishers:place", tradings.NewFishers().Place)
	mux.HandleFunc("binance:spot:tradings:scalping:flush", tradings.NewScalping().Flush)
	return nil
}
