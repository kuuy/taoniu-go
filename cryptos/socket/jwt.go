package socket

import (
  "errors"

  repositories "taoniu.local/cryptos/repositories/account"
)

type JwtHandler struct{}

func (h *JwtHandler) Authenticate(accessToken string) (string, error) {
  if accessToken == "" {
    return "", errors.New("access_token is required")
  }

  repository := &repositories.TokenRepository{}
  uid, err := repository.Uid(accessToken)
  if err != nil {
    if uid != "" {
      return uid, err
    }
    return "", errors.New("access not allowed")
  }
  return uid, nil
}
