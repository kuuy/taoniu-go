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
	val := os.Getenv(key)
	result, err := strconv.Atoi(val)
	if err != nil {
		panic(err.Error())
	}
	return result
}

func GetEnvArray(key string) []string {
	var result []string
	i := 1
	for {
		val := os.Getenv(fmt.Sprintf("%v_%v", key, i))
		if "" == val {
			return result
		}
		result = append(result, val)
		i++
	}
}
