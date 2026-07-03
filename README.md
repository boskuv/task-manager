# Task Manager

REST API для управления задачами в командах: регистрация, JWT-аутентификация, роли в команде, CRUD задач, история изменений, аналитические SQL-запросы, кеш списка задач в Redis.

**Стек:** Go, MySQL, Redis, Docker Compose, goose (миграции).

## Требования

| Инструмент | Зачем |
|------------|--------|
| **Go 1.25+** | сборка и локальный запуск (`go.mod`) |
| **Docker** | `docker compose`, интеграционные тесты и `make test-cover-core` (testcontainers) |
| **golangci-lint** | опционально, `make lint` |
| **jq** | опционально, для примеров curl ниже |

## Структура

```
cmd/api/                    — точка входа
internal/
  app/                      — сборка зависимостей, роутер, конфиг, health, миграции
  domain/                   — сущности и бизнес-ошибки
  dto/                      — контракты API
  model/                    — модели БД, маппинг domain ↔ model
  repository/               — интерфейсы, MySQL и Redis
  usecase/                  — бизнес-логика
  handler/                  — HTTP-обработчики и middleware
  openapi/                  — OpenAPI-спека (embed в бинарник)
  pkg/                      — jwt, password, logging, circuitbreaker, email
  testutil/integration/     — хелперы testcontainers для интеграционных тестов
migrations/                 — SQL-миграции (goose)
configs/                    — YAML-конфиг
docker/                     — Dockerfile
.github/workflows/          — CI (unit + integration)
```

Слои: `handler → usecase → repository`. `domain` не зависит от HTTP и БД.

## Быстрый старт

### Docker (MySQL + Redis + API)

```bash
docker compose up -d
```

API слушает `http://localhost:8080`. MySQL — `localhost:3306`, Redis — `localhost:6379`.  
При старте API автоматически применяются миграции goose (`database.auto_migrate: true`).  
В compose задан dev-секрет `JWT_SECRET` — для production замените на свой.

Проверка:

```bash
curl http://localhost:8080/health          # liveness
curl http://localhost:8080/health/ready    # readiness (MySQL + Redis)
```

### Локально (без контейнера API)

1. Поднять MySQL и Redis (`docker compose up -d mysql redis` или свои инстансы).
2. Скопировать конфиг:

```bash
cp configs/config.example.yaml configs/config.yaml
```

3. Запустить сервер (миграции применятся при старте):

```bash
make run
# или
make build && ./bin/task-manager
```

Проверка — те же `curl` на `/health` и `/health/ready`, что выше.

## Миграции

