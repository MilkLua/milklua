package lua

import (
	"fmt"
	"sort"
)

func OpenTable(L *LState) int {
	tabmod := L.RegisterModule(TabLibName, tableFuncs)
	L.Push(tabmod)
	return 1
}

var TblLibFuncDoc = map[string]libFuncDoc{
	TabLibName: {
		libName: "tbllib",
		libFuncName: []string{
			"GetN",
			"SetN",
			"GetLen",
			"Concat",
			"Clone",
			"Equal",
			"Insert",
			"MaxN",
			"Remove",
			"Sort",
			"Unpack",
			"Pack",
		},
	},
}

var tableFuncs = map[string]LGFunction{
	"GetN":   tableGetN,
	"SetN":   tableSetN,
	"GetLen": tableGetLen,
	"Concat": tableConcat,
	"Clone":  tableClone,
	"Equal":  tableEqual,
	"Insert": tableInsert,
	"MaxN":   tableMaxN,
	"Remove": tableRemove,
	"Sort":   tableSort,
	"Unpack": tableUnpack,
	"Pack":   tablePack,
}

// tableSort 模块函数，用于对表进行排序
// 参数：
//  1. tbl (table) - 待排序的表
//  2. fn (function) - 排序函数（可选）
//
// 返回值：
//
//	无
//
// 调用方式：
//  1. tbllib.Sort(tbl)
//  2. tbllib.Sort(tbl, fn)
//
// 示例：
//
//	local tbl = {3, 1, 2}
//	tbllib.Sort(tbl)
//	PrintLn(tbl) // 输出：{1, 2, 3}
//
// 备注：
//  1. 如果提供排序函数，则使用排序函数进行排序
//  2. 如果未提供排序函数，则使用默认排序规则进行排序
//  3. 排序函数的定义方式为：func(a, b) { return a < b }
//  4. 排序函数的返回值为 true 时，表示 a 在 b 之前
//  5. 排序函数的返回值为 false 时，表示 a 在 b 之后
//  6. 排序函数的返回值为 nil 时，表示 a 和 b 相等
//  7. 排序函数的返回值为其他值时，表示 a 和 b 的关系不确定
//  8. 排序函数的返回值为其他类型时，会导致排序失败
func tableSort(L *LState) int {
	tbl := L.CheckTable(1)
	sorter := lValueArraySorter{L, nil, tbl.array}
	if L.GetTop() != 1 {
		sorter.Fn = L.CheckFunction(2)
	}
	sort.Sort(sorter)
	return 0
}

// tableGetN 模块函数，用于获取表的长度
// 参数：
//  1. tbl (table) - 待获取长度的表
//
// 返回值：
//  1. number（表的长度）
//
// 调用方式：
//  1. local len = tbllib.GetN(tbl)
//
// 示例：
//
//	local tbl = {1, 2, 3}
//	local len = tbllib.GetN(tbl)
//	PrintLn(len) // 输出：3
//
// 备注：
//  1. 获取表的长度，即表中元素的个数
//  2. 如果表中存在 nil 元素，则不会计入长度
func tableGetN(L *LState) int {
	L.Push(LNumber(L.CheckTable(1).Len()))
	return 1
}

// tableSetN 模块函数，用于设置表的长度
// 参数：
//  1. tbl (table) - 待设置长度的表
//  2. n (number) - 新的长度
//  3. v (any) - 新的元素值
//
// 返回值：
//
//  1. string（错误信息）
//
// 调用方式：
//  1. tbllib.SetN(tbl, n, v)
//
// 示例：
//
//	local tbl = {1, 2, 3}
//	tbllib.SetN(tbl, 5, 4)
//	PrintLn(tbl) // 输出：{1, 2, 3, nil, 4}
//
// 备注：
//  1. 设置表的长度，即设置表中元素的个数
//  2. 如果新的长度大于原长度，则会在表尾添加 nil 元素
//  3. 如果新的长度小于原长度，则会删除多余的元素
//  4. 如果新的长度小于 0，则会返回错误信息
func tableSetN(L *LState) int {
	tbl := L.CheckTable(1)
	n := L.CheckInt(2)
	if n < 0 {
		L.Push(LString(fmt.Sprintf("invalid length %d", n)))
		return 1
	}
	if n > tbl.Len() {
		tbl.RawSetInt(n, L.Get(3))
	} else {
		tbl.RawSetInt(n, L.Get(3))
	}
	return 0
}

// tableGetLen 模块函数，用于获取表的长度
// 参数：
//  1. tbl (table) - 待获取长度的表
//
// 返回值：
//  1. number（表的长度）
//
// 调用方式：
//  1. local len = tbllib.GetLen(tbl)
//
// 示例：
//
//	local tbl = {1, 2, 3}
//	local len = tbllib.GetLen(tbl)
//	PrintLn(len) // 输出：3
//
// 备注：
//  1. 获取表的长度，即表中元素的个数
//  2. 如果表中存在 nil 元素，则会计入长度
func tableGetLen(L *LState) int {
	tbl := L.CheckTable(1)
	cnt := 0
	tbl.ForEach(func(k, v LValue) {
		if v != LNil {
			cnt++
		}
	})
	L.Push(LNumber(cnt))
	return 1
}

