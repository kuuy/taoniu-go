package tradingview

import (
	"errors"
	"gorm.io/gorm"
	"strings"
	models "taoniu.local/cryptos/models/tradingview"
)

type AnalysisRepository struct {
	Db *gorm.DB
}

func NewAnalysisRepository(db *gorm.DB) *AnalysisRepository {
	return &AnalysisRepository{
		Db: db,
	}
}

func (r *AnalysisRepository) Signal(symbol string) (int64, error) {
	var entity models.Analysis
	result := r.Db.Where(
		"symbol=?",
		symbol,
	).Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, result.Error
	}
	var signal int64 = 0
	if strings.Contains(entity.Summary["RECOMMENDATION"].(string), "BUY") {
		signal = 1
	}
	if strings.Contains(entity.Summary["RECOMMENDATION"].(string), "SELL") {
		signal = 2
	}

	return signal, nil
}
