package lua

import (
	"fmt"
	"strings"
)

func OpenDebug(L *LState) int {
	dbgmod := L.RegisterModule(DebugLibName, debugFuncs)
	L.Push(dbgmod)
	return 1
}

var DbgLibFuncDoc = map[string]libFuncDoc{
	DebugLibName: {
		libName: DebugLibName,
		libFuncName: []string{
			"GetFEnv",
			"GetInfo",
			"GetLocal",
			"GetMetatable",
			"GetUpvalue",
			"SetFEnv",
			"SetLocal",
			"SetMetatable",
			"SetUpvalue",
			"Traceback",
		},
	},
}

var debugFuncs = map[string]LGFunction{
	"GetFEnv":      debugGetFEnv,
	"GetInfo":      debugGetInfo,
	"GetLocal":     debugGetLocal,
	"GetMetatable": debugGetMetatable,
	"GetUpvalue":   debugGetUpvalue,
	"SetFEnv":      debugSetFEnv,
	"SetLocal":     debugSetLocal,
	"SetMetatable": debugSetMetatable,
	"SetUpvalue":   debugSetUpvalue,
	"Traceback":    debugTraceback,
}

// debugGetFEnv 模块函数，用于获取函数的环境
// 参数：
//  1. any - 任意类型
//
// 返回值：
//
//  1. table - 函数的环境
//
// 调用方式：
//  1. debuglib.GetFEnv(any)
//
// 示例：
//  1. local env = debuglib.GetFEnv(Print)
//  2. local env = debuglib.GetFEnv(1)
//
// 注意：
//  1. 如果参数是一个函数，则返回函数的环境
func debugGetFEnv(L *LState) int {
	L.Push(L.GetFEnv(L.CheckAny(1)))
	return 1
}

// debugGetInfo 模块函数，用于获取函数的信息
// 参数：
//  1. function|number - 函数或调用栈层级
//  2. string - 信息类型（可选）
//
// 返回值：
//
//  1. table - 函数信息
//
// 调用方式：
//  1. debuglib.GetInfo(fn)
//  2. debuglib.GetInfo(fn, what)
//  3. debuglib.GetInfo(1)
//  4. debuglib.GetInfo(1, what)
func debugGetInfo(L *LState) int {
	L.CheckTypes(1, LTFunction, LTNumber)
	arg1 := L.Get(1)
	what := L.OptString(2, "Slunf")
	var dbg *Debug
	var fn LValue
	var err error
	var ok bool
	switch lv := arg1.(type) {
	case *LFunction:
		dbg = &Debug{}
		fn, err = L.GetInfo(">"+what, dbg, lv)
	case LNumber:
		dbg, ok = L.GetStack(int(lv))
		if !ok {
			L.Push(LNil)
			return 1
		}
		fn, err = L.GetInfo(what, dbg, LNil)
	}

	if err != nil {
		L.Push(LNil)
		return 1
	}
	tbl := L.NewTable()
	if len(dbg.Name) > 0 {
		tbl.RawSetString("name", LString(dbg.Name))
	} else {
		tbl.RawSetString("name", LNil)
	}
	tbl.RawSetString("what", LString(dbg.What))
	tbl.RawSetString("source", LString(dbg.Source))
	tbl.RawSetString("currentline", LNumber(dbg.CurrentLine))
	tbl.RawSetString("nups", LNumber(dbg.NUpvalues))
	tbl.RawSetString("linedefined", LNumber(dbg.LineDefined))
	tbl.RawSetString("lastlinedefined", LNumber(dbg.LastLineDefined))
	tbl.RawSetString("func", fn)
	L.Push(tbl)
	return 1
}

// debugGetLocal 模块函数，用于获取函数的局部变量
// 参数：
//  1. number - 调用栈层级
//  2. number - 局部变量索引
//
// 返回值：
//
//  1. string - 变量名
//  2. any - 变量值
//
// 调用方式：
//  1. debuglib.GetLocal(level, idx)
//
// 示例：
//  1. local name, value = debuglib.GetLocal(1, 1)
//  2. local name, value = debuglib.GetLocal(1, 2)
//
// 注意：
//  1. 如果局部变量不存在，则返回nil
func debugGetLocal(L *LState) int {
	level := L.CheckInt(1)
	idx := L.CheckInt(2)
	dbg, ok := L.GetStack(level)
	if !ok {
		L.ArgError(1, "level out of range")
	}
	name, value := L.GetLocal(dbg, idx)
	if len(name) > 0 {
		L.Push(LString(name))
		L.Push(value)
		return 2
	}
	L.Push(LNil)
	return 1
}

// debugGetMetatable 模块函数，用于获取对象的元表
// 参数：
//  1. any - 任意类型
//
// 返回值：
//
//  1. table - 对象的元表
//
// 调用方式：
//  1. debuglib.GetMetatable(any)
//
// 示例：
//  1. local mt = debuglib.GetMetatable(tbl)
//  2. local mt = debuglib.GetMetatable(1)
//
// 注意：
//  1. 如果对象没有元表，则返回nil
func debugGetMetatable(L *LState) int {
	L.Push(L.GetMetatable(L.CheckAny(1)))
	return 1
}

