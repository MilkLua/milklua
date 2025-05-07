package lua

import (
	"strings"
	"time"

	mrpcg32 "github.com/MilkLua/milkrandom/pcg32"
	mrpcg64 "github.com/MilkLua/milkrandom/pcg64"
	mrsplitmix64 "github.com/MilkLua/milkrandom/splitmix64"
	mrxoshiro256starstar "github.com/MilkLua/milkrandom/xoshiro256starstar"
)

var pcg64rand *mrpcg64.PCG64
var pcg32rand *mrpcg32.PCG32
var splitmix64rand *mrsplitmix64.SplitMix64
var mrxoshiro256starstarrand *mrxoshiro256starstar.Xoshiro256StarStar

func init() {
	pcg64rand = mrpcg64.New()
	pcg32rand = mrpcg32.New()
	splitmix64rand = mrsplitmix64.New()
	mrxoshiro256starstarrand = mrxoshiro256starstar.New()
}

func OpenRandom(L *LState) int {
	mod := L.RegisterModule(RandomLibName, randomFuncs)
	L.Push(mod)
	return 1
}

var RandomLibFuncDoc = map[string]libFuncDoc{
	RandomLibName: {
		libName: RandomLibName,
		libFuncName: []string{
			"Seed",
			"Next",
		},
	},
}

var randomFuncs = map[string]LGFunction{
	"Seed": randomSeed,
	"Next": randomNext,
}

// randomSeed 模块函数，用于设置随机数种子
// 参数：
//  1. seed (number) - 随机数种子
//  2. prnggenname (string) - PRNG生成器名称（可选）
//
// 返回值：
//
//  1. string - 错误信息
//
// 调用方式：
//  1. randomlib.Seed(seed)
//  2. randomlib.Seed(seed, prnggenname)
//
// PRNG生成器名称：
//  1. pcg32
//  2. pcg64
//  3. splitmix64
//  4. xoshiro256starstar
//
// 示例：
//
//	randomlib.Seed(123456)
//	randomlib.Seed(123456, "pcg64")
//
//	randomlib.Seed(123456, "pcg32")
//	randomlib.Seed(123456, "splitmix64")
//	randomlib.Seed(123456, "xoshiro256starstar")
//
//	randomlib.Seed(123456, "unknown")
//
//	randomlib.Seed(timelib.Uinx(), "pcg64")
func randomSeed(L *LState) int {
	seed := L.OptNumber(1, LNumber(time.Now().UnixMicro()))
	prnggenname := L.OptString(2, "pcg64")
	switch strings.ToLower(prnggenname) {
	case "pcg32":
		pcg32rand.Seed(uint64(seed))
	case "pcg64":
		pcg64rand.Seed(uint64(seed))
	case "splitmix64":
		splitmix64rand.Seed(uint64(seed))
	case "xoshiro256starstar":
		mrxoshiro256starstarrand.Seed(uint64(seed))
	default:
		L.Push(LString("Unknown PRNG generator"))
		return 1
	}
	return 0
}

// randomNext 模块函数，用于获取随机数
// 参数：
//  1. min (number) - 最小值（可选）
//  2. max (number) - 最大值（可选）
//  3. prnggenname (string) - PRNG生成器名称（可选）
//
// 返回值：
//
//  1. number - 随机数
//  2. string - 错误信息
//
// 调用方式：
//  1. randomlib.Next()
//  2. randomlib.Next(min)
//  3. randomlib.Next(min, max)
//  4. randomlib.Next(min, max, prnggenname)
//
// PRNG生成器名称：
//  1. pcg32
//  2. pcg64
//  3. splitmix64
//  4. xoshiro256starstar
//
// 示例：
//
//	local num = randomlib.Next()
//	local num = randomlib.Next(10)
//	local num = randomlib.Next(10, 100)
//	local num = randomlib.Next(10, 100, "pcg64")
//
// 备注：
//  1. 获取[min, max]范围内的随机数
//  2. 如果未提供min和max，则默认为[0, 1]
//  3. 如果未提供PRNG生成器名称，则默认为pcg64
//  4. 如果提供的PRNG生成器名称不合法，则返回错误信息
func randomNext(L *LState) int {
	min := L.OptNumber(1, 0)
	max := L.OptNumber(2, 1)
	prnggenname := L.OptString(3, "pcg64")
	switch strings.ToLower(prnggenname) {
	case "pcg32":
		randnum := pcg32rand.Float64()*float64(max-min) + float64(min)
		L.Push(LNumber(randnum))
	case "pcg64":
		randnum := pcg64rand.Float64()*float64(max-min) + float64(min)
		L.Push(LNumber(randnum))
	case "splitmix64":
		randnum := splitmix64rand.Float64()*float64(max-min) + float64(min)
		L.Push(LNumber(randnum))
	case "xoshiro256starstar":
		randnum := mrxoshiro256starstarrand.Float64()*float64(max-min) + float64(min)
		L.Push(LNumber(randnum))
	default:
		L.Push(LNil)
		L.Push(LString("Unknown PRNG generator"))
		return 2
	}
	return 1
}
