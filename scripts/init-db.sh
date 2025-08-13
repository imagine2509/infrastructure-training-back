#!/bin/bash

# Скрипт для инициализации базы данных PostgreSQL

set -e

# Загружаем переменные окружения из .env если файл существует
if [ -f .env ]; then
    echo "Loading environment variables from .env file..."
    export $(cat .env | grep -v '#' | awk '/=/ {print $1}')
fi

# Установка значений по умолчанию
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-infrastructure_training}
DB_ADMIN_USER=${DB_ADMIN_USER:-postgres}
DB_ADMIN_PASSWORD=${DB_ADMIN_PASSWORD:-postgres}

echo "Initializing database with the following settings:"
echo "Host: $DB_HOST"
echo "Port: $DB_PORT"
echo "Database: $DB_NAME"
echo "User: $DB_USER"
echo ""

# Функция для выполнения SQL команд от имени администратора
execute_sql() {
    PGPASSWORD=$DB_ADMIN_PASSWORD /Library/PostgreSQL/17/bin/psql -h $DB_HOST -p $DB_PORT -U $DB_ADMIN_USER -d postgres -c "$1"
}

# Функция для проверки существования базы данных
database_exists() {
    PGPASSWORD=$DB_ADMIN_PASSWORD /Library/PostgreSQL/17/bin/psql -h $DB_HOST -p $DB_PORT -U $DB_ADMIN_USER -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'" | grep -q 1
}

# Функция для проверки существования пользователя
user_exists() {
    PGPASSWORD=$DB_ADMIN_PASSWORD /Library/PostgreSQL/17/bin/psql -h $DB_HOST -p $DB_PORT -U $DB_ADMIN_USER -d postgres -tAc "SELECT 1 FROM pg_user WHERE usename='$DB_USER'" | grep -q 1
}

# Проверка соединения с PostgreSQL
echo "Checking PostgreSQL connection..."
if ! PGPASSWORD=$DB_ADMIN_PASSWORD /Library/PostgreSQL/17/bin/psql -h $DB_HOST -p $DB_PORT -U $DB_ADMIN_USER -d postgres -c '\q' 2>/dev/null; then
    echo "Error: Cannot connect to PostgreSQL server"
    echo "Please make sure PostgreSQL is running and credentials are correct"
    exit 1
fi

echo "✓ PostgreSQL connection successful"

# Создание пользователя если он не существует
if user_exists; then
    echo "✓ User '$DB_USER' already exists"
else
    echo "Creating user '$DB_USER'..."
    execute_sql "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';"
    echo "✓ User '$DB_USER' created successfully"
fi

# Создание базы данных если она не существует
if database_exists; then
    echo "✓ Database '$DB_NAME' already exists"
else
    echo "Creating database '$DB_NAME'..."
    execute_sql "CREATE DATABASE $DB_NAME OWNER $DB_USER;"
    echo "✓ Database '$DB_NAME' created successfully"
fi

# Предоставление прав пользователю
echo "Granting privileges to user '$DB_USER'..."
execute_sql "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"
execute_sql "ALTER USER $DB_USER CREATEDB;"

# Подключение к созданной базе и предоставление прав на схему public
PGPASSWORD=$DB_ADMIN_PASSWORD /Library/PostgreSQL/17/bin/psql -h $DB_HOST -p $DB_PORT -U $DB_ADMIN_USER -d $DB_NAME -c "GRANT ALL ON SCHEMA public TO $DB_USER;"
PGPASSWORD=$DB_ADMIN_PASSWORD /Library/PostgreSQL/17/bin/psql -h $DB_HOST -p $DB_PORT -U $DB_ADMIN_USER -d $DB_NAME -c "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USER;"
PGPASSWORD=$DB_ADMIN_PASSWORD /Library/PostgreSQL/17/bin/psql -h $DB_HOST -p $DB_PORT -U $DB_ADMIN_USER -d $DB_NAME -c "GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $DB_USER;"
PGPASSWORD=$DB_ADMIN_PASSWORD /Library/PostgreSQL/17/bin/psql -h $DB_HOST -p $DB_PORT -U $DB_ADMIN_USER -d $DB_NAME -c "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DB_USER;"
PGPASSWORD=$DB_ADMIN_PASSWORD /Library/PostgreSQL/17/bin/psql -h $DB_HOST -p $DB_PORT -U $DB_ADMIN_USER -d $DB_NAME -c "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $DB_USER;"

echo "✓ Privileges granted successfully"

# Проверка соединения с новой базой данных от имени созданного пользователя
echo "Testing connection with new user..."
if PGPASSWORD=$DB_PASSWORD /Library/PostgreSQL/17/bin/psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c '\q' 2>/dev/null; then
    echo "✓ Connection test successful"
else
    echo "⚠ Warning: Could not connect with new user credentials"
fi

echo ""
echo "🎉 Database initialization completed successfully!"
echo ""
echo "Database details:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo ""
echo "Next steps:"
echo "  1. Run 'make migrate-up' to apply database migrations"
echo "  2. Start the application with 'make run'"