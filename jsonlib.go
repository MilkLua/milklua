package lua

import (
	"encoding/json"
	"fmt"
)

func OpenJson(L *LState) int {
	jsonmod := L.RegisterModule(JsonLibName, jsonFuncs)
	L.Push(jsonmod)
	return 1
}

var JsonLibFuncDoc = map[string]libFuncDoc{
	JsonLibName: {
		libName: JsonLibName,
		libFuncName: []string{
			"Encode",
			"Decode",
		},
	},
}

var jsonFuncs = map[string]LGFunction{
	"Encode": jsonEncode,
	"Decode": jsonDecode,
}

// jsonEncode 模块函数，用于将 table 转换为 JSON 格式字符串
// 参数：
//  1. tbl (table)：表示要转换的 table
//
// 返回值：
//  1. string（转换后的 JSON 字符串）
//  2. string（转换过程中出现的错误信息）
//
// 调用方式：local str, err = jsonlib.jsonEncode(tbl)
// 备注：
//  1. 如果转换过程中出现错误，会返回 nil 和错误信息
//  2. 转换成功后，返回转换得到的 JSON 字符串
func jsonEncode(L *LState) int {
	tbl := L.CheckTable(1)
	goValue := tableToGo(L, tbl)

	data, err := json.Marshal(goValue)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("JSON encode error: %v", err)))
		return 2
	}

	L.Push(LString(data))
	return 1
}

// jsonDecode 模块函数，用于解析 JSON 格式字符串
// 参数：
//  1. data (string)：表示要解析的 JSON 字符串
//
// 返回值：
//  1. table（根据 JSON 解析结果转换为对应的 table）
//  2. string（解析过程中出现的错误信息）
//
// 调用方式：
//  1. local tbl, err = jsonlib.jsonDecode(data)
//
// 备注：
//  1. 如果解析过程中出现错误，则会返回 nil 和错误信息
//  2. 返回的 table可以是 table、字符串、数值或布尔值等，具体取决于 JSON 内容
func jsonDecode(L *LState) int {
	data := L.CheckString(1)
	var goValue interface{}
	if err := json.Unmarshal([]byte(data), &goValue); err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("JSON decode error in parsing JSON: %v", err)))
		return 2
	}
	lv, err := goToLValue(L, goValue)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("JSON decode error in converting to LValue: %v", err)))
		return 0
	}
	L.Push(lv)
	return 1
}
