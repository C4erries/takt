# Backend & Go Конвенции (`docs/conventions/backend.md`)

> Этот документ определяет правила архитектуры, проектирования и написания серверного кода на **Go 1.23+**.
> Обязателен к прочтению субагентами роли `backend_developer` и всеми инженерами бэкенда.

---

## 🏛️ 1. Архитектурные слои и Закон трех подпакетов

В директории `backend/internal/` разрешены **СТРОГО 3 корневых подпакета**:
`domain`, `application`, `infrastructure`.

> **Закон архитектуры**: Создавать другие папки на верхнем уровне `internal/` (например, `delivery/`, `config/`, `app/`, `container/`) **запрещено** и блокируется скриптом `make check-arch`. Все адаптеры внешнего мира и Composition Root приложения (`container/` или `di/`) размещаются строго внутри подпакетов `infrastructure/*`.

```text
backend/internal/
├── domain/              # Чистый домен (сущности, инварианты, доменные ошибки и события)
│   ├── shared/          # Базовые примитивы (UserID, Money, Pagination, базовые коды ошибок)
│   ├── user/            # entity.go, value_objects.go, errors.go, repository.go (mocks/mock_*.go)
│   └── billing/         # invoice.go, subscription.go, errors.go, repository.go (mocks/mock_*.go)
├── application/         # Сценарии использования (Use Cases) + интерфейсы зависимостей (ports.go)
│   ├── user/            # register.go, profile.go, ports.go (mocks/mock_*.go)
│   └── billing/         # charge.go, webhook.go, ports.go (mocks/mock_*.go)
└── infrastructure/      # ВСЕ внешние адаптеры, драйверы и Composition Root
    ├── container/       # DI и сборка приложения (repos.go, services.go, router.go, container.go)
    ├── config/          # Загрузка и валидация .env (config.go)
    ├── postgres/        # userrepo, billingrepo
    └── http/            # middleware, userhandler, billinghandler
```

### Матрица ответственности слоев и правила импортов

| Слой | Назначение | Что может импортировать | Чего НЕ должно быть внутри |
|---|---|---|---|
| **`domain/<pkg>`** | Чистые бизнес-сущности, Value Objects, доменные ошибки и интерфейсы репозиториев своего домена. | Только стандартную библиотеку и `internal/domain/shared`. | Никаких импортов SQL, HTTP, сторонних библиотек и **никаких импортов соседних `domain/<other>`**. |
| **`application/<pkg>`** | Сценарии использования (Use Cases), оркестрация доменной логики, проверка прав доступа. | Свой `internal/domain/<pkg>` и `internal/domain/shared`. | Никаких прямых SQL-запросов, HTTP-роутеров и **никаких прямых импортов соседних `application/<other>`**. |
| **`infrastructure/<pkg>`** | Адаптеры внешнего мира (PostgreSQL, sqlc, Echo, S3, Config) и сборка (`container/`). | `domain/*` + `application/*` + сторонние драйверы. | Чистая бизнес-логика здесь запрещена: здесь живут SQL-запросы, маппинг DTO, роутинг, чтение ENV и DI-сборка (`container/`). |

---

## 🔗 2. Защита от циклических зависимостей в Go (`No Import Cycles Rule`)

1. **Интерфейсы объявляются на стороне потребителя (`ports.go`)**:
   Пакет `application/billing` **НЕ ИМЕЕТ ПРАВА** импортировать пакет `application/user` (и наоборот). Если сценарию в `application/billing` нужно проверить статус пользователя, он объявляет в своем файле `internal/application/billing/ports.go` минимальный интерфейс:
   ```go
   package billingapp

   import (
       "context"
       "app/internal/domain/shared"
   )

   // UserChecker описывает внешнюю зависимость, нужную биллингу от домена пользователей.
   type UserChecker interface {
       IsUserActive(ctx context.Context, userID shared.UserID) (bool, error)
   }
   ```
