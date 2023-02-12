package convutil

import "strconv"

// SliceInt2Str int64切片转成字符串
func SliceInt2Str(src []int64) []string {
	str := make([]string, 0, len(src))
	for _, v := range src {
		str = append(str, strconv.FormatInt(v, 10))
	}
	return str
}
