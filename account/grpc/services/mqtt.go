package services

import (
  "taoniu.local/account/common"
)

type Mqtt struct {
  GrpcContext *common.GrpcContext
}

func NewMqtt(
  grpcContext *common.GrpcContext,
) *Mqtt {
  return &Mqtt{
    GrpcContext: grpcContext,
  }
}

func (s *Mqtt) Register() error {
  //mqtt.NewPublishers(s.GrpcContext).Register()
  return nil
}
