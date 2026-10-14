# Текущий спринт / Оперативный центр управления (`CURRENT_SPRINT.md`)

> Этот файл — **единственная точка правды** о том, какие решения приняты по текущей задаче и что команда и AI-агенты делают прямо сейчас.
> Работа над каждым спринтом строго делится на **Фазу А (Штаб: Принятие решений)** и **Фазу Б (Конвейер: Реализация и Верификация)**.
> Завершенные спринты архивируются в каталог `docs/sprints/sprint-X.Y.Z.md` с записью в реестр `docs/sprints/README.md`.

## Текущий спринт: Спринт 0.1.1 (Этап 0.1: Инженерный фундамент, Архитектурный линтер и UI-Kit «Soft Editorial GTD») ✅ ЗАВЕРШЕН (Готов к архивации)

---

## 🧠 ФАЗА А: ШТАБ — Аналитика и Принятые решения (Роль: CEO, CTO & Design Lead)

### 1. Бизнес-обоснование (CEO View)
* **Проблема**: Запуск нового продукта Takt требует надежного, скоростного и эстетически выверенного фундамента. Без четких границ пакетов, линтера архитектуры и готового UI-Kit агенты начнут верстать страницы вразнобой сырыми цветами и сваливать код в монолитные файлы, вызывая перегрузку контекста и ошибки.
* **Решение**: Развернуть каркас монорепозитория, настроить автоматический архитектурный линтер `make check-arch`, мультифайловый OpenAPI, единый Docker Compose с Nginx и создать базовый атомарный UI-Kit в эстетике японской канцелярии («Soft Editorial GTD»).
* **Бизнес-эффект / Критерий успеха**: Любой субагент может мгновенно разрабатывать изолированные фичи, опираясь на готовые токены, интерфейсы и компоненты, а система собирается и проверяется одной командой `make check-arch && make test`.

### 2. Принятые микро-решения и технические решения спринта (Micro-Decisions)
* **Стек Backend**: Go 1.23+, роутер Echo v4. Тонкий `cmd/server/main.go` (< 50 строк), точка сборки зависимостей в `internal/infrastructure/container/`. База данных PostgreSQL 16 с пулом `pgxpool`, транзакции через контекст (`Transactor`), генератор запросов `sqlc` (`sqlc.yaml`).
* **Контракты API**: Мультифайловый OpenAPI 3.0 в `api/openapi/` (разделение на `paths/` и `schemas/`), сборка через `@redocly/cli bundle` и генерация хендлеров через `oapi-codegen`.
* **Стек Frontend**: React 18/19, TypeScript, Vite, Tailwind CSS. Шрифты Google Fonts (`Newsreader` Serif, `Geist Sans` UI, `JetBrains Mono`). Наложение тактильного SVG-шума бумаги на `--bg-canvas`. Пружинная физика J/K переключения задач (`cubic-bezier(0.22, 1, 0.36, 1)`, 150ms).
* **UI-Kit**: Атомарные компоненты строго по Дизайн-паспорту в `frontend/src/shared/components/ui/` (`Button`, `Input`, `Checkbox` 18px круглый со стилями Midori, `Card`, `Badge` капсульный, `Modal`, `Skeleton`, `EmptyState`, `ErrorBanner`).
* **Архитектурный контроль**: Скрипт `scripts/check_arch.sh` проверяет лимит строк (< 400), лимит файлов в пакетах (<= 7), чистоту слоев DDD и отсутствие сырых цветов Tailwind в `features/` и `widgets/`.
* **QA & Инфраструктура**: Docker Compose (Postgres, Backend, Frontend, Nginx), Makefile, каталог `tests/` с тест-раннером `pytest` + `httpx`.

### 3. Архитектура изменений и затрагиваемые директории
* **Инфраструктура**: `docker-compose.yml`, `Makefile`, `deploy/nginx/default.conf`, `scripts/check_arch.sh`.
* **API Контракты**: `api/openapi/openapi.yaml`, `api/openapi/paths/health.yaml`, `api/openapi/schemas/error.yaml`.
* **Backend**: `backend/cmd/server/main.go`, `backend/internal/infrastructure/container/`, `backend/internal/infrastructure/config/`, `backend/internal/infrastructure/http/`, `backend/sqlc.yaml`.
* **Frontend**: `frontend/src/app/`, `frontend/src/shared/components/ui/`, `frontend/tailwind.config.js`.
* **Tests**: `tests/requirements.txt`, `tests/conftest.py`, `tests/e2e/test_health.py`.

