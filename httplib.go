package lua

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

// 默认超时时间和时间单位配置
var (
	// 初始超时设为 30 秒
	requestTimeout = 30 * time.Second
)

// OpenHttp 模块入口，注册所有 http 模块函数
func OpenHttp(L *LState) int {
	httpmod := L.RegisterModule(HttpLibName, httpModuleFuncs)
	L.Push(httpmod)
	return 1
}

// 模块文档说明
var HttpLibFuncDoc = map[string]libFuncDoc{
	HttpLibName: {
		libName: HttpLibName,
		libFuncName: []string{
			"Get",
			"Post",
			"Put",
			"Patch",
			"Delete",
			"Head",
			"Options",
			"SetTimeout",
		},
	},
}

// 模块函数映射
var httpModuleFuncs = map[string]LGFunction{
	"Get":        httpGet,
	"Post":       httpPost,
	"Put":        httpPut,
	"Patch":      httpPatch,
	"Delete":     httpDelete,
	"Head":       httpHead,
	"Options":    httpOptions,
	"SetTimeout": httpSetTimeout,
}

// httpGet 模块函数，用于发送 HTTP GET 请求
func httpGet(L *LState) int {
	url := L.CheckString(1)
	headers := L.OptTable(2, nil)

	// 使用 context.WithTimeout 控制请求超时
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		L.RaiseError("create request failed: %v", err)
		return 0
	}

	if headers != nil && isValidHeader(headers) {
		req.Header = tableToHeader(L, headers)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		L.RaiseError("HTTP get error: %v", err)
		return 0
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		L.RaiseError("HTTP read error: %v", err)
		return 0
	}
	L.Push(LString(string(data)))
	return 1
}

// httpPost 模块函数，用于发送 HTTP POST 请求
func httpPost(L *LState) int {
	url := L.CheckString(1)
	body := L.CheckString(2)
	headers := L.OptTable(3, nil)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if err != nil {
		L.RaiseError("create request failed: %v", err)
		return 0
	}

	if headers != nil && isValidHeader(headers) {
		req.Header = tableToHeader(L, headers)
	}

	// 默认 Content-Type
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		L.RaiseError("HTTP post error: %v", err)
		return 0
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		L.RaiseError("HTTP read error: %v", err)
		return 0
	}
	L.Push(LString(string(data)))
	return 1
}

// httpPut 模块函数，用于发送 HTTP PUT 请求
func httpPut(L *LState) int {
	url := L.CheckString(1)
	body := L.CheckString(2)
	headers := L.OptTable(3, nil)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "PUT", url, strings.NewReader(body))
	if err != nil {
		L.RaiseError("create request failed: %v", err)
		return 0
	}

	if headers != nil && isValidHeader(headers) {
		req.Header = tableToHeader(L, headers)
	}

	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		L.RaiseError("HTTP put error: %v", err)
		return 0
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		L.RaiseError("HTTP read error: %v", err)
		return 0
	}
	L.Push(LString(string(data)))
	return 1
}

// httpPatch 模块函数，用于发送 HTTP PATCH 请求
func httpPatch(L *LState) int {
	url := L.CheckString(1)
	body := L.CheckString(2)
	headers := L.OptTable(3, nil)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, strings.NewReader(body))
	if err != nil {
		L.RaiseError("create request failed: %v", err)
		return 0
	}

	if headers != nil && isValidHeader(headers) {
		req.Header = tableToHeader(L, headers)
	}

	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		L.RaiseError("HTTP patch error: %v", err)
		return 0
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		L.RaiseError("HTTP read error: %v", err)
		return 0
	}
	L.Push(LString(string(data)))
	return 1
}

// httpDelete 模块函数，用于发送 HTTP DELETE 请求
func httpDelete(L *LState) int {
	url := L.CheckString(1)
	headers := L.OptTable(2, nil)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		L.RaiseError("create request failed: %v", err)
		return 0
	}

	if headers != nil && isValidHeader(headers) {
		req.Header = tableToHeader(L, headers)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		L.RaiseError("HTTP delete error: %v", err)
		return 0
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		L.RaiseError("HTTP read error: %v", err)
		return 0
	}
	L.Push(LString(string(data)))
	return 1
}

// httpHead 模块函数，用于发送 HTTP HEAD 请求
func httpHead(L *LState) int {
	url := L.CheckString(1)
	headers := L.OptTable(2, nil)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		L.RaiseError("create request failed: %v", err)
		return 0
	}

	if headers != nil && isValidHeader(headers) {
		req.Header = tableToHeader(L, headers)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		L.RaiseError("HTTP head error: %v", err)
		return 0
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	L.Push(LString(string(data)))
	return 1
}

// httpOptions 模块函数，用于发送 HTTP OPTIONS 请求
func httpOptions(L *LState) int {
	url := L.CheckString(1)
	headers := L.OptTable(2, nil)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "OPTIONS", url, nil)
	if err != nil {
		L.RaiseError("create request failed: %v", err)
		return 0
	}

	if headers != nil && isValidHeader(headers) {
		req.Header = tableToHeader(L, headers)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		L.RaiseError("HTTP options error: %v", err)
		return 0
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	L.Push(LString(string(data)))
	return 1
}

// httpSetTimeout 模块函数，用于设置 HTTP 请求的超时时间
// 接受两个参数：超时时间长度（数字）和可选的时间单位（默认 "s"）
func httpSetTimeout(L *LState) int {
	timelength := L.OptNumber(1, LNumber(float64(requestTimeout/time.Second)))
	timeunit := L.OptString(2, defaultTimeUnit)
	dur, ok := timeUnit[timeunit]
	if !ok {
		L.RaiseError("invalid time unit %q", timeunit)
		return 0
	}
	requestTimeout = time.Duration(float64(timelength) * float64(dur))
	return 0
}
