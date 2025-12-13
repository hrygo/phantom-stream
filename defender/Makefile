.PHONY: help all build install clean \
	test test-unit test-integration test-race test-all test-coverage bench \
	lint fmt vet check \
	deps tidy dev-deps \
	sign verify

# 尝试加载 .env 文件 (如果存在)
-include .env

# --- 全局变量 ---
BINARY_NAME := phantom-guard
BUILD_DIR := ./bin
GO := go
GOFLAGS := -v
LDFLAGS := -ldflags "-s -w"

# 测试配置
TEST_TIMEOUT := 30s
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html
INTEGRATION_TAG := integration

# 默认测试数据
# 默认测试数据
# 优先使用环境变量 DEFAULT_KEY，否则使用硬编码默认值
DEFAULT_KEY ?= 12345678901234567890123456789012
TEST_PDF := testdata/2511.17467v2.pdf

# 颜色定义
COLOR_RESET   := \033[0m
COLOR_BOLD    := \033[1m
COLOR_RED     := \033[31m
COLOR_GREEN   := \033[32m
COLOR_YELLOW  := \033[33m
COLOR_BLUE    := \033[34m
COLOR_PURPLE  := \033[35m
COLOR_CYAN    := \033[36m
COLOR_WHITE   := \033[37m

# 默认目标
.DEFAULT_GOAL := help

##@ General

help: ## 显示帮助信息
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Defender - PhantomGuard 开发构建工具$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)使用说明:$(COLOR_RESET)"
	@echo "  make [command] [VARIABLE=value]"
	@echo ""
	@echo "$(COLOR_BOLD)可用命令:$(COLOR_RESET)"
	@awk 'BEGIN {FS = ":.*##"; printf ""} \
		/^##@/ { printf "\n$(COLOR_PURPLE)%s$(COLOR_RESET)\n", substr($$0, 5) } \
		/^[a-zA-Z_-]+:.*?##/ { printf "  $(COLOR_CYAN)%-20s$(COLOR_RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""

all: clean deps check build ## 执行完整 CI 流程 (清理 -> 依赖 -> 检查 -> 编译)

clean: ## 清理构建产物、缓存和临时文件
	@echo "$(COLOR_YELLOW)🧹 清理环境...$(COLOR_RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@rm -f *_signed.pdf *_temp*.pdf
	@$(GO) clean -cache -testcache
	@echo "$(COLOR_GREEN)✅ 清理完成$(COLOR_RESET)"

##@ Build & Install

build: ## 编译发布版本 (stripped & optimized)
	@echo "$(COLOR_YELLOW)🔨 编译发布版本...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(COLOR_GREEN)✅ 编译成功: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

build-dev: ## 编译开发版本 (保留调试符号)
	@echo "$(COLOR_YELLOW)🔨 编译开发版本...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(COLOR_GREEN)✅ 开发版本编译成功$(COLOR_RESET)"

install: build ## 安装到系统路径 (/usr/local/bin)
	@echo "$(COLOR_YELLOW)📦 安装到系统...$(COLOR_RESET)"
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(COLOR_GREEN)✅ 安装完成$(COLOR_RESET)"

##@ Development & Testing

test: test-unit ## 运行标准单元测试

test-unit: ## 运行单元测试 (排除集成测试)
	@echo "$(COLOR_YELLOW)🧪 运行单元测试...$(COLOR_RESET)"
	@$(GO) test ./... -v -timeout $(TEST_TIMEOUT) -short

test-integration: ## 运行集成测试 (需环境支持)
	@echo "$(COLOR_YELLOW)🧪 运行集成测试...$(COLOR_RESET)"
	@if [ ! -f $(TEST_PDF) ]; then \
		echo "$(COLOR_RED)⚠️  警告: 测试文件 $(TEST_PDF) 不存在$(COLOR_RESET)"; \
	fi
	@$(GO) test ./... -v -tags=$(INTEGRATION_TAG) -timeout 2m

test-race: ## 运行竞态检测
	@echo "$(COLOR_YELLOW)🏃 运行竞态检测...$(COLOR_RESET)"
	@$(GO) test ./... -race -v -timeout $(TEST_TIMEOUT)

test-all: test-unit test-integration ## 运行所有测试 (单元 + 集成)

test-coverage: ## 生成并查看测试覆盖率报告
	@echo "$(COLOR_YELLOW)📊 生成覆盖率报告...$(COLOR_RESET)"
	@$(GO) test ./... -tags=$(INTEGRATION_TAG) -coverprofile=$(COVERAGE_FILE) -covermode=atomic -timeout 2m
	@$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(COLOR_GREEN)✅ 报告已生成: $(COVERAGE_HTML)$(COLOR_RESET)"
	@$(GO) tool cover -func=$(COVERAGE_FILE) | tail -n 1

bench: ## 运行基准测试
	@echo "$(COLOR_YELLOW)⚡ 运行基准测试...$(COLOR_RESET)"
	@$(GO) test ./... -tags=$(INTEGRATION_TAG) -bench=. -benchmem -run=^$$

##@ Quality Control

lint: ## 运行代码检查 (golangci-lint)
	@echo "$(COLOR_YELLOW)🔍 运行代码检查...$(COLOR_RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(COLOR_YELLOW)⚠️  警告: golangci-lint 未安装, 跳过 lint 检查$(COLOR_RESET)"; \
		echo "   提示: 运行 'make dev-deps' 安装开发工具"; \
	fi

fmt: ## 格式化代码 (go fmt)
	@echo "$(COLOR_YELLOW)✨ 格式化代码...$(COLOR_RESET)"
	@$(GO) fmt ./...

vet: ## 静态分析 (go vet)
	@echo "$(COLOR_YELLOW)🔍 运行静态分析...$(COLOR_RESET)"
	@$(GO) vet ./...

check: fmt vet test-unit ## 运行完整质量检查流程 (fmt -> vet -> test)
	@echo "$(COLOR_GREEN)✅ 基础检查通过$(COLOR_RESET)"

check-full: fmt vet lint test-unit ## 运行完整质量检查流程 (包括 lint)
	@echo "$(COLOR_GREEN)✅ 所有检查通过$(COLOR_RESET)"

##@ Dependency Management

deps: ## 下载项目依赖
	@echo "$(COLOR_YELLOW)📥 下载依赖...$(COLOR_RESET)"
	@$(GO) mod download

tidy: ## 整理模块依赖 (go mod tidy)
	@echo "$(COLOR_YELLOW)🔧 整理依赖...$(COLOR_RESET)"
	@$(GO) mod tidy

dev-deps: ## 安装开发工具 (golangci-lint)
	@echo "$(COLOR_YELLOW)📥 安装开发依赖...$(COLOR_RESET)"
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(COLOR_GREEN)✅ 安装完成$(COLOR_RESET)"

##@ Operations (Running)

sign: ## 运行签名工具 (默认使用测试文件)
	@echo "$(COLOR_YELLOW)🛡️  运行签名...$(COLOR_RESET)"
	@$(eval FILE := $(if $(FILE),$(FILE),$(TEST_PDF)))
	@$(eval KEY := $(if $(KEY),$(KEY),$(DEFAULT_KEY)))
	@$(eval MSG := $(if $(MSG),$(MSG),Protected-$(shell date +%s)))
	@echo "   📂 文件: $(FILE)"
	@echo "   🔑 密钥: $(KEY)"
	@echo "   📝 消息: $(MSG)"
	@if [ ! -f "$(FILE)" ]; then echo "$(COLOR_RED)❌ 错误: 文件 $(FILE) 不存在$(COLOR_RESET)"; exit 1; fi
	@$(GO) run . sign -f $(FILE) -m "$(MSG)" -k "$(KEY)"

verify: ## 运行验证工具 (默认验证生成的测试文件)
	@echo "$(COLOR_YELLOW)🔍 运行验证...$(COLOR_RESET)"
	@$(eval TARGET_FILE := $(if $(FILE),$(FILE),$(patsubst %.pdf,%_signed.pdf,$(TEST_PDF))))
	@$(eval KEY := $(if $(KEY),$(KEY),$(DEFAULT_KEY)))
	@echo "   📂 文件: $(TARGET_FILE)"
	@echo "   🔑 密钥: $(KEY)"
	@if [ ! -f "$(TARGET_FILE)" ]; then echo "$(COLOR_RED)❌ 错误: 文件 $(TARGET_FILE) 不存在. 请先运行 'make sign'$(COLOR_RESET)"; exit 1; fi
	@$(GO) run . verify -f $(TARGET_FILE) -k "$(KEY)"

gen-key: ## 生成随机 32 位密钥并写入 .env 文件
	@echo "$(COLOR_YELLOW)🔑 生成新密钥...$(COLOR_RESET)"
	@# Generate 32-char alphanumeric key
	$(eval NEW_KEY := $(shell openssl rand -base64 128 | tr -dc 'a-zA-Z0-9' | head -c 32))
	@touch .env
	@if grep -q "DEFAULT_KEY=" .env; then \
		echo "$(COLOR_CYAN)ℹ️  发现旧密钥: $$(grep "DEFAULT_KEY=" .env | cut -d '=' -f2)$(COLOR_RESET)"; \
		sed -i '' 's/DEFAULT_KEY=.*/DEFAULT_KEY=$(NEW_KEY)/' .env; \
		echo "$(COLOR_GREEN)✅ 密钥已更新: $(NEW_KEY)$(COLOR_RESET)"; \
	else \
		echo "DEFAULT_KEY=$(NEW_KEY)" >> .env; \
		echo "$(COLOR_GREEN)✅ 密钥已写入 .env: $(NEW_KEY)$(COLOR_RESET)"; \
	fi
	@echo "$(COLOR_CYAN)💡 提示: Make 命令会自动读取 .env，无需手动 export$(COLOR_RESET)"