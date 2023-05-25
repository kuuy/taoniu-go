package nats

import "github.com/nats-io/nats.go"

type Workers struct{}

func NewWorkers() *Workers {
  return &Workers{}
}

func (h *Workers) Subscribe(nc *nats.Conn) error {
  return nil
}
