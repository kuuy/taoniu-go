package tradingview

import (
	scanner "github.com/dematron/go-tvscanner"
)

type ScannerRepository struct{}

func (r *ScannerRepository) Scan(exchange string, symbol string, interval string) (*scanner.RecommendSummary, error) {
	s := scanner.New()
	analysis, err := s.GetAnalysis("crypto", exchange, symbol, interval)
	if err != nil {
		return nil, err
	}

	return &analysis, nil
}
