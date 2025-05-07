ifeq ($(OS),Windows_NT)
    $(shell chcp 65001 >nul 2>&1)
endif

# 检测操作系统
ifeq ($(OS),Windows_NT)
    detected_OS := Windows
    RM := if exist "$(1)" rmdir /S /Q "$(1)"
    MKDIR := if not exist "$(1)" mkdir "$(1)"
    MKDIR_P = $(foreach dir,$(subst /,\,$(1)),if not exist $(dir) mkdir $(dir) &)
    SHELL := cmd.exe
    .SHELLFLAGS := /d /c
    SET_ENV := set
    UNSET := set
    NULL := nul
else
    detected_OS := $(shell uname -s)
    RM := rm -rf
    MKDIR := mkdir -p
    MKDIR_P = mkdir -p $(1)
    SET_ENV := export
    UNSET := unset
    NULL := /dev/null
endif

.PHONY: build test lua run debug release clean help system_info

# 通用变量
GOYACC := goyacc
GO := go
MAIN := cmd/milk/milk.go
PARSER := ./parse/parser.go
PARSER_SRC := ./parse/parser.go.y

# 获取操作系统和架构
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# 目标操作系统和架构（用于交叉编译）
TAROS ?= $(GOOS)
TARARCH ?= $(GOARCH)

# 输出目录
DEBUG_DIR := build/$(TAROS)/debug/$(TARARCH)
RELEASE_DIR := build/$(TAROS)/release/$(TARARCH)
DEBUG_BIN := $(DEBUG_DIR)/bin/milk$(if $(filter windows,$(TAROS)),.exe)
RELEASE_BIN := $(RELEASE_DIR)/bin/milk$(if $(filter windows,$(TAROS)),.exe)

# Debug 标志
DEBUG_FLAGS := -gcflags="all=-N -l" -ldflags="-X main.debug=true"

# Release 标志
RELEASE_FLAGS := -ldflags="-s -w" -trimpath

# 系统信息目标
system_info:
	@echo 当前系统信息:
	@echo 操作系统: $(shell go env GOOS)
	@echo 架构: $(shell go env GOARCH)
ifeq ($(detected_OS),Windows)
	@echo 内存: $(shell powershell -command "[math]::Round((Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory / 1GB, 2)" 2>$(NULL)) GB
	@echo CPU频率: $(shell powershell -command "Get-WmiObject Win32_Processor | Select-Object -ExpandProperty MaxClockSpeed" 2>$(NULL)) MHz
	@echo CPU核心数: $(shell powershell -command "Get-WmiObject Win32_Processor | Select-Object -ExpandProperty NumberOfCores" 2>$(NULL))
else ifeq ($(detected_OS),Darwin)
	@echo 内存: $(shell sysctl -n hw.memsize | awk '{print $$1/1024/1024/1024 " GB"}')
	@echo CPU频率: $(shell sysctl -n hw.cpufrequency | awk '{print $$1/1000000 " GHz"}')
	@echo CPU核心数: $(shell sysctl -n hw.ncpu)
else
	@echo 内存: $(shell free -h | awk '/^Mem:/{print $$2}')
	@echo CPU频率: $(shell lscpu | grep 'CPU MHz' | awk '{print $$3 " MHz"}')
	@echo CPU核心数: $(shell nproc)
endif
	@echo 目标平台: $(TAROS)/$(TARARCH)

# 清除 TAROS 和 TARARCH 变量的函数
define clear_vars
$(if $(filter Windows,$(detected_OS)),\
    $(shell set TAROS= & set TARARCH=),\
    $(shell $(UNSET) TAROS TARARCH))
endef

# 默认目标
all: release

# 生成解析器
$(PARSER): $(PARSER_SRC)
	$(GOYACC) -o $@ $<

# Release 构建
release: system_info $(PARSER)
	@echo 正在为 $(TAROS)/$(TARARCH) 构建发布版本...
	$(call MKDIR_P,$(RELEASE_DIR)/bin)
	$(GO) fmt .
ifeq ($(detected_OS),Windows)
	$(SET_ENV) GOOS=$(TAROS)&& $(SET_ENV) GOARCH=$(TARARCH)&& $(GO) build -o $(RELEASE_BIN) $(RELEASE_FLAGS) $(MAIN)
