# Отчет о завершении Спринта 0.7.2 (`sprint-0.7.2.md`)

> **Фаза 7**: Вывод к MVP 1.0 (UX Elevation, Движок анимаций, Темы и Палитры)  
> **Спринт**: `0.7.2` — Междосочный перенос задач (вручную и автоматизацией) и персональные правила для задач  
> **Даты проведения**: 2026-10-19  
> **Статус**: `✅ Завершен (100% DoD, 0 ошибок make check-arch, 100% PASS тестов)`

---

## 🎯 1. Цели и Результаты Спринта

### 1.1 Бизнес-цели и Продуктовый эффект
- **Междосочный перенос задач вручную**: Возможность быстро переместить задачу из любой колонки на любую доску текущего проекта — как через группированный селектор колонок в шторке (`TaskDrawer` и `TaskInspector`), так и через контекстное меню карточки (`...`) на Канбан-доске.
- **Междосочный перенос автоматизациями**: Расширение действия `move_to_column` в Low-Code конструкторе правил: теперь можно выбрать колонку любой доски проекта, что позволяет автоматически маршрутизировать карточки между досками (например, из доски «Разработка» в доску «Релиз»).
- **Персональные правила для конкретной задачи (`task_id`)**: Возможность повесить автоматизацию на конкретную задачу с триггером `task_timer` (например, «через 24 часа переместить задачу в колонку Релиз» или «через N часов повысить приоритет»).
- **Секция автоматизаций задачи в TaskDrawer**: Отображение персональных правил задачи со статусом, тумблером Вкл/Выкл, удалением, кнопкой добавления правила через `RuleBuilderModal` и кнопкой разового применения действий («⚡ Применить»).

---

## 🧠 2. Принятые микро-решения спринта (Режим «Штаб»)

- **D-072-1 (Персональные автоматизации конкретной задачи / `task_id`)**:
  - В таблице `automations` добавлено поле `task_id UUID NULL REFERENCES tasks(id) ON DELETE CASCADE` с индексом `idx_automations_task`.
  - В схемах OpenAPI (`automation.yaml`) и чистом домене Go добавлены поле `taskId` и тип триггера `task_timer`.
  - В `queries/automations.sql` добавлены запрос `ListAutomationsByTask` и обновление `ListAllEnabledTimerAutomations`.
- **D-072-2 (Междосочный перенос задачи вручную в UI)**:
  - В `QuickActionBar` шторки задачи селектор колонок группируется по доскам через `<optgroup label={boardName}>`.
  - В `KanbanCard` добавлено контекстное меню `...` («Переместить на доску...») со списком досок проекта и колонок с индикацией текущей колонки `(здесь)`.
- **D-072-3 (Междосочный перенос в автоматизациях / Cross-Board Actions)**:
  - В конструкторе правил `RuleBuilderModal` при выборе действия «Переместить в колонку» доступен выбор колонок всех досок проекта.
- **D-072-4 (Секция автоматизаций задачи в TaskDrawer и ручной запуск)**:
  - Добавлен компонент `TaskAutomationsSection` в `TaskDrawer` и `TaskInspector`.
  - Возможность мгновенного ручного применения действий правила к задаче («⚡ Применить») с визуальной обратной связью.

---

## 📦 3. Реализованные слои и пакеты

### 3.1 Backend
- `backend/migrations/000011_task_automations.up.sql` и `.down.sql`: миграция внешнего ключа `task_id` и индекса.
- `api/openapi/schemas/automation.yaml`: схемы `taskId` и триггера `task_timer`.
- `backend/queries/automations.sql`: запросы `ListAutomationsByTask`, Create/Update с `task_id`.
- `backend/internal/domain/automation/`: чистый домен `TaskID`, `TriggerTaskTimer = "task_timer"`, валидация инвариантов, табличные тесты.
- `backend/internal/application/automation/`: Use Cases `ListTaskAutomations`, `ExecuteTaskAutomation`, DTO с `TaskID`/`ClearTaskID`, воркер `processTaskTimerRule`.
- `backend/internal/infrastructure/repository/automation_repository.go`: маппинг `task_id`, `ListAutomationsByTask`, `GetTaskByID` в `TaskReader`.
- `backend/internal/infrastructure/http/automations_handler.go`: эндпоинты, фильтрация `?taskId=` в `ListWorkspaceAutomations`.

### 3.2 Frontend
- `frontend/src/entities/automation/`: типы `taskId`, `useTaskAutomationsQuery`.
- `frontend/src/entities/project/columnQueries.ts`: `useProjectAllColumnsQuery` (загрузка колонок всех досок проекта).
- `frontend/src/widgets/task-drawer/QuickActionBar.tsx`: группированный селектор по доскам `<optgroup>`.
- `frontend/src/features/kanban-board/KanbanCard.tsx`: контекстное меню «Переместить на доску...».
- `frontend/src/features/automations/`: поддержка `taskId`, триггера `task_timer` и группировки колонок в действии переноса.
- `frontend/src/widgets/task-drawer/TaskAutomationsSection.tsx`: секция автоматизаций задачи с тумблерами и ручным запуском.

---

## 🛡️ 4. Метрики контроля качества и верификация

| Проверка | Инструмент / Команда | Результат | Примечания |
|---|---|---|---|
| **Архитектурный линтер** | `make check-arch` | `✅ 0 errors` | Лимиты < 400 строк и <= 7 файлов строго соблюдены во всех пакетах |
| **Backend Unit Tests** | `make test` | `✅ 100% PASS` | Все тесты domain, application, http, storage пройдены |
| **Frontend Lint & Types** | `npm run lint` | `✅ 0 errors` | Строгая проверка TypeScript без предупреждений |
| **Frontend Build** | `npm run build` | `✅ 0 errors` | Vite production bundle собран успешно |
| **E2E Browser Verification** | Playwright Chromium | `✅ PASS` | Создано 4 фото-артефакта UI |

---

## 📸 5. Фото-артефакты спринта

1. `sprint_072_1_card_cross_board_menu.png`: Меню карточки Канбана с каскадным списком досок и колонок для быстрого переноса.
2. `sprint_072_2_task_drawer_cross_board_and_automations.png`: Шторка задачи `TaskDrawer` с группированным селектором колонок и секцией «⚡ Автоматизации задачи».
3. `sprint_072_3_task_rule_builder_modal.png`: Конструктор Low-Code правил в режиме задачи с триггером «Таймер задачи» (отложенный перенос).
4. `sprint_072_4_created_task_automation_in_drawer.png`: Созданное персональное правило в шторке задачи с бейджем триггера, тумблером Вкл/Выкл и кнопкой «▶ Применить».

