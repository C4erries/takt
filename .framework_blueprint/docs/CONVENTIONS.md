# Принципы и инженерные конвенции проекта (`CONVENTIONS.md`)

> Этот документ — **свод инженерных и дизайнерских законов** для всех, кто пишет код в репозитории (включая разработчиков и AI-агентов).
> Соблюдение этих правил контролируется автоматически скриптом **`make check-arch`** и главным агентом (Gatekeeper).

---

## 🏗️ 1. Общие принципы разработки и Модульность (Защита контекста ИИ)

### 12-Factor App (Практическая адаптация)
1. **Конфигурация через окружение**: Все настройки, порты, DSN баз данных и секреты читаются через переменные окружения (`env`). В Git хранится только `.env.example`; реальные `.env` файлы **никогда не коммитятся**.
2. **Stateless сервисы**: Серверные процессы не хранят состояние в памяти процесса (состояние делегируется в PostgreSQL, Valkey/Redis, S3).
3. **Структурированное логирование**: Логи пишутся в `stdout`/`stderr` (JSON в продакшне, читаемый формат в dev) без утечки паролей, токенов и персональных данных.
4. **Dev/Prod Parity**: Все окружение поднимается одной командой `docker compose up`.

### Двойной лимит гранулярности (Против «свалки файлов» и перегрузки LLM)
1. **Лимит размера файла: строго `< 300–400 строк`**
   * *Почему*: Файлы свыше 400 строк перегружают контекстное окно агентов, вызывают `truncated output` и повышают риск сломать соседний код.
   * *Правило*: При приближении файла к 300 строкам он **обязан** быть декомпозирован. Единственное исключение — автосгенерированный код (`*.gen.go`, `*_gen.go`, `*.gen.ts`, `*.pb.go`).
2. **Лимит размера пакета / папки: строго `<= 5–7 рабочих файлов`**
   * *Почему*: Когда в одной папке (`internal/domain` или `internal/application`) скапливается 20–30 файлов, в Go происходит загрязнение пространства имен пакета (конфликты приватных функций и ошибок), а агенты тратят лишние токены на поиск нужного кода.
   * *Правило*: **Запрещено сваливать файлы разных сущностей в один общий пакет.** Каждый архитектурный слой обязательно нарезается на **доменные подпакеты** (глубина вложенности = 1). В одной доменной директории не может находиться более 5–7 рабочих файлов (не считая `*_test.go`). Исключение: плоский каталог базовых UI-примитивов `frontend/src/shared/components/ui`.

### Правила работы с Git
* **Ветки**: `main` (стабильный релиз), `dev` (ветка текущего спринта), `feature/<name>` / `fix/<name>`.
* **Абсолютное табу для AI-агентов**:
  > Агентам **СТРОГО ЗАПРЕЩЕНО** выполнять любые мутирующие команды Git (`git commit`, `git push`, `git checkout`, `git stash`, `git rebase`), если в текущем промпте пользователя нет прямого указания. Разрешены только read-only команды (`git status`, `git diff`, `git log`).
