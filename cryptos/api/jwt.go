package api

import (
  "net/http"
  "strings"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/account"
)

func Authenticator(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    response := &ResponseHandler{}
    response.Jwe = &common.Jwe{}

    bearer := r.Header.Get("Authorization")
    if len(bearer) <= 7 || (strings.ToUpper(bearer[0:6]) != "TAONIU" && strings.ToUpper(bearer[0:6]) != "BEARER") {
      response.Error(w, http.StatusForbidden, 403, "authorization header missing or invalid")
      return
    }

    repository := &repositories.TokenRepository{}
    uid, err := repository.Uid(bearer[7:])
    if err != nil {
      if uid != "" {
        response.Error(w, http.StatusUnauthorized, 401, "invalid token")
      } else {
        response.Error(w, http.StatusForbidden, 403, "invalid token")
      }
      return
    }
    r.Header.Set("uid", uid)

    next.ServeHTTP(w, r)
  })
}
