package lua

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"milklua/pm"
)

const emptyLString LString = LString("")

func OpenString(L *LState) int {
	var mod *LTable
	//_, ok := L.G.builtinMts[int(LTString)]
	//if !ok {
	mod = L.RegisterModule(StringLibName, strFuncs).(*LTable)
	gmatch := L.NewClosure(strGmatch, L.NewFunction(strGmatchIter))
	mod.RawSetString("GMatch", gmatch)
	mod.RawSetString("GFind", gmatch)
	mod.RawSetString("__index", mod)
	L.G.builtinMts[int(LTString)] = mod
	//}
	L.Push(mod)
	return 1
}

var StrLibFuncDoc = map[string]libFuncDoc{
	StringLibName: {
		libName: "stringlib",
		libFuncName: []string{
			"Byte",
			"Char",
			"Find",
			"Format",
			"GSub",
			"Len",
			"Lower",
			"Match",
			"Rep",
			"Reverse",
			"Sub",
			"Upper",
		},
	},
}

var strFuncs = map[string]LGFunction{
	"Byte":    strByte,
	"Char":    strChar,
	"Find":    strFind,
	"Format":  strFormat,
	"GSub":    strGsub,
	"Len":     strLen,
	"Lower":   strLower,
	"Match":   strMatch,
	"Rep":     strRep,
	"Reverse": strReverse,
	"Sub":     strSub,
	"Upper":   strUpper,
}

// strByte 模块函数，用于返回字符串的字节值
// 参数：
//  1. str (string) - 待处理的字符串
//  2. i (number) - 起始位置（可选）
//  3. j (number) - 结束位置（可选）
//
// 返回值：
//  1. number（字节值）
//  2. ...（多个字节值）
//
// 调用方式：
//  1. local b = strlib.Byte(str)
//  2. local b1, b2 = strlib.Byte(str, i, j)
//
// 示例：
//
//	local str = "abc"
//	local b = strlib.Byte(str) // b = 97
//	local b1, b2 = strlib.Byte(str, 2, 3) // b1 = 98, b2 = 99
//
// 备注：
//  1. 如果只提供一个参数，则返回该字符的字节值
//  2. 如果提供两个参数，则返回从第一个参数到第二个参数的所有字符的字节值
//  3. 如果起始位置小于 1，则从第一个字符开始
//  4. 如果结束位置小于 1，则返回空值
//  5. 如果结束位置大于字符串长度，则返回字符串长度
//  6. 如果结束位置小于起始位置，则返回空值
func strByte(L *LState) int {
	str := L.CheckString(1)
	runes := []rune(str)
	length := len(runes)

	start := L.OptInt(2, 1)
	end := L.OptInt(3, start)

	// handle negative indexes
	if start < 0 {
		start = length + start + 1
	}
	if end < 0 {
		end = length + end + 1
	}

	// adjust the indexes
	start = intMax(1, intMin(start, length)) - 1
	end = intMax(start+1, intMin(end, length))

	// if only one index, return the code of the character
	if L.GetTop() == 2 {
		L.Push(LNumber(runes[start]))
		return 1
	}

	// return multiple values
	for i := start; i < end; i++ {
		L.Push(LNumber(runes[i]))
	}
	return end - start
}

// strChar 模块函数，用于返回字符
// 参数：
//  1. c1 (number) - 字符的字节值
//  2. ...（多个字节值）
//
// 返回值：
//  1. string（字符串）
//
// 调用方式：
//  1. local str = strlib.Char(c1)
//  2. local str = strlib.Char(c1, c2)
//  3. local str = strlib.Char(c1, c2, c3)
//
// 示例：
//
//	local str = strlib.Char(97) // str = "a"
//	local str = strlib.Char(97, 98) // str = "ab"
//	local str = strlib.Char(97, 98, 99) // str = "abc"
//
// 备注：
//  1. 返回由参数指定的字符组成的字符串
func strChar(L *LState) int {
	top := L.GetTop()
	chars := make([]rune, top)
	for i := 1; i <= top; i++ {
		chars[i-1] = rune(L.CheckInt(i))
	}
	L.Push(LString(string(chars)))
	return 1
}