* При явном запросе коммита используется формат **Conventional Commits** (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`).

---

## 🐹 2. Backend конвенции (Clean Architecture / DDD + Подпакеты без `import cycle`)

#### Архитектурные слои и структура подпакетов (СТРОГО 4 пакета в internal/)

> **Закон архитектуры**: В директории `backend/internal/` разрешены **СТРОГО 4 корневых подпакета**: `container` (или `di`), `application`, `domain`, `infrastructure`.  
> Создавать другие папки на верхнем уровне `internal/` (например, `delivery/`, `config/`, `app/`) **запрещено** и блокируется скриптом `make check-arch`. Все внешние адаптеры (HTTP-хендлеры, чтение конфига, БД) размещаются строго внутри подпакетов `infrastructure/*`.

```text
backend/internal/
├── container/           # DI и сборка приложения (Composition Root): repos.go, services.go, router.go, container.go
├── domain/              # Чистый домен (сущности, инварианты, доменные ошибки и события)
│   ├── shared/          # Базовые примитивы (UserID, Money, Pagination, базовые коды ошибок)
│   ├── user/            # Подпакет user: entity.go, value_objects.go, errors.go, repository.go
│   └── billing/         # Подпакет billing: invoice.go, subscription.go, errors.go, repository.go
├── application/         # Сценарии использования (Use Cases) + интерфейсы зависимостей (ports.go)
│   ├── user/            # Подпакет userapp: register.go, profile.go, ports.go
│   └── billing/         # Подпакет billingapp: charge.go, webhook.go, ports.go
└── infrastructure/      # ВСЕ внешние адаптеры и драйверы
    ├── config/          # Загрузка и валидация .env (config.go)
    ├── postgres/        # userrepo, billingrepo
    └── http/            # middleware, userhandler, billinghandler
```

| Слой | Назначение | Что может импортировать | Чего НЕ должно быть внутри |
|---|---|---|---|
| **`domain/<pkg>`** | Чистые бизнес-сущности, Value Objects, доменные ошибки и интерфейсы репозиториев своего домена. | Только стандартную библиотеку и `internal/domain/shared`. | Никаких импортов SQL, HTTP, сторонних библиотек и **никаких импортов соседних `domain/<other>`**. |
| **`application/<pkg>`** | Сценарии использования (Use Cases), оркестрация доменной логики, проверка прав доступа. | Свой `internal/domain/<pkg>` и `internal/domain/shared`. | Никаких прямых SQL-запросов, HTTP-роутеров и **никаких прямых импортов соседних `application/<other>`**. |
| **`infrastructure/<pkg>`** | Адаптеры внешнего мира (PostgreSQL, Valkey/Redis, HTTP handlers, S3, Config). | `domain/*` + `application/*` + сторонние драйверы. | Здесь живут SQL-запросы, маппинг DTO, роутинг и чтение ENV. Чистая бизнес-логика здесь запрещена. |
| **`container/` (DI Wiring)** | Сборка всех репозиториев, сервисов, адаптеров и роутов. | Все слои (`domain`, `application`, `infrastructure`). | Никакой бизнес-логики — только инициализация и передача зависимостей (`cmd/main.go < 50 строк`). |

### Защита от циклических зависимостей в Go (`No Import Cycles Rule`) и `cmd/main.go`
При разбиении кода на подпакеты в Go действуют два закона:
1. **Запрет горизонтальных импортов (`ports.go`)**: Пакет `application/billing` **НЕ ИМЕЕТ ПРАВА** импортировать пакет `application/user` (и наоборот). Если сценарию в `application/billing` нужно проверить статус пользователя, он объявляет в своем файле `internal/application/billing/ports.go` минимальный интерфейс:
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
2. **Разделенная сборка в `internal/app/` (против разрастания `cmd/main.go`)**:
   * `cmd/main.go` содержит только чтение конфига, запуск `app.Run(ctx, cfg)` и Graceful Shutdown (< 50 строк).
   * Инициализация репозиториев живет в `internal/app/repos.go`, сервисов — в `internal/app/services.go`, регистрация HTTP-хендлеров — в `internal/app/router.go`.

### Контрактная разработка (Contract-First) и Обработка ошибок
* Все эндпоинты сначала фиксируются в `backend/api/openapi/api.yaml`, после чего запускается кодогенерация (`make oapi`).
* Доменные ошибки типизированы (`ErrNotFound`, `ErrForbidden`, `ErrUnauthorized`, `ErrValidation`, `ErrConflict`, `ErrRateLimited`).
* Сервисы оборачивают ошибки с контекстом: `fmt.Errorf("billingapp.Charge: %w", err)`.
* Единый HTTP Error Mapper переводит доменные ошибки в статус-коды (`404`, `403`, `401`, `400`, `409`, `429`), а любые неизвестные ошибки логирует в `stdout` и возвращает клиенту безопасный `500 Internal Server Error` без деталей внутренней кухни.

---

## 🎨 3. Frontend конвенции и Дизайн-паспорт (React + TypeScript)

### Структура каталогов (Feature-Sliced + Лимит `<= 7 файлов в папке фичи`)
```text
frontend/src/
├── app/                     # Инициализация, провайдеры, глобальный роутер
├── pages/                   # Тонкие страницы-контейнеры (до 100 строк, только композиция фич)
├── features/                # Доменные модули (auth, billing, dashboard)
│   └── [feature_name]/
│       ├── components/      # Внутренние UI-компоненты фичи (<= 7 файлов, каждый < 300 строк)
│       ├── hooks/           # Кастомные хуки бизнес-логики и запросов (use[Feature].ts)
│       ├── types.ts         # Локальные типы фичи
│       └── utils.ts         # Чистые функции-хелперы фичи
├── shared/
│   ├── components/ui/       # АТОМАРНЫЙ UI-KIT ПРОЕКТА (Button, Input, Card, Modal, Skeleton...)
│   ├── hooks/               # Универсальные хуки (useDebounce, useMediaQuery)
│   └── utils/               # Форматирование дат, валют, константы
└── api/                     # Типизированный API-клиент, сгенерированный из OpenAPI
```

### 🎨 Дизайн-паспорт проекта (Заполняется на старте через `BOOTSTRAP.md`)
> **Закон для фронтенд-агентов**: Запрещено придумывать свой визуальный стиль на отдельных страницах. Весь интерфейс строится строго по параметрам ниже.

* **Визуальный архетип и референсы (Vibe)**: `[Например: Минималистичный инженерный SaaS в духе Linear / Stripe. Высокий контраст, тонкие рамки 1px, акцент на читаемости данных без визуального мусора]`
* **Базовая тема**: `[Dark Mode First / Light Mode / Системная]`
* **Семантические цветовые токены (Tailwind / CSS Variables)**:
  > **ТАБУ**: Внутри `src/features/` и `src/pages/` **строго запрещено** использовать сырые палитры (`bg-blue-500`, `text-gray-400`, `#1a2b3c`). Это проверяется автоматически через `make check-arch`. Используются **только** семантические классы дизайн-системы:
  - Фон страницы и поверхностей: `bg-background`, `bg-surface`, `bg-surface-elevated`
  - Текст: `text-foreground` (основной), `text-muted` (вторичный)
  - Акценты и интерактив: `bg-primary`, `text-primary-foreground`, `hover:bg-primary/90`, `bg-secondary`
  - Границы и разделители: `border-subtle`, `border-focus`
  - Статусы: `text-status-success`, `text-status-warning`, `text-status-error`
* **Геометрия, Сетка и Типографика**:
  - Скругления: карточки и модалки — `[например: rounded-xl]`, кнопки и инпуты — `[например: rounded-lg]`.
  - Отступы: кратные `4px/8px` (`gap-2`, `gap-4`, `gap-6`, `p-4`, `p-6`).
  - Шрифты: основной UI — `[например: Inter, sans-serif]`, числа/код/ID — `[например: JetBrains Mono, monospace]`.

### Правило Атомарного UI-Kit (`src/shared/components/ui/`)
1. В **Фазе 0** создается базовый набор переиспользуемых компонентов в `src/shared/components/ui/`:
   `Button`, `Input`, `Select`, `Checkbox`, `Card`, `Badge`, `Modal`, `Skeleton`, `EmptyState`, `ErrorBanner`.
2. При разработке любой фичи в `src/features/*` субагенты **обязаны импортировать эти компоненты**, а не верстать кастомные `<button>` или `<input>` с нуля.
3. **4 обязательных состояния каждого экрана (Mandatory UI States)**:
   - `Loading`: аккуратный `Skeleton` из UI-Kit (без резких полноэкранных спиннеров).
   - `Error`: компонент `ErrorBanner` с понятным текстом и кнопкой «Повторить».
   - `Empty`: компонент `EmptyState` с иконкой, пояснением и кнопкой целевого действия.
   - `Success`: отображение данных с мобильной адаптивностью (`sm:`, `md:`, `lg:`).

---

## 🔒 4. Конвенции Безопасности и Надежности (Security by Design)

1. **Tenant Isolation & Защита от IDOR**: Каждый запрос к данным в сервисном слое обязан проверять владельца ресурса (`resource.OwnerID == currentUserID` или роль `admin`). Никогда не доверяй `user_id`, переданному в теле запроса или query-параметре — бери его только из верифицированного контекста сессии/JWT.
2. **Валидация входных данных**: Все входные DTO валидируются на уровне схемы OpenAPI и доменных конструкторов (ограничение длины строк, проверка форматов UUID/Email, числовых диапазонов).
3. **Защита БД и XSS**: Только параметризованные SQL-запросы (`$1, $2`). На фронтенде запрещено использование `dangerouslySetInnerHTML` без санитизации.
4. **Rate Limiting и Timeouts**: Публичные и ресурсоемкие эндпоинты защищаются Rate Limiter'ом (например, Token Bucket). Все внешние вызовы и запросы к БД обязаны передавать `context.Context` с таймаутом.

---

## 🧪 5. Автоматический архитектурный контроль (`make check-arch`) и Тестирование

Чтобы правила модульности не забывались субагентами, в проекте в **Фазе 0** настраивается скрипт `scripts/check_arch.sh`, вызываемый через `make check-arch`.

### Эталонный скрипт `scripts/check_arch.sh` (Устанавливается в Фазе 0):
```bash
#!/usr/bin/env bash
set -euo pipefail

MAX_LINES=400
MAX_FILES_PER_DIR=7
ERRORS=0

echo "🔍 [check-arch] 1. Проверка лимита строк в файлах (< ${MAX_LINES})..."
while IFS= read -r file; do
    # Пропускаем строго файлы автогенерации
    [[ "$file" == *.gen.go || "$file" == *_gen.go || "$file" == *.pb.go || "$file" == *.gen.ts || "$file" == *.gen.tsx || "$file" == *generated* ]] && continue
    lines=$(wc -l < "$file")
    if (( lines > MAX_LINES )); then
        echo "❌ FAIL [Размер файла]: $file содержит $lines строк (лимит: $MAX_LINES). Декомпозируйте файл!"
        ERRORS=$((ERRORS + 1))
    fi
done < <(find backend frontend \
    -not -path "*/node_modules/*" \
    -not -path "*/dist/*" \
    -not -path "*/build/*" \
    -not -path "*/.git/*" \
    -not -path "*/vendor/*" \
    -not -path "*/.next/*" \
    -type f \( -name "*.go" -o -name "*.ts" -o -name "*.tsx" \) 2>/dev/null || true)

echo "🔍 [check-arch] 2. Проверка количества файлов в доменных пакетах и папках фич (<= ${MAX_FILES_PER_DIR})..."
while IFS= read -r dir; do
    # Исключаем плоский каталог атомарного UI-Kit из лимита 7 файлов
    [[ "$dir" == *"shared/components/ui"* ]] && continue
    count=$(find "$dir" -maxdepth 1 -type f \( \
        -name "*.go" ! -name "*_test.go" ! -name "*.gen.go" ! -name "*_gen.go" -o \
        -name "*.ts" ! -name "*.test.ts" ! -name "*.spec.ts" ! -name "*.gen.ts" -o \
        -name "*.tsx" ! -name "*.test.tsx" ! -name "*.spec.tsx" ! -name "*.gen.tsx" \
    \) | wc -l)
    if (( count > MAX_FILES_PER_DIR )); then
        echo "❌ FAIL [Свалка файлов]: В директории $dir находится $count рабочих файлов (лимит: $MAX_FILES_PER_DIR). Выделите подпакет!"
        ERRORS=$((ERRORS + 1))
    fi
done < <(find backend/internal frontend/src/features \
    -not -path "*/node_modules/*" \
    -type d 2>/dev/null || true)

echo "🔍 [check-arch] 3. Проверка чистоты доменного слоя (Clean Architecture)..."
if [ -d "backend/internal/domain" ]; then
    if grep -rnE 'internal/(application|infrastructure|app)|net/http|database/sql|pgx' backend/internal/domain/ 2>/dev/null; then
        echo "❌ FAIL [Чистота Domain]: Слой internal/domain импортирует внешние слои или инфраструктурные пакеты!"
        ERRORS=$((ERRORS + 1))
    fi
fi

echo "🔍 [check-arch] 4. Проверка отсутствия сырых цветов Tailwind в фичах (Дизайн-паспорт)..."
if [ -d "frontend/src/features" ]; then
    if grep -rnE '\b(bg|text|border)-(red|blue|green|indigo|purple|pink|yellow|gray|slate|zinc|neutral|stone)-[0-9]{2,3}\b' frontend/src/features/ 2>/dev/null; then
        echo "❌ FAIL [Дизайн-паспорт]: В frontend/src/features/ обнаружены сырые цвета Tailwind! Используйте семантические токены (bg-surface, bg-primary, text-muted и т.д.)."
        ERRORS=$((ERRORS + 1))
    fi
fi

if (( ERRORS > 0 )); then
    echo "⛔ Архитектурная проверка НЕ ПРОЙДЕНА (ошибок: $ERRORS)."
    exit 1
fi
echo "✅ [check-arch] Архитектурные конвенции соблюдены (0 ошибок)!"
```

### Конвенции тестирования
1. **Unit-тесты (`make test`)**: Запускаются локально за секунды без внешних контейнеров, обязательно с флагом `-race` (`go test -race ./...`). Покрывают доменные инварианты и сервисный слой.
2. **E2E-тесты (`make test-e2e`)**: Запускаются в изолированном окружении поверх тестовой БД. Рекомендуемый стек: Python + **`uv`** (`uv run pytest`) + `httpx` в каталоге `tests/`. Проверяют полный цикл пользовательских сценариев и негативные сценарии безопасности (401/403/404).

---

## ⚙️ 6. Специфичные конвенции текущего проекта (Project-Specific Rules)

> *Этот раздел заполняется при старте проекта на Шаге 3 сценария `BOOTSTRAP.md` (индивидуальные требования CTO / разработчика для конкретного репозитория).*

* **Правило 1 (Работа с БД / ORM / Генераторы)**: `[Например: Используем чистый pgx v5 + sqlc, ORM GORM строго запрещен]`
* **Правило 2 (State Management и запросы на фронте)**: `[Например: Используем TanStack Query (React Query) для серверного стейта, Zustand для UI-стейта; Redux не используем]`
* **Правило 3 (E2E-тестирование)**: `[Например: Используем Python с менеджером uv (uv run pytest) и httpx в папке tests/]`
* **Правило 4 (Установка Go инструментов)**: `[Например: golang-migrate, mockery v3, oapi-codegen устанавливаются через go install в $HOME/go/bin]`
* **Правило 5 (Дополнительные ограничения)**: `[Любые ваши личные предпочтения по коду проекта]`
