package lua

import (
	"encoding/hex"
	"fmt"
)

func OpenHex(L *LState) int {
	mod := L.RegisterModule(HexLibName, hexFuncs)
	L.Push(mod)
	return 1
}

var HexLibFuncDoc = map[string]libFuncDoc{
	HexLibName: {
		libName: HexLibName,
		libFuncName: []string{
			"Encode",
			"Decode",
		},
	},
}

var hexFuncs = map[string]LGFunction{
	"Encode": hexEncode,
	"Decode": hexDecode,
}

// hexEncode 模块函数，用于将 Lua 字符串编码为 Hex 格式的字符串
// 参数：
//  1. str (string) - 需要编码的 Lua 字符串
//
// 返回值：
//  1. string（编码后的 Hex 字符串）
//
// 调用方式：
//  1. local encoded = hexlib.Encode(str)
//
// 备注：
//  1. 返回的字符串即为编码后的 Hex 格式内容
func hexEncode(L *LState) int {
	str := L.CheckString(1)
	encoded := hex.EncodeToString([]byte(str))
	L.Push(LString(encoded))
	return 1
}

// hexDecode 模块函数，用于解析 Hex 格式的字符串
// 参数：
//  1. str (string) - 需要解析的 Hex 字符串
//
// 返回值：
//  1. string（解码后的字符串）
//  2. string（解码过程中出现的错误信息）
//
// 调用方式：
//  1. local decoded, err = hexlib.Decode(str)
//
// 备注：
//  1. 返回的字符串即为解码后的内容
//  2. 如果解码过程中出现错误，则会返回 nil 和错误信息
func hexDecode(L *LState) int {
	str := L.CheckString(1)
	decoded, err := hex.DecodeString(str)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("Hex decode error: %v", err)))
		return 2
	}
	L.Push(LString(decoded))
	return 1
}
