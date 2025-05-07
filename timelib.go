package lua

import (
	"fmt"
	"strings"
	"time"
)

var defaultTimeUnit = "s"

var defaultTimeFormat = "%c"

var timeUnit = map[string]time.Duration{
	"h":  time.Hour,
	"m":  time.Minute,
	"s":  time.Second,
	"ms": time.Millisecond,
}

func OpenTime(L *LState) int {
	mod := L.RegisterModule(TimeLibName, timeFuncs).(*LTable)
	L.Push(mod)
	return 1
}

var TimeLibFuncDoc = map[string]libFuncDoc{
	TimeLibName: {
		libName: "timelib",
		libFuncName: []string{
			"Unix",
			"Sleep",
			"Date",
			"Time",

			"SetDefaultUnit",
			"SetDefaultFormat",
		},
	},
}

var timeFuncs = map[string]LGFunction{
	"Unix":  timeUnix,
	"Sleep": timeSleep,
	"Date":  timeDate,
	"Time":  timeTime,

	"SetDefaultUnit":   timeSetDefaultUnit,
	"SetDefaultFormat": timeSetDefaultFormat,
}

// timeUnix 模块函数，用于获取当前时间戳
// 参数：
//  1. unit (string) - 时间单位（可选，默认为 "s"）
//
// 返回值：
//  1. number（当前时间戳）
//  2. string（错误信息）
//
// 调用方式：
//  1. local ts, err = timelib.Unix(unit)
//
// 备注：
//  1. 如果 unit 为 "s"，则返回秒级时间戳
//  2. 如果 unit 为 "ms"，则返回毫秒级时间戳
//  3. 如果 unit 为 "m"，则返回分钟级时间戳
//  4. 如果 unit 为 "h"，则返回小时级时间戳
//  5. 如果 unit 为其他值，则会返回错误信息
//  6. 如果不传入 unit 参数，则默认为 "s"
func timeUnix(L *LState) int {
	unit := L.OptString(1, defaultTimeUnit)
	dur, ok := timeUnit[unit]
	if !ok {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("invalid time unit %q", unit)))
		return 2
	}
	L.Push(LNumber(time.Now().UnixNano() / int64(dur)))
	return 1
}

// timeSleep 模块函数，用于休眠指定时间
// 参数：
//  1. duration (number) - 休眠时间
//  2. unit (string) - 时间单位（可选，默认为 "s"）
//
// 返回值：
//  1. string（错误信息）
//
// 调用方式：
//  1. timelib.Sleep(duration, unit)
//
// 备注：
//  1. 如果 unit 为 "s"，则休眠秒数
//  2. 如果 unit 为 "ms"，则休眠毫秒数
//  3. 如果 unit 为 "m"，则休眠分钟数
//  4. 如果 unit 为 "h"，则休眠小时数
//  5. 如果 unit 为其他值，则会返回错误信息
//  6. 如果不传入 unit 参数，则默认为 "s"
func timeSleep(L *LState) int {
	duration := L.CheckNumber(1)
	unit := L.OptString(2, defaultTimeUnit)
	dur, ok := timeUnit[unit]
	if !ok {
		L.Push(LString(fmt.Sprintf("invalid time unit %q", unit)))
		return 1
	}
	time.Sleep(time.Duration(duration) * dur)
	return 0
}

// timeDate 模块函数，用于获取时间日期字符串
// 参数：
//  1. format (string) - 时间格式（可选，默认为 "%c"）
//  2. timestamp (number) - 时间戳（可选，默认为当前时间）
//
// 返回值：
//  1. string|table（时间日期字符串或时间字段表）
//
// 调用方式：
//  1. local str = timelib.Date(format, timestamp)
//  2. local tbl = timelib.Date("*t", timestamp)
//
// 备注：
//  1. 如果 format 以 "*t" 开头，则返回一个包含时间字段的 table
//  2. 如果 format 以 "!" 开头，则返回 UTC 时间
//  3. 如果不传入 format 参数，则默认为 "%c"
//  4. 如果不传入 timestamp 参数，则默认为当前时间
func timeDate(L *LState) int {
	// default format is "%c"
	format := L.OptString(1, defaultTimeFormat)
	// detect if UTC time is requested
	isUTC := false
	if strings.HasPrefix(format, "!") {
		format = strings.TrimPrefix(format, "!")
		isUTC = true
	}

	// get timestamp from argument
	var t time.Time
	if L.GetTop() >= 2 {
		t = time.Unix(L.CheckInt64(2), 0)
	} else {
		t = time.Now()
	}
	if isUTC {
		t = t.UTC()
	}

	// if format starts with "*t" return a table with time fields
	if strings.HasPrefix(format, "*t") {
		ret := L.NewTable()
		ret.RawSetString("year", LNumber(t.Year()))
		ret.RawSetString("month", LNumber(t.Month()))
		ret.RawSetString("day", LNumber(t.Day()))
		ret.RawSetString("hour", LNumber(t.Hour()))
		ret.RawSetString("min", LNumber(t.Minute()))
		ret.RawSetString("sec", LNumber(t.Second()))
		ret.RawSetString("wday", LNumber(int(t.Weekday())+1))
		ret.RawSetString("yday", LNumber(t.YearDay()))
		ret.RawSetString("isdst", LBool(t.IsDST()))
		L.Push(ret)
	} else {
		L.Push(LString(strftime(t, format)))
	}
	return 1
}

