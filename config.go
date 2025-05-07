package lua

import (
	"os"
)

var CompatVarArg = true
var FieldsPerFlush = 50
var RegistrySize = 256 * 20
var RegistryGrowStep = 32
var CallStackSize = 256
var MaxTableGetLoop = 100
var MaxArrayIndex = 67108864
var MaxLoopUnrollCount = 8

type LNumber float64

const LNumberBit = 64
const LNumberScanFormat = "%f"
const LuaCompVersion = "Lua 5.1"

var LuaPath = "MILK_PATH"
var MilkLDir string
var MilkPathDefault string
var MilkOS string
var MilkDirSep string
var MilkPathSep = ";"
var MilkPathMark = "?"
var MilkExecDir = "!"
var MilkIgMark = "-"

func init() {
	if os.PathSeparator == '/' { // unix-like
		MilkOS = "unix"
		MilkLDir = "/usr/local/share/milk"
		MilkDirSep = "/"
		MilkPathDefault = "./?.mlk;" + MilkLDir + "/?.mlk;" + MilkLDir + "/?/init.mlk"
	} else { // windows
		MilkOS = "windows"
		MilkLDir = "!\\milk"
		MilkDirSep = "\\"
		MilkPathDefault = ".\\?.mlk;" + MilkLDir + "\\?.mlk;" + MilkLDir + "\\?\\init.mlk"
	}
}
