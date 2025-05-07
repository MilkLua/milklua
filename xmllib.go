package lua

import (
	"fmt"

	"github.com/clbanning/mxj"
)

func OpenXml(L *LState) int {
	xmlmod := L.RegisterModule(XmlLibName, xmlFuncs)
	L.Push(xmlmod)
	return 1
}

var XmlLibFuncDoc = map[string]libFuncDoc{
	XmlLibName: {
		libName: XmlLibName,
		libFuncName: []string{
			"Encode",
			"Decode",
		},
	},
}

var xmlFuncs = map[string]LGFunction{
	"Encode": xmlEncode,
	"Decode": xmlDecode,
}

// xmlEncode 模块函数，用于将 table 转换为 XML 格式字符串
// 参数：
//  1. tbl (table)：表示要转换的 table
//
// 返回值：
//  1. string（转换后的 XML 字符串）
//
// 调用方式：local str, err = xmllib.xmlEncode(tbl)
// 备注：
//  1. 此函数要求传入的 Lua 表必须是一个根层级 map/dict 结构，否则会返回错误信息
//  2. 如果转换过程中出现错误，会返回 nil 和错误信息
//  3. 转换成功后，返回转换得到的 XML 字符串
func xmlEncode(L *LState) int {
	tbl := L.CheckTable(1)
	goValue := tableToGo(L, tbl)

	m, ok := goValue.(map[string]any)
	if !ok {
		L.Push(LNil)
		L.Push(LString("XML encode error: root table must be a map/dict"))
		return 2
	}

	data, err := mxj.Map(m).Xml()
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("XML encode error: %v", err)))
		return 2
	}

	L.Push(LString(data))
	return 1
}

// xmlDecode 模块函数，用于解析 XML 格式字符串
// 参数：
//  1. data (string)：表示要解析的 XML 字符串
//
// 返回值：
//  1. Lua 表（根据 XML 解析结果转换为对应的 Lua 表）
//
// 调用方式：
//  1. local tbl, err = xmllib.xmlDecode(data)
//
// 备注：
//  1. 如果解析过程中出现错误，会返回 nil 和错误信息
//  2. 返回的 table 可以是 map/dict 结构，也可以是数组结构，具体取决于 XML 内容
func xmlDecode(L *LState) int {
	data := L.CheckString(1)

	m, err := mxj.NewMapXml([]byte(data))
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("XML decode error in parsing XML: %v", err)))
		return 2
	}

	val, err := goToLValue(L, m)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("XML decode error in converting to MilkValue: %v", err)))
		return 2
	}
	L.Push(val)
	return 1
}