---

## 📋 ФАЗА Б: КОНВЕЙЕР — Таблица задач спринта

| # | Задача | Статус | Приоритет | Ответственный / Субагент | Заметки и контракты |
|---|--------|--------|-----------|---------------------------|---------------------|
| 1 | Каркас монорепозитория, Makefile и Docker Compose с Nginx | ✅ Done | Критический | `devops_engineer` | `docker-compose.yml`, `deploy/nginx/default.conf`, `Makefile` |
| 2 | Скрипт архитектурного линтера `scripts/check_arch.sh` | ✅ Done | Критический | `lead_architect` | Проверка строк < 400, файлов <= 7, слоев DDD, токенов |
| 3 | Каркас мультифайлового OpenAPI 3.0 и Redocly бандлер | ✅ Done | Высокий | `backend_developer` | `api/openapi/*`, скрипт бандлинга в Makefile |
| 4 | Каркас Go-сервиса (Echo v4, config, infrastructure/container, sqlc.yaml) | ✅ Done | Критический | `backend_developer` | Тонкий main.go, DI-сборка, healthcheck эндпоинт |
| 5 | Фронтенд-каркас Vite + React + Tailwind (Editorial токен-паспорт) | ✅ Done | Критический | `frontend_developer` | Палитра Washi/Matcha, шрифты Newsreader/Geist, SVG grain |
| 6 | Базовый атомарный UI-Kit в `shared/components/ui/` | ✅ Done | Критический | `frontend_developer` | Button, Input, Checkbox (18px круг), Card, Badge, Modal, Skeleton... |
| 7 | Настройка Pytest E2E тест-контура в `tests/` | ✅ Done | Высокий | `qa_engineer` | `requirements.txt`, `conftest.py`, `test_health.py` |

> **Статусы задач**: `🔲 Planned` (Запланировано) · `🔵 In Progress` (В работе) · `🛑 Blocked` (Блокер) · `✅ Done` (Завершено)

---

## 🛠️ Детальное Микро-ТЗ спринта

### Задача 1: Каркас монорепозитория, Docker Compose и Nginx
* **Файлы**: `docker-compose.yml`, `Makefile`, `deploy/nginx/default.conf`, `deploy/Dockerfile.backend`, `deploy/Dockerfile.frontend`.
* **Требования**:
  - `docker compose up` поднимает PostgreSQL 16 (с healthcheck), backend на порту 8080, frontend dev-сервер и Nginx на порту 80 (проксирует `/api/*` $\rightarrow$ backend:8080, `/*` $\rightarrow$ frontend).
  - В `Makefile` добавлены команды: `check-arch`, `build`, `run`, `test`, `test-e2e`, `generate`.

---

### Задача 2: Скрипт архитектурного контроля `scripts/check_arch.sh`
* **Файлы**: `scripts/check_arch.sh` (права `chmod +x`).
* **Требования**:
  - Проверка лимита строк в файлах: `< 400 строк` (пропуск `*.gen.*`).
  - Проверка количества рабочих файлов в пакетах `backend/internal/` и `frontend/src/features/`: `<= 7 файлов`.
  - Проверка чистоты `internal/domain`: отсутствие импортов HTTP, SQL, pgx, echo.
  - Проверка отсутствия сырых цветов Tailwind в `frontend/src/features` и `widgets`.

---

### Задача 3: Мультифайловый OpenAPI манифест (`api/openapi/`)
* **Файлы**: `api/openapi/openapi.yaml`, `api/openapi/paths/health.yaml`, `api/openapi/schemas/error.yaml`, `api/openapi/schemas/health.yaml`.
* **Требования**:
  - Валидная спецификация OpenAPI 3.0.
  - Команда `make generate` собирает единый бандл через `@redocly/cli` и генерирует роуты через `oapi-codegen`.

---