// strFind 模块函数，用于查找字符串中的子串
// 参数：
//  1. str (string) - 待处理的字符串
//  2. pattern (string) - 模式
//  3. init (number) - 起始位置（可选，默认为 1）
//  4. plain (boolean) - 是否为普通模式（可选，默认为 false）
//
// 返回值：
//  1. number（起始位置）
//  2. number（结束位置）
//  3. ...（多个子串）
//
// 调用方式：
//  1. local s, e = strlib.Find(str, pattern)
//  2. local s, e = strlib.Find(str, pattern, init)
//  3. local s, e = strlib.Find(str, pattern, init, plain)
//
// 示例：
//
//	local str = "hello world"
//	local s, e = strlib.Find(str, "world") // s = 7, e = 11
//	local s, e = strlib.Find(str, "world", 8) // s = nil, e = nil
//	local s, e = strlib.Find(str, "world", 8, true) // s = nil, e = nil
//
// 备注：
//  1. 返回字符串中匹配模式的起始位置和结束位置
//  2. 如果未找到匹配，则返回 nil
//  3. 如果提供了多个子串，则返回多个子串
//  4. 如果提供了起始位置，则从该位置开始查找
//  5. 如果提供了普通模式，则使用普通模式查找
func strFind(L *LState) int {
	str := L.CheckString(1)
	pattern := L.CheckString(2)
	if len(pattern) == 0 {
		L.Push(LNumber(1))
		L.Push(LNumber(0))
		return 2
	}
	init := luaIndex2StringIndex(str, L.OptInt(3, 1), true)
	plain := false
	if L.GetTop() == 4 {
		plain = LVAsBool(L.Get(4))
	}

	if plain {
		pos := strings.Index(str[init:], pattern)
		if pos < 0 {
			L.Push(LNil)
			return 1
		}
		L.Push(LNumber(init+pos) + 1)
		L.Push(LNumber(init + pos + len(pattern)))
		return 2
	}

	mds, err := pm.Find(pattern, unsafeFastStringToReadOnlyBytes(str), init, 1)
	if err != nil {
		L.RaiseError("%s", err.Error())
	}
	if len(mds) == 0 {
		L.Push(LNil)
		return 1
	}
	md := mds[0]
	L.Push(LNumber(md.Capture(0) + 1))
	L.Push(LNumber(md.Capture(1)))
	for i := 2; i < md.CaptureLength(); i += 2 {
		if md.IsPosCapture(i) {
			L.Push(LNumber(md.Capture(i)))
		} else {
			L.Push(LString(str[md.Capture(i):md.Capture(i+1)]))
		}
	}
	return md.CaptureLength()/2 + 1
}

// strFormat 模块函数，用于格式化字符串
// 参数：
//  1. str (string) - 待处理的字符串
//  2. ...（多个参数）
//
// 返回值：
//  1. string（格式化后的字符串）
//
// 调用方式：
//  1. local str = strlib.Format(str)
//  2. local str = strlib.Format(str, ...)
//
// 示例：
//
//	local str = "hello %s"
//	local str = strlib.Format(str, "world") // str = "hello world"
//
// 备注：
//  1. 返回格式化后的字符串
func strFormat(L *LState) int {
	str := L.CheckString(1)
	args := make([]interface{}, L.GetTop()-1)
	top := L.GetTop()
	for i := 2; i <= top; i++ {
		args[i-2] = L.Get(i)
	}
	npat := strings.Count(str, "%") - strings.Count(str, "%%")
	L.Push(LString(fmt.Sprintf(str, args[:intMin(npat, len(args))]...)))
	return 1
}

func strGsub(L *LState) int {
	str := L.CheckString(1)
	pat := L.CheckString(2)
	L.CheckTypes(3, LTString, LTTable, LTFunction)
	repl := L.CheckAny(3)
	limit := L.OptInt(4, -1)

	mds, err := pm.Find(pat, unsafeFastStringToReadOnlyBytes(str), 0, limit)
	if err != nil {
		L.RaiseError("%s", err.Error())
	}
	if len(mds) == 0 {
		L.SetTop(1)
		L.Push(LNumber(0))
		return 2
	}
	switch lv := repl.(type) {
	case LString:
		L.Push(LString(strGsubStr(L, str, string(lv), mds)))
	case *LTable:
		L.Push(LString(strGsubTable(L, str, lv, mds)))
	case *LFunction:
		L.Push(LString(strGsubFunc(L, str, lv, mds)))
	}
	L.Push(LNumber(len(mds)))
	return 2
}

type replaceInfo struct {
	Indicies []int
	String   string
}

