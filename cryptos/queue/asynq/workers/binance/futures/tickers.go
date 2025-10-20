package futures

import (
  "context"
  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type Tickers struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.TickersRepository
}

func NewTickers(ansqContext *common.AnsqServerContext) *Tickers {
  h := &Tickers{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.TickersRepository{
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  return h
}

func (h *Tickers) Flush(ctx context.Context, t *asynq.Task) error {
  h.Repository.Flush()
  return nil
}

func (h *Tickers) Register() error {
  h.AnsqContext.Mux.HandleFunc("binance:futures:tickers:flush", h.Flush)
  return nil
}
