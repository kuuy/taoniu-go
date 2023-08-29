package binance

import (
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/queue/asynq/workers/binance/spot"
)

type Spot struct{}

func NewSpot() *Spot {
  return &Spot{}
}

func (h *Spot) Register(mux *asynq.ServeMux) error {
  spot.NewTickers().Register(mux)
  spot.NewKlines().Register(mux)
  spot.NewDepth().Register(mux)
  spot.NewIndicators().Register(mux)
  spot.NewStrategies().Register(mux)
  spot.NewPlans().Register(mux)
  spot.NewAccount().Register(mux)
  spot.NewOrders().Register(mux)
  spot.NewTradings().Register(mux)
  spot.NewMargin().Register(mux)
  return nil
}
