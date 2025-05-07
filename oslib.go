package lua

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var startedAt time.Time

func init() {
	startedAt = time.Now()
}

func getIntField(_ *LState, tb *LTable, key string, v int) int {
	ret := tb.RawGetString(key)

	switch lv := ret.(type) {
	case LNumber:
		return int(lv)
	case LString:
		slv := string(lv)
		slv = strings.TrimLeft(slv, " ")
		if strings.HasPrefix(slv, "0") && !strings.HasPrefix(slv, "0x") && !strings.HasPrefix(slv, "0X") {
			// 标准 Lua 解释器仅支持十进制和十六进制
			slv = strings.TrimLeft(slv, "0")
			if slv == "" {
				return 0
			}
		}
		if num, err := parseNumber(slv); err == nil {
			return int(num)
		}
	default:
		return v
	}

	return v
}

func getBoolField(_ *LState, tb *LTable, key string, v bool) bool {
	ret := tb.RawGetString(key)
	if lb, ok := ret.(LBool); ok {
		return bool(lb)
	}
	return v
}

func OpenOs(L *LState) int {
	osmod := L.RegisterModule(OsLibName, osFuncs)
	L.Push(osmod)
	return 1
}

var OsLibFuncDoc = map[string]libFuncDoc{
	OsLibName: {
		libName: OsLibName,
		libFuncName: []string{
			"Execute",
			"Exit",
			"GetEnv",
			"Remove",
			"Rename",
			"SetEnv",
			"TmpName",
			"PathJoin",
			"AbsPath",
			"GetCurrWorkingDir",
			"DirName",
			"RelPath",
			"GetPID",
			"GetPPID",
			"MCpus",
			"MkdirAll",
			"Symlink",
			"Stat",
			"Exists",
			"GetOSName",
		},
	},
}

var osFuncs = map[string]LGFunction{
	"Execute":           osExecute,
	"Exit":              osExit,
	"GetEnv":            osGetEnv,
	"Remove":            osRemove,
	"Rename":            osRename,
	"SetEnv":            osSetEnv,
	"TmpName":           osTmpname,
	"PathJoin":          osPathJoin,
	"AbsPath":           osAbsPath,
	"GetCurrWorkingDir": osGetCurrWorkingDir,
	"DirName":           osDirName,
	"RelPath":           osRelPath,
	"GetPID":            osGetPID,
	"GetPPID":           osGetPPID,
	"MCpus":             osMCpus,
	"MkdirAll":          osMkdirAll,
	"Symlink":           osSymlink,
	"Stat":              osStat,
	"Exists":            osExists,
	"GetOSName":         osGetOSName,
}

// osExecute 模块函数，用于执行外部命令。
// 参数：
//  1. cmd (string) - 需要执行的命令（可包含完整路径）。
//  2. ... (string) - 命令行参数。
//
// 返回值：
//  1. int - 执行结果码，其中 0 表示成功，1 表示失败。
//
// 调用方式：
//  1. local result = oslib.Execute(cmd, ...)
func osExecute(L *LState) int {
	var procAttr os.ProcAttr
	procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	cmd, args := popenArgs(L.CheckString(1))
	args = append([]string{cmd}, args...)
	process, err := os.StartProcess(cmd, args, &procAttr)
	if err != nil {
		L.Push(LNumber(1))
		return 1
	}

	ps, err := process.Wait()
	if err != nil || !ps.Success() {
		L.Push(LNumber(1))
		return 1
	}
	L.Push(LNumber(0))
	return 1
}

// osExit 模块函数，用于退出当前进程
// 参数：
//  1. code (int) - 退出码
//
// 返回值：
//
//	无
//
// 调用方式：
//  1. oslib.Exit(code)
//
// 备注：
//  1. 退出码为 0 表示正常退出，非 0 表示异常退出
//  2. 该函数会立即终止当前进程
//  3. 该函数默认退出码为 0
func osExit(L *LState) int {
	L.Close()
	os.Exit(L.OptInt(1, 0))
	return 1
}

