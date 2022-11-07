package hilo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"taoniu.local/gamblings/common"
	"time"

	config "taoniu.local/gamblings/config/wolf"
)

type BetRepository struct {
	UseProxy bool
}

type BetRequest struct {
	Currency   string  `json:"currency"`
	Multiplier string  `json:"multiplier"`
	Amount     string  `json:"amount"`
	Rule       string  `json:"rule"`
	BetValue   float64 `json:"bet_value"`
	SubNonce   int     `json:"sub_nonce"`
	WinChance  float64 `json:"win_chance"`
}

func (r *BetRepository) Place(request *BetRequest, limit int) (float64, int, int, error) {
	var betValue float64
	var status, subNonce int
	var err error

	subNonce = request.SubNonce
	for i := subNonce; i < limit; i++ {
		if betValue != 0 {
			request.BetValue = betValue
		}
		betValue, status, err = r.Play(request)
		if err != nil {
			return 0, 0, 0, err
		}
		if status != 3 {
			return betValue, status, 0, nil
		}
		request.SubNonce += 1
	}
	err = r.Finish()
	if err != nil {
		return 0, 0, 0, err
	}

	return betValue, status, 0, nil
}

func (r *BetRepository) Status(request *BetRequest) (string, float64, int, error) {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	if r.UseProxy {
		session := &common.ProxySession{
			Proxy: "socks5://127.0.0.1:1080?timeout=5s",
		}
		tr.DialContext = session.DialContext
	} else {
		session := &net.Dialer{}
		tr.DialContext = session.DialContext
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}

	url := "https://wolf.bet/api/v1/user/hilo/status"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("X-Client-Type", "Web-Application")
	req.Header.Set("X-Hash-Api", config.LOGIN_HASH)
	req.Header.Set("Referer", "https://wolf.bet/casino/hilo")
	req.Header.Set("Cookie", config.LOGIN_COOKIE)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.LOGIN_TOKEN))
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, 0, errors.New(fmt.Sprintf("request error: status[%s] code[%d]", resp.Status, resp.StatusCode))
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	if _, ok := data["bet"]; !ok {
		return "", 0, 0, errors.New("bet not started")
	}
	bet := data["bet"].(map[string]interface{})

	if _, ok := bet["hash"]; !ok {
		return "", 0, 0, errors.New("hash not exists")
	}
	hash, _ := bet["hash"].(string)

	if _, ok := bet["initial_bet_value"]; !ok {
		return "", 0, 0, errors.New("initial bet value not exists")
	}
	betValue := bet["initial_bet_value"].(float64)

	if _, ok := bet["status"]; !ok {
		return "", 0, 0, errors.New("status not exists")
	}
	status := int(bet["status"].(float64))
	if status != 3 && status != 4 {
		return "", 0, 0, errors.New("bet status not valid")
	}

	if _, ok := bet["bets"]; !ok {
		return "", 0, 0, errors.New("bets not exists")
	}
	bets := bet["bets"].([]interface{})
	subNonce := len(bets)

	if len(bets) > 0 {
		result := bets[len(bets)-1].(map[string]interface{})
		if _, ok := result["result_value"]; !ok {
			return "", 0, 0, errors.New("bet value not exists")
		}
		betValue, _ = result["result_value"].(float64)
	}

	return hash, betValue, subNonce, nil
}

