# Принципы и инженерные конвенции проекта (`CONVENTIONS.md`)

> Этот документ — **свод инженерных и дизайнерских законов** проекта **Takt** для всех, кто пишет код в репозитории (включая разработчиков и AI-агентов).
> Соблюдение этих правил контролируется автоматически скриптом **`make check-arch`** и главным агентом (Gatekeeper).

---

## 🏗️ 1. Общие принципы разработки и Модульность (Защита контекста ИИ)

### 12-Factor App (Практическая адаптация)
1. **Конфигурация через окружение**: Все настройки, порты, DSN баз данных и секреты читаются через переменные окружения (`env`). В Git хранится только `.env.example`; реальные `.env` файлы **никогда не коммитятся**.
2. **Stateless сервисы**: Серверные процессы не хранят состояние в памяти процесса (состояние делегируется в PostgreSQL, а в будущем — в Redis/Kafka).
3. **Структурированное логирование**: Логи пишутся в `stdout`/`stderr` (JSON в продакшне, структурированный цветной лог в dev) без утечки паролей, токенов и персональных данных.
4. **Dev/Prod Parity**: Все окружение (PostgreSQL, Backend, Frontend, Nginx) поднимается одной командой `docker compose up`.

### Двойной лимит гранулярности (Против «свалки файлов» и перегрузки LLM)
1. **Лимит размера файла: строго `< 300–400 строк`**
   * *Почему*: Файлы свыше 400 строк перегружают контекстное окно агентов, вызывают `truncated output` и повышают риск сломать соседний код.
   * *Правило*: При приближении файла к 300 строкам он **обязан** быть декомпозирован. Единственное исключение — автосгенерированный код (`*.gen.go`, `*_gen.go`, `*.gen.ts`).
2. **Лимит размера пакета / папки: строго `<= 5–7 рабочих файлов`**
   * *Почему*: Когда в одной папке (`internal/domain` или `internal/application`) скапливается 20–30 файлов, в Go происходит загрязнение пространства имен пакета (конфликты приватных функций и ошибок), а агенты тратят лишние токены на поиск нужного кода.
   * *Правило*: **Запрещено сваливать файлы разных сущностей в один общий пакет.** Каждый архитектурный слой обязательно нарезается на **доменные подпакеты** (`workspace`, `project`, `task`, `automation`). В одной доменной директории не может находиться более 5–7 рабочих файлов (не считая `*_test.go`). Исключение: плоский каталог базовых UI-примитивов `frontend/src/shared/components/ui`.

### Правила работы с Git
* **Ветки**: `main` (стабильный релиз), `dev` (ветка текущего спринта), `feature/<name>` / `fix/<name>`.
* **Абсолютное табу для AI-агентов**:
  > Агентам **СТРОГО ЗАПРЕЩЕНО** выполнять любые мутирующие команды Git (`git commit`, `git push`, `git checkout`, `git stash`, `git rebase`), если в текущем промпте пользователя нет прямого указания. Разрешены только read-only команды (`git status`, `git diff`, `git log`).