// osGetEnv 模块函数，用于获取环境变量
// 参数：
//  1. name (string) - 环境变量名称
//
// 返回值：
//  1. string（环境变量值）
//
// 调用方式：
//  1. local value = oslib.GetEnv(name)
//
// 备注：
//  1. 如果环境变量不存在，则返回 nil
func osGetEnv(L *LState) int {
	v := os.Getenv(L.CheckString(1))
	if len(v) == 0 {
		L.Push(LNil)
	} else {
		L.Push(LString(v))
	}
	return 1
}

// osRemove 模块函数，用于删除指定路径的文件或空目录。
// 参数：
//  1. path (string) - 要删除的文件或空目录路径。
//
// 返回值：
//  1. bool（是否删除成功）
//  2. string（错误信息）
//
// 调用方式：
//  1. local success, errmsg = oslib.Remove(path)
//
// 备注：
//  1. 本函数不会抛出异常。
func osRemove(L *LState) int {
	err := os.Remove(L.CheckString(1))
	if err != nil {
		L.Push(LNil)
		L.Push(LString(err.Error()))
		return 2
	} else {
		L.Push(LTrue)
		return 1
	}
}

// osRename 模块函数，用于重命名文件或目录
// 参数：
//  1. oldname (string) - 旧文件名
//  2. newname (string) - 新文件名
//
// 返回值：
//  1. bool（是否重命名成功）
//  2. string（错误信息）
//
// 调用方式：
//  1. local success, err = oslib.Rename(oldname, newname)
//
// 备注：
//  1. 如果重命名成功，则返回 true，否则返回 false
//  2. 如果重命名失败，则返回错误信息
//  3. 该函数不会抛出异常
//  4. 该函数重命名的是相对源码文件"./"目录下的文件或目录
func osRename(L *LState) int {
	err := os.Rename(L.CheckString(1), L.CheckString(2))
	if err != nil {
		L.Push(LNil)
		L.Push(LString(err.Error()))
		return 2
	} else {
		L.Push(LTrue)
		return 1
	}
}

// osSetEnv 模块函数，用于设置环境变量
// 参数：
//  1. name (string) - 环境变量名称
//  2. value (string) - 环境变量值
//
// 返回值：
//  1. bool（是否设置成功）
//  2. string（错误信息）
//
// 调用方式：
//  1. local success, err = oslib.SetEnv(name, value)
//
// 备注：
//  1. 如果设置成功，则返回 true，否则返回 false
//  2. 如果设置失败，则返回错误信息
//  3. 该函数不会抛出异常
func osSetEnv(L *LState) int {
	err := os.Setenv(L.CheckString(1), L.CheckString(2))
	if err != nil {
		L.Push(LNil)
		L.Push(LString(err.Error()))
		return 2
	} else {
		L.Push(LTrue)
		return 1
	}
}

// osTmpname 模块函数，用于生成一个临时文件名
// 参数：
//
//	无
//
// 返回值：
//  1. string（临时文件名）
//  2. string（错误信息）
//
// 调用方式：
//  1. local tmpname, err = oslib.TmpName()
//
// 备注：
//  1. 生成的文件名是唯一的
//  2. 生成的文件名是相对源码文件"./"目录下的文件
func osTmpname(L *LState) int {
	file, err := os.CreateTemp("", "")
	if err != nil {
		L.Push(LNil)
		L.Push(LString("unable to generate a unique filename"))
		return 2
	}
	file.Close()
	os.Remove(file.Name()) // 忽略错误
	L.Push(LString(file.Name()))
	return 1
}

// osPathJoin 模块函数，用于连接多个路径
// 参数：
//  1. ... (string) - 多个路径
//
// 返回值：
//  1. string（连接后的路径）
//
// 调用方式：
//  1. local path = oslib.PathJoin(...)
//
// 备注：
//  1. 该函数会自动处理路径分隔符
//  2. 该函数会自动处理路径中的"."和".."等特殊字符
//  3. 该函数会自动处理路径中的多个分隔符
func osPathJoin(L *LState) int {
	n := L.GetTop()
	if n == 0 {
		L.Push(LString(""))
		return 1
	}
	parts := make([]string, 0, n)
	for i := 1; i <= n; i++ {
		part := L.CheckString(i)
		if part != "" {
			parts = append(parts, part)
		}
	}

	joined := filepath.Clean(filepath.Join(parts...))
	L.Push(LString(joined))
	return 1
}

