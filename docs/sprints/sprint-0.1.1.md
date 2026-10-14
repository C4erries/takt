# Спринт 0.1.1: Инженерный фундамент, Архитектурный линтер и UI-Kit «Soft Editorial GTD»

* **Даты проведения**: 2026-10-14 — 2026-10-14
* **Статус**: ✅ Завершен (100% Definition of Done)
* **Связанные глобальные ADR**: [ADR-0001](../decisions/0001-initial-architecture-and-stack.md)

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

* **Бизнес-цель**: Развернуть надежный монорепозиторий, настроить автоматический архитектурный линтер `make check-arch`, мультифайловый OpenAPI, единый Docker Compose с Nginx и создать базовый атомарный UI-Kit в эстетике японской канцелярии («Soft Editorial GTD») для быстрого и изолированного подключения субагентов.
* **Ключевые продуктовые правила и Edge Cases**:
  - Палитра «Бамбук, Рисовая бумага и Матча»: `--bg-canvas` (`#E8E5DA`), `--bg-surface` (`#F7F5EE`), `--border-subtle` (`#DCD7C8`), `--text-primary` (`#21241E`), `--text-muted` (`#74786C`), `--accent-surface` (`#D6DEC8`), `--accent-solid` (`#566B44`), `--accent-solid-hover` (`#465737`), теги: `--tag-persimmon` (`#EBD5C3`), `--tag-denim` (`#CFD8E0`), `--tag-straw` (`#EAE4CC`).
  - Шрифты: `Newsreader` (Serif), `Geist Sans` (UI), `JetBrains Mono` (шорткаты/время).
  - SVG-шум бумаги (SVG-noise grain) на базовом фоне окна.
  - Пружинная физика J/K: `cubic-bezier(0.22, 1, 0.36, 1)`, 150ms.
* **Локальные технические решения**:
  - Строго 3 корневых подпакета в `backend/internal/`: `domain/`, `application/`, `infrastructure/`.
  - Composition Root вынесен в `backend/internal/infrastructure/container/` (тонкий `cmd/server/main.go` < 50 строк).
  - HTTP-доставка в `backend/internal/infrastructure/http/`, конфигурация в `backend/internal/infrastructure/config/`.
  - Мультифайловый OpenAPI 3.0 в `api/openapi/` со сборкой через `@redocly/cli bundle` в `api/dist/openapi.json`.
  - Сквозные E2E-тесты на Python (`pytest` + `httpx`) в каталоге `tests/`.

---

## 🎯 2. Ключевые результаты и Затронутые подпакеты

1. **Инфраструктура и Дев-окружение**:
   - `docker-compose.yml`: PostgreSQL 16 (с healthcheck), backend (:8080), frontend (:5173), Nginx (:80 reverse proxy).
   - `Makefile`: цели `check-arch`, `build`, `run`, `test`, `test-e2e`, `generate`, `docker-up`, `docker-down`.
   - `.env.example`, `.gitignore`.
2. **Архитектурный контроль (`scripts/check_arch.sh`)**:
   - Проверка 1: лимит строк `< 400` во всех файлах.
   - Проверка 2: лимит рабочих файлов в пакетах `<= 7`.
   - Проверка 3: чистота `internal/domain` (отсутствие импортов SQL, HTTP, application).
   - Проверка 4: запрет сырых цветов Tailwind в `features/` и `widgets/`.
   - Проверка 5: строго 3 корневых папки в `backend/internal/` (`domain`, `application`, `infrastructure`).
3. **API Контракты (`api/openapi/`)**:
   - `openapi.yaml`, `paths/health.yaml`, `schemas/health.yaml`, `schemas/error.yaml`.
4. **Backend (`backend/`)**:
   - `cmd/server/main.go`: точка входа (25 строк).
   - `internal/infrastructure/container/`: сборка `app.go`, `router.go`, `repos.go`, `services.go`.
   - `internal/infrastructure/http/`: хендлеры `/health` и `/api/v1/health`, middleware (CORS, Recover, RequestID, Logging, ErrorHandler), юнит-тесты `health_test.go`.
   - `internal/infrastructure/config/`: парсер настроек окружения.
   - `internal/infrastructure/db/`: пул `pgxpool.Pool` с методом `Ping()`.
   - `sqlc.yaml`, `migrations/000001_init_schema.up.sql`, `queries/health.sql`.
5. **Frontend и UI-Kit («Soft Editorial GTD»)**:
   - Vite + React 18 + TypeScript + Tailwind CSS.
   - Атомарные компоненты в `src/shared/components/ui/`: `Button`, `Input`, `Checkbox` (18px круг со стилем Midori), `Badge` (капсулы), `Card`, `Modal`, `Skeleton`, `EmptyState`, `ErrorBanner`.
   - Интерактивная витрина `src/app/App.tsx` с демонстрацией 4 состояний (Skeleton/Error/Empty/Success).
6. **E2E Тестирование (`tests/`)**:
   - `tests/requirements.txt`, `tests/conftest.py` (фикстура `api_client` с fallback), `tests/e2e/test_health.py`.

---

## 📊 3. Метрики качества, Архитектурный контроль и Безопасность

* **Архитектурный линтер (`make check-arch`)**:
  - `PASS (0 ошибок)`: 5/5 проверок удовлетворены на 100%.
* **Юнит-тесты бэкенда (`make test`)**:
  - `100% PASS` (`go test -v -race ./...` — гонок данных нет).
* **Сборка фронтенда (`cd frontend && npm run build`)**:
  - `0 ошибок, 0 предупреждений` (Vite production bundle собран за 3.19s).
* **Сквозные автотесты (`make test-e2e`)**:
  - `3/3 passed (100% PASS)`.
* **Аудит безопасности (Security Gate)**:
  - Безопасная обработка 500 ошибок без утечки деталей в HTTP-ответах.
  - CORS настроен без небезопасных wildcard заголовков с credentials.
  - Реальные секреты в репозиторий не добавлены.