// timeTime 模块函数，用于获取时间戳
// 参数：
//  1. tbl (table) - 时间字段表
//
// 返回值：
//  1. number（时间戳）
//  2. string（错误信息）
//
// 调用方式：
//  1. local ts = timelib.Time(tbl)
//
// 备注：
//  1. tbl 必须包含 "year"、"month"、"day"、"hour"、"min"、"sec" 字段
//  2. 如果 tbl 包含 "isdst" 字段，则表示是否为夏令时
//  3. 如果 tbl 不包含 "isdst" 字段，则默认为 false
func timeTime(L *LState) int {
	// return current timestamp if no argument is given
	if L.GetTop() == 0 || L.CheckAny(1) == LNil {
		L.Push(LNumber(time.Now().Unix()))
		return 1
	}

	tbl, ok := L.CheckAny(1).(*LTable)
	if !ok {
		L.TypeError(1, LTTable)
	}
	// read time fields from table
	sec := getIntField(L, tbl, "sec", 0)
	min := getIntField(L, tbl, "min", 0)
	hour := getIntField(L, tbl, "hour", 12)
	day := getIntField(L, tbl, "day", -1)
	month := getIntField(L, tbl, "month", -1)
	year := getIntField(L, tbl, "year", -1)

	// if day, month or year is missing or invalid, raise an error
	if day <= 0 || month <= 0 || year <= 0 {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("invalid date: %d-%d-%d", year, month, day)))
		return 2
	}
	isdst := getBoolField(L, tbl, "isdst", false)
	t := time.Date(year, time.Month(month), day, hour, min, sec, 0, time.Local)
	// adjust time if DST is different
	if isdst != t.IsDST() {
		if isdst {
			t = t.Add(time.Hour)
		} else {
			t = t.Add(-time.Hour)
		}
	}
	L.Push(LNumber(t.Unix()))
	return 1
}

// timeSetDefaultUnit 模块函数，用于设置默认时间单位
// 参数：
//  1. unit (string) - 时间单位
//
// 返回值：
//  1. string（错误信息）
//
// 调用方式：
//  1. local err = timelib.SetDefaultUnit(unit)
//
// 备注：
//  1. 如果 unit 为 "s"，则默认时间单位为秒
//  2. 如果 unit 为 "ms"，则默认时间单位为毫秒
//  3. 如果 unit 为 "m"，则默认时间单位为分钟
//  4. 如果 unit 为 "h"，则默认时间单位为小时
//  5. 如果 unit 为其他值，则会返回错误信息
//  6. 如果不传入 unit 参数，则默认为 "s"
func timeSetDefaultUnit(L *LState) int {
	unit := L.OptString(1, "s")
	if _, ok := timeUnit[unit]; !ok {
		L.Push(LString(fmt.Sprintf("invalid time unit %q", unit)))
	}
	defaultTimeUnit = unit
	return 0
}

// timeSetDefaultFormat 模块函数，用于设置默认时间格式
// 参数：
//  1. format (string) - 时间格式
//
// 返回值：
//  1. string（错误信息）
//
// 调用方式：
//  1. timelib.SetDefaultFormat(format)
//
// 备注：
//  1. 如果不传入 format 参数，则默认为 "%c"
func timeSetDefaultFormat(L *LState) int {
	format := L.OptString(1, "%c")
	defaultTimeFormat = format
	return 0
}
