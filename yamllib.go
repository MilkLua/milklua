package lua

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func OpenYml(L *LState) int {
	ymlmod := L.RegisterModule(YamlLibName, yamlFuncs)
	L.Push(ymlmod)
	return 1
}

var YamlLibFuncDoc = map[string]libFuncDoc{
	YamlLibName: {
		libName: YamlLibName,
		libFuncName: []string{
			"Encode",
			"Decode",
		},
	},
}

var yamlFuncs = map[string]LGFunction{
	"Encode": yamlEncode,
	"Decode": yamlDecode,
}

// yamlEncode 模块函数，用于将 MilkValue 编码为 YAML 格式的字符串
// 参数：
//  1. tbl (table) - 需要编码的 table
//
// 返回值：
//  1. string（编码后的 YAML 字符串）
//
// 调用方式：local encoded = yamllib.yamlEncode(tbl)
// 备注：
//  1. 如果编码过程中出现错误，则会返回 nil 和错误信息
//  2. 返回的字符串即为编码后的 YAML 格式内容
//  3. 编码完成后可将结果写入文件、发送到网络等
func yamlEncode(L *LState) int {
	tbl := L.CheckTable(1)
	goValue := tableToGo(L, tbl)

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()

	encoder := yaml.NewEncoder(buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(goValue); err != nil {
		encoder.Close()
		bufferPool.Put(buf)
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("YAML encode error: %v", err)))
		return 2
	}
	encoder.Close()

	result := buf.String()
	bufferPool.Put(buf)
	L.Push(LString(result))
	return 1
}

// yamlDecode 模块函数，用于解析 YAML 格式的字符串
// 参数：
//  1. data (string) - 需要解析的 YAML 字符串
//
// 返回值：
//  1. any（根据 YAML 解析结果转换为对应的 milk 类型）
//
// 调用方式：local decoded, err = yamllib.yamlDecode(data)
// 备注：
//  1. 如果解析过程中出现错误，则会返回 nil 和错误信息
//  2. 返回的 milk 值类型可以是 table、字符串、数值或布尔值等，具体取决于 YAML 内容
//  3. 解析完成后可将结果用于后续的 Lua 逻辑处理
func yamlDecode(L *LState) int {
	data := L.CheckString(1)
	decoder := yaml.NewDecoder(strings.NewReader(data))

	var goValue interface{}
	if err := decoder.Decode(&goValue); err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("YAML decode error in parsing YAML: %v", err)))
		return 2
	}

	val, err := goToLValue(L, goValue)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("YAML decode error in converting to MilkValue: %v", err)))
		return 2
	}
	L.Push(val)
	return 1
}
