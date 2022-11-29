package common

import (
	"fmt"
	"math"
	"time"
)

func FormatDatetime(datetime time.Time) string {
	now := time.Now()
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
	days := math.Floor((diff.Hours() + float64(23-now.Hour())) / 24)
	if days < 1 {
		return datetime.Format("15:04")
	}
	if days == 1 {
		return "昨天"
	}
	if days == 2 {
		return "前天"
	}
	if days == 3 {
		return "大前天"
	}
	if days > 10 {
		return "十天前"
	}
	if days > 5 {
		return "五天前"
	}
	if days > 3 {
		return "三天前"
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