func checkCaptureIndex(L *LState, m *pm.MatchData, idx int) {
	if idx <= 2 {
		return
	}
	if idx >= m.CaptureLength() {
		L.RaiseError("invalid capture index")
	}
}

func capturedString(L *LState, m *pm.MatchData, str string, idx int) string {
	checkCaptureIndex(L, m, idx)
	if idx >= m.CaptureLength() && idx == 2 {
		idx = 0
	}
	if m.IsPosCapture(idx) {
		return fmt.Sprint(m.Capture(idx))
	} else {
		return str[m.Capture(idx):m.Capture(idx+1)]
	}

}

func strGsubDoReplace(str string, info []replaceInfo) string {
	offset := 0
	buf := []byte(str)
	for _, replace := range info {
		oldlen := len(buf)
		b1 := append([]byte(""), buf[0:offset+replace.Indicies[0]]...)
		b2 := []byte("")
		index2 := offset + replace.Indicies[1]
		if index2 <= len(buf) {
			b2 = append(b2, buf[index2:]...)
		}
		buf = append(b1, replace.String...)
		buf = append(buf, b2...)
		offset += len(buf) - oldlen
	}
	return string(buf)
}

func strGsubStr(L *LState, str string, repl string, matches []*pm.MatchData) string {
	// make a slice of replaceInfo
	infoList := make([]replaceInfo, 0, len(matches))
	// define a function to process the replacement
	processReplacement := func(match *pm.MatchData) string {
		scanner := newFlagScanner('%', "", "", repl)
		for c, eos := scanner.Next(); !eos; c, eos = scanner.Next() {
			// if the flag is changed, skip the current character
			if scanner.ChangeFlag {
				continue
			}
			if scanner.HasFlag {
				if c >= '0' && c <= '9' {
					// match index
					scanner.AppendString(capturedString(L, match, str, 2*(int(c)-int('0'))))
				} else {
					scanner.AppendChar('%')
					scanner.AppendChar(c)
				}
				scanner.HasFlag = false
			} else {
				scanner.AppendChar(c)
			}
		}
		return scanner.String()
	}

	// iterate over the matches and process the replacement
	for _, match := range matches {
		start, end := match.Capture(0), match.Capture(1)
		replacement := processReplacement(match)
		infoList = append(infoList, replaceInfo{[]int{start, end}, replacement})
	}

	return strGsubDoReplace(str, infoList)
}

func strGsubTable(L *LState, str string, repl *LTable, matches []*pm.MatchData) string {
	infoList := make([]replaceInfo, 0, len(matches))
	for _, match := range matches {
		idx := 0
		if match.CaptureLength() > 2 { // has captures
			idx = 2
		}
		var value LValue
		if match.IsPosCapture(idx) {
			value = L.GetTable(repl, LNumber(match.Capture(idx)))
		} else {
			value = L.GetField(repl, str[match.Capture(idx):match.Capture(idx+1)])
		}
		if !LVIsFalse(value) {
			infoList = append(infoList, replaceInfo{[]int{match.Capture(0), match.Capture(1)}, LVAsString(value)})
		}
	}
	return strGsubDoReplace(str, infoList)
}

func strGsubFunc(L *LState, str string, repl *LFunction, matches []*pm.MatchData) string {
	infoList := make([]replaceInfo, 0, len(matches))
	for _, match := range matches {
		start, end := match.Capture(0), match.Capture(1)
		L.Push(repl)
		nargs := 0
		if match.CaptureLength() > 2 { // has captures
			for i := 2; i < match.CaptureLength(); i += 2 {
				if match.IsPosCapture(i) {
					L.Push(LNumber(match.Capture(i)))
				} else {
					L.Push(LString(capturedString(L, match, str, i)))
				}
				nargs++
			}
		} else {
			L.Push(LString(capturedString(L, match, str, 0)))
			nargs++
		}
		L.Call(nargs, 1)
		ret := L.reg.Pop()
		if !LVIsFalse(ret) {
			infoList = append(infoList, replaceInfo{[]int{start, end}, LVAsString(ret)})
		}
	}
	return strGsubDoReplace(str, infoList)
}

type strMatchData struct {
	str     string
	pos     int
	matches []*pm.MatchData
}

