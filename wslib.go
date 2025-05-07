package lua

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	wsConnClass = "WS*"
)

// wsModuleFuncs 定义模块级别的函数（这里只包含 Connect 和 SetTimeout）
var wsModuleFuncs = map[string]LGFunction{
	"Connect":    wsConnect,
	"SetTimeout": wsSetTimeout,
}

// wsConnMethods 定义 websocket 连接的实例方法（面向对象调用）
var wsConnMethods = map[string]LGFunction{
	"Send":    wsConnSend,
	"Receive": wsConnReceive,
	"Close":   wsConnClose,
}

// WsLibFuncDoc 记录模块文档信息（仅供生成文档或调试使用）
var WsLibFuncDoc = map[string]libFuncDoc{
	WsLibName: {
		libName: WsLibName,
		libFuncName: []string{
			"Connect",
			"SetTimeout",
		},
	},
}

// OpenWs 模块入口，除了注册模块函数外，还需要注册 websocket 连接 userdata 的元方法
func OpenWs(L *LState) int {
	// 注册模块函数
	wsmod := L.RegisterModule(WsLibName, wsModuleFuncs).(*LTable)
	// 建立 wsConn 类型的元表
	mt := L.NewTypeMetatable(wsConnClass)
	mt.RawSetString("__index", mt)
	L.SetFuncs(mt, wsConnMethods)
	L.Push(wsmod)
	return 1
}

// wsConn 用于封装 websocket.Conn 对象
type wsConn struct {
	conn *websocket.Conn
}

// 默认握手超时和心跳超时时间
var wsDialTimeout = 60 * time.Second

// wsConnect 模块函数，用于建立 websocket 连接
// 参数：
//  1. url (string)
//  2. 可选的 headers (table)，转换为 http.Header
//
// 返回值：
//  1. userdata（封装了 *wsConn 对象，并设置了正确的元表，使其可调用 Send、Receive、Close 方法）
//
// 调用方式：local wsconn = wslib.Connect(url, headers)
// 备注：
//  1. headers 为可选参数，如果不传则默认为空表
//  2. headers 表中的值只能是 string 类型，否则会导致连接失败
//  3. 返回的 userdata 可以调用 Send、Receive、Close 方法，分别用于发送消息、接收消息、关闭连接
//  4. 连接建立后，需要定期发送心跳包，以保持连接活跃
func wsConnect(L *LState) int {
	url := L.CheckString(1)
	var hdr http.Header
	if L.GetTop() >= 2 {
		headers := L.OptTable(2, nil)
		if headers != nil && isValidHeader(headers) {
			hdr = tableToHeader(L, headers)
		} else {
			hdr = http.Header{}
		}
	} else {
		hdr = http.Header{}
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: wsDialTimeout,
	}
	conn, _, err := dialer.Dial(url, hdr)
	if err != nil {
		L.RaiseError("failed to connect to %q: %v", url, err)
		return 0
	}

	// 设置读超时和 pong 回调，保持连接活跃，避免服务器主动断开连接
	conn.SetReadDeadline(time.Now().Add(wsDialTimeout))
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(wsDialTimeout))
		return nil
	})

	ud := L.NewUserData()
	ud.Value = &wsConn{conn: conn}
	L.SetMetatable(ud, L.GetTypeMetatable(wsConnClass))
	L.Push(ud)
	return 1
}

// wsConnSend 为 wsConn 的实例方法，用于发送消息
// 参数：
//  1. 消息内容（string）
//
// 返回值：无
// 调用方式：
//  1. wsconn:Send(message)
//
// 备注：
//  1. 发送消息失败时，会抛出错误信息
//  2. 发送消息成功后，不会有返回值
func wsConnSend(L *LState) int {
	ud := L.CheckUserData(1)
	ws, ok := ud.Value.(*wsConn)
	if !ok || ws == nil {
		L.RaiseError("invalid websocket connection")
		return 0
	}
	message := L.CheckString(2)
	if err := ws.conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		L.RaiseError("send message failed: %v", err)
		return 0
	}
	return 0
}

// wsConnReceive 为 wsConn 的实例方法，用于接收消息
// 参数：无
// 返回值：
//  1. string（消息内容）
//
// 调用方式：
//  1. local msg = wsconn:Receive()
//
// 备注：
//  1. 接收消息失败时，会抛出错误信息
//  2. 接收消息成功后，返回消息内容
func wsConnReceive(L *LState) int {
	ud := L.CheckUserData(1)
	ws, ok := ud.Value.(*wsConn)
	if !ok || ws == nil {
		L.RaiseError("invalid websocket connection")
		return 0
	}
	_, message, err := ws.conn.ReadMessage()
	if err != nil {
		L.RaiseError("receive message failed: %v", err)
		return 0
	}
	L.Push(LString(string(message)))
	return 1
}

// wsConnClose 为 wsConn 的实例方法，用于关闭 websocket 连接
// 参数：无
// 返回值：无
// 调用方式：wsconn:Close()
func wsConnClose(L *LState) int {
	ud := L.CheckUserData(1)
	ws, ok := ud.Value.(*wsConn)
	if !ok || ws == nil {
		L.RaiseError("invalid websocket connection")
		return 0
	}
	if err := ws.conn.Close(); err != nil {
		L.RaiseError("close connection failed: %v", err)
		return 0
	}
	return 0
}

// wsSetTimeout 模块函数，用于设置建立 websocket 连接时的握手超时时间
// 参数：
//  1. 时间长度（number）
//  2. 时间单位（string）
//
// 返回值：无
// 调用方式：
//  1. wslib.SetTimeout(timelength, timeunit)
//
// 备注：
//  1. timeunit 为可选参数，如果不传则默认为 "s"
//  2. timeunit 默认为 "s"，即秒，支持的单位有 "ms"、"s"、"m"、"h"
//  3. 设置后的超时时间会应用到后续的所有 websocket 连接
func wsSetTimeout(L *LState) int {
	timelength := L.OptNumber(1, defaultTimeOutSec)
	timeunit := L.OptString(2, defaultTimeUnit)
	dur, ok := timeUnit[timeunit]
	if !ok {
		L.RaiseError("invalid time unit %q", timeunit)
		return 0
	}
	wsDialTimeout = time.Duration(timelength) * dur
	return 0
}
