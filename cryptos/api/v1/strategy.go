package v1

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	pool "taoniu.local/cryptos/common"
	"taoniu.local/cryptos/repositories"
	"time"
)

type StrategyHandler struct {
	repository *repositories.StrategyRepository
}

type Strategy struct {
	ID              string  `json:"id"`
	Symbol          string  `json:"symbol"`
	Indicator       string  `json:"indicator"`
	Price           float64 `json:"price"`
	Signal          int64   `json:"signal"`
	CreatedAt       string  `json:"created_at"`
	CreatedAtFormat string  `json:"created_at_fmt"`
}

type StrategyDetail struct {
	ID        string  `json:"id"`
	Symbol    string  `json:"symbol"`
	Indicator string  `json:"indicator"`
	Price     float64 `json:"price"`
	Signal    int64   `json:"signal"`
	CreatedAt string  `json:"created_at"`
}

type ListStrategyResponse struct {
	Success  bool       `json:"success"`
	Data     []Strategy `json:"data"`
	Total    int64      `json:"total"`
	PageSize int        `json:"pageSize"`
	Current  int        `json:"current"`
}

func NewStrategyRouter() http.Handler {
	db := pool.NewDB()
	repository := repositories.NewStrategyRepository(db)

	handler := StrategyHandler{
		repository: repository,
	}

	r := chi.NewRouter()
	r.Get("/", handler.Listings)

	return r
}

func (h *StrategyHandler) Listings(
	w http.ResponseWriter,
	r *http.Request,
) {
	current, err := strconv.Atoi(r.URL.Query().Get("current"))
	if err != nil {
		return
	}
	if current < 1 {
		return
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil {
		return
	}
	if pageSize > 500 {
		return
	}

	total, err := h.repository.Count()
	if err != nil {
		return
	}
	strategies, err := h.repository.Listings(current, pageSize)
	if err != nil {
		return
	}

	var response ListStrategyResponse
	response.Success = true
	response.Data = make([]Strategy, 0)
	response.Total = total
	response.PageSize = pageSize
	response.Current = current

	for _, entity := range strategies {
		var strategy Strategy
		strategy.ID = entity.ID
		strategy.Symbol = entity.Symbol
		strategy.Indicator = entity.Indicator
		strategy.Price = entity.Price
		strategy.Signal = entity.Signal
		strategy.CreatedAt = entity.CreatedAt.Format("2006-01-02 15:04:05")
		strategy.CreatedAtFormat = getTimeFormatText(entity.CreatedAt)

		response.Data = append(response.Data, strategy)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return
	}

	w.Write(jsonResponse)
}

func getTimeFormatText(datetime time.Time) string {
	now := time.Now()
	yestoday := now.AddDate(0, 0, -1)
	lastMonth := now.AddDate(0, -1, 0)
	diff := now.Sub(datetime)
	if diff.Seconds() <= 30 {
		return fmt.Sprintf("%.f秒前", diff.Seconds())
	}
	if diff.Minutes() < 1 {
		return "刚刚"
	}
	if diff.Minutes() < 30 {
		return fmt.Sprintf("%.f分钟前", diff.Minutes())
	}
	if diff.Hours() < 1 {
		return "半小时前"
	}
	if diff.Hours() < 3 {
		return fmt.Sprintf("%.f小时前", diff.Hours())
	}
	if diff.Hours() < 24 && now.Day() == datetime.Day() {
		return "今天"
	}
	if diff.Hours() < 48 && yestoday.Day() == datetime.Day() {
		return "昨天"
	}
	if now.Year() == datetime.Year() && now.Month() == datetime.Month() {
		return "当月"
	}
	if now.Year() == datetime.Year() && lastMonth.Month() == datetime.Month() {
		return "上个月"
	}
	if now.Year() == datetime.Year() {
		return "今年"
	}
	if now.Year() == datetime.Year()-1 {
		return "去年"
	} else {
		return fmt.Sprintf("%d", datetime.Year())
	}
}
