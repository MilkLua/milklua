package lua

const (
	// BaseLibName is here for consistency; the base functions have no namespace/library.
	BaseLibName = ""
	// LoadLibName is here for consistency; the loading system has no namespace/library.
	LoadLibName = "pkglib"
	// TabLibName is the name of the table Library.
	TabLibName = "tbllib"
	// IoLibName is the name of the io Library.
	IoLibName = "iolib"
	// OsLibName is the name of the os Library.
	OsLibName = "oslib"
	// StringLibName is the name of the string Library.
	StringLibName = "strlib"
	// MathLibName is the name of the math Library.
	MathLibName = "matlib"
	// DebugLibName is the name of the debug Library.
	DebugLibName = "dbglib"
	// ChannelLibName is the name of the channel Library.
	ChannelLibName = "chnlib"
	// CoroutineLibName is the name of the coroutine Library.
	CoroutineLibName = "coroutlib"
	// TimeLibName is the name of the time Library.
	TimeLibName = "timelib"
	// RandomLibName is the name of the random Library.
	RandomLibName = "randlib"

	// JsonLibName is the name of the json Library.
	JsonLibName = "jsonlib"
	// YamlLibName is the name of the yaml Library.
	YamlLibName = "yamllib"
	// XmlLibName is the name of the xml Library.
	XmlLibName = "xmllib"
	// TomlLibName is the name of the toml Library.
	TomlLibName = "tomllib"
	// Base64LibName is the name of the base64 Library.
	Base64LibName = "b64lib"
	// Base32LibName is the name of the base32 Library.
	Base32LibName = "b32lib"
	// Base62XLibName is the name of the base62x Library.
	Base62XLibName = "b62xlib"
	// HexLibName is the name of the hex Library.
	HexLibName = "hexlib"
	// UrlLibName is the name of the url Library.
	UrlLibName = "urllib"

	// HttpLibName is the name of the http Library.
	HttpLibName = "httplib"
	// WsLibName is the name of the websocket Library.
	WsLibName = "wslib"
)

type luaLib struct {
	libName string
	libFunc LGFunction
}

type libFuncDoc struct {
	libName     string
	libFuncName []string
}

var luaLibs = []luaLib{
	{LoadLibName, OpenPackage},
	{BaseLibName, OpenBase},
	{TabLibName, OpenTable},
	{IoLibName, OpenIo},
	{OsLibName, OpenOs},
	{StringLibName, OpenString},
	{MathLibName, OpenMath},
	{DebugLibName, OpenDebug},
	{ChannelLibName, OpenChannel},
	{CoroutineLibName, OpenCoroutine},
	{TimeLibName, OpenTime},
	{RandomLibName, OpenRandom},

	// --- Encoding/Decoding Libraries ---
	{JsonLibName, OpenJson},
	{YamlLibName, OpenYml},
	{XmlLibName, OpenXml},
	{TomlLibName, OpenToml},
	{Base64LibName, OpenBase64},
	{Base32LibName, OpenBase32},
	{Base62XLibName, OpenBase62X},
	{HexLibName, OpenHex},
	{UrlLibName, OpenURLLib},

	// --- network Libraries ---
	{HttpLibName, OpenHttp},
	{WsLibName, OpenWs},
}

func ShowFuncDoc() string {
	var LibFuncDoc = map[string]libFuncDoc{}
	LibFuncDoc[BaseLibName] = BaseLibFuncDoc[BaseLibName]
	LibFuncDoc[LoadLibName] = LoLibFuncDoc[LoadLibName]
	LibFuncDoc[TabLibName] = TblLibFuncDoc[TabLibName]
	LibFuncDoc[IoLibName] = IoLibFuncDoc[IoLibName]
	LibFuncDoc[OsLibName] = OsLibFuncDoc[OsLibName]
	LibFuncDoc[StringLibName] = StrLibFuncDoc[StringLibName]
	LibFuncDoc[MathLibName] = MatLibFuncDoc[MathLibName]
	LibFuncDoc[DebugLibName] = DbgLibFuncDoc[DebugLibName]
	LibFuncDoc[ChannelLibName] = ChnLibFuncDoc[ChannelLibName]
	LibFuncDoc[CoroutineLibName] = CoroutLibFuncDoc[CoroutineLibName]
	LibFuncDoc[TimeLibName] = TimeLibFuncDoc[TimeLibName]
	LibFuncDoc[RandomLibName] = RandomLibFuncDoc[RandomLibName]

	LibFuncDoc[JsonLibName] = JsonLibFuncDoc[JsonLibName]
	LibFuncDoc[YamlLibName] = YamlLibFuncDoc[YamlLibName]
	LibFuncDoc[XmlLibName] = XmlLibFuncDoc[XmlLibName]
	LibFuncDoc[TomlLibName] = TomlLibFuncDoc[TomlLibName]
	LibFuncDoc[Base64LibName] = Base64LibFuncDoc[Base64LibName]
	LibFuncDoc[Base32LibName] = Base32LibFuncDoc[Base32LibName]
	LibFuncDoc[Base62XLibName] = Base62XLibFuncDoc[Base62XLibName]
	LibFuncDoc[HexLibName] = HexLibFuncDoc[HexLibName]
	LibFuncDoc[UrlLibName] = URLLibFuncDoc[UrlLibName]

	LibFuncDoc[HttpLibName] = HttpLibFuncDoc[HttpLibName]
	LibFuncDoc[WsLibName] = WsLibFuncDoc[WsLibName]
	var doc string
	doc += PackageCopyRight + "\n"
	for _, lib := range luaLibs {
		if doc != "" {
			doc += "\n"
		}
		if lib.libName == "" {
			doc += "Base library:\n"
		} else {
			doc += lib.libName + " library:\n"
		}
		for _, funcName := range LibFuncDoc[lib.libName].libFuncName {
			if lib.libName == "" {
				doc += "\t" + funcName + "()\n"
			} else {
				doc += "\t" + lib.libName + "." + funcName + "()\n"
			}
		}
	}

	return doc
}

// OpenLibs loads the built-in libraries. It is equivalent to running OpenLoad,
// then OpenBase, then iterating over the other OpenXXX functions in any order.
func (ls *LState) OpenLibs() {
	// NB: Map iteration order in Go is deliberately randomised, so must open Load/Base
	// prior to iterating.
	for _, lib := range luaLibs {
		ls.Push(ls.NewFunction(lib.libFunc))
		ls.Push(LString(lib.libName))
		ls.Call(1, 0)
	}
}
