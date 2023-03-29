package spot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	models "taoniu.local/cryptos/models/binance/spot"
)

type DepthRepository struct {
	Db       *gorm.DB
	UseProxy bool
}

func (r *DepthRepository) Flush(symbol string) error {
	depth, err := r.Request(symbol)
	if err != nil {
		return err
	}
	r.Db.Model(&models.Symbol{}).Where("symbol", symbol).Update("depth", depth)
	return nil
}

func (r *DepthRepository) Request(symbol string) (map[string]interface{}, error) {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	if r.UseProxy {
		session := &common.ProxySession{
			Proxy: "socks5://127.0.0.1:1088?timeout=8s",
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

	url := "https://api.binance.com/api/v3/depth"
	req, _ := http.NewRequest("GET", url, nil)
	q := req.URL.Query()
	q.Add("symbol", symbol)
	q.Add("limit", "1000")
	req.URL.RawQuery = q.Encode()
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(
			fmt.Sprintf(
				"request error: status[%s] code[%d]",
				resp.Status,
				resp.StatusCode,
			),
		)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}
