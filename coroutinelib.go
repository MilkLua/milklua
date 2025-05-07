package lua

func OpenCoroutine(L *LState) int {
	// TODO: Tie module name to contents of linit.go?
	mod := L.RegisterModule(CoroutineLibName, coFuncs)
	L.Push(mod)
	return 1
}

var CoroutLibFuncDoc = map[string]libFuncDoc{
	CoroutineLibName: {
		libName: CoroutineLibName,
		libFuncName: []string{
			"Create",
			"Yield",
			"Resume",
			"Running",
			"Status",
			"Wrap",
		},
	},
}

var coFuncs = map[string]LGFunction{
	"Create":  coCreate,
	"Yield":   coYield,
	"Resume":  coResume,
	"Running": coRunning,
	"Status":  coStatus,
	"Wrap":    coWrap,
}

// coCreate 模块函数，用于创建一个新的协程
// 参数：
//  1. fn (function) - 协程函数
//
// 返回值：
//
//  1. thread - 协程
//
// 调用方式：
//  1. coroutlib.Create(fn)
//
// 示例：
//  1. local co = coroutlib.Create(func(){ PrintLn("Hello, World!") })
//
// 注意：
//  1. 协程函数的参数和返回值与普通函数相同
//  2. 协程函数的返回值将作为协程的返回值
func coCreate(L *LState) int {
	fn := L.CheckFunction(1)
	newthread, _ := L.NewThread()
	base := 0
	newthread.stack.Push(callFrame{
		Fn:         fn,
		Pc:         0,
		Base:       base,
		LocalBase:  base + 1,
		ReturnBase: base,
		NArgs:      0,
		NRet:       MultRet,
		Parent:     nil,
		TailCall:   0,
	})
	L.Push(newthread)
	return 1
}

// coYield 模块函数，用于挂起当前协程
// 参数：
//
//	无
//
// 返回值：
//
//	无
//
// 调用方式：
//  1. coroutlib.Yield()
//
// 示例：
//  1. coroutlib.Yield()
//
// 注意：
//  1. 挂起当前协程，将控制权交给父协程
//  2. 挂起当前协程后，将无法再恢复执行
func coYield(L *LState) int {
	return -1
}

func coResume(L *LState) int {
	th := L.CheckThread(1)
	if L.G.CurrentThread == th {
		msg := "can not resume a running thread"
		if th.wrapped {
			L.RaiseError("%s", msg)
			return 0
		}
		L.Push(LFalse)
		L.Push(LString(msg))
		return 2
	}
	if th.Dead {
		msg := "can not resume a dead thread"
		if th.wrapped {
			L.RaiseError("%s", msg)
			return 0
		}
		L.Push(LFalse)
		L.Push(LString(msg))
		return 2
	}
	th.Parent = L
	L.G.CurrentThread = th
	if !th.isStarted() {
		cf := th.stack.Last()
		th.currentFrame = cf
		th.SetTop(0)
		nargs := L.GetTop() - 1
		L.XMoveTo(th, nargs)
		cf.NArgs = nargs
		th.initCallFrame(cf)
		th.Panic = panicWithoutTraceback
	} else {
		nargs := L.GetTop() - 1
		L.XMoveTo(th, nargs)
	}
	top := L.GetTop()
	threadRun(th)
	return L.GetTop() - top
}

// coRunning 模块函数，用于获取当前正在运行的协程
// 参数：
//
//	无
//
// 返回值：
//
//  1. thread - 当前正在运行的协程
//
// 调用方式：
//  1. coroutlib.Running()
//
// 示例：
//  1. local co = coroutlib.Running()
//
// 注意：
//  1. 获取当前正在运行的协程
func coRunning(L *LState) int {
	if L.G.MainThread == L {
		L.Push(LNil)
		return 1
	}
	L.Push(L.G.CurrentThread)
	return 1
}

// coStatus 模块函数，用于获取协程状态
// 参数：
//  1. co (thread) - 协程
//
// 返回值：
//
//  1. string - 协程状态
//
// 调用方式：
//  1. coroutlib.Status(co)
//
// 示例：
//  1. local status = coroutlib.Status(co)
//
// 注意：
//  1. 获取协程的状态
//  2. 协程状态包括：running, suspended, normal, dead
func coStatus(L *LState) int {
	L.Push(LString(L.Status(L.CheckThread(1))))
	return 1
}

func wrapaux(L *LState) int {
	L.Insert(L.ToThread(UpvalueIndex(1)), 1)
	return coResume(L)
}

// coWrap 模块函数，用于包装一个函数为协程
// 参数：
//  1. fn (function) - 函数
//
// 返回值：
//
//  1. function - 包装后的协程
//
// 调用方式：
//  1. coroutlib.Wrap(fn)
//
// 示例：
//  1. local co = coroutlib.Wrap(func(){ PrintLn("Hello, World!") })
//
// 注意：
//  1. 包装一个函数为协程
//  2. 包装后的协程可以通过coroutlib.Resume()来执行
func coWrap(L *LState) int {
	coCreate(L)
	L.CheckThread(L.GetTop()).wrapped = true
	v := L.Get(L.GetTop())
	L.Pop(1)
	L.Push(L.NewClosure(wrapaux, v))
	return 1
}

//
