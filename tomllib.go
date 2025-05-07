package lua

import (
	"fmt"

	"github.com/pelletier/go-toml"
)

func OpenToml(L *LState) int {
	tomlmod := L.RegisterModule(TomlLibName, tomlFuncs)
	L.Push(tomlmod)
	return 1
}

var TomlLibFuncDoc = map[string]libFuncDoc{
	TomlLibName: {
		libName: TomlLibName,
		libFuncName: []string{
			"Encode",
			"Decode",
		},
	},
}

var tomlFuncs = map[string]LGFunction{
	"Encode": tomlEncode,
	"Decode": tomlDecode,
}

// tomlEncode 模块函数，用于将 MilkValue 编码为 TOML 格式的字符串
// 参数：
//  1. tbl (table) - 需要编码的 table
//
// 返回值：
//  1. string（编码后的 TOML 字符串）
//  2. string（编码过程中出现的错误信息）
//
// 调用方式：local encoded, err = tomlib.Encode(tbl)
// 备注：
//  1. 如果编码过程中出现错误，则会返回 nil 和错误信息
//  2. 返回的字符串即为编码后的 TOML 格式内容
func tomlEncode(L *LState) int {
	tbl := L.CheckTable(1)
	goValue := tableToGo(L, tbl)

	data, err := toml.Marshal(goValue)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("TOML encode error: %v", err)))
		return 2
	}

	L.Push(LString(data))
	return 1
}

// tomlDecode 模块函数，用于解析 TOML 格式的字符串
// 参数：
//  1. data (string) - 需要解析的 TOML 字符串
//
// 返回值：
//  1. any（根据 TOML 解析结果转换为对应的 milk 类型）
//  2. string（解析过程中出现的错误信息）
//
// 调用方式：local decoded, err = tomlib.Decode(data)
// 备注：
//  1. 如果解析过程中出现错误，则会返回 nil 和错误信息
//  2. 返回的 Lua 值可以是 table、字符串、数值或布尔值等，具体取决于 TOML 内容
func tomlDecode(L *LState) int {
	data := L.CheckString(1)

	var goValue interface{}
	err := toml.Unmarshal([]byte(data), &goValue)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("TOML decode error in parsing TOML: %v", err)))
		return 2
	}

	lv, err := goToLValue(L, goValue)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("TOML decode error in converting to MilkValue: %v", err)))
		return 2
	}
	L.Push(lv)
	return 1
}
