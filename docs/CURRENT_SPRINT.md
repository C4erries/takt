# Текущий спринт / Оперативный центр управления (`CURRENT_SPRINT.md`)

> Этот файл — **единственная точка правды** о том, какие решения приняты по текущей задаче и что команда и AI-агенты делают прямо сейчас.
> Работа над каждым спринтом строго делится на **Фазу А (Штаб: Принятие решений)** и **Фазу Б (Конвейер: Реализация и Верификация)**.
> Завершенные спринты архивируются в каталог `docs/sprints/sprint-X.Y.Z.md` с записью в реестр `docs/sprints/README.md`.

## Текущий спринт: Спринт 0.7.2 (Фаза 7: Междосочный перенос задач вручную и автоматизацией, персональные правила для задач) 🟢 ФАЗА Б (Конвейер)

---

## 🧠 1. Аналитика и Зафиксированные микро-решения спринта (Режим «Штаб»)

* **Бизнес-цель**: Реализовать перенос задач между досками проекта как вручную (из карточки и шторки задачи), так и автоматизацией (междосочный перенос `move_to_column` с выбором целевой доски). Внедрить персональные автоматизации для конкретной задачи (`task_id`), позволяющие настраивать таймеры и сценарии над конкретной карточкой (например, «через 24 часа переместить задачу в колонку X») с возможностью ручного применения действий из шторки.

* **Микро-решения спринта (Решения Штаба)**:
  - **D-072-1 (Персональные автоматизации конкретной задачи / `task_id`)**:
    - В таблице `automations`: миграция `000011_task_automations.up.sql`:
      `ALTER TABLE automations ADD COLUMN task_id UUID NULL REFERENCES tasks(id) ON DELETE CASCADE;`
      `CREATE INDEX IF NOT EXISTS idx_automations_task ON automations(task_id);`
    - В OpenAPI спецификации `api/openapi/schemas/automation.yaml`:
      * В `Automation`, `CreateAutomationRequest`, `UpdateAutomationRequest`: добавить поле `taskId` (uuid, nullable: true).
      * В `Trigger.type` и `triggerType` enum: добавить `task_timer` (таймер задачи).
    - В `queries/automations.sql`:
      * Поддержка `task_id` в `CreateAutomation`, `CreateColumnAutomation`, `UpdateAutomation`.
      * Запрос `ListAutomationsByTask` (`SELECT * FROM automations WHERE task_id = $1 AND workspace_id = $2 ORDER BY created_at ASC;`).
      * Обновление `ListAllEnabledTimerAutomations`: включение `'task_timer'` в выборку.
    - В чистом домене `internal/domain/automation/`:
      * Поле `TaskID *uuid.UUID` в структуре `Automation`.
      * Константа `TriggerTaskTimer = "task_timer"`.
      * Валидация инвариантов и табличные тесты.
  - **D-072-2 (Междосочный перенос задачи вручную в UI)**:
    - В шторке задачи (`TaskDrawer` и `TaskInspector`):
      * В блоке выбора статуса/колонки: группированный селектор по доскам проекта (`optgroup label="Доска X"`), либо выбор целевой доски и колонки.
      * При выборе колонки с другой доски: вызов `projectApi.moveTask(workspaceId, taskId, { columnId: newColId, position: Date.now(), projectId })` с мгновенным оптимистичным переносом.
    - В контекстном меню карточки Канбана (`KanbanCard.tsx`):
      * Добавить пункт меню «Переместить на доску...» со списком других досок проекта и их колонок.
  - **D-072-3 (Междосочный перенос в автоматизациях / Cross-Board Actions)**:
    - В `Action` (`move_to_column`): поддержка параметров `{ column_id, target_board_id }`.
    - В UI конструктора правил (`useRuleBuilder.ts` и `RuleBuilderSections.tsx`): при выборе действия «Переместить в колонку» список колонок группируется по доскам проекта, позволяя автоматически маршрутизировать карточку на любую доску.
  - **D-072-4 (Секция автоматизаций задачи в TaskDrawer и ручной запуск)**:
    - В `TaskDrawer.tsx` (и `TaskInspector.tsx`):
      * Секция «⚡ Автоматизации задачи»: отображение списка правил, привязанных к данной задаче (`task_id`), с тумблером Вкл/Выкл и удалением.
      * Кнопка «+ Добавить правило для задачи» $\rightarrow$ открытие `RuleBuilderModal` с предзаполненным `taskId` и `columnId`.
      * Кнопка «⚡ Применить действия» (разовый запуск) для немедленного выполнения действий правила над задачей.

---

## 📋 Таблица задач спринта 0.7.2

| # | Задача | Статус | Приоритет | Ответственный | Заметки и контракты |
|---|--------|--------|-----------|---------------|---------------------|
| 1 | **[Backend: Step 1 Contracts & Domain]**: Миграция `000011_task_automations`, схемы OpenAPI (`automation.yaml`), sqlc-запросы, чистый домен `internal/domain/automation/` с тестами | ✅ Завершено | Критический | `backend_developer` | `000011_task_automations.up.sql`, `automation.yaml`, `automations.sql` |
| 2 | **[Backend: Step 2 Application & HTTP]**: Use Cases для привязки правил к задачам, эндпоинты `ListAutomationsByTask`, исполнение действий задачи и таймера `task_timer` | ✅ Завершено | Высокий | `backend_developer` | `application/automation/`, `automations_handler.go` |
| 3 | **[Frontend: Cross-Board Task Move UI]**: Междосочный перенос карточки в `TaskDrawer` (группированный селектор по доскам) и в меню карточки Канбана (`KanbanCard.tsx`) | ✅ Завершено | Высокий | `frontend_developer` | `TaskDrawer.tsx`, `KanbanCard.tsx` |
| 4 | **[Frontend: Task Automations & Rule Builder Integration]**: Секция правил задачи в `TaskDrawer`, создание правила с `taskId` через `RuleBuilderModal`, выбор доски в действии `move_to_column` | ✅ Завершено | Высокий | `frontend_developer` | `TaskDrawer.tsx`, `RuleBuilderModal.tsx`, `RuleBuilderSections.tsx` |
| 5 | **[Verification & E2E Testing]**: Независимая проверка гейткипера (`make check-arch`, `make test`, `npm run build`, Playwright скриншоты) | ✅ Завершено | Критический | `qa_engineer` & Lead | Защита от регрессий, 4 скриншота-артефакта |

---

## 🎯 Definition of Done (DoD)

- [x] В БД выполнена миграция `000011_task_automations.up.sql` с внешним ключом `task_id` в таблице `automations`.
- [x] В OpenAPI схемах и Go домене поддержан `taskId` и тип триггера `task_timer`.
- [x] Задачу в `TaskDrawer` можно перенести в любую колонку любой доски текущего проекта.
- [x] В меню карточки на Канбан-доске доступен перенос на другую доску.
- [x] В `TaskDrawer` есть секция «Автоматизации задачи» с возможностью создания персонального правила (например, перенос через время).
- [x] В действии автоматизации `move_to_column` можно выбрать колонку любой доски проекта.
- [x] `make check-arch` проходит с 0 ошибок.
- [x] `cd frontend && npm run build` и `make test` проходят успешно.
