.PHONY: build clean version cross-build lint fmt tidy help

# 获取当前git版本号
VERSION := $(shell git describe --tags 2>/dev/null || echo "v0.0.0")

run:
	@echo "Running..."
	go run cmd/main.go

# 默认构建目标
build:
	@echo "Building Go binary (version: ${VERSION}) with optimization flags..."
	go build -o textool.exe cmd/main.go
	@echo "Build done"

# 代码质量检查
lint:
	@if ! command -v golangci-lint >/dev/null; then \
		echo "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@golangci-lint run ./...

# 代码格式化检查
fmt:
	@echo "Checking code formatting..."
	go fmt ./... && go vet ./...
	@echo "Code is properly formatted"; \


# 依赖管理
tidy:
	@echo "Tidying module dependencies..."
	go mod tidy -v

