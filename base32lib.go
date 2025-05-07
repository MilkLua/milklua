package lua

import (
	"encoding/base32"
	"fmt"
)

func OpenBase32(L *LState) int {
	mod := L.RegisterModule(Base32LibName, base32Funcs)
	L.Push(mod)
	return 1
}

var Base32LibFuncDoc = map[string]libFuncDoc{
	Base32LibName: {
		libName: Base32LibName,
		libFuncName: []string{
			"Encode",
			"Decode",
		},
	},
}

var base32Funcs = map[string]LGFunction{
	"Encode": base32Encode,
	"Decode": base32Decode,
}

// base32Encode 模块函数，用于将 Lua 字符串编码为 Base32 格式的字符串
// 参数：
//  1. str (string) - 需要编码的 Lua 字符串
//
// 返回值：
//  1. string（编码后的 Base32 字符串）
//
// 调用方式：
//  1. local encoded = base32lib.Encode(str)
//
// 备注：
//  1. 返回的字符串即为编码后的 Base32 格式内容
func base32Encode(L *LState) int {
	str := L.CheckString(1)
	encoded := base32.StdEncoding.EncodeToString([]byte(str))
	L.Push(LString(encoded))
	return 1
}

// base32Decode 模块函数，用于解析 Base32 格式的字符串
// 参数：
//  1. str (string) - 需要解析的 Base32 字符串
//
// 返回值：
//  1. string（解码后的字符串）
//  2. string（解码过程中出现的错误信息）
//
// 调用方式：
//  1. local decoded, err = base32lib.Decode(str)
//
// 备注：
//  1. 返回的字符串即为解码后的内容
func base32Decode(L *LState) int {
	str := L.CheckString(1)
	decoded, err := base32.StdEncoding.DecodeString(str)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("Base32 decode error: %v", err)))
		return 2
	}
	L.Push(LString(decoded))
	return 1
}
