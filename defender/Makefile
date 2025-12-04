.PHONY: help build clean test test-verbose test-integration bench-integration run-sign run-verify install lint fmt vet check dev-deps deliver-red notify-red

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
BINARY_NAME=defender
BUILD_DIR=./bin
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"

# 红方交付配置
SOURCE_PDF=../docs/2511.17467v2.pdf
OUTPUT_DIR=../docs
DEFAULT_KEY=12345678901234567890123456789012
ROUND?=R3

# 可覆盖的密钥（默认使用 DEFAULT_KEY）
KEY?=$(DEFAULT_KEY)

# 路径派生变量（基于 SOURCE_PDF 派生签名文件路径）
SRC_DIR := $(dir $(SOURCE_PDF))
SRC_BASE := $(notdir $(SOURCE_PDF))
SRC_NAME := $(basename $(SRC_BASE))
SIGNED_FILE := $(SRC_DIR)$(SRC_NAME)_$(ROUND)_signed.pdf

# 输出目录默认与源文件目录一致（如需另存，可覆盖 OUTPUT_DIR）
OUTPUT_DIR?=$(SRC_DIR)

# 颜色输出
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

help: ## 显示帮助信息
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Defender - 可用命令:$(COLOR_RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-15s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""

build: ## 编译项目
	@echo "$(COLOR_YELLOW)🔨 编译 $(BINARY_NAME)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(COLOR_GREEN)✅ 编译完成: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

build-dev: ## 编译开发版本（包含调试信息）
	@echo "$(COLOR_YELLOW)🔨 编译开发版本...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(COLOR_GREEN)✅ 开发版本编译完成$(COLOR_RESET)"

clean: ## 清理编译产物和缓存
	@echo "$(COLOR_YELLOW)🧹 清理编译产物...$(COLOR_RESET)"
	@rm -rf $(BUILD_DIR)
	@$(GO) clean -cache -testcache -modcache
	@echo "$(COLOR_GREEN)✅ 清理完成$(COLOR_RESET)"

test: ## 运行测试
	@echo "$(COLOR_YELLOW)🧪 运行测试...$(COLOR_RESET)"
	@$(GO) test ./... -cover
	@echo "$(COLOR_GREEN)✅ 测试完成$(COLOR_RESET)"

test-verbose: ## 运行详细测试
	@echo "$(COLOR_YELLOW)🧪 运行详细测试...$(COLOR_RESET)"
	@$(GO) test ./... -v -cover -coverprofile=coverage.out
	@echo "$(COLOR_GREEN)✅ 测试完成$(COLOR_RESET)"

# 集成测试（需要启用 build tag）
test-integration: ## 运行集成测试（需启用 build tag）
	@echo "$(COLOR_YELLOW)🧪 运行集成测试...$(COLOR_RESET)"
	@$(GO) test ./... -tags=integration -v
	@echo "$(COLOR_GREEN)✅ 集成测试完成$(COLOR_RESET)"

bench-integration: ## 运行集成基准测试（需启用 build tag）
	@echo "$(COLOR_YELLOW)⚡ 运行集成基准测试...$(COLOR_RESET)"
	@$(GO) test ./... -tags=integration -bench=. -benchmem
	@echo "$(COLOR_GREEN)✅ 集成基准测试完成$(COLOR_RESET)"

test-coverage: ## 生成测试覆盖率报告
	@echo "$(COLOR_YELLOW)📊 生成覆盖率报告...$(COLOR_RESET)"
	@$(GO) test ./... -coverprofile=coverage.out
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✅ 覆盖率报告已生成: coverage.html$(COLOR_RESET)"

bench: ## 运行基准测试
	@echo "$(COLOR_YELLOW)⚡ 运行基准测试...$(COLOR_RESET)"
	@$(GO) test ./... -bench=. -benchmem

run-sign: ## 运行签名命令示例 (需要设置 FILE, MSG, KEY 变量)
	@echo "$(COLOR_YELLOW)🛡️  执行签名操作...$(COLOR_RESET)"
	@if [ -z "$(FILE)" ] || [ -z "$(MSG)" ] || [ -z "$(KEY)" ]; then \
		echo "$(COLOR_BOLD)使用方式:$(COLOR_RESET) make run-sign FILE=path/to/file.pdf MSG='UserID:123' KEY='your-32-byte-key-here!!!!!!!!' [ROUND=11]"; \
		exit 1; \
	fi
	@$(GO) run . sign -f $(FILE) -m "$(MSG)" -k "$(KEY)" $(if $(ROUND),-r $(ROUND),)

run-verify: ## 运行验证命令示例 (需要设置 FILE, KEY 变量)
	@echo "$(COLOR_YELLOW)🔍 执行验证操作...$(COLOR_RESET)"
	@if [ -z "$(FILE)" ] || [ -z "$(KEY)" ]; then \
		echo "$(COLOR_BOLD)使用方式:$(COLOR_RESET) make run-verify FILE=path/to/file_signed.pdf KEY='your-32-byte-key-here!!!!!!!!'"; \
		exit 1; \
	fi
	@$(GO) run . verify -f $(FILE) -k "$(KEY)"

install: build ## 安装到系统路径
	@echo "$(COLOR_YELLOW)📦 安装 $(BINARY_NAME)...$(COLOR_RESET)"
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(COLOR_GREEN)✅ 已安装到 /usr/local/bin/$(BINARY_NAME)$(COLOR_RESET)"

lint: ## 运行 golangci-lint 代码检查
	@echo "$(COLOR_YELLOW)🔍 运行代码检查...$(COLOR_RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(COLOR_YELLOW)⚠️  golangci-lint 未安装，使用 'make dev-deps' 安装$(COLOR_RESET)"; \
	fi

fmt: ## 格式化代码
	@echo "$(COLOR_YELLOW)✨ 格式化代码...$(COLOR_RESET)"
	@$(GO) fmt ./...
	@echo "$(COLOR_GREEN)✅ 代码格式化完成$(COLOR_RESET)"

vet: ## 运行 go vet 检查
	@echo "$(COLOR_YELLOW)🔍 运行 vet 检查...$(COLOR_RESET)"
	@$(GO) vet ./...
	@echo "$(COLOR_GREEN)✅ Vet 检查完成$(COLOR_RESET)"

check: fmt vet test ## 运行所有检查（格式化、vet、测试）
	@echo "$(COLOR_GREEN)✅ 所有检查通过$(COLOR_RESET)"

deps: ## 下载依赖
	@echo "$(COLOR_YELLOW)📥 下载依赖...$(COLOR_RESET)"
	@$(GO) mod download
	@echo "$(COLOR_GREEN)✅ 依赖下载完成$(COLOR_RESET)"

tidy: ## 整理依赖
	@echo "$(COLOR_YELLOW)🔧 整理依赖...$(COLOR_RESET)"
	@$(GO) mod tidy
	@echo "$(COLOR_GREEN)✅ 依赖整理完成$(COLOR_RESET)"

dev-deps: ## 安装开发依赖工具
	@echo "$(COLOR_YELLOW)📥 安装开发工具...$(COLOR_RESET)"
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(COLOR_GREEN)✅ 开发工具安装完成$(COLOR_RESET)"

quick-test: ## 快速测试（使用示例文件）
	@echo "$(COLOR_YELLOW)⚡ 快速测试...$(COLOR_RESET)"
	@if [ ! -f testdata/2511.17467v2.pdf ]; then \
		echo "$(COLOR_BOLD)错误:$(COLOR_RESET) 测试文件不存在"; \
		exit 1; \
	fi
	@echo "$(COLOR_BLUE)1. 签名测试...$(COLOR_RESET)"
	@$(GO) run . sign -f testdata/2511.17467v2.pdf -m "TEST:QuickTest" -k "TestKey12345678901234567890123" -r "test"
	@echo "$(COLOR_BLUE)2. 验证测试...$(COLOR_RESET)"
	@$(GO) run . verify -f testdata/2511.17467v2_test_signed.pdf -k "TestKey12345678901234567890123"
	@echo "$(COLOR_GREEN)✅ 快速测试完成$(COLOR_RESET)"

all: clean deps build test ## 完整构建流程（清理、下载依赖、编译、测试）
	@echo "$(COLOR_GREEN)✅ 完整构建完成$(COLOR_RESET)"

deliver-red: build ## 生成红方交付文件 (可通过 ROUND=R3 指定轮次)
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)        🛡️  PhantomStream 攻防演习 - 蓝方交付文件生成$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_YELLOW)📋 配置信息:$(COLOR_RESET)"
	@echo "   轮次: $(ROUND)"
	@echo "   源文件: $(SOURCE_PDF)"
	@echo "   消息: Defender:Phase7-$(ROUND)"
	@echo ""
	@echo "$(COLOR_YELLOW)🔨 步骤 1/3: 生成签名文件...$(COLOR_RESET)"
	@$(BUILD_DIR)/$(BINARY_NAME) sign -f $(SOURCE_PDF) -m "Defender:Phase7-$(ROUND)" -k "$(KEY)" -r $(ROUND)
	@echo ""
	@echo "$(COLOR_YELLOW)🔍 步骤 2/3: 验证签名...$(COLOR_RESET)"
	@$(BUILD_DIR)/$(BINARY_NAME) verify -f $(SIGNED_FILE) -k "$(KEY)"
	@echo ""
	@echo "$(COLOR_YELLOW)📊 步骤 3/3: 文件信息...$(COLOR_RESET)"
	@ls -lh $(SIGNED_FILE) | awk '{print "   文件大小:", $$5}'
	@echo ""
	@echo "$(COLOR_GREEN)✅ 交付文件生成完成!$(COLOR_RESET)"
	@echo "$(COLOR_GREEN)📁 文件路径: $(SIGNED_FILE)$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)💡 下一步: 运行 'make notify-red ROUND=$(ROUND)' 发送通知消息$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"

