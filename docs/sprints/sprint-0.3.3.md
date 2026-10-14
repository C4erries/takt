# Спринт 0.3.3: Движок событий (EventBus) и 3 Пресетные Автоматизации v1

* **Даты проведения**: 2026-10-14 — 2026-10-14
* **Статус**: ✅ Завершен (100% Definition of Done)
* **Связанные документы**: [docs/CURRENT_SPRINT.md](../CURRENT_SPRINT.md), [docs/conventions/backend.md](../conventions/backend.md), [docs/conventions/frontend.md](../conventions/frontend.md)

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

1. **Асинхронный in-memory `EventBus`**:
   - Интерфейс `EventBus` в чистом домене `internal/domain/events` (`Publish(ctx, event)`, `Subscribe(eventName, handler)`).
   - Инфраструктурная реализация `internal/infrastructure/eventbus` с пулом из 4 воркеров-горутин, буферизованными каналами и graceful shutdown (`Close(ctx)`).
   - Запросы клиентов не блокируются выполнением фоновых правил и возвращаются моментально (< 5мс).
2. **Мягкий «Inbox Zero»**:
   - Перемещение задачи во входящих (`column_id IS NULL`) в первую колонку проекта со статусом `todo` («К выполнению») происходит только при явном назначении `project_id`.
   - Задачи, которым назначен только дедлайн без проекта, остаются во «Входящих», но автоматически всплывают в представлении «Сегодня».
3. **Защита от зацикливания (Recursion Loop Guard)**:
   - Ограничение глубины срабатывания `Max Depth = 1`.
   - Все действия, вызванные автоматизацией, помечаются маркером контекста (`context.WithValue` / `TriggeredByAutomation: true` и `IsFromAutomation: true`), блокируя каскадный вызов повторных автоматизаций.
4. **Стартовое состояние 3 правил для воркспейса**:
   - `Done & Archive` (`task_completed` $\rightarrow$ перенос в `done` с фиксацией `completed_at`): **ВКЛЮЧЕНО (`true`)**.
   - `Auto-Focus` (`task_updated` с `due_date <= today` $\rightarrow$ `focus_order = 1`): **ВКЛЮЧЕНО (`true`)**.
   - `Inbox Zero` (`task_updated` с `project_id` и `column_id IS NULL` $\rightarrow$ перенос в первую колонку `todo`): **ВЫКЛЮЧЕНО (`false`)**.
5. **Автоматическое сидирование правил**:
   - При регистрации нового пользователя и создании воркспейса автоматически генерируются 3 пресетных правила.

---

## 🎯 2. Выполненные задачи и Затронутые подпакеты

1. **Миграции БД и SQL**:
   - `backend/migrations/000005_automations.up.sql`: таблица `automations`, составной индекс `(workspace_id, trigger_type, is_enabled)`, сидирование 3 правил для всех существующих воркспейсов.
   - `backend/queries/automations.sql`: sqlc-запросы `ListWorkspaceAutomations`, `GetAutomationByID`, `ToggleAutomation`, `CreateAutomation`, `ListEnabledAutomationsByTrigger`.
2. **OpenAPI Контракты и Кодогенерация**:
   - `api/openapi/schemas/automation.yaml`: схемы `Automation`, `ToggleAutomationRequest`.
   - `api/openapi/paths/automations.yaml`: эндпоинты `GET /workspaces/{id}/automations`, `PATCH /workspaces/{id}/automations/{id}/toggle`.
   - Сгенерирован типобезопасный `ServerInterface` (`internal/infrastructure/http/generated/api.gen.go`).
3. **Чистый доменный слой (`internal/domain/events` и `internal/domain/automation`)**:
   - `events/event.go`, `events/bus.go`, `events/context.go`, `events/event_test.go`: интерфейсы событий и шины, маркеры контекста Loop Guard.
   - `automation/automation.go`, `automation/evaluator.go`, `automation/errors.go`, `automation/automation_test.go`, `automation/evaluator_test.go`: чистый RuleEvaluator для сопоставления условий правил с контекстом задачи.
4. **Use Cases и Сервисы (`internal/application/`)**:
   - `application/automation`: `ports.go`, `usecases.go`, `listener.go` (обработчик событий `TaskCompleted`, `TaskUpdated`, `TaskCreated` с вызовом действий и проверкой `IsAutomationTriggered`).
   - `application/task` и `application/project`: внедрение `EventPublisher` для эмита событий задач.
   - `application/workspace`: автоматическое сидирование 3 пресетов при создании нового воркспейса (`SeedDefaultAutomations`).
5. **Инфраструктура (`internal/infrastructure/`)**:
   - `eventbus/inmemory_bus.go`: асинхронный воркер-пул шины событий с graceful shutdown.
   - `repository/automation_repository.go`: методы хранения и выборки автоматизаций.
   - `http/automations_handler.go`: хендлеры `ListWorkspaceAutomations` и `ToggleAutomation` с валидацией и проверкой прав доступа воркспейса (IDOR protection).
   - `container/`: сборка `InMemoryEventBus`, регистрация слушателей, проброс зависимостей и остановка шины в `container.Close()`.
6. **Фронтенд (`frontend/src/`)**:
   - `entities/automation/`: типы `Automation`, API методы и хуки TanStack Query v5 с оптимистичным UI (`useAutomationsQuery`, `useToggleAutomationMutation`).
   - `features/automations-toggle/AutomationsModal.tsx`: диалоговое окно в эстетике Washi/Matcha с 3 карточками пресетов и тактильными тумблерами (Toggle Switch).
   - `widgets/sidebar/Sidebar.tsx` и `app/App.tsx`: кнопка «Автоматизации» с иконкой Sparkles и бейджем, шорткат закрытия по `Escape`.
7. **E2E-тестирование (`tests/e2e/test_automations.py`)**:
   - 8 сквозных E2E-тестов на Pytest: сидирование при онбординге, переключение тумблеров, отработка `Done & Archive`, отработка `Auto-Focus`, отработка `Inbox Zero`, защита от зацикливания `Recursion Loop Guard`, проверка Tenant Isolation (IDOR) и проверка 401 Unauthorized.

---

## 📊 3. Метрики качества и Приемочный контроль

* **`make check-arch`**: `PASS (0 ошибок)` по всем 5 проверкам архитектурного линтера.
* **`make test`**: `100% PASS` с флагом `-race` во всех пакетах бэкенда.
* **`make test-e2e`**: `25/25 PASS (100%)` в Docker-окружении (8 новых тестов автоматизаций).
* **`cd frontend && npm run build`**: `0 ошибок, 0 предупреждений` (Vite build собран за 4.09s).
