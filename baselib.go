package lua

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

/* basic functions {{{ */

func OpenBase(L *LState) int {
	global := L.Get(GlobalsIndex).(*LTable)
	L.SetGlobal("_G", global)
	L.SetGlobal("_VERSION", LString(LuaCompVersion))
	basemod := L.RegisterModule("_G", baseFuncs)
	global.RawSetString("IPairs", L.NewClosure(baseIpairs, L.NewFunction(ipairsaux)))
	global.RawSetString("Pairs", L.NewClosure(basePairs, L.NewFunction(pairsaux)))
	L.Push(basemod)
	return 1
}

var BaseLibFuncDoc = map[string]libFuncDoc{
	BaseLibName: {
		libName: BaseLibName,
		libFuncName: []string{
			"Assert",
			"DoFile",
			"Error",
			"GetFEnv",
			"GetMetatable",
			"Load",
			"LoadFile",
			"LoadString",
			"Next",
			"PCall",
			"PrintLn",
			"Print",
			"Input",
			"RawEqual",
			"RawGet",
			"RawSet",
			"Select",
			"SetFEnv",
			"SetMetatable",
			"ToNumber",
			"ToString",
			"Type",
			"XpCall",
			"Module",
			"Require",
			"NewProxy",
		},
	},
}

var baseFuncs = map[string]LGFunction{
	"Assert":       baseAssert,
	"DoFile":       baseDoFile,
	"Error":        baseError,
	"GetFEnv":      baseGetFEnv,
	"GetMetatable": baseGetMetatable,
	"Load":         baseLoad,
	"LoadFile":     baseLoadFile,
	"LoadString":   baseLoadString,
	"Next":         baseNext,
	"PCall":        basePCall,
	"PrintLn":      basePrintln,
	"Print":        basePrint,
	"Input":        baseInput,
	"RawEqual":     baseRawEqual,
	"RawGet":       baseRawGet,
	"RawSet":       baseRawSet,
	"Select":       baseSelect,
	"SetFEnv":      baseSetFEnv,
	"SetMetatable": baseSetMetatable,
	"ToNumber":     baseToNumber,
	"ToString":     baseToString,
	"Type":         baseType,
	"XpCall":       baseXPCall,
	// loadlib
	"Module":  loModule,
	"Require": loRequire,
	// hidden features
	"NewProxy": baseNewProxy,
}

// baseAssert 模块函数，用于断言
// 参数：
//  1. bool - 断言条件
//  2. string - 断言失败时的错误信息（可选）
//
// 返回值：
//
//  1. 无
//
// 调用方式：
//  1. Assert(true)
//  2. Assert(false, "assertion failed!")
//
// 示例：
//  1. Assert(true)
//  2. Assert(false, "assertion failed!")
//
// 注意：
//  1. 如果第一个参数为 false，则抛出错误
//  2. 如果第一个参数为 true，则不做任何操作
//  3. 如果第二个参数存在，则抛出第二个参数的错误信息
//  4. 如果第二个参数不存在，则抛出默认错误信息
func baseAssert(L *LState) int {
	if !L.ToBool(1) {
		L.RaiseError("%s", L.OptString(2, "assertion failed!"))
		return 0
	}
	return L.GetTop()
}

// baseDoFile 模块函数，用于执行文件
// 参数：
//  1. string - 文件路径
//
// 返回值：
//
//  1. function - 函数
//
// 调用方式：
//  1. DoFile("test.mlk")
//
// 示例：
//  1. DoFile("test.mlk")
//
// 注意：
//  1. 执行文件
func baseDoFile(L *LState) int {
	src := L.ToString(1)
	top := L.GetTop()
	fn, err := L.LoadFile(src)
	if err != nil {
		L.Push(LString(err.Error()))
		L.Panic(L)
	}
	L.Push(fn)
	L.Call(0, MultRet)
	return L.GetTop() - top
}

