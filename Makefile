# Go Agentic Patterns - Makefile

.PHONY: help run-chain run-routing run-parallelization run-reflection run-tools run-planning tidy

help: ## Mevcut komutları listeler
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run-chain: ## Prompt Chaining desenini çalıştırır
	@echo "🚀 Prompt Chaining (Blog Yazarı) başlatılılyor..."
	@go run 01-prompt-chaining-pattern/main.go

run-routing: ## Routing desenini çalıştırır
	@echo "🚀 Routing (Müşteri Destek) başlatılıyor."
	@go run 02-routing-pattern/main.go

run-parallelization: ## Paralellization desenini çalıştırır
	@echo "🚀 Parallelization başlatılıyor..."
	@go run 03-parallelization-pattern/main.go

run-reflection: ## Reflection desenini çalıştırır
	@echo "🚀 Reflection (Kendini Düzeltme) başlatılıyor..."
	@go run 04-reflection-pattern/main.go

run-tools: ## Tool Use desenini çalıştırır
	@echo "🚀 Tool Use başlatılıyor..."
	@ro run 05-tool-use-pattern/main.go

run-planning:
	@echo "🚀 Planning (Planlayıcı) desenini çalıştırır"
	@go run 06-planning-pattern/main.go

tidy: ## Go modüllerini temizler ve indirir
	@go mod tidy