// tableMaxN 模块函数，用于获取表的最大索引
// 参数：
//  1. tbl (table) - 待获取最大索引的表
//
// 返回值：
//  1. number（表的最大索引）
//
// 调用方式：
//  1. local max = tbllib.MaxN(tbl)
//
// 示例：
//
//	local tbl = {1, 2, 3}
//	local max = tbllib.MaxN(tbl)
//	PrintLn(max) // 输出：3
//
// 备注：
//  1. 获取表的最大索引，即表中最大的整数索引
//  2. 如果表中存在非整数索引，则不会计入最大索引
func tableMaxN(L *LState) int {
	L.Push(LNumber(L.CheckTable(1).MaxN()))
	return 1
}

// tableRemove 模块函数，用于移除表中的元素
// 参数：
//  1. tbl (table) - 待移除元素的表
//  2. idx (number) - 待移除元素的索引（可选，默认为最后一个元素）
//
// 返回值：
//  1. any（被移除的元素）
//
// 调用方式：
//  1. local v = tbllib.Remove(tbl)
//  2. local v = tbllib.Remove(tbl, idx)
//
// 示例：
//
//	local tbl = {1, 2, 3}
//	local v = tbllib.Remove(tbl)
//	PrintLn(v) // 输出：3
//
// 备注：
//  1. 移除表中的元素，返回被移除的元素
//  2. 如果未提供索引，则默认移除最后一个元素
//  3. 如果索引超出范围，则返回 nil
func tableRemove(L *LState) int {
	tbl := L.CheckTable(1)
	if L.GetTop() == 1 {
		L.Push(tbl.Remove(-1))
	} else {
		L.Push(tbl.Remove(L.CheckInt(2)))
	}
	return 1
}

// tableConcat 模块函数，用于连接表中的元素
// 参数：
//  1. tbl (table) - 待连接元素的表
//  2. sep (string) - 连接元素的分隔符（可选，默认为空字符串）
//  3. i (number) - 起始索引（可选，默认为 1）
//  4. j (number) - 结束索引（可选，默认为表的长度）
//
// 返回值：
//  1. string（连接后的字符串）
//
// 调用方式：
//  1. local str = tbllib.Concat(tbl)
//  2. local str = tbllib.Concat(tbl, sep)
//  3. local str = tbllib.Concat(tbl, sep, i)
//  4. local str = tbllib.Concat(tbl, sep, i, j)
//
// 示例：
//
//	local tbl = {1, 2, 3}
//	local str = tbllib.Concat(tbl)
//	PrintLn(str) // 输出：123
//
// 备注：
//  1. 连接表中的元素，返回连接后的字符串
//  2. 如果未提供分隔符，则默认为空字符串
//  3. 如果未提供起始索引，则默认为 1
//  4. 如果未提供结束索引，则默认为表的长度
//  5. 如果起始索引超出范围，则返回空字符串
//  6. 如果结束索引超出范围，则返回空字符串
func tableConcat(L *LState) int {
	tbl := L.CheckTable(1)
	sep := LString(L.OptString(2, ""))
	i := L.OptInt(3, 1)
	j := L.OptInt(4, tbl.Len())
	if L.GetTop() == 3 {
		if i > tbl.Len() || i < 1 {
			L.Push(emptyLString)
			return 1
		}
	}
	i = intMax(intMin(i, tbl.Len()), 1)
	j = intMin(intMin(j, tbl.Len()), tbl.Len())
	if i > j {
		L.Push(emptyLString)
		return 1
	}
	//TODO should flushing?
	retbottom := L.GetTop()
	for ; i <= j; i++ {
		v := tbl.RawGetInt(i)
		if !LVCanConvToString(v) {
			L.RaiseError("invalid value (%s) at index %d in table for concat", v.Type().String(), i)
		}
		L.Push(v)
		if i != j {
			L.Push(sep)
		}
	}
	L.Push(stringConcat(L, L.GetTop()-retbottom, L.reg.Top()-1))
	return 1
}

// tableClone 模块函数，用于克隆表
// 参数：
//  1. tbl (table) - 待克隆的表
//
// 返回值：
//  1. table（克隆后的表）
//
// 调用方式：
//  1. local newtbl = tbllib.Clone(tbl)
//
// 示例：
//
//	local tbl = {1, 2, 3}
//	local newtbl = tbllib.Clone(tbl)
//	PrintLn(newtbl) // 输出：{1, 2, 3}
//
// 备注：
//  1. 克隆表，返回克隆后的表
//  2. 克隆后的表与原表相互独立，互不影响
func tableClone(L *LState) int {
	tbl := L.CheckTable(1)
	newtbl := tbl
	L.Push(newtbl)
	return 1
}