// baseError 模块函数，用于抛出错误
// 参数：
//  1. any - 错误信息
//  2. number - 错误等级（可选）
//
// 返回值：
//
//  1. 无
//
// 调用方式：
//  1. Error("error message")
//  2. Error("error message", 2)
//
// 示例：
//  1. Error("error message")
//  2. Error("error message", 2)
//
// 注意：
//  1. 抛出错误
//  2. 如果提供错误等级，则使用提供的错误等级
//  3. 如果未提供错误等级，则默认为 1
func baseError(L *LState) int {
	obj := L.CheckAny(1)
	level := L.OptInt(2, 1)
	L.Error(obj, level)
	return 0
}

// baseGetFEnv 模块函数，用于获取函数环境
// 参数：
//  1. function - 函数
//  2. number - 层级（可选）
//
// 返回值：
//
//  1. table - 函数环境
//
// 调用方式：
//  1. GetFEnv(fn)
//  2. GetFEnv(fn, 2)
//
// 示例：
//  1. local env = GetFEnv(fn)
//
// 注意：
//  1. 获取函数的环境
//  2. 如果未提供层级，则默认为 1
func baseGetFEnv(L *LState) int {
	var value LValue
	if L.GetTop() == 0 {
		value = LNumber(1)
	} else {
		value = L.Get(1)
	}

	if fn, ok := value.(*LFunction); ok {
		if !fn.IsG {
			L.Push(fn.Env)
		} else {
			L.Push(L.G.Global)
		}
		return 1
	}

	if number, ok := value.(LNumber); ok {
		level := int(float64(number))
		if level <= 0 {
			L.Push(L.Env)
		} else {
			cf := L.currentFrame
			for i := 0; i < level && cf != nil; i++ {
				cf = cf.Parent
			}
			if cf == nil || cf.Fn.IsG {
				L.Push(L.G.Global)
			} else {
				L.Push(cf.Fn.Env)
			}
		}
		return 1
	}

	L.Push(L.G.Global)
	return 1
}

// baseGetMetatable 模块函数，用于获取元表
// 参数：
//  1. any - 任意类型
//
// 返回值：
//
//  1. table - 元表
//
// 调用方式：
//  1. GetMetatable(any)
//
// 示例：
//  1. local mt = GetMetatable(tbl)
//  2. local mt = GetMetatable(1)
//
// 注意：
//  1. 获取对象的元表
//  2. 如果对象没有元表，则返回 nil
func baseGetMetatable(L *LState) int {
	L.Push(L.GetMetatable(L.CheckAny(1)))
	return 1
}

func ipairsaux(L *LState) int {
	tb := L.CheckTable(1)
	i := L.CheckInt(2)
	i++
	v := tb.RawGetInt(i)
	if v == LNil {
		return 0
	} else {
		L.Pop(1)
		L.Push(LNumber(i))
		L.Push(LNumber(i))
		L.Push(v)
		return 2
	}
}

// baseIpairs 模块函数，用于迭代表
// 参数：
//  1. table - 表
//
// 返回值：
//
//  1. function - 迭代器
//
// 调用方式：
//  1. IPairs(tbl)
//
// 示例：
//  1. IPairs(tbl)
//
// 注意：
//  1. 迭代表
func baseIpairs(L *LState) int {
	tb := L.CheckTable(1)
	L.Push(L.Get(UpvalueIndex(1)))
	L.Push(tb)
	L.Push(LNumber(0))
	return 3
}

func loadaux(L *LState, reader io.Reader, chunkname string) int {
	if fn, err := L.Load(reader, chunkname); err != nil {
		L.Push(LNil)
		L.Push(LString(err.Error()))
		return 2
	} else {
		L.Push(fn)
		return 1
	}
}

