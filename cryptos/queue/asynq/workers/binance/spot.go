package binance

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/binance/spot"
)

type Spot struct {
  AnsqContext *common.AnsqServerContext
}

func NewSpot(ansqContext *common.AnsqServerContext) *Spot {
  return &Spot{
    AnsqContext: ansqContext,
  }
}

func (h *Spot) Register() error {
  spot.NewTickers(h.AnsqContext).Register()
  spot.NewKlines(h.AnsqContext).Register()
  spot.NewDepth(h.AnsqContext).Register()
  spot.NewIndicators(h.AnsqContext).Register()
  spot.NewStrategies(h.AnsqContext).Register()
  spot.NewPlans(h.AnsqContext).Register()
  spot.NewAccount(h.AnsqContext).Register()
  spot.NewOrders(h.AnsqContext).Register()
  spot.NewPositions(h.AnsqContext).Register()
  spot.NewScalping(h.AnsqContext).Register()
  spot.NewTradings(h.AnsqContext).Register()
  return nil
}
