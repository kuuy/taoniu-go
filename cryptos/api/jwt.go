package api

import (
	"log"
	"net/http"
)

type JwtHandler struct{}

func Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("auth before")
		next.ServeHTTP(w, r)
		log.Println("auth after")
	})
}
