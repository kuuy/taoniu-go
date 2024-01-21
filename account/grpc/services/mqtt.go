package services

import (
  "taoniu.local/account/common"
  "taoniu.local/account/grpc/services/mqtt"
)

type Mqtt struct {
  GrpcContext *common.GrpcContext
}

func NewMqtt(grpcContext *common.GrpcContext) *Mqtt {
  return &Mqtt{
    GrpcContext: grpcContext,
  }
}

func (srv *Mqtt) Register() error {
  mqtt.NewPublishers(srv.GrpcContext).Register()
  return nil
}
