package tradingview

import (
	scanner "github.com/dematron/go-tvscanner"
	"net"
	"net/http"
	"taoniu.local/cryptos/common"
	"time"
)

type ScannerRepository struct {
	UseProxy bool
}

func (r *ScannerRepository) Scan(exchange string, symbol string, interval string) (*scanner.RecommendSummary, error) {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	if r.UseProxy {
		session := &common.ProxySession{
			Proxy: "socks5://127.0.0.1:1088?timeout=5s",
		}
		tr.DialContext = session.DialContext
	} else {
		session := &net.Dialer{}
		tr.DialContext = session.DialContext
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(8) * time.Second,
	}

	s := scanner.NewWithCustomHttpClient(httpClient)
	analysis, err := s.GetAnalysis("crypto", exchange, symbol, interval)
	if err != nil {
		return nil, err
	}

	return &analysis, nil
}
