package lua

import (
	"fmt"

	base62 "github.com/jxskiss/base62"
)

func OpenBase62X(L *LState) int {
	mod := L.RegisterModule(Base62XLibName, base62xFuncs)
	L.Push(mod)
	return 1
}

var Base62XLibFuncDoc = map[string]libFuncDoc{
	Base62XLibName: {
		libName: Base62XLibName,
		libFuncName: []string{
			"Encode",
			"Decode",
		},
	},
}

var base62xFuncs = map[string]LGFunction{
	"Encode": base62xEncode,
	"Decode": base62xDecode,
}

// base62xEncode 模块函数，用于将 Lua 字符串编码为 Base62x 格式的字符串
// 参数：
//  1. str (string) - 需要编码的 Lua 字符串
//
// 返回值：
//  1. string（编码后的 Base62x 字符串）
//
// 调用方式：
//  1. local encoded = b62xlib.Encode(str)
//
// 备注：
//  1. 返回的字符串即为编码后的 Base62x 格式内容
func base62xEncode(L *LState) int {
	data := []byte(L.CheckString(1))
	encoded := base62.EncodeToString(data)
	L.Push(LString(encoded))
	return 1
}

// base62xDecode 模块函数，用于解析 Base62x 格式的字符串
// 参数：
//  1. str (string) - 需要解析的 Base62x 字符串
//
// 返回值：
//  1. string（解码后的字符串）
//  2. string（解码过程中出现的错误信息）
//
// 调用方式：
//  1. local decoded, err = b62xlib.Decode(str)
//
// 备注：
//  1. 返回的字符串即为解码后的内容
func base62xDecode(L *LState) int {
	str := L.CheckString(1)
	decoded, err := base62.DecodeString(str)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("Base62x decode error: %v", err)))
		return 2
	}
	L.Push(LString(string(decoded)))
	return 1
}
