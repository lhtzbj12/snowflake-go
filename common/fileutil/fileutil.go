package fileutil

import "os"

// Exists 判断文件是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
