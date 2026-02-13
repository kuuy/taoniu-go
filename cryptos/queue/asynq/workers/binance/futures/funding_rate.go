package futures

import (
  "context"
  "github.com/hibiken/asynq"
  config "taoniu.local/cryptos/config/binance/futures"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type FundingRate struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.FundingRateRepository
}

func NewFundingRate(ansqContext *common.AnsqServerContext) *FundingRate {
  h := &FundingRate{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.FundingRateRepository{
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  return h
}

func (h *FundingRate) Flush(ctx context.Context, t *asynq.Task) error {
  h.Repository.Flush()
  return nil
}

func (h *FundingRate) Register() error {
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_FUNDING_RATE_FLUSH, h.Flush)
  return nil
}
