# Backend & Go Конвенции (`docs/conventions/backend.md`)

> Этот документ определяет правила архитектуры, проектирования и написания серверного кода на **Go 1.23+** для проекта **Takt**.
> Обязателен к прочтению субагентами роли `backend_developer` и всеми инженерами бэкенда.

---

## 🏛️ 1. Архитектурные слои и Закон трех подпакетов

В директории `backend/internal/` разрешены **СТРОГО 3 корневых подпакета**:
`domain`, `application`, `infrastructure`.

> **Закон архитектуры**: Создавать другие папки на верхнем уровне `internal/` (например, `delivery/`, `config/`, `app/`, `container/`) **запрещено** и блокируется скриптом `make check-arch`. Все адаптеры внешнего мира и Composition Root приложения (`container/` или `di/`) размещаются строго внутри подпакетов `infrastructure/*`.

```text
backend/internal/
├── domain/              # Чистый домен (сущности, инварианты, доменные ошибки и события)
│   ├── shared/          # Базовые примитивы (ID, Pagination, системные доменные ошибки)
│   ├── workspace/       # entity.go, errors.go, repository.go (mocks/mock_*.go)
│   ├── project/         # project.go, board.go, column.go, repository.go
│   ├── task/            # task.go, tag.go, errors.go, repository.go (mocks/mock_*.go)
│   ├── automation/      # rule.go, evaluator.go, repository.go
│   └── events/          # task_events.go, eventbus.go
├── application/         # Сценарии использования (Use Cases) + интерфейсы зависимостей (ports.go)
│   ├── workspace/       # usecases.go, ports.go (mocks/mock_*.go)
│   ├── project/         # usecases.go, ports.go
│   ├── task/            # usecases.go, ports.go (mocks/mock_*.go)
│   └── automation/      # runner.go, ports.go
└── infrastructure/      # ВСЕ внешние адаптеры, драйверы и Composition Root
    ├── container/       # DI и сборка приложения (repos.go, services.go, router.go, app.go)
    ├── config/          # Загрузка и валидация .env (config.go)
    ├── http/            # Echo роутер, middleware, хендлеры
    ├── db/              # pgxpool, транзактор через Context, sqlc queries
    ├── eventbus/        # In-memory диспетчер событий (задел под Kafka)
    └── repository/      # Реализации репозиториев поверх sqlc
```

### Двойной лимит и культура декомпозиции (Лимит < 500 строк, <= 7 файлов)
- **Лимит строк**: Строго `< 500 строк` на файл (проверяется `make check-arch`). Целевой ориентир для комфортного чтения и работы агентов — `< 300–400 строк`.
- **Запрет на втискивание**: Если файл приближается к 350-400 строкам, **категорически запрещено** зацикливаться на микро-оптимизации строк в одном файле. Действуй проактивно:
  1. Выноси чистые парсеры, мапперы и вспомогательные функции (`toFloat`, `extractParams`, маппинг строк sqlc в домен) в `<file>_helpers.go`.
  2. Разделяй исполнение условий и действий на отдельные файлы (`evaluator_conditions.go`, `evaluator_actions.go`).
  3. **Выделяй вложенные подпакеты**: сложная изолированная логика (например, движок вычисления правил, планировщик расписаний, процессинг событий) должна жить в собственном подпакете (`application/automation/evaluator/`, `application/automation/scheduler/`).
- **Лимит пакета**: Не более 5–7 рабочих файлов в одной директории (не считая `*_test.go`). Если файлов становится больше — создавай вложенный подпакет.

| Слой | Назначение | Что может импортировать | Чего НЕ должно быть внутри |
|---|---|---|---|
| **`domain/<pkg>`** | Чистые бизнес-сущности, Value Objects, доменные ошибки и интерфейсы репозиториев своего домена. | Только стандартную библиотеку и `internal/domain/shared`. | Никаких импортов SQL, HTTP, драйверов и **никаких импортов соседних `domain/<other>`**. |
| **`application/<pkg>`** | Сценарии использования (Use Cases), оркестрация доменной логики, проверка прав доступа. | Свой `internal/domain/<pkg>` и `internal/domain/shared`. | Никаких прямых SQL-запросов, HTTP-роутеров и **никаких прямых импортов соседних `application/<other>`**. |
| **`infrastructure/<pkg>`** | Адаптеры внешнего мира (PostgreSQL, sqlc, Echo, EventBus, Config) и сборка (`container/`). | `domain/*` + `application/*` + сторонние библиотеки (pgx, echo). | Чистая бизнес-логика здесь запрещена: здесь живут SQL-запросы, маппинг DTO, роутинг, чтение ENV и DI-сборка (`container/`). |

