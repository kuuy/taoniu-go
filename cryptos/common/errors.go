package common

import "fmt"

type BinanceAPIError struct {
  Code    int64  `json:"code"`
  Message string `json:"msg"`
}

func (e BinanceAPIError) Error() string {
  return fmt.Sprintf("<APIError> code=%d, msg=%s", e.Code, e.Message)
}

func IsBinanceAPIError(e error) bool {
  _, ok := e.(*BinanceAPIError)
  return ok
}
