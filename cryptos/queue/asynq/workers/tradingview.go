package workers

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/tradingview"
)

type Tradingview struct {
  AnsqContext *common.AnsqServerContext
}

func NewTradingview(ansqContext *common.AnsqServerContext) *Tradingview {
  return &Tradingview{
    AnsqContext: ansqContext,
  }
}

func (h *Tradingview) Register() error {
  h.AnsqContext.Mux.HandleFunc("tradingview:analysis:flush", tradingview.NewAnalysis(h.AnsqContext).Flush)
  return nil
}
