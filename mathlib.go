package lua

import (
	"fmt"
	"math"
)

const (
	piInDegree = 180
)

func OpenMath(L *LState) int {
	mod := L.RegisterModule(MathLibName, mathFuncs).(*LTable)
	mod.RawSetString("pi", LNumber(math.Pi))
	mod.RawSetString("huge", LNumber(math.MaxFloat64))
	L.Push(mod)
	return 1
}

var MatLibFuncDoc = map[string]libFuncDoc{
	MathLibName: {
		libName: "mathlib",
		libFuncName: []string{
			"Abs",
			"Acos",
			"Asin",
			"Atan",
			"Atan2",
			"Ceil",
			"Cos",
			"Cosh",
			"Deg",
			"Exp",
			"Floor",
			"Fmod",
			"Frexp",
			"Ldexp",
			"Ln",
			"Log",
			"Max",
			"Min",
			"Mod",
			"Modf",
			"Pow",
			"Rad",
			"Sin",
			"Sinh",
			"Sqrt",
			"Tan",
			"Tanh",
		},
	},
}

var mathFuncs = map[string]LGFunction{
	"Abs":   mathAbs,
	"Acos":  mathAcos,
	"Asin":  mathAsin,
	"Atan":  mathAtan,
	"Atan2": mathAtan2,
	"Ceil":  mathCeil,
	"Cos":   mathCos,
	"Cosh":  mathCosh,
	"Deg":   mathDeg,
	"Exp":   mathExp,
	"Floor": mathFloor,
	"Fmod":  mathFmod,
	"Frexp": mathFrexp,
	"Ldexp": mathLdexp,
	"Ln":    mathLn,
	"Log":   mathLog,
	"Max":   mathMax,
	"Min":   mathMin,
	"Mod":   mathMod,
	"Modf":  mathModf,
	"Pow":   mathPow,
	"Rad":   mathRad,
	"Sin":   mathSin,
	"Sinh":  mathSinh,
	"Sqrt":  mathSqrt,
	"Tan":   mathTan,
	"Tanh":  mathTanh,
}

func mathAbs(L *LState) int {
	L.Push(LNumber(math.Abs(float64(L.CheckNumber(1)))))
	return 1
}

func mathAcos(L *LState) int {
	L.Push(LNumber(math.Acos(float64(L.CheckNumber(1)))))
	return 1
}

func mathAsin(L *LState) int {
	L.Push(LNumber(math.Asin(float64(L.CheckNumber(1)))))
	return 1
}

func mathAtan(L *LState) int {
	L.Push(LNumber(math.Atan(float64(L.CheckNumber(1)))))
	return 1
}

func mathAtan2(L *LState) int {
	L.Push(LNumber(math.Atan2(float64(L.CheckNumber(1)), float64(L.CheckNumber(2)))))
	return 1
}

func mathCeil(L *LState) int {
	L.Push(LNumber(math.Ceil(float64(L.CheckNumber(1)))))
	return 1
}

func mathCos(L *LState) int {
	L.Push(LNumber(math.Cos(float64(L.CheckNumber(1)))))
	return 1
}

func mathCosh(L *LState) int {
	L.Push(LNumber(math.Cosh(float64(L.CheckNumber(1)))))
	return 1
}

func mathDeg(L *LState) int {
	L.Push(LNumber(float64(L.CheckNumber(1)) * piInDegree / math.Pi))
	return 1
}

func mathExp(L *LState) int {
	L.Push(LNumber(math.Exp(float64(L.CheckNumber(1)))))
	return 1
}

func mathFloor(L *LState) int {
	L.Push(LNumber(math.Floor(float64(L.CheckNumber(1)))))
	return 1
}

func mathFmod(L *LState) int {
	L.Push(LNumber(math.Mod(float64(L.CheckNumber(1)), float64(L.CheckNumber(2)))))
	return 1
}

func mathFrexp(L *LState) int {
	v1, v2 := math.Frexp(float64(L.CheckNumber(1)))
	L.Push(LNumber(v1))
	L.Push(LNumber(v2))
	return 2
}

func mathLdexp(L *LState) int {
	L.Push(LNumber(math.Ldexp(float64(L.CheckNumber(1)), L.CheckInt(2))))
	return 1
}

func mathLn(L *LState) int {
	L.Push(LNumber(math.Log(float64(L.CheckNumber(1)))))
	return 1
}

func mathLog(L *LState) int {
	v1 := L.CheckNumber(1)
	base := LNumber(10)
	if L.GetTop() == 2 {
		base = L.CheckNumber(2)
	}
	L.Push(LNumber(math.Log(float64(v1)) / math.Log(float64(base))))
	return 1
}

func mathMax(L *LState) int {
	if L.GetTop() == 0 {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("wrong number of arguments")))
		return 2
	}
	max := L.CheckNumber(1)
	top := L.GetTop()
	for i := 2; i <= top; i++ {
		v := L.CheckNumber(i)
		if v > max {
			max = v
		}
	}
	L.Push(max)
	return 1
}

func mathMin(L *LState) int {
	if L.GetTop() == 0 {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("wrong number of arguments")))
		return 2
	}
	min := L.CheckNumber(1)
	top := L.GetTop()
	for i := 2; i <= top; i++ {
		v := L.CheckNumber(i)
		if v < min {
			min = v
		}
	}
	L.Push(min)
	return 1
}

func mathMod(L *LState) int {
	lhs := L.CheckNumber(1)
	rhs := L.CheckNumber(2)
	L.Push(luaModulo(lhs, rhs))
	return 1
}

func mathModf(L *LState) int {
	v1, v2 := math.Modf(float64(L.CheckNumber(1)))
	L.Push(LNumber(v1))
	L.Push(LNumber(v2))
	return 2
}

func mathPow(L *LState) int {
	L.Push(LNumber(math.Pow(float64(L.CheckNumber(1)), float64(L.CheckNumber(2)))))
	return 1
}

func mathRad(L *LState) int {
	L.Push(LNumber(float64(L.CheckNumber(1)) * math.Pi / piInDegree))
	return 1
}

func mathSin(L *LState) int {
	L.Push(LNumber(math.Sin(float64(L.CheckNumber(1)))))
	return 1
}

func mathSinh(L *LState) int {
	L.Push(LNumber(math.Sinh(float64(L.CheckNumber(1)))))
	return 1
}

func mathSqrt(L *LState) int {
	L.Push(LNumber(math.Sqrt(float64(L.CheckNumber(1)))))
	return 1
}

func mathTan(L *LState) int {
	L.Push(LNumber(math.Tan(float64(L.CheckNumber(1)))))
	return 1
}

func mathTanh(L *LState) int {
	L.Push(LNumber(math.Tanh(float64(L.CheckNumber(1)))))
	return 1
}
