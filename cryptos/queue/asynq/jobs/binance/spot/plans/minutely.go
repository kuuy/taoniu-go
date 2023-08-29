package plans

import "github.com/hibiken/asynq"

type Minutely struct{}

func (h *Minutely) Flush() (*asynq.Task, error) {
  return asynq.NewTask("binance:spot:plans:1m:flush", nil), nil
}