func strGmatchIter(L *LState) int {
	md := L.CheckUserData(1).Value.(*strMatchData)
	str := md.str
	matches := md.matches
	idx := md.pos
	md.pos += 1
	if idx == len(matches) {
		return 0
	}
	L.Push(L.Get(1))
	match := matches[idx]
	if match.CaptureLength() == 2 {
		L.Push(LString(str[match.Capture(0):match.Capture(1)]))
		return 1
	}

	for i := 2; i < match.CaptureLength(); i += 2 {
		if match.IsPosCapture(i) {
			L.Push(LNumber(match.Capture(i)))
		} else {
			L.Push(LString(str[match.Capture(i):match.Capture(i+1)]))
		}
	}
	return match.CaptureLength()/2 - 1
}

// strGmatch 模块函数，用于迭代字符串中的子串
// 参数：
//  1. str (string) - 待处理的字符串
//  2. pattern (string) - 模式
//
// 返回值：
//  1. function（迭代器）
//  2. userdata（迭代器数据）
//  3. string (错误信息)
//
// 调用方式：
//  1. local iter, data, err = strlib.GMatch(str, pattern)
//
// 示例：
//
//	local str = "hello world"
//	local iter, data = strlib.GMatch(str, "%w+")
//	for s in iter, data do
//		PrintLn(s) // 输出：hello, world
//	end
//
// 备注：
//  1. 返回一个迭代器和迭代器数据
//  2. 迭代器数据用于保存迭代器状态
//  3. 迭代器用于迭代字符串中的子串
func strGmatch(L *LState) int {
	str := L.CheckString(1)
	pattern := L.CheckString(2)
	mds, err := pm.Find(pattern, []byte(str), 0, -1)
	if err != nil {
		L.Push(LNil)
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("Failed to compile pattern: %s", err.Error())))
		return 3
	}
	L.Push(L.Get(UpvalueIndex(1)))
	ud := L.NewUserData()
	ud.Value = &strMatchData{str, 0, mds}
	L.Push(ud)
	return 2
}

// strLen 模块函数，用于返回字符串的长度
// 参数：
//  1. str (string) - 待处理的字符串
//
// 返回值：
//  1. number（字符串长度）
//
// 调用方式：
//  1. local len = strlib.Len(str)
//
// 示例：
//
//	local str = "hello world"
//	local len = strlib.Len(str) // len = 11
//
// 备注：
//  1. 返回字符串的长度
func strLen(L *LState) int {
	str := L.CheckString(1)
	L.Push(LNumber(utf8.RuneCountInString(str)))
	return 1
}

// strLower 模块函数，用于将字符串转换为小写
// 参数：
//  1. str (string) - 待处理的字符串
//
// 返回值：
//  1. string（小写字符串）
//
// 调用方式：
//  1. local str = strlib.Lower(str)
//
// 示例：
//
//	local str = "HELLO WORLD"
//	local str = strlib.Lower(str) // str = "hello world"
//
// 备注：
//  1. 返回字符串的小写形式
func strLower(L *LState) int {
	str := L.CheckString(1)
	L.Push(LString(strings.ToLower(str)))
	return 1
}

// strMatch 模块函数，用于匹配字符串
// 参数：
//  1. str (string) - 待处理的字符串
//  2. pattern (string) - 模式
//  3. init (number) - 起始位置（可选，默认为 1）
//
// 返回值：
//  1. string（子串）
//  2. ...（多个子串）
//
// 调用方式：
//  1. local s = strlib.Match(str, pattern)
//  2. local s1, s2 = strlib.Match(str, pattern, init)
//
// 示例：
//
//	local str = "hello world"
//	local s = strlib.Match(str, "world") // s = "world"
//	local s1, s2 = strlib.Match(str, "world", 8) // s1 = nil, s2 = nil
//
// 备注：
//  1. 返回字符串中匹配模式的子串
//  2. 如果未找到匹配，则返回 nil
//  3. 如果提供了多个子串，则返回多个子串
//  4. 如果提供了起始位置，则从该位置开始查找
func strMatch(L *LState) int {
	str := L.CheckString(1)
	pattern := L.CheckString(2)
	offset := L.OptInt(3, 1)
	l := len(str)
	if offset < 0 {
		offset = l + offset + 1
	}
	offset--
	if offset < 0 {
		offset = 0
	}

	mds, err := pm.Find(pattern, unsafeFastStringToReadOnlyBytes(str), offset, 1)
	if err != nil {
		L.RaiseError("%s", err.Error())
	}
	if len(mds) == 0 {
		L.Push(LNil)
		return 0
	}
	md := mds[0]
	nsubs := md.CaptureLength() / 2
	switch nsubs {
	case 1:
		L.Push(LString(str[md.Capture(0):md.Capture(1)]))
		return 1
	default:
		for i := 2; i < md.CaptureLength(); i += 2 {
			if md.IsPosCapture(i) {
				L.Push(LNumber(md.Capture(i)))
			} else {
				L.Push(LString(str[md.Capture(i):md.Capture(i+1)]))
			}
		}
		return nsubs - 1
	}
}

