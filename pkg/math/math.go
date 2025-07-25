package math

import (
	"fmt"
	"strconv"
)

// MustParsePrecFloat64 按小数位截取float
func MustParsePrecFloat64(value float64, prec int) float64 {
	format := fmt.Sprintf("%%.%df", prec)
	v, err := strconv.ParseFloat(fmt.Sprintf(format, value), 64)
	if err != nil {
		return 0
	}
	return v
}

// MustParsePrecFloat64 按小数位截取float
func MustParsePrecFloat32(value float32, prec int) float32 {
	format := fmt.Sprintf("%%.%df", prec)
	v, err := strconv.ParseFloat(fmt.Sprintf(format, value), 32)
	if err != nil {
		return 0
	}
	return float32(v)
}
