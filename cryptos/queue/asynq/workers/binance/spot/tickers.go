package spot

import (
  "context"
  "encoding/json"
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
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
  var payload TickersFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Flush(payload.Symbols)
  return nil
}

func (h *Tickers) Register() error {
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_TICKERS_FLUSH, h.Flush)
  return nil
}
