package repositories

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

type TrackersRepository struct{}

func (r *TrackersRepository) Black() ([]string, error) {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	session := &net.Dialer{}
	tr.DialContext = session.DialContext
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	url := "https://raw.githubusercontent.com/ngosang/trackerslist/master/blacklist.txt"
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

	var trackers []string
	for _, tracker := range strings.Split(string(body), "\n") {
		tracker = strings.Trim(tracker, "\r\n ")
		if len(tracker) != 0 {
			trackers = append(trackers, tracker)
		}
	}
	return trackers, nil
}

func (r *TrackersRepository) Crawl() ([]string, error) {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	session := &net.Dialer{}
	tr.DialContext = session.DialContext
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	url := "https://raw.githubusercontent.com/ngosang/trackerslist/master/trackers_all.txt"
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

	var trackers []string
	for _, tracker := range strings.Split(string(body), "\n") {
		tracker = strings.Trim(tracker, "\r\n ")
		if len(tracker) != 0 {
			trackers = append(trackers, tracker)
		}
	}
	return trackers, nil
}
