package wolf

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"taoniu.local/gamblings/common"
	config "taoniu.local/gamblings/config/wolf"
)

type AccountRepository struct{}

func (r *AccountRepository) Balance() error {
	//session := &net.Dialer{}
	session := &common.ProxySession{
		Proxy: "socks5://127.0.0.1:1080?timeout=5s",
	}
	tr := &http.Transport{
		DialContext:       session.DialContext,
		DisableKeepAlives: true,
	}
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}

	url := "https://wolf.bet/api/v1/user/balances"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.API_TOKEN))

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("request error: status[%s] code[%d]", resp.Status, resp.StatusCode))
	}

	log.Println("body", resp.Body)

	return nil
}