// osAbsPath 返回给定路径的绝对路径
// 参数：
//  1. path (string) - 给定路径
//
// 返回值：
//  1. string（绝对路径）
//  2. string（错误信息）
//
// 调用方式：
//  1. local abspath, err = oslib.AbsPath(path)
//
// 备注：
//  1. 如果获取绝对路径失败，则返回错误信息
//  2. 该函数返回的路径是绝对路径
//  3. 该函数不会抛出异常
func osAbsPath(L *LState) int {
	path := L.CheckString(1)
	absPath, err := filepath.Abs(path)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(err.Error()))
		return 2
	}
	L.Push(LString(absPath))
	return 1
}

// osGetCurrWorkingDir 模块函数，用于获取当前工作目录
// 参数：
//
//	无
//
// 返回值：
//  1. string（当前工作目录）
//
// 调用方式：
//  1. local cwd = oslib.GetCurrWorkingDir()
//
// 备注：
//  1. 返回的路径是当前工作目录
//  2. 该函数不会抛出异常
func osGetCurrWorkingDir(L *LState) int {
	cwd, err := os.Getwd()
	if err != nil {
		L.Push(LNil)
		L.Push(LString(err.Error()))
		return 2
	}
	L.Push(LString(cwd))
	return 1
}

// osDirName 模块函数，用于获取给定路径的目录部分
// 参数：
//  1. path (string) - 给定路径
//
// 返回值：
//  1. string（目录部分）
//
// 调用方式：
//  1. local dir = oslib.DirName(path)
//
// 备注：
//  1. 返回的路径是给定路径的目录部分
//  2. 该函数不会抛出异常
func osDirName(L *LState) int {
	path := L.CheckString(1)
	dir := filepath.Dir(path)
	L.Push(LString(dir))
	return 1
}

// osRelPath 模块函数，用于获取基础路径到目标路径的相对路径
// 参数：
//  1. base (string) - 基础路径
//  2. target (string) - 目标路径
//
// 返回值：
//  1. string（相对路径）
//  2. string（错误信息）
//
// 调用方式：
//  1. local relpath, err = oslib.RelPath(base, target)
//
// 备注：
//  1. 如果获取相对路径失败，则返回错误信息
//  2. 该函数返回的路径是相对路径
//  3. 该函数不会抛出异常
func osRelPath(L *LState) int {
	base := L.CheckString(1)
	target := L.CheckString(2)
	rel, err := filepath.Rel(base, target)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(err.Error()))
		return 2
	}
	L.Push(LString(rel))
	return 1
}

// osGetPID 模块函数，用于获取当前进程 ID
// 参数：
//
//	无
//
// 返回值：
//  1. int（进程 ID）
//
// 调用方式：
//  1. local pid = oslib.GetPID()
//
// 备注：
//  1. 返回的 ID 是当前进程的 ID
func osGetPID(L *LState) int {
	L.Push(LNumber(os.Getpid()))
	return 1
}

// osGetPPID 模块函数，用于获取当前进程的父进程 ID
// 参数：
//
//	无
//
// 返回值：
//  1. int（父进程 ID）(Windows 下可能返回 nil)
//  2. string（错误信息）
//
// 调用方式：
//  1. local ppid = oslib.GetPPID()
//
// 备注：
//  1. 返回的 ID 是当前进程的父进程 ID
func osGetPPID(L *LState) int {
	if MilkOS == "windows" {
		if ppid := getWindowsPPID(); ppid != 0 {
			L.Push(LNumber(ppid))
			return 1
		} else {
			L.Push(LNil)
			L.Push(LString("GetPPID not supported on Windows"))
			return 2
		}
	}
	L.Push(LNumber(syscall.Getppid()))
	return 1
}

func getWindowsPPID() int {
	cmd := exec.Command("tasklist", "/fi", "imagename eq cmd.exe", "/fo", "csv", "/nh")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(string(output), "\r\n")
	if len(lines) < 2 {
		return 0
	}

	fields := strings.Split(lines[0], ",")
	if len(fields) < 2 {
		return 0
	}

	ppidStr := strings.Trim(fields[1], "\"")
	ppid, err := strconv.Atoi(ppidStr)
	if err != nil {
		return 0
	}

	return ppid
}

