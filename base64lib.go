package lua

import (
	base64 "encoding/base64"
	"fmt"
)

func OpenBase64(L *LState) int {
	mod := L.RegisterModule(Base64LibName, base64Funcs)
	L.Push(mod)
	return 1
}

var Base64LibFuncDoc = map[string]libFuncDoc{
	Base64LibName: {
		libName: Base64LibName,
		libFuncName: []string{
			"Encode",
			"Decode",
		},
	},
}

var base64Funcs = map[string]LGFunction{
	"Encode": base64Encode,
	"Decode": base64Decode,
}

// base64Encode 模块函数，用于将 Lua 字符串编码为 Base64 格式的字符串
// 参数：
//  1. str (string) - 需要编码的 Lua 字符串
//
// 返回值：
//  1. string（编码后的 Base64 字符串）
//
// 调用方式：
//  1. local encoded = b64lib.Encode(str)
//
// 备注：
//  1. 返回的字符串即为编码后的 Base64 格式内容
func base64Encode(L *LState) int {
	str := L.CheckString(1)
	encoded := base64.StdEncoding.EncodeToString([]byte(str))
	L.Push(LString(encoded))
	return 1
}

// base64Decode 模块函数，用于解析 Base64 格式的字符串
// 参数：
//  1. str (string) - 需要解析的 Base64 字符串
//
// 返回值：
//  1. string（解码后的字符串）
//  2. string（解码过程中出现的错误信息）
//
// 调用方式：
//  1. local decoded, err = b64lib.Decode(str)
//
// 备注：
//  1. 返回的字符串即为解码后的内容
func base64Decode(L *LState) int {
	str := L.CheckString(1)
	decoded, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("Base64 decode error: %v", err)))
		return 2
	}
	L.Push(LString(decoded))
	return 1
}