notify-red: ## 显示红方通知消息 (可通过 ROUND=R3 指定轮次)
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)╔══════════════════════════════════════════════════════════════════════════════╗$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)║                    📨 PhantomStream 攻防演习 - 蓝队通知                      ║$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)╚══════════════════════════════════════════════════════════════════════════════╝$(COLOR_RESET)"
	@echo ""
	@echo "致：红方（Attacker）"
	@echo "发自：蓝方（Defender）"
	@date '+时间：%Y-%m-%d %H:%M:%S'
	@echo "状态：✅ 防御升级完成"
	@echo ""
	@echo "$(COLOR_BOLD)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)## 📋 本轮交付信息$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)轮次$(COLOR_RESET)：$(ROUND)"
	@echo "$(COLOR_BOLD)交付文件$(COLOR_RESET)：$(SIGNED_FILE)"
	@ls -lh $(SIGNED_FILE) 2>/dev/null | awk '{print "$(COLOR_BOLD)文件大小$(COLOR_RESET)：", $$5}' || echo "$(COLOR_BOLD)文件大小$(COLOR_RESET)：文件未生成，请先运行 'make deliver-red'"
	@echo "$(COLOR_BOLD)状态$(COLOR_RESET)：已完成防护处理"
	@echo ""
	@echo "$(COLOR_BOLD)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)## 🛡️  技术对抗$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)本方措施$(COLOR_RESET)："
	@echo "- 已对文件进行增强防护处理"
	@echo "- 采用符合 PDF 标准的技术方案"
	@echo "- 文件结构完全合法，可被所有阅读器正常打开"
	@echo ""
	@echo "$(COLOR_BOLD)验证方式$(COLOR_RESET)："
	@echo "- 样本文件已就绪，请自行测试"
	@echo "- 文件路径：docs/2511.17467v2_$(ROUND)_signed.pdf"
	@echo ""
	@echo "$(COLOR_BOLD)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)## 📊 技术要求$(COLOR_RESET)"
	@echo ""
	@echo "请红方完成以下任务："
	@echo "1. ✅ 确保 PDF 文件结构完全合法"
	@echo "2. ✅ 确保所有 PDF 阅读器均可正常打开"
	@echo "3. ✅ 确保页面内容和视觉效果无异常"
	@echo "4. ✅ 尝试清除本方防护机制"
	@echo ""
	@echo "$(COLOR_BOLD)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)## 🔍 等待反馈$(COLOR_RESET)"
	@echo ""
	@echo "请红方对本轮样本进行测试，并提供处理结果。"
	@echo ""
	@echo "期待您的挑战！🎯"
	@echo ""
	@echo "$(COLOR_BOLD)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo ""
	@echo "                              $(COLOR_BOLD)蓝方 (Defender)$(COLOR_RESET)"
	@echo "                           PhantomStream 项目组"
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)╚══════════════════════════════════════════════════════════════════════════════╝$(COLOR_RESET)"
	@echo ""

deliver-and-notify: deliver-red notify-red ## 一键生成交付文件并显示通知消息
	@echo ""
	@echo "$(COLOR_GREEN)✅ 所有步骤完成！$(COLOR_RESET)"

.PHONY: build-dev test-coverage bench deps tidy quick-test all deliver-and-notify
