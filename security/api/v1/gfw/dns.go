package gfw

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"net/http"
	"strings"
	"taoniu.local/security/api"
	"taoniu.local/security/common"
	repositories "taoniu.local/security/repositories/gfw"
)

type DnsHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Response   *api.ResponseHandler
	Repository *repositories.DnsRepository
}

type DnsInfo struct {
	Hash   string  `json:"hash"`
	Number float64 `json:"number"`
}

func NewDnsRouter() http.Handler {
	h := DnsHandler{
		Rdb: common.NewRedis(),
		Ctx: context.Background(),
	}
	h.Repository = &repositories.DnsRepository{}

	r := chi.NewRouter()
	r.Get("/", h.Gets)

	return r
}

func (h *DnsHandler) Gets(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.Response = &api.ResponseHandler{
		Writer: w,
	}

	if r.URL.Query().Get("domains") == "" {
		h.Response.Error(http.StatusForbidden, 1004, "domains is empty")
		return
	}

	domains := strings.Split(r.URL.Query().Get("domains"), ",")

	data := h.Repository.Gets(domains)

	h.Response.Json(data)
}
