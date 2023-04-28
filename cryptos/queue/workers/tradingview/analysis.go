package tradingview

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	repositories "taoniu.local/cryptos/repositories/tradingview"
)

type Analysis struct {
	Db                *gorm.DB
	Repository        *repositories.AnalysisRepository
	SymbolsRepository *spotRepositories.SymbolsRepository
}

func NewAnalysis() *Analysis {
	h := &Analysis{
		Db: common.NewDB(),
	}
	h.Repository = &repositories.AnalysisRepository{
		Db: h.Db,
	}
	h.SymbolsRepository = &spotRepositories.SymbolsRepository{
		Db: h.Db,
	}
	return h
}

type AnalysisFlushPayload struct {
	Exchange string
	Symbol   string
	Interval string
	UseProxy bool
}

func (h *Analysis) Flush(ctx context.Context, t *asynq.Task) error {
	var payload AnalysisFlushPayload
	json.Unmarshal(t.Payload(), &payload)

	if payload.UseProxy {
		h.Repository.UseProxy = true
	}

	h.Repository.Flush(payload.Exchange, payload.Symbol, payload.Interval)

	return nil
}
