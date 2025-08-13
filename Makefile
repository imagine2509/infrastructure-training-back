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

# Инициализация базы данных
db-init:
	./scripts/init-db.sh

# Миграции базы данных
migrate-up:
	cd cmd/migrate && go run . -action=up

migrate-status:
	cd cmd/migrate && go run . -action=status

migrate-build:
	cd cmd/migrate && go build -o ../../bin/migrate .

# Создание новой миграции
migrate-create:
	@read -p "Введите название миграции: " name; \
	timestamp=$$(date +%Y%m%d%H%M%S); \
	filename="migrations/$${timestamp}_$${name}.sql"; \
	touch $$filename; \
	echo "Создана миграция: $$filename"

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  run          - Запуск приложения"
	@echo "  dev          - Запуск в режиме разработки"
	@echo "  build        - Сборка приложения"
	@echo "  clean        - Очистка собранных файлов"
	@echo "  test         - Запуск тестов"
	@echo "  lint         - Проверка форматирования и линтинг"
	@echo "  deps         - Установка зависимостей"
	@echo "  start        - Запуск собранного приложения"
	@echo "  build-prod   - Полная сборка с проверками"
	@echo "  db-init      - Инициализация базы данных"
	@echo "  migrate-up   - Запуск миграций"
	@echo "  migrate-status - Статус миграций"
	@echo "  migrate-build - Сборка утилиты миграций"
	@echo "  migrate-create - Создание новой миграции"
	@echo "  help         - Показать эту справку"