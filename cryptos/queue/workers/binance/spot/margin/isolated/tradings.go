package isolated

import (
	"github.com/hibiken/asynq"
	"taoniu.local/cryptos/queue/workers/binance/spot/margin/isolated/tradings"
)

type Tradings struct{}

func NewTradings() *Tradings {
	return &Tradings{}
}

func (h *Tradings) Register(mux *asynq.ServeMux) error {
	mux.HandleFunc("binance:spot:margin:isolated:tradings:fishers:flush", tradings.NewFishers().Flush)
	mux.HandleFunc("binance:spot:margin:isolated:tradings:fishers:place", tradings.NewFishers().Place)
	return nil
}
