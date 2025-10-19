package services

import (
  "context"
  "fmt"
  "math"
  "os"

  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials/insecure"

  "taoniu.local/security/grpc/aes"
)

type Aes struct {
  Ctx       context.Context
  AesClient aes.AesClient
}

func (srv *Aes) Client() aes.AesClient {
  if srv.AesClient == nil {
    conn, err := grpc.NewClient(
      fmt.Sprintf("%v:%v", os.Getenv("SECURITY_AES_GRPC_HOST"), os.Getenv("SECURITY_AES_GRPC_PORT")),
      grpc.WithTransportCredentials(insecure.NewCredentials()),
      grpc.WithDefaultCallOptions(
        grpc.MaxCallRecvMsgSize(math.MaxInt64),
        grpc.MaxCallSendMsgSize(math.MaxInt64),
      ),
    )
    if err != nil {
      panic(err.Error())
    }
    srv.AesClient = aes.NewAesClient(conn)
  }
  return srv.AesClient
}

func (srv *Aes) Decrypt(a, b, c string) (r *aes.DecryptReply, err error) {
  return srv.Client().Decrypt(srv.Ctx, &aes.DecryptRequest{A: a, B: b, C: c})
}