2. **Разделенная сборка в `internal/infrastructure/container/` (или `di/`)**:
   * `cmd/main.go` содержит только чтение конфига, запуск `container.Run(ctx, cfg)` и Graceful Shutdown (< 50 строк).
   * Инициализация репозиториев живет в `internal/infrastructure/container/repos.go`, сервисов — в `container/services.go`, регистрация HTTP-хендлеров — в `container/router.go`.

---

## 💳 3. Транзакции через `context.Context` (Unit of Work)

* Управление транзакциями строится через интерфейс `Transactor`:
  ```go
  type Transactor interface {
      WithinTx(ctx context.Context, fn func(txCtx context.Context) error) error
  }
  ```
* Сервисный слой оборачивает атомарные операции в `transactor.WithinTx(ctx, ...)`.
* Репозитории БД извлекают транзакцию из переданного контекста. Если транзакции в контексте нет — запрос выполняется штатно через пул соединений.

---

## 📜 4. Контрактная разработка (Contract-First) и Ошибки

* Все эндпоинты сначала фиксируются в `backend/api/openapi/api.yaml`, после чего запускается кодогенерация (`make oapi` / `make generate`).
* Доменные ошибки типизированы (`ErrNotFound`, `ErrForbidden`, `ErrUnauthorized`, `ErrValidation`, `ErrConflict`).
* Сервисы оборачивают ошибки с контекстом: `fmt.Errorf("billingapp.Charge: %w", err)`.
* Единый HTTP Error Mapper переводит доменные ошибки в статус-коды (`404`, `403`, `401`, `400`, `409`, `429`), а неизвестные ошибки логирует в `stderr` и возвращает клиенту безопасный `500 Internal Server Error` без раскрытия стека.

---

## 🧪 5. Стандарт тестирования Go-кода

### 1. Преимущественно табличные тесты (Table-Driven Tests)
Все тесты сервисов, валидаторов и обработчиков структурируются по табличному паттерну с явным именованием тест-кейсов:

```go
func TestService_Method(t *testing.T) {
    testCases := []struct {
        name      string
        input     InputDTO
        setupMock func(repo *mocks.MockRepository)
        wantErr   error
    }{
        {
            name: "успешный сценарий",
            ...
        },
        {
            name: "ошибка: сущность не найдена",
            ...
        },
    }

    for _, tc := range testCases {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            repo := mocks.NewMockRepository(t)
            tc.setupMock(repo)

            svc := pkgapp.NewService(repo)
            err := svc.Method(context.Background(), tc.input)

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
* `require.*` (`github.com/stretchr/testify/require`) — для прерывающих проверок (ошибки, nil-указатели, длины срезов перед чтением по индексу).
* `assert.*` (`github.com/stretchr/testify/assert`) — для проверки возвращаемых значений, эквивалентности DTO, булевых флагов и статусов.

### 3. Моки через `mockery v3` (строго шаблон `testify`)
* **Автогенерация моков**: Все моки генерируются строго утилитой `mockery` со стандартным шаблоном `template: testify`.
* **Приоритет моков над фейками**: Преимущественно используются **моки**, проверяющие контракт вызова (`EXPECT().Method(...)` или `On("Method", ...).Return(...)`) и автоматически верифицирующие ожидания через `mock.AssertExpectations(t)` в `Cleanup`.
* **Фейки и стабы**: допустимы исключительно там, где они объективно оправданы (например, эмулятор времени или in-memory хранилище), но дефолтный инструмент изоляции — моки.
* **Расположение моков**: Моки располагаются в подпакете `mocks/` строго внутри того пакета, где объявлен мокаемый интерфейс:
  - Интерфейс `internal/domain/<pkg>/repository.go` $\rightarrow$ мок в `internal/domain/<pkg>/mocks/mock_<Interface>.go`.
  - Интерфейс `internal/application/<pkg>/ports.go` $\rightarrow$ моки в `internal/application/<pkg>/mocks/mock_*.go`.
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
    app:
      config:
        recursive: true
        all: true
  ```
* **Запуск генерации**: команда `make mock` (или `cd backend && mockery`).

