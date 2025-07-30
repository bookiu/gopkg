package unit

import (
	"strconv"
	"strings"
)

// Htokb 将字符串单位转换为kb
func Htokb(h string) (int, error) {
	h = strings.ToLower(h)
	if strings.HasSuffix(h, "t") {
		trimmed := strings.TrimSuffix(h, "t")
		gb, err := strconv.Atoi(trimmed)
		if err != nil {
			return 0, err
		}
		return gb * 1024 * 1024 * 1024, nil
	} else if strings.HasSuffix(h, "g") {
		trimmed := strings.TrimSuffix(h, "g")
		gb, err := strconv.Atoi(trimmed)
		if err != nil {
			return 0, err
		}
		return gb * 1024 * 1024, nil
	} else if strings.HasSuffix(h, "m") {
		trimmed := strings.TrimSuffix(h, "m")
		mb, err := strconv.Atoi(trimmed)
		if err != nil {
			return 0, err
		}
		return mb * 1024, nil
	} else if strings.HasSuffix(h, "k") {
		trimmed := strings.TrimSuffix(h, "k")
		kb, err := strconv.Atoi(trimmed)
		if err != nil {
			return 0, err
		}
		return kb, nil
	}
	return strconv.Atoi(h)
}
