**Структура конфигурационного файла**
```toml
env: local # dev, prod

storage:
  host: localhost
  port: 5432
  database: some_db
  user: some_user
  password: some_password

http_server:
  address: localhost:8082
  timeout: 4s
  idle_timeout: 60s

secret_key: some_key
```