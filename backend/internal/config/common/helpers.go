package common

import (
	"os"
	"strconv"
	"strings"
)

func GetenvTrim(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func WithDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func GetInt(key string, fallback int) int {
	val := GetenvTrim(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return parsed
}

func GetPositiveInt(key string, fallback int) int {
	val := GetenvTrim(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(val)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func GetBool(key string, fallback bool) bool {
	val := GetenvTrim(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return parsed
}
