#!/bin/bash
set -e

# Создание базы данных и пользователя для приложения
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER app_user WITH PASSWORD 'app_password';
    GRANT ALL PRIVILEGES ON DATABASE $POSTGRES_DB TO app_user;
    GRANT ALL ON SCHEMA public TO app_user;
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO app_user;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO app_user;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO app_user;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO app_user;
EOSQL

echo "Database and user created successfully"