Инструмент: [goose](https://github.com/pressly/goose). SQL-файлы в каталоге `migrations/`.

| Способ | Когда |
|--------|--------|
| **Автоматически при старте API** | по умолчанию (`database.auto_migrate: true` в конфиге или `DATABASE_AUTO_MIGRATE=true`) |
| **`make migrate`** | ручной прогон без перезапуска API (DSN берётся из `configs/config.yaml`) |
| **`make migrate-down`** | откат последней миграции |
| **`make migrate-create NAME=...`** | создать новый SQL-файл в `migrations/` |

Отключить автомиграции (например, если миграции гоняет отдельный job в CI/CD):

```bash
DATABASE_AUTO_MIGRATE=false make run
```

Каталог миграций переопределяется через `database.migrations_dir` / `DATABASE_MIGRATIONS_DIR` (по умолчанию `migrations`). В Docker-образе файлы копируются в `/app/migrations`.

## Конфигурация

Файл: `configs/config.yaml` (путь переопределяется через `CONFIG_PATH`).

Основные переменные окружения (перекрывают YAML):

| Переменная | Назначение |
|------------|------------|
| `CONFIG_PATH` | путь к YAML |
| `DATABASE_DSN` | DSN MySQL |
| `DATABASE_AUTO_MIGRATE` | применять миграции при старте (`true` по умолчанию) |
| `DATABASE_MIGRATIONS_DIR` | каталог миграций (по умолчанию `migrations`) |
| `REDIS_ADDR` | адрес Redis |
| `JWT_SECRET` | секрет для access token |
| `RATE_LIMIT_REQUESTS_PER_MINUTE` | лимит запросов на пользователя (по умолчанию 100) |
| `LOG_LEVEL` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `text` или `json` |

Пример DSN для docker compose:

```
taskmanager:taskmanager@tcp(localhost:3306)/taskmanager?parseTime=true&charset=utf8mb4
```

## API

Спецификация: [`internal/openapi/openapi.yaml`](internal/openapi/openapi.yaml).  
Интерактивная документация при запущенном сервере: [http://localhost:8080/swagger/](http://localhost:8080/swagger/)

Базовый префикс: `/api/v1`. Защищённые эндпоинты требуют заголовок:

```
Authorization: Bearer <token>
```

Ошибки возвращаются в формате:

```json
{"error": "описание", "code": "invalid_input"}
```

### Служебные

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/health` | liveness: процесс жив (`ok`) |
| GET | `/health/ready` | readiness: ping MySQL и Redis, JSON `{"status":"ready","checks":{...}}` |
| GET | `/metrics` | метрики Prometheus |

### Аутентификация

| Метод | Путь | Тело |
|-------|------|------|
| POST | `/api/v1/register` | `{"email":"...","password":"..."}` |
| POST | `/api/v1/login` | `{"email":"...","password":"..."}` |
| GET | `/api/v1/me` | — (нужен токен) |

Ответ register/login: `{"token":"..."}`. Пароль — минимум 8 символов.

### Команды

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/teams` | создать команду (создатель = owner) |
| GET | `/api/v1/teams` | список команд текущего пользователя |
| POST | `/api/v1/teams/{id}/invite` | пригласить пользователя по email |

Тело invite: `{"email":"...","role":"member"}` или `"admin"`. Приглашать могут только owner и admin. Пользователь с таким email должен быть уже зарегистрирован.

### Задачи

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/tasks` | создать задачу |
| GET | `/api/v1/tasks` | список с фильтрами и пагинацией |
| PUT | `/api/v1/tasks/{id}` | обновить задачу |
| GET | `/api/v1/tasks/{id}/history` | история изменений |

Статусы: `todo`, `in_progress`, `done`.

Query-параметры списка: `team_id` (обязателен), `status`, `assignee_id`, `page`, `page_size`.

Создание задачи — только для участника команды. `assignee_id`, если указан, должен быть участником той же команды.

### Аналитика

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/v1/analytics/teams/stats` | статистика по командам (участники, done за 7 дней) |
| GET | `/api/v1/analytics/top-creators` | топ-3 создателей задач по командам за месяц |

В ответ попадают только команды, в которых состоит текущий пользователь.

## Пример сценария (curl)

```bash
BASE=http://localhost:8080

# регистрация
TOKEN=$(curl -s -X POST "$BASE/api/v1/register" \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"password1"}' \
  | jq -r .token)

# команда
TEAM_ID=$(curl -s -X POST "$BASE/api/v1/teams" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Backend"}' \
  | jq -r .id)

# задача
curl -s -X POST "$BASE/api/v1/tasks" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"team_id\":$TEAM_ID,\"title\":\"Первая задача\",\"description\":\"\"}"

# список задач
curl -s "$BASE/api/v1/tasks?team_id=$TEAM_ID&page=1&page_size=20" \
  -H "Authorization: Bearer $TOKEN"
```

## Makefile

```bash
make help              # список всех целей
make build             # собрать bin/task-manager
make run               # запустить API
make fmt               # go fmt + gofmt -s
make lint              # golangci-lint
make clean             # удалить bin/, coverage.out, coverage.html

make migrate           # применить миграции вручную
make migrate-down      # откатить последнюю миграцию
make migrate-create NAME=create_foo  # новая миграция

make test              # unit-тесты (race)
make test-integration  # интеграционные тесты (-tags=integration, нужен Docker)
make test-cover        # покрытие всех пакетов, без порога
make test-cover-core   # покрытие usecase + repository, порог 85%
make test-cover-html   # HTML-отчёт coverage.html (после test-cover-core)
```

## Тесты

### Что где тестируется

| Слой | Unit | Integration |
|------|------|-------------|
| usecase (auth, team, task, analytics) | моки репозиториев | task usecase + MySQL |
| repository (mysql, redis) | — | testcontainers |
| handler + middleware | httptest, моки usecase | — |
| app (роутер, конфиг, health, migrate) | unit + smoke (`router_smoke_test.go`) | app + MySQL/Redis |
| pkg (jwt, password, logging, …) | unit | — |

Хелперы контейнеров: `internal/testutil/integration/`.

### Покрытие

Критичные слои — **usecase** и **repository**. Целевой порог: **85%**.

```bash
make test-cover-core   # ~86.8% на текущей версии; падает, если ниже 85%
make test-cover-html   # открыть coverage.html в браузере
```

`make test-cover` считает покрытие по всем пакетам (`cmd`, `handler`, `app`, …) без проверки порога.

### CI

Файл: [`.github/workflows/ci.yml`](.github/workflows/ci.yml). Триггер: push/PR в `main` и `dev`.

| Job | Команда |
|-----|---------|
| Unit tests | `go test ./cmd/... ./internal/... -race -count=1` |
| Integration tests | то же с `-tags=integration` |

Проверка покрытия (`test-cover-core`) в CI **не** запускается — только локально перед релизом или в отдельном workflow.

## База данных

Таблицы: `users`, `teams`, `team_members`, `tasks`, `task_history`, `task_comments`.

Роли в команде: `owner`, `admin`, `member`.

Таблица `task_comments` есть в схеме; API комментариев в текущей версии не реализован.

## Прочее

- Список задач кешируется в Redis (TTL 5 мин), инвалидируется при создании и обновлении задачи.
- Rate limit: скользящее окно в Redis, лимит на пользователя.
- При invite вызывается mock email-сервис через circuit breaker (логирование, сбой email не блокирует добавление в команду).
- Graceful shutdown по SIGINT/SIGTERM.
