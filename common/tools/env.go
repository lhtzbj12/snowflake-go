package tools

import (
	"os"
)

// GetEnv 获取环境变量值
func GetEnv(name, defaultValue string) string {
	value := os.Getenv(name)
	if value != "" {
		return value
	} else {
		return defaultValue
	}
}

// SetEnv 设置环境变量值
func SetEnv(name, value string) error {
	return os.Setenv(name, value)
}