else
	GOOS=$(TAROS) GOARCH=$(TARARCH) $(GO) build -o $(RELEASE_BIN) $(RELEASE_FLAGS) $(MAIN)
endif
	@echo 构建完成: $(RELEASE_BIN)
	$(call clear_vars)
	@echo TAROS 和 TARARCH 变量已清除。

# Debug 构建
debug: system_info $(PARSER)
	@echo 正在为 $(TAROS)/$(TARARCH) 构建调试版本...
	$(call MKDIR_P,$(DEBUG_DIR)/bin)
	$(call MKDIR_P,$(DEBUG_DIR)/obj)
	$(GO) fmt .
ifeq ($(detected_OS),Windows)
	$(SET_ENV) GOOS=$(TAROS)&& $(SET_ENV) GOARCH=$(TARARCH)&& $(GO) build -o $(DEBUG_BIN) $(DEBUG_FLAGS) $(MAIN)
else
	GOOS=$(TAROS) GOARCH=$(TARARCH) $(GO) build -o $(DEBUG_BIN) $(DEBUG_FLAGS) $(MAIN)
endif
ifeq ($(TAROS),$(GOOS))
ifeq ($(TARARCH),$(GOARCH))
	$(GO) tool objdump $(DEBUG_BIN) > $(DEBUG_DIR)/obj/lua.s
	$(GO) tool nm $(DEBUG_BIN) > $(DEBUG_DIR)/obj/lua.o
endif
endif
	@echo 构建完成: $(DEBUG_BIN)
	$(call clear_vars)
	@echo TAROS 和 TARARCH 变量已清除。


# 运行 Lua
lua: release
ifeq ($(detected_OS),Windows)
	$(shell powershell -command "start '$(CURDIR)/$(RELEASE_BIN)'")
else
	$(RELEASE_BIN)
endif


# 构建（默认为 release）
build: release

# 运行
run: $(PARSER)
	$(GO) fmt .
	$(GO) run $(MAIN)

# 测试
test: $(PARSER)
	$(GO) fmt .
	$(GO) test

# 清理
clean:
ifeq ($(detected_OS),Windows)
	if exist build rmdir /S /Q build
	$(call clear_vars)
else
	rm -rf build
	$(call clear_vars)
endif
	@echo 清理完成: build 文件夹已删除。


# 显示帮助信息
help:
	@echo 可用的 make 目标:
	@echo   debug    - 构建调试版本
	@echo   release  - 构建发布版本
	@echo   lua      - 构建并运行 Lua（发布版本）
	@echo   run      - 直接运行（不生成可执行文件）
	@echo   test     - 运行测试
	@echo   clean    - 清理生成的文件
	@echo   help     - 显示此帮助信息
	@echo.
	@echo 可用的操作系统:
	@echo   linux   - Linux
	@echo   darwin  - macOS
	@echo   windows - Windows
	@echo.
	@echo 可用的架构:
	@echo   amd64   - x86-64
	@echo   386     - x86
	@echo   arm     - ARM
	@echo   arm64   - ARM64
	@echo.
	@echo 当前系统信息:
	@echo 操作系统: $(shell go env GOOS)
	@echo 架构: $(shell go env GOARCH)
	@echo 目标平台: $(TAROS)/$(TARARCH)
	@echo.
	@echo 交叉编译选项:
ifeq ($(detected_OS),Windows)
	@echo   Windows CMD: set "TAROS=^<os^>" ^& set "TARARCH=^<arch^>" ^& make ^<target^>
	@echo   Windows PowerShell: $$env:TAROS="^<os^>"; $$env:TARARCH="^<arch^>"; make ^<target^>
else
	@echo   TAROS=<os> TARARCH=<arch> make <target>
endif
	@echo 例如:
ifeq ($(detected_OS),Windows)
	@echo   Windows CMD: set "TAROS=linux" ^& set "TARARCH=amd64" ^& make release
	@echo   Windows PowerShell: $$env:TAROS="linux"; $$env:TARARCH="amd64"; make release
else
	@echo   TAROS=linux TARARCH=amd64 make release
endif