// strRep 模块函数，用于重复字符串
// 参数：
//  1. str (string) - 待处理的字符串
//  2. n (number) - 重复次数（可选，默认为 1）
//
// 返回值：
//  1. string（重复后的字符串）
//
// 调用方式：
//  1. local str = strlib.Rep(str)
//  2. local str = strlib.Rep(str, n)
//
// 示例：
//
//	local str = "hello"
//	local str = strlib.Rep(str) // str = "hello"
//	local str = strlib.Rep(str, 3) // str = "hellohellohello"
//
// 备注：
//  1. 返回重复 n 次后的字符串
//  2. 如果 n 小于 0，则返回空字符串
func strRep(L *LState) int {
	str := L.CheckString(1)
	n := L.OptInt(2, 1)
	if n < 0 {
		L.Push(emptyLString)
	} else {
		L.Push(LString(strings.Repeat(str, n)))
	}
	return 1
}

// strReverse 模块函数，用于反转字符串
// 参数：
//  1. str (string) - 待处理的字符串
//
// 返回值：
//  1. string（反转后的字符串）
//
// 调用方式：
//  1. local str = strlib.Reverse(str)
//
// 示例：
//
//	local str = "hello"
//	local str = strlib.Reverse(str) // str = "olleh"
//
// 备注：
//  1. 返回反转后的字符串
func strReverse(L *LState) int {
	str := L.CheckString(1)
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	L.Push(LString(string(runes)))
	return 1
}

// strSub 模块函数，用于截取字符串
// 参数：
//  1. str (string) - 待处理的字符串
//  2. i (number) - 起始位置
//  3. j (number) - 结束位置（可选，默认为 -1）
//
// 返回值：
//  1. string（截取后的字符串）
//
// 调用方式：
//  1. local str = strlib.Sub(str, i)
//  2. local str = strlib.Sub(str, i, j)
//
// 示例：
//
//	local str = "hello world"
//	local str = strlib.Sub(str, 1) // str = "hello world"
//	local str = strlib.Sub(str, 1, 5) // str = "hello"
//
// 备注：
//  1. 返回从起始位置到结束位置的子串
//  2. 如果起始位置小于 1，则从第一个字符开始
//  3. 如果结束位置小于 1，则返回空字符串
//  4. 如果结束位置大于字符串长度，则返回字符串长度
//  5. 如果结束位置小于起始位置，则返回空字符串
func strSub(L *LState) int {
	str := L.CheckString(1)
	start := L.CheckInt(2)
	end := L.OptInt(3, -1)

	runes := []rune(str)
	length := len(runes)

	// adjust start index
	if start < 0 {
		start = length + start + 1
	}
	if start < 1 {
		start = 1
	}
	start-- // convert to 0-based index

	// adjust end index
	if end < 0 {
		end = length + end + 1
	}
	if end > length {
		end = length
	}

	// check if the indexes are valid
	if start >= length || end <= start {
		L.Push(emptyLString)
	} else {
		L.Push(LString(string(runes[start:end])))
	}
	return 1
}

// strUpper 模块函数，用于将字符串转换为大写
// 参数：
//  1. str (string) - 待处理的字符串
//
// 返回值：
//  1. string（大写字符串）
//
// 调用方式：
//  1. local str = strlib.Upper(str)
//
// 示例：
//
//	local str = "hello world"
//	local str = strlib.Upper(str) // str = "HELLO WORLD"
//
// 备注：
//  1. 返回字符串的大写形式
func strUpper(L *LState) int {
	str := L.CheckString(1)
	L.Push(LString(strings.ToUpper(str)))
	return 1
}

func luaIndex2StringIndex(str string, i int, start bool) int {
	runes := []rune(str)
	if start && i != 0 {
		i -= 1
	}
	l := len(runes)
	if i < 0 {
		i = l + i + 1
	}
	i = intMax(0, i)
	if !start && i > l {
		i = l
	}
	return i
}

//
