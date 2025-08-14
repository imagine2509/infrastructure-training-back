FROM postgres:15-alpine

# Создание директорий для инициализации
RUN mkdir -p /docker-entrypoint-initdb.d

# Копирование миграций и скриптов инициализации
COPY migrations/ /docker-entrypoint-initdb.d/
COPY init-db.sh /docker-entrypoint-initdb.d/01-init.sh

# Установка прав на выполнение
RUN chmod +x /docker-entrypoint-initdb.d/01-init.sh

# Установка переменных окружения по умолчанию
ENV POSTGRES_DB=infrastructure_training
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=postgres

# Экспорт порта
EXPOSE 5432