// osMkdirAll 模块函数，用于创建目录
// 参数：
//  1. path (string) - 目录路径
//  2. mode (int) - 目录权限
//
// 返回值：
//  1. bool（是否创建成功）
//  2. string（错误信息）
//
// 调用方式：
//  1. local success, err = oslib.MkdirAll(path, mode)
//
// 备注：
//  1. 如果创建成功，则返回 true，否则返回 false
//  2. 如果创建失败，则返回错误信息
//  3. 该函数不会抛出异常
func osMkdirAll(L *LState) int {
	path := L.CheckString(1)
	mode := L.OptInt(2, 0755)
	if err := os.MkdirAll(path, os.FileMode(mode)); err != nil {
		L.Push(LFalse)
		L.Push(LString(fmt.Sprintf("error creating directory: %s", err.Error())))
		return 2
	}
	L.Push(LTrue)
	return 1
}

// osSymlink 模块函数，用于创建符号链接
// 参数：
//  1. oldname (string) - 源文件名
//  2. newname (string) - 链接文件名
//
// 返回值：
//  1. bool（是否创建成功）
//  2. string（错误信息）
//
// 调用方式：
//  1. local success, err = oslib.Symlink(oldname, newname)
//
// 备注：
//  1. 如果创建成功，则返回 true，否则返回 false
//  2. 如果创建失败，则返回错误信息
//  3. 该函数不会抛出异常
func osSymlink(L *LState) int {
	oldname := L.CheckString(1)
	newname := L.CheckString(2)
	if err := os.Symlink(oldname, newname); err != nil {
		L.Push(LFalse)
		L.Push(LString(fmt.Sprintf("error creating symlink: %s", err.Error())))
		return 2
	}
	L.Push(LTrue)
	return 1
}

// osStat 模块函数，用于获取文件信息
// 参数：
//  1. path (string) - 文件路径
//
// 返回值：
//  1. table（文件信息）
//
// 调用方式：
//  1. local info = oslib.Stat(path)
//
// 备注：
//  1. 返回的信息包括文件大小、文件权限、修改时间、是否为目录等
func osStat(L *LState) int {
	path := L.CheckString(1)
	info, err := os.Stat(path)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(err.Error()))
		return 2
	}
	tb := L.NewTable()
	tb.RawSetString("size", LNumber(info.Size()))
	tb.RawSetString("mode", LNumber(info.Mode()))
	tb.RawSetString("modifytime", LNumber(info.ModTime().Unix()))
	tb.RawSetString("isdir", LBool(info.IsDir()))
	L.Push(tb)
	return 1
}

// osExists 模块函数，用于判断文件是否存在
// 参数：
//  1. path (string) - 文件路径
//
// 返回值：
//  1. bool（是否存在）
//
// 调用方式：
//  1. local exists = oslib.Exists(path)
//
// 备注：
//  1. 如果文件存在，则返回 true，否则返回 false
func osExists(L *LState) int {
	path := L.CheckString(1)
	if _, err := os.Stat(path); err != nil {
		L.Push(LFalse)
		return 1
	}
	L.Push(LTrue)
	return 1
}

// osMCpus 模块函数，用于获取 CPU 相关信息
// 参数：
//
//	无
//
// 返回值：
//  1. table（CPU 信息）
//
// 调用方式：
//  1. local cpus = oslib.MCpus()
//
// 备注：
//  1. 返回的信息包括 CPU 数量和 GOMAXPROCS 值
func osMCpus(L *LState) int {
	tbl := L.NewTable()
	tbl.RawSetString("num", LNumber(runtime.NumCPU()))
	tbl.RawSetString("gomaxprocs", LNumber(runtime.GOMAXPROCS(0)))
	L.Push(tbl)
	return 1
}

// osGetOSName 模块函数，用于获取操作系统名称
// 参数：
//
//	无
//
// 返回值：
//  1. string（操作系统名称）
//
// 调用方式：
//  1. local osname = oslib.GetOSName()
//
// 备注：
//  1. 返回的名称包括 "windows"、"linux"、"darwin" 等
func osGetOSName(L *LState) int {
	L.Push(LString(runtime.GOOS))
	return 1
}