// debugGetUpvalue 模块函数，用于获取函数的上值
// 参数：
//  1. function - 函数
//  2. number - 上值索引
//
// 返回值：
//
//  1. string - 上值名
//  2. any - 上值值
//
// 调用方式：
//  1. debuglib.GetUpvalue(fn, idx)
//
// 示例：
//  1. local name, value = debuglib.GetUpvalue(fn, 1)
//  2. local name, value = debuglib.GetUpvalue(fn, 2)
//
// 注意：
//  1. 如果上值不存在，则返回nil
func debugGetUpvalue(L *LState) int {
	fn := L.CheckFunction(1)
	idx := L.CheckInt(2)
	name, value := L.GetUpvalue(fn, idx)
	if len(name) > 0 {
		L.Push(LString(name))
		L.Push(value)
		return 2
	}
	L.Push(LNil)
	return 1
}

// debugSetFEnv 模块函数，用于设置函数的环境
// 参数：
//  1. function - 函数
//  2. table - 环境
//
// 返回值：
//
//	无
//
// 调用方式：
//  1. debuglib.SetFEnv(fn, env)
//
// 示例：
//  1. debuglib.SetFEnv(fn, env)
//
// 注意：
//  1. 设置函数的环境
func debugSetFEnv(L *LState) int {
	L.SetFEnv(L.CheckAny(1), L.CheckAny(2))
	return 0
}

// debugSetLocal 模块函数，用于设置函数的局部变量
// 参数：
//  1. number - 调用栈层级
//  2. number - 局部变量索引
//  3. any - 局部变量值
//
// 返回值：
//
//  1. string - 变量名
//  2. string - 错误信息
//
// 调用方式：
//  1. debuglib.SetLocal(level, idx, value)
//
// 示例：
//  1. local name, err = debuglib.SetLocal(1, 1, 10)
//  2. local name, err = debuglib.SetLocal(1, 2, "hello")
//
// 注意：
//  1. 如果局部变量不存在，则返回错误信息
func debugSetLocal(L *LState) int {
	level := L.CheckInt(1)
	idx := L.CheckInt(2)
	value := L.CheckAny(3)
	dbg, ok := L.GetStack(level)
	if !ok {
		L.Push(LNil)
		L.Push(LString("level out of range"))
		return 2
	}
	name := L.SetLocal(dbg, idx, value)
	if len(name) > 0 {
		L.Push(LString(name))
	} else {
		L.Push(LNil)
	}
	return 1
}

// debugSetMetatable 模块函数，用于设置对象的元表
// 参数：
//  1. any - 任意类型
//  2. table - 元表
//
// 返回值：
//
//	无
//
// 调用方式：
//  1. debuglib.SetMetatable(any, mt)
//
// 示例：
//  1. debuglib.SetMetatable(tbl, mt)
//  2. debuglib.SetMetatable(1, mt)
//
// 注意：
//  1. 设置对象的元表
func debugSetMetatable(L *LState) int {
	L.CheckTypes(2, LTNil, LTTable)
	obj := L.Get(1)
	mt := L.Get(2)
	if mtastbl, ok := mt.(*LTable); ok {
		if _, ok := mtastbl.RawGetString("__index").(*LTable); !ok {
			mt.(*LTable).RawSetString("__index", mt)
		}
	}
	L.SetMetatable(obj, mt)
	L.SetTop(1)
	return 1
}

// debugSetUpvalue 模块函数，用于设置函数的上值
// 参数：
//  1. function - 函数
//  2. number - 上值索引
//  3. any - 上值值
//
// 返回值：
//
//  1. string - 上值名
//
// 调用方式：
//  1. debuglib.SetUpvalue(fn, idx, value)
//
// 示例：
//  1. local name = debuglib.SetUpvalue(fn, 1, 10)
//  2. local name = debuglib.SetUpvalue(fn, 2, "hello")
//
// 注意：
//  1. 设置函数的上值
func debugSetUpvalue(L *LState) int {
	fn := L.CheckFunction(1)
	idx := L.CheckInt(2)
	value := L.CheckAny(3)
	name := L.SetUpvalue(fn, idx, value)
	if len(name) > 0 {
		L.Push(LString(name))
	} else {
		L.Push(LNil)
	}
	return 1
}

// debugTraceback 模块函数，用于获取调用栈信息
// 参数：
//  1. string - 错误信息（可选）
//  2. number - 调用栈层级（可选）
//
// 返回值：
//
//  1. string - 调用栈信息
//
// 调用方式：
//  1. debuglib.Traceback()
//  2. debuglib.Traceback(msg)
//  3. debuglib.Traceback(level)
//  4. debuglib.Traceback(msg, level)
//
// 示例：
//  1. local traceback = debuglib.Traceback()
//  2. local traceback = debuglib.Traceback("error")
func debugTraceback(L *LState) int {
	msg := ""
	level := L.OptInt(2, 1)
	ls := L
	if L.GetTop() > 0 {
		if s, ok := L.Get(1).(LString); ok {
			msg = string(s)
		}
		if l, ok := L.Get(1).(*LState); ok {
			ls = l
			msg = ""
		}
	}

	traceback := strings.TrimSpace(ls.stackTrace(level))
	if len(msg) > 0 {
		traceback = fmt.Sprintf("%s\n%s", msg, traceback)
	}
	L.Push(LString(traceback))
	return 1
}
