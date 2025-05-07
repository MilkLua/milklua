package lua

import (
	"fmt"
	"net/url"
)

func OpenURLLib(L *LState) int {
	mod := L.RegisterModule(UrlLibName, urlFuncs)
	L.Push(mod)
	return 1
}

var URLLibFuncDoc = map[string]libFuncDoc{
	UrlLibName: {
		libName: UrlLibName,
		libFuncName: []string{
			"Encode",
			"Decode",
		},
	},
}

var urlFuncs = map[string]LGFunction{
	"Encode": urlEncode,
	"Decode": urlDecode,
}

// urlEncode 模块函数，用于将字符串进行 URL 编码
// 参数：
//  1. str (string) - 需要编码的字符串
//
// 返回值：
//  1. string（编码后的 URL 字符串）
//
// 调用方式：
//  1. local encoded = urllib.Encode(str)
//
// 备注：
//  1. 返回的字符串即为编码后的 URL 格式内容
func urlEncode(L *LState) int {
	str := L.CheckString(1)
	encoded := url.QueryEscape(str)
	L.Push(LString(encoded))
	return 1
}

// urlDecode 模块函数，用于解析 URL 编码的字符串
// 参数：
//  1. str (string) - 需要解析的 URL 编码字符串
//
// 返回值：
//  1. string（解码后的字符串）
//
// 调用方式：
//  1. local decoded, err = urllib.Decode(str)
//
// 备注：
//  1. 返回的字符串即为解码后的内容
//  2. 如果解码过程中出现错误，则会返回 nil 和错误信息
func urlDecode(L *LState) int {
	str := L.CheckString(1)
	decoded, err := url.QueryUnescape(str)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("URL decode error: %v", err)))
		return 2
	}
	L.Push(LString(decoded))
	return 1
}
