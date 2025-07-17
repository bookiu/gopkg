package str

import "strings"

// SplitByComma 按逗号分隔字符串
func SplitByComma(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}