func baseLoad(L *LState) int {
	fn := L.CheckFunction(1)
	chunkname := L.OptString(2, "?")
	top := L.GetTop()
	buf := []string{}
	for {
		L.SetTop(top)
		L.Push(fn)
		L.Call(0, 1)
		ret := L.reg.Pop()
		if ret == LNil {
			break
		} else if LVCanConvToString(ret) {
			str := ret.String()
			if len(str) > 0 {
				buf = append(buf, string(str))
			} else {
				break
			}
		} else {
			L.Push(LNil)
			L.Push(LString("reader function must return a string"))
			return 2
		}
	}
	return loadaux(L, strings.NewReader(strings.Join(buf, "")), chunkname)
}

func baseLoadFile(L *LState) int {
	var reader io.Reader
	var chunkname string
	var err error
	if L.GetTop() < 1 {
		reader = os.Stdin
		chunkname = "<stdin>"
	} else {
		chunkname = L.CheckString(1)
		reader, err = os.Open(chunkname)
		if err != nil {
			L.Push(LNil)
			L.Push(LString(fmt.Sprintf("can not open file: %v", chunkname)))
			return 2
		}
		defer reader.(*os.File).Close()
	}
	return loadaux(L, reader, chunkname)
}

func baseLoadString(L *LState) int {
	return loadaux(L, strings.NewReader(L.CheckString(1)), L.OptString(2, "<string>"))
}

func baseNext(L *LState) int {
	tb := L.CheckTable(1)
	index := LNil
	if L.GetTop() >= 2 {
		index = L.Get(2)
	}
	key, value := tb.Next(index)
	if key == LNil {
		L.Push(LNil)
		return 1
	}
	L.Push(key)
	L.Push(value)
	return 2
}

func pairsaux(L *LState) int {
	tb := L.CheckTable(1)
	key, value := tb.Next(L.Get(2))
	if key == LNil {
		return 0
	} else {
		L.Pop(1)
		L.Push(key)
		L.Push(key)
		L.Push(value)
		return 2
	}
}

func basePairs(L *LState) int {
	tb := L.CheckTable(1)
	L.Push(L.Get(UpvalueIndex(1)))
	L.Push(tb)
	L.Push(LNil)
	return 3
}

func basePCall(L *LState) int {
	L.CheckAny(1)
	v := L.Get(1)
	if v.Type() != LTFunction && L.GetMetaField(v, "__call").Type() != LTFunction {
		L.Push(LFalse)
		L.Push(LString("attempt to call a " + v.Type().String() + " value"))
		return 2
	}
	nargs := L.GetTop() - 1
	if err := L.PCall(nargs, MultRet, nil); err != nil {
		L.Push(LFalse)
		if aerr, ok := err.(*ApiError); ok {
			L.Push(aerr.Object)
		} else {
			L.Push(LString(err.Error()))
		}
		return 2
	} else {
		L.Insert(LTrue, 1)
		return L.GetTop()
	}
}

func basePrintln(L *LState) int {
	top := L.GetTop()
	for i := 1; i <= top; i++ {
		fmt.Print(L.ToStringMeta(L.Get(i)).String())
		if i != top {
			fmt.Print(" ")
		}
	}
	fmt.Println("")
	return 0
}

func basePrint(L *LState) int {
	top := L.GetTop()
	for i := 1; i <= top; i++ {
		fmt.Print(L.ToStringMeta(L.Get(i)).String())
		if i != top {
			fmt.Print(" ")
		}
	}
	return 0
}

func baseInput(L *LState) int {
	prompt := L.OptString(1, "")
	if prompt != "" {
		fmt.Print(prompt)
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			L.RaiseError("EOF encountered: no input provided")
			return 0
		}
		L.RaiseError("input error: %s", err.Error())
		return 0
	}
	input = strings.TrimRight(input, "\r\n")
	L.Push(LString(input))
	return 1
}

func baseRawEqual(L *LState) int {
	if L.CheckAny(1) == L.CheckAny(2) {
		L.Push(LTrue)
	} else {
		L.Push(LFalse)
	}
	return 1
}

func baseRawGet(L *LState) int {
	L.Push(L.RawGet(L.CheckTable(1), L.CheckAny(2)))
	return 1
}

func baseRawSet(L *LState) int {
	L.RawSet(L.CheckTable(1), L.CheckAny(2), L.CheckAny(3))
	return 0
}

