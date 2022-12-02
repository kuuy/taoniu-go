package gfw

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/bluele/adblock"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

type CrawlRepository struct{}

func (r *CrawlRepository) Flush() (*adblock.Rules, error) {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	session := &net.Dialer{}
	tr.DialContext = session.DialContext
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	url := "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt"
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("request error: status[%s] code[%d]", resp.Status, resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content, _ := base64.StdEncoding.DecodeString(string(body))

	var rules []string
	for _, rule := range strings.Split(string(content), "\n") {
		rule = strings.Trim(rule, "\r\n ")
		if len(rule) != 0 {
			rules = append(rules, rule)
		}
	}

	ab, err := adblock.NewRules(rules, nil)
	if err != nil {
		return nil, err
	}

	return ab, err
}
