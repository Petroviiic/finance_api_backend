package env

import (
	"os"
	"strconv"
	"strings"
)

func GetString(key string, fallback string) string {
	if result, ok := os.LookupEnv(key); ok && len(result) > 0 {
		return result
	}
	return fallback
}

func GetInt(key string, fallback int) int {
	if result, ok := os.LookupEnv(key); ok && len(result) > 0 {
		res, err := strconv.Atoi(result)
		if err == nil {
			return res
		}
	}
	return fallback
}

func GetBool(key string, fallback bool) bool {
	if result, ok := os.LookupEnv(key); ok && len(result) > 0 {
		if strings.ToUpper(result) == "TRUE" {
			return true
		} else if strings.ToUpper(result) == "FALSE" {
			return false
		}
	}
	return fallback
}