func baseSelect(L *LState) int {
	L.CheckTypes(1, LTNumber, LTString)
	switch lv := L.Get(1).(type) {
	case LNumber:
		idx := int(lv)
		num := L.GetTop()
		if idx < 0 {
			idx = num + idx
		} else if idx > num {
			idx = num
		}
		if 1 > idx {
			L.ArgError(1, "index out of range")
		}
		return num - idx
	case LString:
		if string(lv) != "#" {
			L.ArgError(1, "invalid string '"+string(lv)+"'")
		}
		L.Push(LNumber(L.GetTop() - 1))
		return 1
	}
	return 0
}

func baseSetFEnv(L *LState) int {
	var value LValue
	if L.GetTop() == 0 {
		value = LNumber(1)
	} else {
		value = L.Get(1)
	}
	env := L.CheckTable(2)

	if fn, ok := value.(*LFunction); ok {
		if fn.IsG {
			L.RaiseError("cannot change the environment of given object")
		} else {
			fn.Env = env
			L.Push(fn)
			return 1
		}
	}

	if number, ok := value.(LNumber); ok {
		level := int(float64(number))
		if level <= 0 {
			L.Env = env
			return 0
		}

		cf := L.currentFrame
		for i := 0; i < level && cf != nil; i++ {
			cf = cf.Parent
		}
		if cf == nil || cf.Fn.IsG {
			L.RaiseError("cannot change the environment of given object")
		} else {
			cf.Fn.Env = env
			L.Push(cf.Fn)
			return 1
		}
	}

	L.RaiseError("cannot change the environment of given object")
	return 0
}

func baseSetMetatable(L *LState) int {
	L.CheckTypes(2, LTNil, LTTable)
	obj := L.Get(1)
	if obj == LNil {
		L.RaiseError("cannot set metatable to a nil object.")
	}
	mt := L.Get(2)
	if m := L.metatable(obj, true); m != LNil {
		if tb, ok := m.(*LTable); ok && tb.RawGetString("__metatable") != LNil {
			L.RaiseError("cannot change a protected metatable")
		}
	}
	if mtastbl, ok := mt.(*LTable); ok {
		if _, ok := mtastbl.RawGetString("__index").(*LTable); !ok {
			mt.(*LTable).RawSetString("__index", mt)
		}
	}
	L.SetMetatable(obj, mt)
	L.SetTop(1)
	return 1
}

func baseToNumber(L *LState) int {
	base := L.OptInt(2, 10)
	noBase := L.Get(2) == LNil

	switch lv := L.CheckAny(1).(type) {
	case LNumber:
		L.Push(lv)
	case LString:
		str := strings.Trim(string(lv), " \n\t")
		if strings.Index(str, ".") > -1 {
			if v, err := strconv.ParseFloat(str, LNumberBit); err != nil {
				L.Push(LNil)
			} else {
				L.Push(LNumber(v))
			}
		} else {
			if noBase && strings.HasPrefix(strings.ToLower(str), "0x") {
				base, str = 16, str[2:] // Hex number
			} else if noBase && strings.HasPrefix(strings.ToLower(str), "0b") {
				base, str = 2, str[2:] // Binary number
			} else if noBase && strings.HasPrefix(strings.ToLower(str), "0o") {
				base, str = 8, str[2:] // Octal number
			}
			if v, err := strconv.ParseInt(str, base, LNumberBit); err != nil {
				L.Push(LNil)
			} else {
				L.Push(LNumber(v))
			}
		}
	default:
		L.Push(LNil)
	}
	return 1
}

func baseToString(L *LState) int {
	v1 := L.CheckAny(1)
	L.Push(L.ToStringMeta(v1))
	return 1
}

func baseType(L *LState) int {
	L.Push(LString(L.CheckAny(1).Type().String()))
	return 1
}

