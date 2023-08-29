package common

import (
  "fmt"
  "os"
  "strconv"
)

func GetEnvString(key string) string {
  return os.Getenv(key)
}

func GetEnvInt(key string) int {
  value := os.Getenv(key)
  result, err := strconv.Atoi(value)
  if err != nil {
    panic(err.Error())
  }
  return result
}

func GetEnvInt64(key string) int64 {
  value := os.Getenv(key)
  result, err := strconv.ParseInt(value, 10, 64)
  if err != nil {
    panic(err.Error())
  }
  return result
}

func GetEnvArray(key string) []string {
  var result []string
  i := 1
  for {
    value := os.Getenv(fmt.Sprintf("%v_%v", key, i))
    if "" == value {
      return result
    }
    result = append(result, value)
    i++
  }
}