func (r *BetRepository) Start(request *BetRequest) (string, float64, error) {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	if r.UseProxy {
		session := &common.ProxySession{
			Proxy: "socks5://127.0.0.1:1080?timeout=5s",
		}
		tr.DialContext = session.DialContext
	} else {
		session := &net.Dialer{}
		tr.DialContext = session.DialContext
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}

	body, _ := json.Marshal(&BetRequest{
		Currency:   request.Currency,
		Multiplier: "0",
		Amount:     request.Amount,
		Rule:       "start",
		BetValue:   request.BetValue,
		SubNonce:   request.SubNonce,
		WinChance:  0,
	})

	url := "https://wolf.bet/api/v1/user/hilo/play"
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("X-Client-Type", "Web-Application")
	req.Header.Set("X-Hash-Api", config.LOGIN_HASH)
	req.Header.Set("Referer", "https://wolf.bet/casino/hilo")
	req.Header.Set("Cookie", config.LOGIN_COOKIE)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.LOGIN_TOKEN))
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, errors.New(fmt.Sprintf("request error: status[%s] code[%d]", resp.Status, resp.StatusCode))
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	if _, ok := data["bet"]; !ok {
		return "", 0, errors.New("bet not exists")
	}
	bet := data["bet"].(map[string]interface{})

	if _, ok := bet["hash"]; !ok {
		return "", 0, errors.New("hash not exists")
	}
	hash, _ := bet["hash"].(string)

	if _, ok := bet["initial_bet_value"]; !ok {
		return "", 0, errors.New("initial bet value not exists")
	}
	betValue := bet["initial_bet_value"].(float64)

	if _, ok := bet["status"]; !ok {
		return "", 0, errors.New("status not exists")
	}
	status := int(bet["status"].(float64))
	if status != 4 {
		return "", 0, errors.New("bet started failed")
	}

	return hash, betValue, nil
}

func (r *BetRepository) Finish() error {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	if r.UseProxy {
		session := &common.ProxySession{
			Proxy: "socks5://127.0.0.1:1080?timeout=5s",
		}
		tr.DialContext = session.DialContext
	} else {
		session := &net.Dialer{}
		tr.DialContext = session.DialContext
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}

	url := "https://wolf.bet/api/v1/user/hilo/finish"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("X-Client-Type", "Web-Application")
	req.Header.Set("X-Hash-Api", config.LOGIN_HASH)
	req.Header.Set("Referer", "https://wolf.bet/casino/hilo")
	req.Header.Set("Cookie", config.LOGIN_COOKIE)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.LOGIN_TOKEN))
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("request error: status[%s] code[%d]", resp.Status, resp.StatusCode))
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	if _, ok := data["bet"]; !ok {
		return nil
	}

	bet := data["bet"].(map[string]interface{})
	if _, ok := bet["status"]; !ok {
		return nil
	}
	status := int(bet["status"].(float64))
	if status != 1 && status != 3 {
		return errors.New("bet finished failed")
	}

	return nil
}

func (r *BetRepository) Play(request *BetRequest) (float64, int, error) {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	if r.UseProxy {
		session := &common.ProxySession{
			Proxy: "socks5://127.0.0.1:1080?timeout=5s",
		}
		tr.DialContext = session.DialContext
	} else {
		session := &net.Dialer{}
		tr.DialContext = session.DialContext
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}

	body, _ := json.Marshal(request)

	url := "https://wolf.bet/api/v1/user/hilo/play"
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("X-Client-Type", "Web-Application")
	req.Header.Set("X-Hash-Api", config.LOGIN_HASH)
	req.Header.Set("Referer", "https://wolf.bet/casino/hilo")
	req.Header.Set("Cookie", config.LOGIN_COOKIE)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.LOGIN_TOKEN))
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, errors.New(fmt.Sprintf("request error: status[%s] code[%d]", resp.Status, resp.StatusCode))
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	if _, ok := data["bet"]; !ok {
		return 0, 0, errors.New("bet not exists")
	}

	bet := data["bet"].(map[string]interface{})
	if _, ok := bet["status"]; !ok {
		return 0, 0, errors.New("status not exists")
	}
	status := int(bet["status"].(float64))

	if _, ok := bet["bets"]; !ok {
		return 0, 0, errors.New("bets not exists")
	}
	bets := bet["bets"].([]interface{})
	if len(bets) == 0 {
		return 0, 0, errors.New("bets result empty")
	}

	result := bets[len(bets)-1].(map[string]interface{})
	if _, ok := result["result_value"]; !ok {
		return 0, 0, errors.New("bet value not exists")
	}
	betValue, _ := result["result_value"].(float64)

	return betValue, status, nil
}