func baseXPCall(L *LState) int {
	fn := L.CheckFunction(1)
	errfunc := L.CheckFunction(2)

	top := L.GetTop()
	L.Push(fn)
	if err := L.PCall(0, MultRet, errfunc); err != nil {
		L.Push(LFalse)
		if aerr, ok := err.(*ApiError); ok {
			L.Push(aerr.Object)
		} else {
			L.Push(LString(err.Error()))
		}
		return 2
	} else {
		L.Insert(LTrue, top+1)
		return L.GetTop() - top
	}
}

/* }}} */

/* load lib {{{ */

func loModule(L *LState) int {
	name := L.CheckString(1)
	loaded := L.GetField(L.Get(RegistryIndex), "_LOADED")
	tb := L.GetField(loaded, name)
	if _, ok := tb.(*LTable); !ok {
		tb = L.FindTable(L.Get(GlobalsIndex).(*LTable), name, 1)
		if tb == LNil {
			L.RaiseError("name conflict for module: %v", name)
		}
		L.SetField(loaded, name, tb)
	}
	if L.GetField(tb, "_NAME") == LNil {
		L.SetField(tb, "_M", tb)
		L.SetField(tb, "_NAME", LString(name))
		names := strings.Split(name, ".")
		pname := ""
		if len(names) > 1 {
			pname = strings.Join(names[:len(names)-1], ".") + "."
		}
		L.SetField(tb, "_PACKAGE", LString(pname))
	}

	caller := L.currentFrame.Parent
	if caller == nil {
		L.RaiseError("no calling stack.")
	} else if caller.Fn.IsG {
		L.RaiseError("module() can not be called from GFunctions.")
	}
	L.SetFEnv(caller.Fn, tb)

	top := L.GetTop()
	for i := 2; i <= top; i++ {
		L.Push(L.Get(i))
		L.Push(tb)
		L.Call(1, 0)
	}
	L.Push(tb)
	return 1
}

var loopdetection = &LUserData{}

func loRequire(L *LState) int {
	name := L.CheckString(1)
	loaded := L.GetField(L.Get(RegistryIndex), "_LOADED")
	lv := L.GetField(loaded, name)
	if LVAsBool(lv) {
		if lv == loopdetection {
			L.RaiseError("loop or previous error loading module: %s", name)
		}
		L.Push(lv)
		return 1
	}
	loaders, ok := L.GetField(L.Get(RegistryIndex), "_LOADERS").(*LTable)
	if !ok {
		L.RaiseError("pkglib.loaders must be a table")
	}
	messages := []string{}
	var modasfunc LValue
	for i := 1; ; i++ {
		loader := L.RawGetInt(loaders, i)
		if loader == LNil {
			L.RaiseError("module %s not found:\n\t%s, ", name, strings.Join(messages, "\n\t"))
		}
		L.Push(loader)
		L.Push(LString(name))
		L.Call(1, 1)
		ret := L.reg.Pop()
		switch retv := ret.(type) {
		case *LFunction:
			modasfunc = retv
			goto loopbreak
		case LString:
			messages = append(messages, string(retv))
		}
	}
loopbreak:
	L.SetField(loaded, name, loopdetection)
	L.Push(modasfunc)
	L.Push(LString(name))
	L.Call(1, 1)
	ret := L.reg.Pop()
	modv := L.GetField(loaded, name)
	if ret != LNil && modv == loopdetection {
		L.SetField(loaded, name, ret)
		L.Push(ret)
	} else if modv == loopdetection {
		L.SetField(loaded, name, LTrue)
		L.Push(LTrue)
	} else {
		L.Push(modv)
	}
	return 1
}

/* }}} */

/* hidden features {{{ */

func baseNewProxy(L *LState) int {
	ud := L.NewUserData()
	L.SetTop(1)
	if L.Get(1) == LTrue {
		L.SetMetatable(ud, L.NewTable())
	} else if d, ok := L.Get(1).(*LUserData); ok {
		L.SetMetatable(ud, L.GetMetatable(d))
	}
	L.Push(ud)
	return 1
}

/* }}} */

//