* При явном запросе коммита используется формат **Conventional Commits** (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`).

---

## 🐹 2. Backend конвенции (Clean Architecture / DDD + Подпакеты без `import cycle`)

### Архитектурные слои и структура подпакетов (СТРОГО 3 пакета в internal/)

> **Закон архитектуры**: В директории `backend/internal/` разрешены **СТРОГО 3 корневых подпакета**: `domain`, `application`, `infrastructure`.  
> Создавать другие папки на верхнем уровне `internal/` (например, `delivery/`, `config/`, `app/`, `container/`) **запрещено** и блокируется скриптом `make check-arch`. Вся инфраструктура, внешние адаптеры и точка сборки приложения (`container/` или `di/`) размещаются строго внутри подпакетов `infrastructure/*`.

```text
backend/internal/
├── domain/              # Чистый домен (сущности, инварианты, доменные ошибки и события)
│   ├── shared/          # Базовые примитивы (ID, Pagination, системные доменные ошибки)
│   ├── workspace/       # entity.go, errors.go, repository.go
│   ├── project/         # project.go, board.go, column.go, repository.go
│   ├── task/            # task.go, tag.go, errors.go, repository.go
│   ├── automation/      # rule.go, evaluator.go, repository.go
│   └── events/          # task_events.go, eventbus.go
├── application/         # Сценарии использования (Use Cases) + интерфейсы зависимостей (ports.go)
│   ├── workspace/       # usecases.go, ports.go
│   ├── project/         # usecases.go, ports.go
│   ├── task/            # usecases.go, ports.go
│   └── automation/      # runner.go, ports.go
└── infrastructure/      # ВСЕ внешние адаптеры, драйверы и Composition Root
    ├── container/       # DI и сборка приложения (repos.go, services.go, router.go, container.go)
    ├── config/          # Загрузка и валидация .env (config.go)
    ├── http/            # Echo роутер, middleware, oapi-codegen хендлеры
    ├── db/              # pgxpool, транзактор через Context, sqlc queries
    ├── eventbus/        # In-memory диспетчер событий (задел под Kafka)
    └── repository/      # Реализации репозиториев поверх sqlc
```

| Слой | Назначение | Что может импортировать | Чего НЕ должно быть внутри |
|---|---|---|---|
| **`domain/<pkg>`** | Чистые бизнес-сущности, Value Objects, доменные ошибки и интерфейсы репозиториев своего домена. | Только стандартную библиотеку и `internal/domain/shared`. | Никаких импортов SQL, HTTP, драйверов и **никаких импортов соседних `domain/<other>`**. |
| **`application/<pkg>`** | Сценарии использования (Use Cases), оркестрация доменной логики, проверка прав доступа. | Свой `internal/domain/<pkg>` и `internal/domain/shared`. | Никаких прямых SQL-запросов, HTTP-роутеров и **никаких прямых импортов соседних `application/<other>`**. |
| **`infrastructure/<pkg>`** | Адаптеры внешнего мира (PostgreSQL, sqlc, Echo, EventBus, Config) и сборка (`container/`). | `domain/*` + `application/*` + сторонние библиотеки (pgx, echo). | Здесь живут SQL-запросы, маппинг DTO, роутинг, чтение ENV и DI-сборка (`container/`). Чистая бизнес-логика здесь запрещена. |

### Защита от циклических зависимостей в Go (`No Import Cycles Rule`) и `ports.go`
1. **Интерфейсы объявляются на стороне потребителя (`ports.go`)**:
   Пакет `application/automation` **НЕ ИМЕЕТ ПРАВА** напрямую импортировать пакет `application/task`. Если сервису автоматизаций нужно изменить статус задачи, он объявляет в своем файле `internal/application/automation/ports.go` локальный интерфейс:
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

### Транзакции через `context.Context` (Unit of Work)
* Управление транзакциями строится через интерфейс `Transactor`:
  ```go
  type Transactor interface {
      WithinTx(ctx context.Context, fn func(txCtx context.Context) error) error
  }
  ```
* Сервисный слой оборачивает атомарные операции в `transactor.WithinTx(ctx, ...)`.
* Репозитории БД на базе `sqlc` извлекают `pgx.Tx` из переданного контекста и используют `queries.WithTx(tx)`. Если транзакции в контексте нет — запрос выполняется штатно через пул `pgxpool.Pool`.

---

## 🎨 3. Frontend конвенции и Дизайн-паспорт («Soft Editorial GTD»)

### Структура каталогов (Feature-Sliced + Лимит `<= 7 файлов в папке фичи`)
```text
frontend/src/
├── app/                     # Инициализация, провайдеры (QueryClient), темы, роутер
├── pages/                   # Тонкие страницы-контейнеры (до 100 строк, композиция виджетов)
├── widgets/                 # Крупные блоки 3-колоночного каркаса:
│   ├── sidebar/             # Левая навигация (Horizons GTD, Projects) — 220–250px
│   ├── main-sheet/          # Центральный "бумажный лист" (Surface-on-Canvas) — 1fr (max 880px)
│   ├── task-inspector/      # Правый верхний Bento-блок деталей выбранной задачи — 280–320px
│   └── mini-calendar/       # Правый нижний Bento-блок матричного календаря
├── features/                # Интерактивные фичи (quick-capture, task-focus, automations-toggle)
│   └── [feature_name]/
│       ├── components/      # Внутренние UI-компоненты фичи (<= 7 файлов, каждый < 300 строк)
│       ├── hooks/           # Кастомные хуки (use[Feature].ts)
│       └── types.ts
├── entities/                # Доменные UI-сущности (task, project, workspace)
├── shared/
│   ├── components/ui/       # АТОМАРНЫЙ UI-KIT ПРОЕКТА (Button, Input, Checkbox, Badge, Modal...)
│   ├── hooks/               # Универсальные хуки (useDebounce, useHotkeys)
│   ├── lib/                 # Утилиты форматирования, cn хелпер
│   └── api/                 # Сгенерированный клиент OpenAPI
└── tailwind.config.js       # Токены палитры Midori/Washi и типографика
```

### 🎨 Дизайн-паспорт проекта: Концепт «Soft Editorial GTD» (Японская канцелярия Midori MD / Hobonichi + Claude Tech)

> **Закон для фронтенд-агентов**: Запрещено верстать экраны в неоново-стеклянном или «космическом» стиле. Полный отказ от `backdrop-filter: blur()`, кричащих неоновых градиентов и тяжелых теней. Ощущение утреннего кофе за деревянным столом, фактурной бумаги и чернил.

#### 1. Семантические цветовые токены (Палитра «Бамбук, Рисовая бумага и Матча»)
> **ТАБУ**: Внутри `features/` и `widgets/` **строго запрещено** использовать сырые цвета Tailwind (`bg-blue-500`, `text-gray-400`, `bg-zinc-900`). Это проверяется автоматически через `make check-arch`. Используются **только** семантические переменные:

```css
:root {
  /* Поверхности и Холст */
  --bg-canvas: #E8E5DA;         /* Сухой бамбук / неотбеленный лён (базовый фон окна) */
  --bg-surface: #F7F5EE;        /* Рисовая бумага Washi (центральный лист, карточки инспектора) */
  --border-subtle: #DCD7C8;     /* Тонкое рисовое волокно (1px разделители строк и таблиц) */

  /* Текст и чернила */
  --text-primary: #21241E;      /* Чернила Суми с мягким тёплым подтоном */
  --text-muted: #74786C;        /* Сушеный чайный лист (вторичный текст, подписи) */

  /* Акценты и фокус */
  --accent-surface: #D6DEC8;    /* Молочная матча (мягкая заливка фокусной задачи) */
  --accent-solid: #566B44;      /* Густой зеленый чай матча (основная кнопка, активный чекбокс) */
  --accent-solid-hover: #465737;

  /* Пастельные контекстные теги (Pills) */
  --tag-persimmon: #EBD5C3;     /* Хурма / тёплый персиковый (@дом) */
  --tag-denim: #CFD8E0;         /* Индиго деним / приглушенный серо-голубой (проект) */
  --tag-straw: #EAE4CC;         /* Рисовая солома (время: 20 мин) */
}
```

#### 2. Сетка и компоновка (Layout & Spatial System)
* **3-колоночный каркас «Surface-on-Canvas»**:
  - Высота фиксирована: `100vw / 100vh` без глобального скролла окна (скроллятся только внутренние контейнеры).
  - **Левый сайдбар (220–250px)**: лежит прямо на `--bg-canvas` без внешних рамок, давая воздух глазу.
  - **Центральный лист (Main Sheet, max 880px)**: выглядит как отдельный лист бумаги Washi (`--bg-surface`), внешние отступы `margin: 12px 0`, мягкое скругление `border-radius: 16px`, граница `1px solid var(--border-subtle)` и едва заметная деликатная тень `0 1px 3px rgba(0,0,0,0.03)`. Внутренние поля щедрые: `padding: 32px 40px`.
  - **Правая панель (280–320px)**: два Bento-блока с `gap: 12px` (инспектор задачи сверху + матричный календарь снизу).
* **Тактильный SVG-шум (Paper Grain)**: На `--bg-canvas` накладывается сверхлегкая текстура шума (SVG-noise с `opacity: 0.015–0.02`), убирающая эффект плоского экрана.

#### 3. Типографика (Стиль Claude / Kinfolk)
* **Книжный Serif для заголовков (Today, Inbox, Проекты)**:
  - Шрифт: **`Newsreader`** (или `Instrument Serif`, Google Fonts).
  - Параметры: `font-size: 28px–34px`, `font-weight: 400`, `letter-spacing: -0.02em`, `line-height: 1.15`.
* **Основной UI и текст задач (Sans-Serif)**:
  - Шрифт: **`Geist Sans`** (или `Plus Jakarta Sans` / `Inter`).
  - Текст задачи: `15px`, `font-weight: 450`, цвет `--text-primary`.
  - Подзаголовки групп (`⚡ FOCUS OF THE DAY`, `📁 ВХОДЯЩИЕ`): `uppercase`, `11px`, `font-weight: 600`, `letter-spacing: 0.08em`, цвет `--text-muted`.
* **Моноширинный шрифт (Monospace)**:
  - Шрифт: **`JetBrains Mono`** — для времени (`20 мин`), тайм-слотов и клавиатурных хоткеев (`kbd`).

#### 4. Анатомия компонентов (UI-Kit)
* **Строка задачи (Task Row)**:
  - Высота `42px–46px`, горизонтальные тонкие разделители `1px solid var(--border-subtle)`.
  - Фокусная активная задача: мягкая заливка `--accent-surface` со скруглением `8px–10px`, отступом `padding: 0 12px` и быстрой пружинной анимацией (`cubic-bezier(0.22, 1, 0.36, 1)`, 150ms).
  - Поддержка мгновенной клавиатурной навигации `J / K` (как в Linear).
* **Чекбокс**:
  - Круглый `18px × 18px` с границей `1.5px solid var(--border-subtle)`.
  - При закрытии задачи: мягкая заливка `--accent-solid` с белой галочкой внутри.
* **Теги и контексты (Capsule Pills)**:
  - Форма капсулы (`border-radius: 9999px`), `padding: 2px 10px`, шрифт `12px` (`font-weight: 500`).
  - Пастельный фон с темным контрастным текстом внутри.

---

## 🔒 4. Конвенции Безопасности и Надежности (Security by Design)

1. **Tenant Isolation & Защита от IDOR**: Каждый запрос к данным проверяет принадлежность ресурса к активному `workspace_id`. Запрещено доверять `user_id` или `workspace_id` из query-параметров без верификации членства.
2. **Параметризованный SQL**: Вся работа с базой ведется через `sqlc` и `pgx/v5` с позиционными параметрами (`$1, $2`). Прямая конкатенация строк в SQL запрещена.
3. **Безопасная обработка ошибок**: Доменные ошибки транслируются HTTP-мидлварью в чистые коды (`400`, `401`, `403`, `404`, `429`). Любые системные паники и внутренние ошибки БД отдаются наружу как обезличенный `500 Internal Server Error`, а детали пишутся в stderr.

---

## 🧪 5. Автоматический архитектурный контроль (`make check-arch`) и Тестирование

Скрипт `scripts/check_arch.sh` запускается через `make check-arch` и проверяет:
1. Лимит строк во всех файлах: `< 350–400 строк`.
2. Лимит файлов в пакетах `backend/internal/domain/*`, `internal/application/*`, `frontend/src/features/*`: `<= 7 файлов`.
3. Чистоту `internal/domain`: отсутствие импортов HTTP, SQL, application или внешних адаптеров.
4. Отсутствие сырых цветов Tailwind в `frontend/src/features/` и `widgets/`.

### Стратегия тестирования
1. **Unit-тесты Go (`make test`)**: запуск с флагом `-race` (`go test -race ./...`), моки через `mockery v3`.
2. **E2E-тесты на Python (`tests/` с `pytest` и `httpx`)**: сквозные пользовательские сценарии через реальный API (регистрация $\rightarrow$ создание задачи $\rightarrow$ срабатывание автоматизации $\rightarrow$ проверка статуса).

---

## ⚙️ 6. Специфичные конвенции проекта Takt (Project-Specific Rules)

* **Правило 1 (Go + Echo + sqlc + pgx v5)**:
  Роутер — **Echo v4**. Работа с БД — строго **`sqlc` + `pgx/v5`**. Запросы хранятся в `backend/queries/*.sql`. ORM (GORM и аналоги) строго запрещены.
* **Правило 2 (Транзакции через context.Context)**:
  Транзакции передаются через Unit of Work / `transactor.WithinTx(ctx, func(txCtx context.Context) error)`, а `sqlc`-репозитории извлекают транзакцию и используют `q.WithTx(tx)`.
* **Правило 3 (Интерфейсы по месту использования)**:
  Каждый пакет потребителя объявляет свои собственные минимальные интерфейсы зависимостей в `ports.go`. Горизонтальные импорты между соседними подпакетами запрещены.
* **Правило 4 (Мультифайловый OpenAPI)**:
  API спецификация хранится разбитой по файлам в `api/openapi/` (`paths/`, `schemas/`). Сборка выполняется через `@redocly/cli bundle`, генерация Echo-хендлеров — через `oapi-codegen`.
* **Правило 5 (State Management)**:
  На фронтенде для серверных данных используется **`TanStack Query`** с оптимистичными апдейтами; для клиентского UI — **`Zustand`**. Redux не используется.
* **Правило 6 (E2E тесты на Pytest + uv)**:
  Вся E2E-приемка сервиса ведется тест-сьютом на Python (`pytest` + `httpx`) в папке `tests/` с использованием менеджера **`uv`** (`uv run pytest` / `uv sync`) для мгновенного развёртывания и изоляции окружения.
* **Правило 7 (Установка Go CLI инструментов)**:
  Инструменты кодогенерации и миграций устанавливаются строго через `go install`:
  - `golang-migrate`: `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`
  - `mockery v3`: `go install github.com/vektra/mockery/v3@latest`
  - `oapi-codegen`: `go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest`
  Бинарники размещаются в `$HOME/go/bin`.
* **Правило 8 (Задел для конструктора автоматизаций v2)**:
  В коде v1 правила автоматизаций хранятся декларативно в таблице `automations` (`trigger_type`, `conditions_json`, `actions_json`). В v2 конструктор автоматизаций может проектироваться как hybrid/low-code/скриптовый, что закладывается в архитектуру движка событий уже сейчас.
