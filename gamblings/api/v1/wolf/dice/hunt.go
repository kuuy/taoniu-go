package dice

import (
	"context"
	"github.com/go-redis/redis/v8"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"taoniu.local/gamblings/api"
	"taoniu.local/gamblings/common"
	repositories "taoniu.local/gamblings/repositories/wolf/dice"
)

type HuntHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Response   *api.ResponseHandler
	Repository *repositories.HuntRepository
}

type HuntInfo struct {
	Hash   string  `json:"hash"`
	Number float64 `json:"number"`
}

func NewHuntRouter() http.Handler {
	h := HuntHandler{
		Rdb: common.NewRedis(),
		Ctx: context.Background(),
	}
	h.Repository = &repositories.HuntRepository{
		Db:  common.NewDB(),
		Rdb: h.Rdb,
		Ctx: h.Ctx,
	}

	r := chi.NewRouter()
	r.Get("/", h.Gets)
	r.Get("/start", h.Start)

	return r
}

func (h *HuntHandler) Gets(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.Response = &api.ResponseHandler{
		Writer: w,
	}

	conditions := make(map[string]interface{})

	if r.URL.Query().Get("numbers") != "" {
		values := strings.Split(r.URL.Query().Get("numbers"), ",")
		numbers := make([]float64, len(values))
		for i := 0; i < len(values); i++ {
			numbers[i], _ = strconv.ParseFloat(values[i], 64)
		}
		conditions["numbers"] = numbers
	}

	if r.URL.Query().Get("side") != "" {
		side, _ := strconv.Atoi(r.URL.Query().Get("side"))
		conditions["side"] = side
	}

	if r.URL.Query().Get("ipart") != "" {
		var numbers []int
		ranges := strings.Split(r.URL.Query().Get("ipart"), "-")
		if len(ranges) == 2 {
			min, _ := strconv.Atoi(ranges[0])
			max, _ := strconv.Atoi(ranges[1])
			for i := min; i < max; i++ {
				numbers = append(numbers, i)
			}
		} else {
			values := strings.Split(r.URL.Query().Get("ipart"), ",")
			for i := 0; i < len(values); i++ {
				value, _ := strconv.Atoi(values[i])
				numbers = append(numbers, value)
			}
		}
		conditions["ipart"] = numbers
	}

	if r.URL.Query().Get("dpart") != "" {
		var numbers []int
		ranges := strings.Split(r.URL.Query().Get("dpart"), "-")
		if len(ranges) == 2 {
			min, _ := strconv.Atoi(ranges[0])
			max, _ := strconv.Atoi(ranges[1])
			for i := min; i < max; i++ {
				numbers = append(numbers, i)
			}
		} else {
			values := strings.Split(r.URL.Query().Get("dpart"), ",")
			for i := 0; i < len(values); i++ {
				value, _ := strconv.Atoi(values[i])
				numbers = append(numbers, value)
			}
		}
		conditions["dpart"] = numbers
	}

	if r.URL.Query().Get("is_mirror") == "1" {
		conditions["is_mirror"] = true
	}

	if r.URL.Query().Get("is_repeate") == "1" {
		conditions["is_repeate"] = true
	}

	if r.URL.Query().Get("is_neighbor") == "1" {
		conditions["is_neighbor"] = true
	}

	score, _ := h.Rdb.ZScore(
		h.Ctx,
		"wolf:hunts",
		"dice",
	).Result()
	if score > 0 {
		conditions["opentime"] = time.Unix(int64(score), 0)
	}

	hunts := h.Repository.Gets(conditions)
	data := make([]*HuntInfo, len(hunts))
	for i, hunt := range hunts {
		data[i] = &HuntInfo{
			Hash:   hunt.Hash,
			Number: hunt.Number,
		}
	}

	h.Response.Json(data)
}

func (h *HuntHandler) Start(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.Response = &api.ResponseHandler{
		Writer: w,
	}
	h.Repository.Start()
	h.Response.Json(nil)
}