---

## 🔗 2. Защита от циклических зависимостей в Go (`No Import Cycles Rule`)

1. **Интерфейсы объявляются на стороне потребителя (`ports.go`)**:
   Пакет `application/automation` **НЕ ИМЕЕТ ПРАВА** напрямую импортировать пакет `application/task`. Если сервису автоматизаций нужно изменить статус задачи, он объявляет в своем файле `internal/application/automation/ports.go` минимальный интерфейс:
   ```go
   package automapp

   import (
       "context"
       "takt/internal/domain/task"
   )

   type TaskMutator interface {
       MoveTaskToColumn(ctx context.Context, taskID task.ID, columnID task.ColumnID) error
   }
   ```
2. **Разделенная сборка в `internal/infrastructure/container/`**:
   * `cmd/server/main.go` содержит только чтение конфигурации, запуск `container.Run(ctx, cfg)` и Graceful Shutdown (< 50 строк).
   * Инициализация подключений и репозиториев живет в `internal/infrastructure/container/repos.go`, сервисов — в `services.go`, регистрация роутов Echo — в `router.go`.

---

## 💳 3. Транзакции через `context.Context` (Unit of Work)

* Управление транзакциями строится через интерфейс `Transactor`:
  ```go
  type Transactor interface {
      WithinTx(ctx context.Context, fn func(txCtx context.Context) error) error
  }
  ```
* Сервисный слой оборачивает атомарные операции в `transactor.WithinTx(ctx, ...)`.
* Репозитории БД на базе `sqlc` извлекают `pgx.Tx` из переданного контекста и используют `queries.WithTx(tx)`. Если транзакции в контексте нет — запрос выполняется штатно через пул `pgxpool.Pool`.

---

## 📜 4. Контрактная разработка (Contract-First) и генерация Echo v4 через `oapi-codegen`

1. **Единый источник правды — OpenAPI 3.0**:
   * Все эндпоинты, параметры, тела запросов и DTO ответов сначала проектируются в модульной спецификации `api/openapi/` (`paths/`, `schemas/`).
   * Запуск `make generate` выполняет склейку (`@redocly/cli bundle` $\rightarrow$ `api/dist/openapi.json`), генерацию Go-сервера через `oapi-codegen` и генерацию SQL-запросов через `sqlc`.

2. **Кодогенерация Echo v4 (`oapi-codegen v2`)**:
   * Конфигурация хранится в `backend/oapi-codegen.yaml`:
     ```yaml
     package: generated
     generate:
       echo-server: true
       models: true
     output: internal/infrastructure/http/generated/api.gen.go
     ```
   * В пакете `backend` размещена директива кодогенерации `backend/generate.go`:
     ```go
     //go:generate oapi-codegen --config=oapi-codegen.yaml ../api/dist/openapi.json
     ```
   * Сгенерированный код помещается в `internal/infrastructure/http/generated/api.gen.go` и содержит:
     - Все структуры моделей DTO (`generated.CreateInboxTaskRequest`, `generated.Task`, `generated.User` и др.).
     - Интерфейс сервера Echo: `generated.ServerInterface`.
     - Функции автоматической регистрации маршрутов: `generated.RegisterHandlers(e, server)` и `generated.RegisterHandlersWithBaseURL(e, server, "/api/v1")`.

3. **Законы реализации HTTP-слоя (`internal/infrastructure/http/`)**:
   * **Обязательная имплементация `ServerInterface`**: Серверный фасад или хендлеры в `internal/infrastructure/http/` обязаны реализовывать `generated.ServerInterface`. Для гарантии компилятора объявляется явная проверка:
     ```go
     var _ generated.ServerInterface = (*Server)(nil)
     ```
   * **Табу на рукописные DTO**: Запрещено создавать собственные структуры запросов/ответов (`type CreateTaskRequest struct`) вручную в файлах хендлеров. Используются исключительно типы из пакета `generated.*`.
   * **Табу на ручную регистрацию роутов**: Маршрутизация привязывается к роутеру Echo через `generated.RegisterHandlers(apiGroup, server)`. Ручная регистрация `e.POST(...)` / `e.GET(...)` для API-эндпоинтов, описанных в схеме, запрещена.

