.PHONY: build run clean test help

# 默认目标
.DEFAULT_GOAL := help

# 二进制文件名
BINARY_NAME=pixelhub
BINARY_PATH=./bin/$(BINARY_NAME)

# Go 相关变量
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# 主程序路径
MAIN_PATH=./cmd/server

## help: 显示帮助信息
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: 编译项目
build:
	@echo "Building..."
	@mkdir -p $(GOBIN)
	$(GOBUILD) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

## run: 运行项目
run:
	@echo "Running PixelHub..."
	$(GOCMD) run $(MAIN_PATH)/main.go

## clean: 清理构建文件
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(GOBIN)
	@rm -rf data/*.db
	@echo "Clean complete"

## test: 运行测试
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## deps: 下载依赖
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies downloaded"

## dev: 开发模式运行（热重载需要安装 air）
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Run: go install github.com/cosmtrek/air@latest"; \
		echo "Running without hot reload..."; \
		$(MAKE) run; \
	fi

## install: 安装到系统
install: build
	@echo "Installing..."
	@cp $(BINARY_PATH) /usr/local/bin/$(BINARY_NAME)
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

## uninstall: 从系统卸载
uninstall:
	@echo "Uninstalling..."
	@rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstalled"

## setup: 初始化项目（创建配置文件）
setup:
	@if [ ! -f config.toml ]; then \
		echo "Creating config.toml from template..."; \
		cp config.example.toml config.toml; \
		echo "Please edit config.toml with your settings"; \
	else \
		echo "config.toml already exists"; \
	fi
	@mkdir -p data
	@echo "Setup complete"

