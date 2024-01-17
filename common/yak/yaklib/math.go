package yaklib

import "math"

var MathExport = map[string]interface{}{
	"Round":       Round,
	"Sqrt":        Sqrt,
	"Pow":         Pow,
	"Pow10":       Pow10,
	"Floor":       Floor,
	"Ceil":        Ceil,
	"RoundToEven": RoundToEven,
	"Abs":         Abs,
	"NaN":         NaN,
	"IsNaN":       IsNaN,
	"Pi":          math.Pi,
	"Ln10":        math.Ln10,
	"Ln2":         math.Ln2,
	"E":           math.E,
	"Sqrt2":       math.Sqrt2,
	"SqrtPi":      math.SqrtPi,
	"SqrtE":       math.SqrtE,
}

// Round Returns rounded to the nearest integer
// There are some special cases: Round(±0) = ±0, Round(±Inf) = ±Inf, Round(NaN) = NaN
// Example:
// ```
// math.Round(1.5) // 2
// math.Round(1.4) // 1
// ```
func Round(x float64) float64 {
	return math.Round(x)
}

// Sqrt Returns the square root of a number
// that is not less than x If x < 0, returns NaN
// Example:
// ```
// math.Sqrt(4) // 2
// math.Sqrt(-1) // NaN
// ```
func Sqrt(x float64) float64 {
	return math.Sqrt(x)
}

// Pow Returns x raised to the yth power
// Example:
// ```
// math.Pow(2, 3) // 8
// math.Pow(2, -1) // 0.5
// ```
func Pow(x, y float64) float64 {
	return math.Pow(x, y)
}

// Pow10 returns 10 to the nth power
// Example:
// ```
// math.Pow10(2) // 100
// math.Pow10(-1) // 0.1
// ```
func Pow10(n int) float64 {
	return math.Pow10(n)
}

// Floor Returns the largest integer
// Example:
// ```
// math.Floor(1.5) // 1
// math.Floor(-1.5) // -2
// ```
func Floor(x float64) float64 {
	return math.Floor(x)
}

// Ceil Returns the value
// Example:
// ```
// math.Ceil(1.5) // 2
// math.Ceil(-1.5) // -1
// ```
func Ceil(x float64) float64 {
	return math.Ceil(x)
}

// RoundToEven Returns rounded to the nearest even integer
// Example:
// ```
// math.RoundToEven(1.5) // 2
// math.RoundToEven(2.5) // 2
// math.RoundToEven(3.5) // 4
// math.RoundToEven(4.5) // 4
// ```
func RoundToEven(x float64) float64 {
	return math.RoundToEven(x)
}

// Abs Returns the absolute value of x
// Example:
// ```
// math.Abs(-1) // 1
// math.Abs(1) // 1
// ```
func Abs(x float64) float64 {
	return math.Abs(x)
}

// NaN Returns an IEEE-574 “not greater than x Non-number”of the smallest integer
// Example:
// ```
// math.NaN()
// ```
func NaN() float64 {
	return math.NaN()
}

// IsNaN Determine whether a number is NaN
// Example:
// ```
// math.IsNaN(1) // false
// math.IsNaN(math.NaN()) // true
// ```
func IsNaN(x float64) bool {
	return math.IsNaN(x)
}
