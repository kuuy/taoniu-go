package api

import (
	"net/http"
	"strings"
	repositories "taoniu.local/cryptos/repositories/account"
)

type JwtHandler struct{}

func Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := &ResponseHandler{
			Writer: w,
		}

		bearer := r.Header.Get("Authorization")
		if len(bearer) <= 7 || strings.ToUpper(bearer[0:6]) != "TAONIU" {
			response.Error(http.StatusInternalServerError, 403, "access not allowed")
			return
		}

		repository := &repositories.TokenRepository{}
		uid, err := repository.Uid(bearer[7:])
		if err != nil {
			response.Error(http.StatusInternalServerError, 403, "access not valid")
			return
		}
		r.Header.Set("uid", uid)

		next.ServeHTTP(w, r)
	})
}
