package utils

import (
	"math"
)

// Round 保留小数点计算
func Round(f float64, n int) float64 {
	shift := math.Pow(10, float64(n))
	fv := 0.0000000001 + f //对浮点数产生.xxx999999999 计算不准进行处理

	return math.Floor(fv*shift+.5) / shift
}
