# Maphraohom Backoffice

[Maphraohom] REST + gRpc using Fiber + OpenTelemetry + exception tracking by Sentry.io

## Extra feature includes

1. Object relational mapping (ORM)
2. In-memory Caching
3. Error exception handling
4. Log level + structured
5. Observability (tracing)
6. API Documentation (Swagger)
7. Scheduler Tasks (Background job)

## Documentation

- Fiber Framework (<https://docs.gofiber.io/>)
- GORM (<https://gorm.io/>)
- Go Redis (<https://redis.uptrace.dev/guide/go-redis.html>)
- Sentry (<https://docs.sentry.io/platforms/go/>)
- OpenTelemetry (<https://opentelemetry.io/docs/instrumentation/go/>)
- Swaggo (<https://github.com/swaggo/swag>)
- Fiber Swagger (<https://github.com/gofiber/swagger>)
- Go Graylog (<https://github.com/Devatoria/go-graylog>)
- GoCron (<https://github.com/go-co-op/gocron>)

## System requirements

- Golang 1.23
- Database:
  - PostgreSQL 15+
  - MySQL 8+
  - MariaDB 10+
  - SQL Server 2008+
  - SQLite 3.6.0+
- Redis 6+

## Go commands

### Install dependencies

```bash
go install
```

### Update dependencies

```bash
go mod tidy
```

### Build go executable file

```bash
go build -o main
```

### Start development server

```bash
go run .

# or
go run main.go
```

The server will listen by default at <http://localhost:8000>

Use `air` command for live-reloading, `go install github.com/air-verse/air@latest` for install it.

## Generator

### Module (CRUD)

The framework with generate new module directory. then inside will have route, dtos, controller, service and repository files.

### Folder structure

```text
📁 src/modules
├── 📁 team_module
│   ├── 📁 dtos
│   │   ├── 📄 dtos.go
│   ├── 📄 team_controller.go
│   ├── 📄 team_module.go
│   ├── 📄 team_repository.go
│   ├── 📄 team_service.go
```

### Command

```bash
go run cmd/generate.go module <module_name>

# Example
go run cmd/generate.go module team
```

### Register module routes

Don't forgot to register module route at `src/routes/http.go` at the `HTTPRoutes` method.

```go
func HTTPRoutes(s *http_server.HttpServer) {
  ...
  // Register routes
  ...
  team_module.NewModule().Routes(api, middleware)
}
```

---

### Model (GORM - declaring models)

The framework with generate new model file.

### Folder structure

```text
📁 src
├── 📄 models
│   ├── 📄 team_model.go
```

### Command

```bash
go run cmd/generate.go model <model_name>

# Example
go run cmd/generate.go model team
```

### Sample model file

```go
package models

type Team struct {
 BaseModel
}
```

## Swagger

### Install Swaggo (for use `swag` command)

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### Generate Swagger document files

```bash
swag init
```

### Format the Swagger comment

```bash
swag fmt
```

## Generate RSA certificates (using OpenSSL 3+)

```bash
# Generate private key
openssl genrsa -out private.pem 2048

# Generate public key
openssl rsa -in private.pem -pubout > public.pem
```

## Environment Variables

```sh
# App config
ENV="local"
HTTP_PORT=8000
GRPC_PORT=8001
APP_NAME="Maphraohom - Backoffice Service"
SERVICE_NAME="maphraohom-backoffice-service"

# Fiber config
FIBER_PREFORK=true
CORS_ALLOW_ORIGINS="*"
RATE_LIMIT=60

# Database config
# Note: Database is support 4 engines
# 1) postgresql (default)
# 2) mysql
# 3) sqlserver
# 4) sqlite [local development only]
DATABASE_DRIVER="postgresql"
DATABASE_HOST="localhost"
DATABASE_PORT=5432
DATABASE_NAME="postgres"
DATABASE_USER="postgres"
DATABASE_PASSWORD="mysecretpassword"
DATABASE_TIMEZONE="Asia/Bangkok"
DATABASE_MAX_IDLE_CONNS=2
DATABASE_MAX_OPEN_CONNS=3
DATABASE_SSL_MODE="disable"

# Redis config
REDIS_HOST="127.0.0.1"
REDIS_PORT="6379"
REDIS_USERNAME="default"
REDIS_PASSWORD=""
REDIS_CACHE_PREFIX="http-service"
REDIS_CACHE_DURATION=5

# Sentry config
SENTRY_DSN=""
SENTRY_ERROR_TRACING=false
SENTRY_TRACES_SAMPLE_RATE=0.2

# OpenTelemetry config
OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4317"
OTEL_INSECURE_MODE=true

# Auth config
AUTH_KEYCLOAK_ENDPOINT="http://localhost:8080"
AUTH_KEYCLOAK_REALM="master"
AUTH_SECRET_KEY="mysecret"

# Logger
# Note: Logger is support 7 modes
# 1) none = Print to console
# 2) file = Write a log file (daily)
# 3) elastic = Write log to Elasticsearch
# 4) database = Write log to table database
# 5) graylog = Write log to Graylog
# 6) s3 = Write log to S3
# 7) gcp = Write log to Google Cloud Logging
LOGGER_MODE="none"
LOGGER_PATH="./logs"
# Logger Graylog
LOGGER_GRAYLOG_ENDPOINT="localhost"
LOGGER_GRAYLOG_PORT=12201
LOGGER_GRAYLOG_TLS_ENABLED=false
LOGGER_GRAYLOG_TLS_SKIP_VERIFY=true
# Logger S3
LOGGER_S3_PATH="logs"
LOGGER_S3_ENDPOINT="localhost:9000"
LOGGER_S3_ACCESS_KEY_ID=""
LOGGER_S3_SECRET_KEY=""
LOGGER_S3_BUCKET_NAME="my-bucket"
LOGGER_S3_REGION="us-east-1"
LOGGER_S3_PREFIX=""
LOGGER_S3_USE_SSL=false
LOGGER_S3_MAX_FILE_KB_SIZE=10240
# Logger Google Cloud Logging | Note: LOGGER_GCP_PARENT_ID such as 'projects/my-project'
LOGGER_GCP_PARENT_ID="my-project"
LOGGER_GCP_LOG_ID="my-log"

# FileSystem
# Note: FileSystem is support 2 modes
# 1) local = Save file to local storage
# 2) s3 = Save file to AWS S3 (or compatible S3 storage like MinIO)
FILE_SYSTEM_DISK="local"
FILE_SYSTEM_LOCAL_PATH="/app/storage" # In case of local disk (development) should be "./storage".
FILE_SYSTEM_S3_ENDPOINT="localhost:9000"
FILE_SYSTEM_S3_ACCESS_KEY_ID=""
FILE_SYSTEM_S3_SECRET_KEY=""
FILE_SYSTEM_S3_BUCKET_NAME="my-bucket"
FILE_SYSTEM_S3_REGION="us-east-1"
FILE_SYSTEM_S3_PREFIX=""
FILE_SYSTEM_S3_USE_SSL=false
 ```
