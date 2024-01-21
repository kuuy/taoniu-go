package binance

import (
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance"
)

type ServerTask struct {
  AnsqContext *common.AnsqClientContext
  Repository  *repositories.ServerRepository
}

func NewServerTask(ansqContext *common.AnsqClientContext) *ServerTask {
  return &ServerTask{
    AnsqContext: ansqContext,
    Repository: &repositories.ServerRepository{
      Rdb: ansqContext.Rdb,
      Ctx: ansqContext.Ctx,
    },
  }
}

func (t *ServerTask) Time() error {
  t.Repository.Time()
  return nil
}
