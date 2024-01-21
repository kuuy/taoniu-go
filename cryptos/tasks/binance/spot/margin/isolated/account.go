package isolated

import (
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
)

type AccountTask struct {
  AnsqContext *common.AnsqClientContext
  Repository  *repositories.AccountRepository
}

func NewAccountTask(ansqContext *common.AnsqClientContext) *AccountTask {
  return &AccountTask{
    AnsqContext: ansqContext,
    Repository: &repositories.AccountRepository{
      Db:  ansqContext.Db,
      Rdb: ansqContext.Rdb,
      Ctx: ansqContext.Ctx,
      TradingsRepository: &repositories.TradingsRepository{
        Db: ansqContext.Db,
        FishersRepository: &tradingsRepositories.FishersRepository{
          Db: ansqContext.Db,
        },
      },
    },
  }
}

func (t *AccountTask) Flush() error {
  return t.Repository.Flush()
}

func (t *AccountTask) Collect() error {
  return t.Repository.Collect()
}

func (t *AccountTask) Liquidate() error {
  return t.Repository.Liquidate()
}
