package lua

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"
)

var tablePool = sync.Pool{
	New: func() interface{} {
		return &LTable{array: make([]LValue, 0, 16), strdict: make(map[string]LValue)}
	},
}

// 递归转换Go数据结构到LValue
func goToLValue(L *LState, v interface{}) (LValue, error) {
	switch v := v.(type) {
	case nil:
		return LNil, nil

	case bool:
		return LBool(v), nil
	case float64:
		return LNumber(v), nil
	case int:
		return LNumber(v), nil
	case string:
		return LString(v), nil

	case []interface{}:
		tbl := tablePool.Get().(*LTable)
		tbl.Clear()
		tbl.array = make([]LValue, 0, len(v))

		defer tablePool.Put(tbl)

		for i, elem := range v {
			lv, err := goToLValue(L, elem)
			if err != nil {
				return nil, fmt.Errorf("index %d: %w", i, err)
			}
			tbl.RawSetInt(i+1, lv)
		}
		return tbl, nil

	case map[string]interface{}:
		tbl := L.CreateTable(0, len(v))
		for key, elem := range v {
			lv, err := goToLValue(L, elem)
			if err != nil {
				return nil, fmt.Errorf("key %q: %w", key, err)
			}
			tbl.RawSetString(key, lv)
		}
		return tbl, nil
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return LNil, nil
			}
			return goToLValue(L, rv.Elem().Interface())
		}
		return nil, fmt.Errorf("unsupported type: %T (kind %v)", v, rv.Kind())
	}
}

func (t *LTable) Clear() {
	t.array = t.array[:0]
	for k := range t.strdict {
		delete(t.strdict, k)
	}
}

// 递归转换Lua table到Go数据结构
func tableToGo(L *LState, tbl *LTable) interface{} {
	// 判断是否为数组式table
	isArray, maxIndex := isArrayTable(tbl)

	if isArray {
		arr := make([]interface{}, maxIndex)
		tbl.ForEach(func(k LValue, v LValue) {
			if idx, ok := k.(LNumber); ok {
				if i := int(idx); i >= 1 && i <= maxIndex {
					arr[i-1] = lvalueToGo(L, v)
				}
			}
		})
		return arr
	}

	// 处理map式table
	m := make(map[string]interface{})
	tbl.ForEach(func(k LValue, v LValue) {
		key := lvalueToString(L, k)
		m[key] = lvalueToGo(L, v)
	})
	return m
}

// 判断是否为数组式table（连续数字索引从1开始）
func isArrayTable(tbl *LTable) (bool, int) {
	maxIndex := 0
	count := 0
	tbl.ForEach(func(k LValue, _ LValue) {
		if n, ok := k.(LNumber); ok && n > 0 {
			index := int(n)
			if index == count+1 {
				count++
				if index > maxIndex {
					maxIndex = index
				}
			} else {
				maxIndex = -1 // 标记非连续
				return
			}
		} else {
			maxIndex = -1
			return
		}
	})
	return maxIndex > 0 && count == maxIndex, maxIndex
}

// 转换Lua值到Go值
func lvalueToGo(L *LState, lv LValue) interface{} {
	switch v := lv.(type) {
	case *LTable:
		return tableToGo(L, v)
	case LString:
		return string(v)
	case LNumber:
		return float64(v)
	case LBool:
		return bool(v)
	case *LNilType:
		return nil
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Lua值转字符串键
func lvalueToString(L *LState, lv LValue) string {
	switch v := lv.(type) {
	case LString:
		return string(v)
	case LNumber:
		return fmt.Sprintf("%v", float64(v))
	case LBool:
		return fmt.Sprintf("%v", bool(v))
	default:
		L.RaiseError("invalid table key type: %T", v)
		return ""
	}
}

// Lua table转http.Header
func tableToHeader(L *LState, tbl *LTable) http.Header {
	headers := make(http.Header, tbl.Len())
	tbl.ForEach(func(k LValue, v LValue) {
		key := lvalueToString(L, k)
		if values, ok := v.(*LTable); ok {
			values.ForEach(func(_, item LValue) { // 批量处理数组值
				headers.Add(key, lvalueToGo(L, item).(string))
			})
		} else {
			headers.Add(key, lvalueToGo(L, v).(string))
		}
	})
	return headers
}

// isValidHeader 检查 headers 是否有效
func isValidHeader(headers *LTable) bool {
	if headers == nil {
		return false
	}
	valid := true
	headers.ForEach(func(k LValue, v LValue) {
		if _, ok := v.(*LTable); ok {
			valid = false
			return
		}
	})
	return valid
}