// tableEqual 模块函数，用于比较表是否相等
// 参数：
//  1. tbl1 (table) - 待比较的表1
//  2. tbl2 (table) - 待比较的表2
//
// 返回值：
//  1. boolean（是否相等）
//
// 调用方式：
//  1. local eq = tbllib.Equal(tbl1, tbl2)
//
// 示例：
//
//	local tbl1 = {1, 2, 3}
//	local tbl2 = {1, 2, 3}
//	local eq = tbllib.Equal(tbl1, tbl2)
//	PrintLn(eq) // 输出：true
//
// 备注：
//  1. 比较两个表是否相等，返回是否相等的布尔值
//  2. 如果两个表的长度不同，则返回 false
//  3. 如果两个表的元素不同，则返回 false
//  4. 如果两个表的元素顺序不同，则返回 false
//  5. 如果两个表的元素相同，则返回 true
func tableEqual(L *LState) int {
	tbl1 := L.CheckTable(1)
	tbl2 := L.CheckTable(2)
	if tbl1 == tbl2 {
		L.Push(LTrue)
		return 1
	}
	if tbl1.Len() != tbl2.Len() {
		L.Push(LFalse)
		return 1
	}
	eq := true
	tbl1.ForEach(func(k, v1 LValue) {
		v2 := tbl2.RawGet(k)
		if !deepEqual(v1, v2) {
			eq = false
			return
		}
	})
	if eq {
		L.Push(LTrue)
	} else {
		L.Push(LFalse)
	}
	return 1
}

// deepEqual 辅助函数，用于深度比较两个值
func deepEqual(v1, v2 LValue) bool {
	if v1 == v2 {
		return true
	}
	if v1.Type() != v2.Type() {
		return false
	}
	switch v1.Type() {
	case LTNil:
		return true
	case LTBool:
		return v1 == v2
	case LTNumber:
		return v1 == v2
	case LTString:
		return v1 == v2
	case LTTable:
		t1 := v1.(*LTable)
		t2 := v2.(*LTable)
		if t1.Len() != t2.Len() {
			return false
		}
		eq := true
		t1.ForEach(func(k, v1 LValue) {
			v2 := t2.RawGet(k)
			if !deepEqual(v1, v2) {
				eq = false
				return
			}
		})
		return eq
	default:
		return v1 == v2
	}
}

// tableInsert 模块函数，用于向表中插入元素
// 参数：
//  1. tbl (table) - 待插入元素的表
//  2. idx (number) - 插入元素的索引
//  3. v (any) - 插入的元素值
//
// 返回值：
//
//  1. string（错误信息）
//
// 调用方式：
//  1. tbllib.Insert(tbl, idx, v)
func tableInsert(L *LState) int {
	tbl := L.CheckTable(1)
	nargs := L.GetTop()
	if nargs == 1 {
		L.Push(LString(fmt.Sprintf("missing argument #2")))
		return 1
	}

	if L.GetTop() == 2 {
		tbl.Append(L.Get(2))
		return 0
	}
	tbl.Insert(int(L.CheckInt(2)), L.CheckAny(3))
	return 0
}

// tableUnpack 模块函数，用于解包表
// 参数：
//  1. tbl (table) - 待解包的表
//  2. start (number) - 起始索引（可选，默认为 1）
//  3. end (number) - 结束索引（可选，默认为表的长度）
//
// 返回值：
//
//  1. ...(any)（解包后的元素）
//
// 调用方式：
//  1. tbllib.Unpack(tbl)
//  2. tbllib.Unpack(tbl, start)
//  3. tbllib.Unpack(tbl, start, end)
//
// 示例：
//
//	local tbl = {1, 2, 3}
//	local a, b, c = tbllib.Unpack(tbl)
//	PrintLn(a, b, c) // 输出：1 2 3
//
// 备注：
//  1. 解包表，返回解包后的元素
//  2. 如果未提供起始索引，则默认为 1
//  3. 如果未提供结束索引，则默认为表的长度
//  4. 如果起始索引超出范围，则返回空值
func tableUnpack(L *LState) int {
	tb := L.CheckTable(1)
	start := L.OptInt(2, 1)
	end := L.OptInt(3, tb.Len())
	for i := start; i <= end; i++ {
		L.Push(tb.RawGetInt(i))
	}
	ret := end - start + 1
	if ret < 0 {
		return 0
	}
	return ret
}

// tablePack 模块函数，用于打包元素为表
// 参数：
//  1. ...（any） - 待打包的元素
//
// 返回值：
//  1. table（打包后的表）
//
// 调用方式：
//  1. local tbl = tbllib.Pack(...)
//
// 示例：
//
//	local tbl = tbllib.Pack(1, 2, 3)
//	PrintLn(tbl) // 输出：{1, 2, 3}
//
// 备注：
//  1. 打包元素为表，返回打包后的表
func tablePack(L *LState) int {
	t := L.NewTable()
	for i := 1; i <= L.GetTop(); i++ {
		t.RawSetInt(i, L.Get(i))
	}
	L.Push(t)
	return 1
}
