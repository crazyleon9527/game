package utils

import (
	"fmt"
	"math"
)

// func AddPrecise(base float64, val float64) float64 {
// 	decimalBase := decimal.NewFromFloat(base)
// 	decimalAdd := decimal.NewFromFloat(val)
// 	decimalResult := decimalBase.Add(decimalAdd).Round(3)
// 	result, _ := decimalResult.Float64() // Handle this error in production code

// 	// if result < 0 {
// 	// 	// Print the name of the function that called AddPrecise
// 	// 	if pc, _, _, ok := runtime.Caller(1); ok {
// 	// 		f := runtime.FuncForPC(pc)
// 	// 		fmt.Printf("AddPrecise was called by: %s\n", f.Name())
// 	// 	}
// 	// }

// 	if result <= 0 {
// 		result = 0.0000001 // Set to a minimum non-zero value to ensure DB update
// 	}
// 	return result
// }

// 六次方程计算器 y=ax^6+bx^4+cx^3+dx^2+ex+f
type HexEquation struct {
	a, b, c, d, e, f float64
}

func NewHexEquation(a, b, c, d, e, f float64) *HexEquation {
	return &HexEquation{a: a, b: b, c: c, d: d, e: e, f: f}
}

// 计算方程值
func (h *HexEquation) ff(x, y float64) float64 {
	return h.a*math.Pow(x, 6) + h.b*math.Pow(x, 4) + h.c*math.Pow(x, 3) +
		h.d*math.Pow(x, 2) + h.e*x + h.f - y
}

// 计算导数值
func (h *HexEquation) dff(x float64) float64 {
	return 6*h.a*math.Pow(x, 5) + 4*h.b*math.Pow(x, 3) + 3*h.c*math.Pow(x, 2) +
		2*h.d*x + h.e
}

// 牛顿法求解器（保留两位小数）
func (h *HexEquation) Solve(y float64) (float64, error) {
	// 动态初始猜测
	x0 := 0.0
	switch {
	case y > 2.0:
		x0 = 50.0
	case y < 0.95:
		x0 = -10.0
	}

	const (
		maxIter = 50
		epsilon = 1e-6
	)

	for i := 0; i < maxIter; i++ {
		fx := h.ff(x0, y)
		dfx := h.dff(x0)

		if math.Abs(fx) < epsilon {
			return math.Round(x0*100) / 100, nil // 保留两位小数‌:ml-citation{ref="7,8" data="citationList"}
		}

		if dfx == 0 {
			return 0, fmt.Errorf("导数为零，无法收敛")
		}

		x0 -= fx / dfx
	}
	return 0, fmt.Errorf("超过最大迭代次数")
}

func (h *HexEquation) Result(x float64) float64 {
	return h.a*math.Pow(x, 6) + h.b*math.Pow(x, 4) + h.c*math.Pow(x, 3) +
		h.d*math.Pow(x, 2) + h.e*x + h.f
}
