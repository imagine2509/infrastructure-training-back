FROM golang:1.24-alpine AS builder

# Установка зависимостей для сборки
RUN apk add --no-cache git

# Создание рабочей директории
WORKDIR /app

# Копирование файлов модуля и загрузка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Финальный образ
FROM postgres:15-alpine

# Установка зависимостей и утилит
RUN apk add --no-cache supervisor bash

# Создание директорий
RUN mkdir -p /app /var/log/supervisor

# Копирование скомпилированного приложения
COPY --from=builder /app/main /app/
COPY --from=builder /app/migrations /app/migrations/

# Копирование конфигурации supervisor
COPY <<EOF /etc/supervisor/conf.d/supervisord.conf
[supervisord]
nodaemon=true
user=root
logfile=/var/log/supervisor/supervisord.log
pidfile=/var/run/supervisord.pid

[program:postgresql]
command=/usr/local/bin/docker-entrypoint.sh postgres
user=postgres
autorestart=true
redirect_stderr=true
stdout_logfile=/var/log/supervisor/postgresql.log
priority=1

[program:app]
command=/app/main
directory=/app
user=root
autorestart=true
redirect_stderr=true
stdout_logfile=/var/log/supervisor/app.log
priority=2
startretries=10
startsecs=5
EOF

# Создание скрипта инициализации
COPY <<EOF /docker-entrypoint-initdb.d/01-init.sql
CREATE DATABASE infrastructure_training;
CREATE USER app_user WITH PASSWORD 'app_password';
GRANT ALL PRIVILEGES ON DATABASE infrastructure_training TO app_user;
\c infrastructure_training;
GRANT ALL ON SCHEMA public TO app_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO app_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO app_user;
EOF

# Создание скрипта запуска
COPY <<EOF /start.sh
#!/bin/bash
set -e

# Инициализация PostgreSQL если данные не существуют
if [ ! -s "$PGDATA/PG_VERSION" ]; then
    echo "Initializing PostgreSQL database..."
    su-exec postgres initdb
    su-exec postgres pg_ctl -D "$PGDATA" -o "-c listen_addresses='*'" -w start
    
    # Выполнение инициализационных скриптов
    for f in /docker-entrypoint-initdb.d/*; do
        case "$f" in
            *.sh)     echo "$0: running $f"; . "$f" ;;
            *.sql)    echo "$0: running $f"; su-exec postgres psql < "$f"; echo ;;
            *.sql.gz) echo "$0: running $f"; gunzip -c "$f" | su-exec postgres psql; echo ;;
            *)        echo "$0: ignoring $f" ;;
        esac
        echo
    done
    
    su-exec postgres pg_ctl -D "$PGDATA" -m fast -w stop
fi

# Запуск supervisor
exec /usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf
EOF

RUN chmod +x /start.sh

# Установка переменных окружения
ENV POSTGRES_DB=infrastructure_training
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=postgres
ENV DB_HOST=localhost
ENV DB_PORT=5432
ENV DB_USER=app_user
ENV DB_PASSWORD=app_password
ENV DB_NAME=infrastructure_training
ENV PORT=8080

# Экспорт портов
EXPOSE 5432 8080

# Запуск
CMD ["/start.sh"]