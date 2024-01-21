package futures

import (
  "context"
  "time"

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
  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    "locks:binance:futures:tickers:flush",
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush()

  return nil
}

func (h *Tickers) Register() error {
  h.AnsqContext.Mux.HandleFunc("binance:futures:tickers:flush", h.Flush)
  return nil
}
