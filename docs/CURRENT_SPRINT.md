# Текущий спринт / Оперативный центр управления (`CURRENT_SPRINT.md`)

> Этот файл — **единственная точка правды** о том, какие решения приняты по текущей задаче и что команда и AI-агенты делают прямо сейчас.
> Работа над каждым спринтом строго делится на **Фазу А (Штаб: Принятие решений)** и **Фазу Б (Конвейер: Реализация и Верификация)**.
> Завершенные спринты архивируются в каталог `docs/sprints/sprint-X.Y.Z.md` с записью в реестр `docs/sprints/README.md`.

## Текущий спринт: Спринт 0.3.3 (Фаза 2: Движок событий EventBus и 3 Пресетные Автоматизации v1) ✅ ЗАВЕРШЕН (Готов к архивации)

---

## 🧠 1. Аналитика и Зафиксированные микро-решения спринта (Режим «Штаб»)

* **Бизнес-цель**: Реализовать асинхронный движок событий (`EventBus`), домен автоматизаций (`internal/domain/automation`), декларативное хранение правил в БД, панель тумблеров Вкл/Выкл 3 пресетов в UI и автосрабатывание правил при мутациях задач.
* **Зафиксированные продуктовые и архитектурные микро-решения**:
  1. **Архитектура движка событий (In-Memory EventBus)**:
     - Чистый интерфейс `EventBus` в `internal/domain/events`: публикация событий (`Publish(ctx, event)`) и подписка (`Subscribe(eventType, handler)`).
     - Реализация в `internal/infrastructure/eventbus`: in-memory шина с буферизованными каналами и пулом горутин-воркеров, контекстный graceful shutdown при остановке сервера.
     - HTTP-запросы клиентов не блокируются выполнением автоматизаций и завершаются мгновенно (< 5мс).
  2. **Мягкий «Inbox Zero»**:
     - Если задаче назначен `project_id`, и она была во «Входящих» (`column_id IS NULL`), автоматизация переносит её в первую колонку проекта со статусом `todo` («К выполнению»).
     - Если задаче назначен только дедлайн (`due_date`), она остается во «Входящих», но автоматически всплывает в представлении «Сегодня».
  3. **Защита от зацикливания (Recursion Loop Guard)**:
     - Ограничение глубины срабатывания `Max Depth = 1`.
     - Все действия, вызванные автоматизацией, помечаются маркером контекста (`context.WithValue` / `TriggeredByAutomation: true`), блокируя каскадный вызов повторных автоматизаций.
  4. **3 Пресетных правила и их начальное состояние при инициализации воркспейса**:
     - `Done & Archive` (`trigger: task_completed`, `action: move_to_column(done)`): **ВКЛЮЧЕНО (`is_enabled: true`)** по умолчанию.
     - `Auto-Focus` (`trigger: task_updated`, `condition: due_date <= today`, `action: set_focus_order(1)`): **ВКЛЮЧЕНО (`is_enabled: true`)** по умолчанию.
     - `Inbox Zero` (`trigger: task_updated`, `condition: project_id IS NOT NULL AND column_id IS NULL`, `action: move_to_first_column(todo)`): **ВЫКЛЮЧЕНО (`is_enabled: false`)** по умолчанию.
  5. **Двухфазная поставка Contract-First**:
     - **Шаг 1**: Миграция `000005_automations.up.sql`, sqlc-запросы `queries/automations.sql`, OpenAPI схемы (`schemas/automation.yaml`, `paths/automations.yaml`), `make generate`, чистый доменный слой `internal/domain/events` и `internal/domain/automation`.
     - **Шаг 2**: Реализация `internal/infrastructure/eventbus`, use cases `internal/application/automation`, публикация событий в `internal/application/task`, HTTP-хендлеры `automations_handler.go` под `ServerInterface`, фронтенд-панель `AutomationsToggle` и тосты.

---

## 📋 Таблица задач спринта 0.3.3

| # | Задача | Статус | Приоритет | Ответственный | Заметки и контракты |
|---|--------|--------|-----------|---------------|---------------------|
| 1 | **[Backend Step 1]**: Миграция 0005 (automations), sqlc-запросы, OpenAPI контракты и домен | ✅ Done | Критический | `backend_developer` | `automations` table, OpenAPI `GET /automations` & `PATCH /toggle`, `internal/domain/automation`, `internal/domain/events` |
| 2 | **[Gatekeeper Checkpoint 1]**: Проверка контрактов и кодогенерации | ✅ Done | Критический | `lead_architect` | `make check-arch && make generate && make test` (0 ошибок) |
| 3 | **[Backend Step 2]**: EventBus воркеры, use cases автоматизаций, эмит событий из задач, HTTP-хендлеры | ✅ Done | Высокий | `backend_developer` | `infrastructure/eventbus`, `application/automation`, `application/task` events, `automations_handler.go` |
| 4 | **[Frontend]**: Панель тумблеров 3 пресетов и визуальная индикация срабатывания | ✅ Done | Высокий | `frontend_developer` | `features/automations-toggle`, TanStack Query `useAutomationsQuery`, `useToggleAutomationMutation` |
| 5 | **[QA & Verification]**: E2E тесты 3 пресетных правил и Loop Guard на Pytest | ✅ Done | Критический | `qa_engineer` | `tests/e2e/test_automations.py`, `make check-arch`, `make test`, `make test-e2e` (25/25 PASS) |

---

## 🎯 Definition of Done (DoD)

- [x] Создана миграция `000005_automations.up.sql` с таблицей `automations` и сидированием 3 дефолтных правил при создании воркспейса.
- [x] Описаны контракты OpenAPI для автоматизаций (`GET /workspaces/{id}/automations`, `PATCH /workspaces/{id}/automations/{id}/toggle`).
- [x] Сгенерирован `api.gen.go` через `make generate`, HTTP-хендлеры строго реализуют `generated.ServerInterface`.
- [x] Реализован чистый домен `internal/domain/events` и `internal/domain/automation` с юнит-тестами.
- [x] Реализован асинхронный `internal/infrastructure/eventbus` с пулом горутин и graceful shutdown.
- [x] В `internal/application/task` встроена публикация доменных событий (`TaskCreated`, `TaskUpdated`, `TaskCompleted`, `TaskMoved`).
- [x] Реализованы слушатели автоматизаций в `internal/application/automation` с Recursion Loop Guard.
- [x] На фронтенде создана панель `AutomationsToggle` в эстетике Washi/Matcha с тумблерами 3 пресетов.
- [x] `make check-arch` — 0 ошибок по всем 5 проверкам.
- [x] `make test` — 100% PASS с флагом `-race`.
- [x] `cd frontend && npm run build` — 0 ошибок сборки.
- [x] Pytest E2E-тесты в `tests/e2e/test_automations.py` проверяют работу всех 3 пресетов и отсутствие зацикливания (8 новых тестов, всего 25/25 PASS).

---

## Лог изменений спринта

| Дата | Что изменилось |
|------|---------------|
| 2026-10-14 | **Спринт 0.3.3 завершен**: Реализован асинхронный движок событий (`internal/domain/events` + `internal/infrastructure/eventbus`), домен и use cases автоматизаций (`internal/domain/automation` + `internal/application/automation`), 3 пресетных правила («Done & Archive», «Auto-Focus», «Inbox Zero») с защитой от зацикливания (`Recursion Loop Guard`), модальное окно управления правилами `AutomationsModal` в эстетике Washi/Matcha и 8 сквозных E2E-тестов на Pytest (25/25 PASS). |
