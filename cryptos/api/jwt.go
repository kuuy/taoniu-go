package api

import (
  "log"
  "net/http"
  "strings"

  "taoniu.local/cryptos/repositories"
  accountRepositories "taoniu.local/cryptos/repositories/account"
)

func Authenticator(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    response := &ResponseHandler{}
    response.JweRepository = &repositories.JweRepository{}
    response.Writer = w

    bearer := r.Header.Get("Authorization")
    if len(bearer) <= 7 || strings.ToUpper(bearer[0:6]) != "TAONIU" {
      response.Error(http.StatusForbidden, 403, "access not allowed")
      return
    }

    repository := &accountRepositories.TokenRepository{}
    uid, err := repository.Uid(bearer[7:])
    if err != nil {
      log.Println("token error", err.Error())
      if uid != "" {
        response.Error(http.StatusUnauthorized, 401, err.Error())
      } else {
        response.Error(http.StatusForbidden, 403, "access not allowed")
      }
      return
    }
    r.Header.Set("uid", uid)

    next.ServeHTTP(w, r)
  })
}