### Задача 4: Go-сервис (Echo v4 + Clean Architecture)
* **Пакеты**: `backend/cmd/server/main.go`, `backend/internal/infrastructure/container/`, `backend/internal/infrastructure/config/`, `backend/internal/infrastructure/http/`, `backend/sqlc.yaml`.
* **Требования**:
  - `main.go` < 50 строк: только чтение конфигурации и вызов `container.Run(ctx, cfg)`.
  - `internal/infrastructure/container/` разделен на логические файлы (`container.go`, `router.go`).
  - Эндпоинты `/health` и `/api/v1/health` возвращают статус `200 OK` и статус подключения к БД.
  - Настроен `sqlc.yaml` для генерации кода из `backend/queries/` с драйвером `pgx/v5`.

---

### Задача 5 & 6: Фронтенд-каркас и Атомарный UI-Kit «Soft Editorial GTD»
* **Пакеты**: `frontend/src/shared/components/ui/`, `frontend/tailwind.config.js`, `frontend/src/index.css`.
* **Требования**:
  - В `tailwind.config.js` заведены семантические токены: `--bg-canvas` (`#E8E5DA`), `--bg-surface` (`#F7F5EE`), `--border-subtle` (`#DCD7C8`), `--text-primary` (`#21241E`), `--text-muted` (`#74786C`), `--accent-surface` (`#D6DEC8`), `--accent-solid` (`#566B44`), теги (`--tag-persimmon`, `--tag-denim`, `--tag-straw`).
  - Подключены шрифты `Newsreader`, `Geist Sans`, `JetBrains Mono`.
  - Реализованы компоненты в `src/shared/components/ui/`:
    1. `Button`: варианты solid (чай матча), outline (тонкая 1px рамка), ghost;
    2. `Input`: мягкая подложка, 1px граница;
    3. `Checkbox`: круглый 18px с теплым серо-коричневым бордером, при чеке — заливка матча с белой галочкой;
    4. `Badge`: капсульная форма (rounded-full) с пастельными тонами;
    5. `Card`: подложка рисовой бумаги Washi (`--bg-surface`), тонкая 1px граница, деликатная тень;
    6. `Modal`: диалоговое окно без тяжелого блюра, с мягким затемнением;
    7. `Skeleton`: деликатная пульсация теплого оттенка;
    8. `EmptyState`: спокойная контурная иконка, заголовок Serif, пояснение и кнопка действия;
    9. `ErrorBanner`: плашка сообщения об ошибке с кнопкой повтора.

---

### Задача 7: Настройка E2E-тестов на Python (Pytest)
* **Пакеты**: `tests/requirements.txt`, `tests/conftest.py`, `tests/e2e/test_health.py`.
* **Требования**:
  - Фикстуры `api_client` на базе `httpx.Client` с базовым URL сервиса.
  - Тест `test_health_endpoint` проверяет успешный ответ `200 OK` от `/api/v1/health`.
  - Запуск через `make test-e2e` (или локально `pytest tests/`).

---

## 🎯 Definition of Done (DoD) Спринта

- [x] **Архитектурный линтер (`make check-arch`) проходит с 0 ошибок**: все файлы `< 400 строк`, в пакетах `<= 7 файлов`, чистые слои DDD, нет сырых цветов Tailwind.
- [x] Сервис и база данных успешно стартуют в `docker compose up` / контейнерах.
- [x] Эндпоинт `/api/v1/health` отвечает `200 OK`.
- [x] UI-Kit задокументирован и доступен на интерактивной демо-витрине `src/app/App.tsx`.
- [x] `pytest tests/` (`make test-e2e`) отрабатывает с результатом `100% PASS` (3/3 теста).
- [x] Карта пакетов в `PROJECT_CONTEXT.md` актуальна.

---

## Лог изменений спринта

| Дата | Что изменилось |
|------|---------------|
| 2026-10-14 | **Фаза А (Штаб)**: завершен этап инициализации BOOTSTRAP.md, сформирован Спринт 0 (Фаза 0: Инженерный фундамент, Архитектурный линтер и UI-Kit «Soft Editorial GTD»). |
| 2026-10-14 | **Фаза Б (Конвейер)**: Завершены Задачи 1–7 Спринта 0.1.1. Независимый гейткипинг пройден: `make check-arch` (0 ошибок), `make test` (PASS), `make test-e2e` (3/3 PASS), сборка фронтенда и бэкенда (`make build`) успешна. |
