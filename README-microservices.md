# Infrastructure Training Backend - Microservices

Этот проект разделен на микросервисы с отдельными контейнерами для базы данных и приложения.

## Структура проекта

```
├── services/
│   ├── app/              # Микросервис приложения
│   │   ├── Dockerfile
│   │   ├── main.go
│   │   ├── database.go
│   │   ├── handlers.go
│   │   ├── middleware.go
│   │   ├── notes.go
│   │   ├── types.go
│   │   ├── utils.go
│   │   ├── migrate.go
│   │   ├── go.mod
│   │   ├── go.sum
│   │   └── migrations/
│   └── database/         # Микросервис базы данных
│       ├── Dockerfile
│       ├── init-db.sh
│       └── migrations/
├── docker-compose.yml
└── .env.example
```

## Быстрый старт

1. **Скопируйте файл конфигурации:**
   ```bash
   cp .env.example .env
   ```

2. **Запустите микросервисы:**
   ```bash
   docker-compose up --build
   ```

3. **Проверьте работу приложения:**
   ```bash
   curl http://localhost:8080/health
   ```

## Описание сервисов

### Database Service

- **Контейнер:** `infrastructure-db`
- **Порт:** 5432
- **База данных:** PostgreSQL 15
- **Автоматическая инициализация:** создание пользователя и схемы
- **Миграции:** применяются автоматически при старте

### Application Service

- **Контейнер:** `infrastructure-app`
- **Порт:** 8080
- **Зависимости:** ждет готовности базы данных
- **Health check:** `/health`

## API Endpoints

- `GET /health` - проверка состояния сервиса
- `GET /api/ping` - простой ping
- `GET /api/notes` - получение всех заметок
- `POST /api/notes` - создание новой заметки
- `DELETE /api/notes/{id}` - удаление заметки

## Команды для работы

```bash
# Запуск в фоновом режиме
docker-compose up -d

# Просмотр логов
docker-compose logs -f

# Остановка сервисов
docker-compose down

# Пересборка образов
docker-compose build

# Очистка данных (включая БД)
docker-compose down -v
```

## Переменные окружения

Основные переменные находятся в файле `.env`:

- `DB_HOST` - хост базы данных (database)
- `DB_PORT` - порт базы данных (5432)
- `DB_USER` - пользователь приложения (app_user)
- `DB_PASSWORD` - пароль пользователя
- `DB_NAME` - название базы данных
- `PORT` - порт приложения (8080)

## Мониторинг

```bash
# Статус контейнеров
docker-compose ps

# Использование ресурсов
docker stats

# Подключение к базе данных
docker-compose exec database psql -U postgres -d infrastructure_training
```

## Разработка

Для разработки вы можете:

1. Запустить только базу данных:
   ```bash
   docker-compose up database
   ```

2. Запустить приложение локально с переменными окружения для localhost:
   ```bash
   export DB_HOST=localhost
   go run .
   ```