4. **Обработка ошибок и маппинг**:
   * Доменные ошибки типизированы (`ErrNotFound`, `ErrForbidden`, `ErrUnauthorized`, `ErrValidation`, `ErrConflict`).
   * Сервисы оборачивают ошибки с контекстом: `fmt.Errorf("taskapp.CreateTask: %w", err)`.
   * Единый HTTP Error Mapper переводит доменные ошибки в статус-коды (`404`, `403`, `401`, `400`, `409`, `429`), а неизвестные ошибки логирует в `stderr` и возвращает клиенту безопасный `500 Internal Server Error` без раскрытия стека.


---

## 🧪 5. Стандарт тестирования Go-кода

### 1. Преимущественно табличные тесты (Table-Driven Tests)
Все тесты сервисов, валидаторов и обработчиков структурируются по табличному паттерну. Каждый кейс явно документирует свое назначение:

```go
func TestService_ToggleTask(t *testing.T) {
    testCases := []struct {
        name        string
        taskID      task.ID
        workspaceID workspace.ID
        setupMock   func(repo *mocks.MockTaskRepository)
        wantErr     error
    }{
        {
            name:        "успешное переключение задачи",
            taskID:      task.NewID(),
            workspaceID: workspace.NewID(),
            setupMock: func(repo *mocks.MockTaskRepository) {
                repo.EXPECT().GetByID(mock.Anything, mock.Anything).Return(validTask, nil).Once()
                repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()
            },
            wantErr: nil,
        },
        {
            name:        "задача не найдена в текущем воркспейсе",
            taskID:      task.NewID(),
            workspaceID: workspace.NewID(),
            setupMock: func(repo *mocks.MockTaskRepository) {
                repo.EXPECT().GetByID(mock.Anything, mock.Anything).Return(nil, task.ErrNotFound).Once()
            },
            wantErr: task.ErrNotFound,
        },
    }

    for _, tc := range testCases {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            repo := mocks.NewMockTaskRepository(t)
            tc.setupMock(repo)

            svc := taskapp.NewService(repo)
            err := svc.ToggleTask(context.Background(), tc.workspaceID, tc.taskID)

            if tc.wantErr != nil {
                require.ErrorIs(t, err, tc.wantErr)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### 2. Библиотека ассертов: `testify`
* `require.*` (`github.com/stretchr/testify/require`) — для прерывающих проверок (ошибки, nil-указатели, длины срезов перед чтением по индексу). Если упал `require`, тест немедленно завершается, предотвращая панику (nil pointer dereference).
* `assert.*` (`github.com/stretchr/testify/assert`) — для проверки возвращаемых значений, эквивалентности DTO, булевых флагов и статусов.

### 3. Моки через `mockery v3` (строго шаблон `testify`)
* **Автогенерация моков**: Все моки генерируются строго утилитой `mockery` со стандартным шаблоном `template: testify`.
* **Приоритет моков над фейками**: Преимущественно используются **моки**, проверяющие контракт вызова (`EXPECT().Method(...)` или `On("Method", ...).Return(...)`) и автоматически верифицирующие ожидания через `mock.AssertExpectations(t)` в `Cleanup`.
* **Фейки и стабы**: допустимы исключительно там, где они объективно оправданы (например, сложный эмулятор времени/часов или in-memory хранилище для интеграционных сценариев), но дефолтный инструмент изоляции — моки.
* **Расположение моков**: Моки располагаются в подпакете `mocks/` строго внутри того пакета, где объявлен мокаемый интерфейс:
  - Интерфейс `internal/domain/task/repository.go` $\rightarrow$ мок в `internal/domain/task/mocks/mock_TaskRepository.go`.
  - Интерфейс `internal/application/task/ports.go` $\rightarrow$ мок в `internal/application/task/mocks/mock_WorkspaceAccessChecker.go`.
* **Конфигурация `.mockery.yaml`**:
  ```yaml
  dir: "{{.InterfaceDir}}/mocks"
  filename: "mock_{{.InterfaceName}}.go"
  structname: "Mock{{.InterfaceName}}"
  pkgname: "mocks"
  template: testify
  force-file-write: true
  recursive: true
  all: true
  packages:
    takt:
      config:
        recursive: true
        all: true
  ```
* **Запуск генерации**: команда `make mock` (или `cd backend && mockery`).

