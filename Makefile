.PHONY: run build clean test dev

# Запуск приложения
run:
	go run .

# Запуск в режиме разработки (с перезапуском при изменениях)
dev:
	go run .

# Сборка приложения
build:
	go build -o bin/app .

# Очистка собранных файлов
clean:
	rm -rf bin/

# Запуск тестов
test:
	go test ./...

# Проверка форматирования и линтинг
lint:
	go fmt ./...
	go vet ./...

# Установка зависимостей
deps:
	go mod tidy
	go mod download

# Запуск приложения в production режиме
start:
	./bin/app

# Полная сборка с проверками
build-prod: clean lint test build

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  run        - Запуск приложения"
	@echo "  dev        - Запуск в режиме разработки"
	@echo "  build      - Сборка приложения"
	@echo "  clean      - Очистка собранных файлов"
	@echo "  test       - Запуск тестов"
	@echo "  lint       - Проверка форматирования и линтинг"
	@echo "  deps       - Установка зависимостей"
	@echo "  start      - Запуск собранного приложения"
	@echo "  build-prod - Полная сборка с проверками"
	@echo "  help       - Показать эту справку"