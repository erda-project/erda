package math

func AbsInt(x int) int {
	y := x >> 31
	return (x ^ y) - y
}

func AbsInt32(x int32) int32 {
	y := x >> 31
	return (x ^ y) - y
}

func AbsInt64(x int64) int64 {
	y := x >> 63
	return (x ^ y) - y
}