package mqtt

import (
  services "taoniu.local/cryptos/grpc/services/account/mqtt"
)

type PublishersRepository struct {
  Service *services.Publishers
}

func (r *PublishersRepository) Token(id string) (token *Token, err error) {
  reply, err := r.Service.Token(id)
  if err == nil {
    token = &Token{
      AccessToken: reply.AccessToken,
    }
  }
  return
}
