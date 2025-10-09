# --- config ---
GO       ?= go
APP      := mini-redis-go
CMD_DIR  := ./cmd/miniredis
PKG      := ./...
ADDR     ?= :6381

# --- meta ---
.DEFAULT_GOAL := help

help: ## Показать доступные цели
	@awk 'BEGIN{FS":.*?## "}/^[a-zA-Z0-9_\-]+:.*?## /{printf "  \033[36m%-14s\033[0m %s\n",$$1,$$2}' $(MAKEFILE_LIST)

# --- dev ---
fmt: ## go fmt
	$(GO) fmt $(PKG)

vet: ## go vet
	$(GO) vet $(PKG)

lint: ## golangci-lint (если установлен)
	@golangci-lint run ./... || echo "tip: install golangci-lint or skip this target"

# --- build & run ---
build: ## Собрать бинарник
	$(GO) build -o bin/$(APP) $(CMD_DIR)

run: ## Запустить сервер (ADDR=:6379 по умолчанию)
	ADDR=$(ADDR) $(GO) run $(CMD_DIR)

# --- tests ---
test: ## Юнит-тесты internal/*
	$(GO) test ./internal/... -v

test-e2e: ## E2E-тесты tests/*
	$(GO) test ./tests -v

race: ## Все тесты с -race
	$(GO) test $(PKG) -race -v

ci: fmt vet test race ## Набор команд для CI

clean: ## Удалить сборки
	rm -rf bin

.PHONY: help fmt vet lint build run test test-e2e race ci clean
