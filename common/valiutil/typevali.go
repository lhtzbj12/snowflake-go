package valiutil

import "regexp"

// IsNumber 是否数字
func IsNumber(str string) bool {
	return Regexp(`\d+`, str)
}

// Regexp 是否匹配正则
func Regexp(pattern, str string) bool {
	v := regexp.MustCompile(pattern)
	return v.MatchString(str)
}
