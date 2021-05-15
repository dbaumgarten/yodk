package number

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Number is a fixed-point 3-digit number type
type Number int64

// Zero as constant
const Zero = Number(0)

// One as constant
const One = Number(1000)

// MaxValue is the largest possible value
const MaxValue = Number(math.MaxInt64)

// MinValue is the smallest possible value
const MinValue = Number(math.MinInt64)

const scale Number = 1000
const decimals = 3

// FromString parses a string into a Number
func FromString(str string) (Number, error) {
	parts := strings.Split(str, ".")
	if len(parts) > 2 {
		return Zero, fmt.Errorf("invalid number %s", str)
	}
	if len(parts) == 1 {
		num, err := strconv.Atoi(parts[0])
		if err != nil {
			return Zero, err
		}
		return FromInt(num), nil
	}
	if len(parts) == 2 {
		if len(parts[1]) > decimals {
			parts[1] = parts[1][:decimals]
		}
		parts[1] += strings.Repeat("0", decimals-len(parts[1]))
		num, err := strconv.Atoi(parts[0] + parts[1])
		if err != nil {
			return Zero, err
		}
		return Number(num), nil
	}
	return Zero, nil
}

// MustFromString parses a string into a Number. Panics on parsing erros
func MustFromString(str string) Number {
	res, err := FromString(str)
	if err != nil {
		panic(err)
	}
	return res
}

// FromInt creates a Number from the given int
func FromInt(in int) Number {
	return Number(int64(in)) * scale
}

// FromFloat64 creates a Number from the given float
func FromFloat64(in float64) Number {
	return Number(int64(in * float64(scale)))
}

// FromFloat32 creates a Number from the given float
func FromFloat32(in float32) Number {
	return Number(int32(in * float32(scale)))
}

// Float64 returns the value of the number as float
func (n Number) Float64() float64 {
	return float64(n) / float64(scale)
}

// Float32 returns the value of the number as float
func (n Number) Float32() float32 {
	return float32(n) / float32(scale)
}

// String returns the value of the number as string
func (n Number) String() string {
	prefix := int(n / scale)
	str := strconv.Itoa(prefix)
	remainder := (n % scale).Abs()
	if remainder != 0 {
		deci := strconv.Itoa(int(remainder))
		str += "." + strings.Repeat("0", 3-len(deci)) + deci
		str = strings.TrimRight(str, "0")
		if n < 0 && prefix >= 0 {
			str = "-" + str
		}
	}
	return str
}

// Int returns the value of the number as int
func (n Number) Int() int {
	return int(n / 1000)
}

// Add adds two numbers
func (n Number) Add(m Number) Number {
	return n + m
}

// Sub substracts two numbers
func (n Number) Sub(m Number) Number {
	return n - m
}

// Mul multiplicates two numbers
func (n Number) Mul(m Number) Number {
	return (n * m) / scale
}

// Div divides two numbers
func (n Number) Div(m Number) (Number, error) {
	if m == 0 {
		return Zero, fmt.Errorf("Division by 0")
	}
	return (n * scale) / m, nil
}

// Abs returns the absolute value of the number
func (n Number) Abs() Number {
	if n >= 0 {
		return n
	}
	return n * -1
}

// Sqrt returns the square root of the number
func (n Number) Sqrt() Number {
	return FromFloat64(math.Sqrt(n.Float64()))
}

// Mod returns the modulus of the number
func (n Number) Mod(m Number) (Number, error) {
	if m == Zero {
		return Zero, fmt.Errorf("Division by 0")
	}
	return n % m, nil
}

// Pow exponentiates the number
func (n Number) Pow(m Number) Number {
	res := math.Pow(n.Float64(), m.Float64())
	if math.IsInf(res, 1) {
		return MaxValue
	}
	if math.IsInf(res, -1) {
		return MinValue
	}
	return FromFloat64(math.Pow(n.Float64(), m.Float64()))
}

// convert given degree to radians
func toRad(n float32) float32 {
	return n * math.Pi / 180
}

// convert given radian to degrees
func toDeg(n float32) float32 {
	return n * 180 / math.Pi
}

// execute the given trigonometric function with the given argument, but only use 32bits of precision (to match the game's implementation)
func reducedPresisionTrig(f func(float64) float64, arg float32) float32 {
	return float32(f(float64(arg)))
}

// Sin returns the sin of the number (in degrees)
func (n Number) Sin() Number {
	return FromFloat32(reducedPresisionTrig(math.Sin, toRad(n.Float32())))
}

// Cos returns the cos of the number (in degrees)
func (n Number) Cos() Number {
	return FromFloat32(reducedPresisionTrig(math.Cos, toRad(n.Float32())))
}

// Tan returns the tan of the number (in degrees)
func (n Number) Tan() Number {
	rads := toRad(n.Float32())
	i := int64(math.Tan(float64(rads)) * 1000)
	return Number(i)
}

// Asin returns the asin of the number in degrees
func (n Number) Asin() Number {
	if n > One || n < -One {
		return MinValue
	}
	return FromFloat32(toDeg(reducedPresisionTrig(math.Asin, n.Float32())))
}

// Acos returns the acos of the number in degrees
func (n Number) Acos() Number {
	if n > One || n < -One {
		return MinValue
	}
	return FromFloat32(toDeg(reducedPresisionTrig(math.Acos, n.Float32())))
}

// Atan returns the atan of the number in degrees
func (n Number) Atan() Number {
	return FromFloat32(toDeg(reducedPresisionTrig(math.Atan, n.Float32())))
}

// Factorial computes the factorial (!)
func (n Number) Factorial() Number {
	if n < 0 {
		return MinValue
	}
	res := 1
	i := 0
	for n > 0 {
		i++
		n = n.Sub(One)
		res *= i
	}
	return FromInt(res)
}
