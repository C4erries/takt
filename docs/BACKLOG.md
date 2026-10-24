# Бэклог задач и будущих спринтов Takt (`BACKLOG.md`)

> В этом документе фиксируются проработанные задачи для будущих спринтов проекта Takt.
> Все решения по каждой задаче зафиксированы в сессии от 2026-10-22.
> Формат: чёткая продуктовая цель, принятые технические решения, конкретные файлы для изменения.
>
> **Легенда статусов**: 🔲 В бэклоге (Ready) · 🔵 В спринте (Active) · ✅ Реализовано (Done)

---

## 📑 Индекс задач по спринтам

| Спринт | ID | Категория | Название задачи | Сложность |
|---|---|---|---|---|
| **0.9.0** | **TASK-B10** | Автоматизации / UX | [Ребрендинг «SLA» в понятные термины](#task-b10) | S |
| **0.9.0** | **TASK-B07** | Автоматизации | [Авто-названия без имени колонки + бейдж #A-01](#task-b07) | S |
| **0.9.0** | **TASK-B11** | Теги / UX | [Клавиатурная навигация в выпадающем списке тегов](#task-b11) | S |
| **0.9.0** | **TASK-B04** | Канбан / UX | [Счётчик только активных задач на карточке доски](#task-b04) | S |
| **0.9.1** | **TASK-B12** | Создание задач / UX | [Хоткей раскрытия заметок при быстром создании](#task-b12) | S |
| **0.9.1** | **TASK-B02** | Задачи / UI | [Эргономика подзадач + привязка существующей задачи](#task-b02) | M |
| **0.9.1** | **TASK-B13** | Теги / Управление | [Централизованный менеджер тегов воркспейса](#task-b13) | M |
| **0.9.2** | **TASK-B03** | Задачи / Core | [Глубокое копирование задачи (Deep Duplicate)](#task-b03) | M |
| **0.9.2** | **TASK-B08** | Автоматизации | [Скоуп автоматизаций: фильтрация по доске в Listener](#task-b08) | S |
| **0.9.3** | **TASK-B05** | Задачи / Архив | [Архив задач: секция, батч-сброс, авто-архивация](#task-b05) | L |
| **0.9.4** | **TASK-B06** | Аудит / UX | [История изменений задачи в шторке](#task-b06) | L |
| **0.9.5** | **TASK-B09** | Профиль | [Профиль пользователя: имя и аватар](#task-b09) | M |
| **Future** | **TASK-B01** | Автоматизации | [Subtasks Automation Suite](#task-b01) | L |

---

## 🏃 Спринт 0.9.0 — «UX Quick Wins» (только фронтенд, без миграций)

> **Цель**: Убрать мелкие UX-раздражители за один спринт. Все задачи — S-размер, только фронтенд, деплоится одним коммитом.

---

### TASK-B10: Ребрендинг «SLA» в понятные термины {#task-b10}

✅ **Статус**: Реализовано (Done) · **Сложность**: S · **Только фронтенд**

**Продуктовая цель**: Убрать техническую аббревиатуру SLA из UI конструктора автоматизаций, заменив на термины которые понимает любой пользователь.

**Принятые решения**:
- Бэкенд **не трогаем** — API-код `column_sla_timer` остаётся, меняем только UI-лейбл.
- `column_sla_timer` → «**Задержка в колонке**» (задача в колонке дольше N часов)
- `task_timer` → «**Таймер задачи**» (без изменений, уже нормально)
- Параметр «Порог нахождения:» → «**Задача в колонке дольше чем:**»

**Файлы**:
- `frontend/src/features/automations/useRuleBuilder.ts` — строка с `column_sla_timer` в `TRIGGER_OPTIONS`

---

### TASK-B07: Авто-названия автоматизаций без имени колонки + бейдж #A-NN {#task-b07}

✅ **Статус**: Реализовано (Done) · **Сложность**: S · **Только фронтенд**

**Продуктовая цель**: Убрать иллюзию привязки правила к конкретной колонке по имени — правило применяется по ID, а не по названию.

**Принятые решения**:
- Формат авто-названия: `«[Триггер] → [Действие]»` — **без** имени колонки
- Текущий код: `` `${triggerLabel} [${col}] → ${actionLabel}` ``
- Новый код: `` `${triggerLabel} → ${actionLabel}` ``
- Бейдж `#A-01` — порядковый номер по индексу в отсортированном по `created_at` списке воркспейса: `#A-${(index+1).toString().padStart(2, '0')}`
- Существующие правила со старыми именами **не переименовываем** автоматически

**Файлы**:
- `frontend/src/features/automations/useRuleBuilder.ts` — логика формирования `name` (убрать блок с `col`)
- `frontend/src/features/automations/AutomationCard.tsx` — добавить бейдж `#A-NN`

---

### TASK-B11: Клавиатурная навигация стрелками в списке тегов {#task-b11}

✅ **Статус**: Реализовано (Done) · **Сложность**: S · **Только фронтенд**

**Продуктовая цель**: Выбирать тег из выпадающего списка без отрыва рук от клавиатуры.

**Принятые решения**:
- `focusedIndex` state: `-1` = ничего не выделено, обновляется при `ArrowDown`/`ArrowUp`
- При вводе в input — сбросить `focusedIndex` в `-1`
- `Enter` при `focusedIndex >= 0` → выбрать `filteredTags[focusedIndex]`
- `Escape` → закрыть dropdown, вернуть фокус на input
- Автоскролл: `useRef` на активный `li` → `scrollIntoView({ block: 'nearest' })`
- Wrap-around: стрелка вниз с последнего → первый, вверх с первого → последний

**Файлы**:
- Найти компонент автокомплита тегов (вероятно `TaskTagSection.tsx` или отдельный `TagInput.tsx` в `task-drawer`)
- Добавить `onKeyDown` обработчик и `focusedIndex` state

---

### TASK-B04: Счётчик только активных задач на карточке доски {#task-b04}

✅ **Статус**: Реализовано (Done) · **Сложность**: S · **Бэкенд + фронтенд**

**Продуктовая цель**: Показывать реальный объём работы на доске — только незакрытые корневые задачи.

**Принятые решения**:
- SQL: `COUNT(*) FILTER (WHERE is_completed = false AND parent_id IS NULL) AS active_tasks_count`
- Подпись в UI: `X колонок · Y в работе` (если Y = 0 → `нет задач`)
- Миграция БД **не нужна** — только изменение SQL-запроса и DTO

**Файлы**:
- `backend/queries/projects.sql` — обновить запрос подсчёта задач по доскам
- `frontend/src/widgets/boards-panel/BoardCard.tsx` — обновить подпись

---

## 🏃 Спринт 0.9.1 — «UX Polish: Создание задач, Подзадачи, Теги»

> **Цель**: Три связанных UX-улучшения вокруг создания задач и управления тегами.

---

### TASK-B12: Хоткей раскрытия заметок при быстром создании задачи {#task-b12}

✅ **Статус**: Реализовано (Done) · **Сложность**: S · **Только фронтенд**

**Продуктовая цель**: Форма быстрого создания задачи остаётся однострочной, но даёт мгновенный доступ к описанию без открытия модалки.

**Принятые решения**:
- По умолчанию: только однострочный input заголовка
- `Shift+Enter` в input заголовка → плавно раскрывает textarea описания (motion/react `AnimatePresence`)
- После раскрытия: фокус автоматически переходит в textarea
- `Ctrl+Enter` → создать задачу с заголовком И описанием
- Обычный `Enter` → создать задачу только с заголовком (textarea сворачивается)
- Анимация: `initial={{ height: 0, opacity: 0 }} animate={{ height: 'auto', opacity: 1 }}`

**Файлы**:
- `frontend/src/features/kanban-board/KanbanColumn.tsx` — форма инлайн-создания

---

### TASK-B02: Эргономика подзадач + привязка существующей задачи {#task-b02}

✅ **Статус**: Реализовано (Done) · **Сложность**: M · **Фронтенд + минимальный бэкенд**

**Продуктовая цель**: Комфортный ввод подзадач и возможность превратить уже существующую задачу в подзадачу.

**Принятые решения**:
- **Эргономика**: `min-h-[34px] py-1.5` в поле ввода названия подзадачи
- **Привязка задачи**: кнопка-иконка `Link` рядом с `+` добавить подзадачу → поповер с поиском
- **Поиск**: использовать существующий `GET /workspaces/{id}/search?q=...`
- **Фильтрация в поиске**: исключить задачи, у которых уже есть `parent_id` (нельзя привязать подзадачу как подзадачу) и задачи у которых есть свои подзадачи (нельзя создать 2-й уровень вложенности), и саму текущую задачу
- **Endpoint**: существующий `PATCH /workspaces/{wid}/tasks/{id}` с `{ parentId: currentTask.id }` — уже реализован
- **После привязки**: инвалидируем query `tasks` — задача исчезает с канбана, появляется в подзадачах родителя

**Файлы**:
- `frontend/src/widgets/task-drawer/TaskSubtasksSection.tsx` — высота инпута + кнопка Link
- `frontend/src/widgets/task-drawer/AttachTaskPopover.tsx` — **новый** компонент (поиск + список)

---

### TASK-B13: Централизованный менеджер тегов воркспейса {#task-b13}

✅ **Статус**: Реализовано (Done) · **Сложность**: M · **Бэкенд + фронтенд**

**Продуктовая цель**: Управлять всем каталогом тегов воркспейса в одном окне — переименовать, сменить цвет, удалить неиспользуемые.

**Принятые решения**:
- Доступ: иконка или пункт «Теги» в сайдбаре (точка входа — нижняя часть сайдбара или настройки воркспейса)
- **Список**: все теги с счётчиком задач (`task_count`) — показывает сколько задач используют тег
- **Действия**: inline-переименование, смена цвета из 12-цветной палитры, удаление с подтверждением
- **Удаление**: диалог «Тег будет отвязан от X задач. Продолжить?»
- **Пакетная очистка**: кнопка «Удалить неиспользуемые» — удаляет все теги с `task_count = 0`
- **Поиск**: строка фильтрации по имени тега внутри модалки

**Новые API**:
- `GET /workspaces/{wid}/tags` — уже есть (проверить)
- `GET /workspaces/{wid}/tags/metrics` — **новый** (теги + `task_count`)
- `PATCH /workspaces/{wid}/tags/{id}` — **новый** (имя, цвет)
- `DELETE /workspaces/{wid}/tags/{id}` — **новый**
- `DELETE /workspaces/{wid}/tags/unused` — **новый** (пакетное удаление)

**Файлы**:
- `backend/internal/application/task/usecases.go` — методы управления тегами
- `backend/internal/infrastructure/http/tasks_handler.go` — новые handlers
- `api/openapi/paths/tags.yaml` — новые endpoints
- `frontend/src/features/tag-manager/TagManagerModal.tsx` — **новый** компонент

---

## 🏃 Спринт 0.9.2 — «Core Features: Дублирование и Скоупы»

> **Цель**: Добавить дублирование задач и починить фильтрацию скоупов автоматизаций.

---

### TASK-B03: Глубокое копирование задачи с подзадачами {#task-b03}

✅ **Статус**: Реализовано (Done) · **Сложность**: M · **Бэкенд + фронтенд**

**Продуктовая цель**: Дублировать задачу вместе со всеми подзадачами — каждая копия это независимая новая запись в БД.

**Принятые решения**:
- **Что копируется**: title (с префиксом «Копия — »), description/notes, priority, tags, подзадачи (новые UUID), вложения (shallow copy — одна S3-ссылка, новые записи в `attachments`)
- **Что НЕ копируется**: due_date, история изменений
- **Вложения**: `INSERT INTO attachments (..., storage_key) VALUES (..., original.storage_key)` — S3 не трогаем
- **Позиция**: `original.position + 0.5` (в той же колонке)
- **UI**: контекстное меню карточки (3 точки) → «Дублировать задачу»

**Новый API**: `POST /workspaces/{wid}/tasks/{task_id}/duplicate`

**Файлы**:
- `backend/internal/application/task/usecases.go` — новый Use Case `DuplicateTask` (транзакция)
- `backend/internal/infrastructure/http/tasks_handler.go` — handler `DuplicateTask`
- `api/openapi/paths/tasks.yaml` — новый endpoint
- `backend/queries/tasks.sql` — запрос дублирования (INSERT ... SELECT)
- `frontend/src/features/kanban-board/KanbanCard.tsx` — пункт «Дублировать» в меню

---

### TASK-B08: Фильтрация скоупа автоматизаций по доске в EventListener {#task-b08}

✅ **Статус**: Реализовано (Done) · **Сложность**: S · **Только бэкенд**

**Продуктовая цель**: Правила с `board_id` применяются только к задачам нужной доски (сейчас фильтрация по `board_id` в listener отсутствует).

**Принятые решения**:
- Схема БД **не трогаем** — `board_id` уже есть в `automations`
- Добавить проверку в `listener.go` для всех handler-ов:
  ```go
  if rule.BoardID != nil && (taskInfo.ProjectID == nil || *rule.BoardID != *taskInfo.ProjectID) {
      continue
  }
  ```
- В `AutomationsHub.tsx` заменить фильтры `'all' / 'column' / 'global'` на `'all' / 'board' / 'column'` с понятными лейблами «Все», «На доску», «На колонку»
- В `RuleBuilderModal` добавить поле «Область применения» → влияет на заполнение `board_id` при сохранении

**Файлы**:
- `backend/internal/application/automation/listener.go` — добавить board_id фильтр в `HandleTask*`
- `frontend/src/features/automations/AutomationsHub.tsx` — обновить категории фильтров
- `frontend/src/features/automations/RuleBuilderModal.tsx` — поле «Область применения»

---

## 🏃 Спринт 0.9.3 — «Архив задач» (Batch SQL ADR-0008)

> **Цель**: Разгрузка досок от выполненных задач без потери истории. Реализует ADR-0008 (Batch SQL Fast Path).

---

### TASK-B05: Архив задач — секция, батч-сброс, авто-архивация {#task-b05}

✅ **Статус**: Реализовано (Done) · **Сложность**: L · **Бэкенд (миграция) + фронтенд**

**Продуктовая цель**: Быстро убрать выполненные задачи с доски — вручную или автоматически — с возможностью восстановить.

**Принятые решения**:

*Схема БД (новая миграция 000014):*
```sql
ALTER TABLE tasks ADD COLUMN is_archived BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE tasks ADD COLUMN archived_at TIMESTAMPTZ NULL;
CREATE INDEX idx_tasks_workspace_archived ON tasks(workspace_id, is_archived) WHERE is_archived = true;
```

*UI — доступ к архиву:*
- Отдельная секция «**Архив**» в боковом сайдбаре (как Inbox) — иконка `Archive` после `Inbox`
- Показывает счётчик архивных задач
- Внутри: список/канбан архивных задач с фильтром по проекту/доске

*Пакетный сброс (Column Flush):*
- Кнопка в меню колонки (3 точки) → «Сгрузить выполненные в архив»
- `POST /workspaces/{wid}/columns/{id}/archive-completed`
- **Реализация**: одним SQL (`UPDATE tasks SET is_archived = true WHERE column_id = $1 AND is_completed = true`)

*Разархивирование:*
- Кнопка «Восстановить» на карточке в архиве → поповер выбора доски + колонки
- `PATCH /workspaces/{wid}/tasks/{id}/unarchive` с `{ columnId: ..., boardId: ... }`

*Авто-архивация (ADR-0008 Batch Fast Path):*
- Добавить `BatchTaskExecutor` интерфейс в `ports.go`
- Реализовать `ArchiveCompletedTasksInColumn(columnID, olderThanHours)` в репозитории
- В `Worker.processSLATimerRule/processScheduledSweepRule` добавить `canUseBatchPath()` проверку

*Фильтрация на канбан-доске:*
- Все существующие запросы задач добавляют `AND is_archived = false`

**Новые API**:
- `GET /workspaces/{wid}/tasks/archived?project_id=...`
- `POST /workspaces/{wid}/columns/{id}/archive-completed`
- `PATCH /workspaces/{wid}/tasks/{id}/unarchive`

**Файлы**:
- `backend/migrations/000014_archive.up.sql` — новая миграция
- `backend/internal/application/automation/ports.go` — `BatchTaskExecutor`
- `backend/internal/infrastructure/repository/task_repo.go` — реализация batch методов
- `backend/internal/application/automation/worker.go` — `canUseBatchPath()` + fast path
- `backend/internal/domain/events/` — `BatchSweepCompletedEvent`
- `frontend/src/widgets/sidebar/` — секция «Архив»
- `frontend/src/features/archive/ArchiveView.tsx` — **новый** компонент

---

## 🏃 Спринт 0.9.4 — «История изменений задачи»

> **Цель**: Прозрачность: когда и кем изменялась задача. Новая таблица + EventBus listener.

---

### TASK-B06: История изменений задачи в шторке (Activity Log) {#task-b06}

✅ **Статус**: Реализовано (Done) · **Сложность**: L · **Бэкенд (миграция + EventBus) + фронтенд**

**Продуктовая цель**: В шторке задачи видеть хронологию — кто и когда перенёс задачу, сменил приоритет, добавил тег.

**Принятые решения**:

*Схема БД (новая миграция 000015):*
```sql
CREATE TABLE task_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    activity_type VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_task_activities_task_id ON task_activities(task_id, created_at DESC);
```

*Типы событий (`activity_type`):*
- `column_moved` — payload: `{ from: "Название", to: "Название" }`
- `status_changed` — payload: `{ completed: true/false }`
- `priority_changed` — payload: `{ from: "high", to: "medium" }`
- `due_date_set` / `due_date_cleared`
- `tag_added` / `tag_removed` — payload: `{ tag: "название" }`
- `title_changed` — payload: `{ from: "...", to: "..." }`
- `subtask_added` — payload: `{ title: "..." }`

*Запись через EventBus:*
- Новый listener `ActivityLogger` подписывается на `EventTaskUpdated`, `EventTaskMoved`, `EventTaskCompleted`
- `ActivityLogger` живёт в `backend/internal/application/task/` или отдельном `activity/` подпакете

*UI — вкладка «История»:*
- Дополнительная вкладка «История» рядом с «Подзадачи» в шторке (`TaskDrawer.tsx`)
- Хронологическая лента событий с относительным временем («2 минуты назад», «вчера в 14:30»)
- Неактивна по умолчанию (не грузит данные пока не открыта)

**Новый API**: `GET /workspaces/{wid}/tasks/{id}/activities`

**Файлы**:
- `backend/migrations/000015_task_activities.up.sql`
- `backend/internal/application/task/activity_logger.go` — **новый** listener
- `backend/internal/infrastructure/http/tasks_handler.go` — handler GetActivities
- `frontend/src/widgets/task-drawer/TaskActivityTab.tsx` — **новый** компонент

---

## 🏃 Спринт 0.9.5 — «UX Polish: Автоматизации, Менеджер тегов и Профиль»

> **Цель**: Структурирование селекторов автоматизаций, полноэкранный менеджер тегов и настройки профиля.

---

### TASK-B14: UX Конструктора автоматизаций — расписание без жесткой колонки и группировка досок/колонок {#task-b14}

✅ **Статус**: Реализовано (Done) · **Сложность**: M · **Фронтенд + Бэкенд (automation worker)**

**Продуктовая цель**: Устранить путаницу в скоупах правил и сделать выбор целевых досок и колонок структурированным и удобным.

**Принятые решения**:
1. **Триггер «Ежедневное расписание» (`scheduled_sweep`)**:
   - При скоупе правила «На всю доску» убрать жесткую привязку к одной колонке — позволить правилу выполнять сброс/проверку по всем колонкам доски (или опционально указывать колонку).
   - В UI триггера поле колонки становится опциональным при выборе скоупа «На доску» («Все колонки доски» по умолчанию).
2. **Иерархическая группировка в селекторах целевой доски и колонки**:
   - Селектор целевой доски: группировать по Проектам (`Проект: <Название>` → список досок).
   - Селектор целевой колонки: группировать по Проектам и Доскам (`<Проект> / <Доска>` → список колонок).
   - Использовать структурированные выпадающие списки (`<optgroup>` или кастомный поповер-селектор с категориями) в `RuleBuilderModal.tsx`, `MoveTaskModal.tsx` и `RestoreTaskPopover.tsx`.

**Файлы**:
- `frontend/src/features/automations/RuleBuilderModal.tsx`
- `frontend/src/features/automations/RuleBuilderSections.tsx`
- `frontend/src/features/automations/useRuleBuilder.ts`
- `frontend/src/features/kanban-board/MoveTaskModal.tsx`
- `backend/internal/application/automation/worker.go` (поддержка sweep для доски без column_id)

---

### TASK-B15: Полноэкранный Менеджер тегов (View-first в MainSheet) {#task-b15}

✅ **Статус**: Реализовано (Done) · **Сложность**: M · **Фронтенд**

**Продуктовая цель**: Перенести управление тегами воркспейса из тесного сайдбар-поповера в просторный полноценный экран (как «Автоматизации», «Архив» и «Сегодня»).

**Принятые решения**:
- В `MainSheet.tsx` добавить вид `tags` (по клику на пункт «Теги» в сайдбаре или URL `/tags`).
- Создать компонент `frontend/src/features/tag-manager/TagManagerView.tsx`:
  - Полноразмерный заголовок с поисковой строкой и кнопкой «Удалить неиспользуемые теги».
  - Просторная сетка/таблица тегов: название, бейдж цвета, счетчик связанных задач, палитра смены цвета, переименование по двойному клику, кнопка удаления.
  - Состояния: загрузка (Skeleton), пустой список (Empty state), ошибка (Error).
- В `Sidebar.tsx` клик по «Теги» открывает полноэкранный `TagManagerView` вместо модального поповера.

**Файлы**:
- `frontend/src/features/tag-manager/TagManagerView.tsx` (новый компонент)
- `frontend/src/features/tag-manager/index.ts`
- `frontend/src/widgets/main-sheet/MainSheet.tsx`
- `frontend/src/widgets/sidebar/Sidebar.tsx`

---

### TASK-B09: Профиль пользователя — имя и аватар {#task-b09}

⏸️ **Статус**: Отложено пользователем · **Сложность**: M · **Бэкенд (миграция) + фронтенд**

**Продуктовая цель**: Персонализация — пользователь видит своё имя и аватар в интерфейсе.

**Принятые решения**:
- **Скоуп**: только `display_name` и `avatar_url`. Шаблоны досок, 2FA, часовой пояс — это Фаза 8 Roadmap
- **Доступ**: кнопка/аватар в нижней части сайдбара → модалка профиля
- **Аватар**: загрузка через существующую S3-инфраструктуру (MinIO), кроп перед загрузкой

*Схема БД (новая миграция 000016):*
```sql
ALTER TABLE users ADD COLUMN display_name VARCHAR(100) NULL;
ALTER TABLE users ADD COLUMN avatar_url TEXT NULL;
```

**Новые API**:
- `PATCH /users/me` — обновить `display_name`
- `POST /users/me/avatar` — загрузить аватар (multipart, через S3)

**Файлы**:
- `backend/migrations/000016_user_profile.up.sql`
- `backend/internal/application/workspace/` или `user/` — Use Case обновления профиля
- `frontend/src/widgets/sidebar/UserProfileButton.tsx` — обновить отображение имени/аватара
- `frontend/src/features/user-profile/UserProfileModal.tsx` — **новый** компонент

---

## ⏳ Future — Subtasks Automation Suite (после стабилизации)

---

### TASK-B01: Автоматизация подзадач (Subtasks Automation Suite) {#task-b01}

🔲 **Статус**: В бэклоге (будущее) · **Сложность**: L

**Продуктовая цель**: Правила автоматизаций реагируют на жизненный цикл подзадач.

**Принятые решения**:

*Новые триггеры в `rule.go`:*
- `all_subtasks_completed` — все подзадачи завершены
- `subtask_created` — добавлена новая подзадача
- `any_subtask_completed` — выполнена очередная подзадача

*Новые условия:*
- `subtasks_all_completed` (bool)
- `subtasks_count` (число с операторами `>`, `==`, `0`)

*Новые действия:*
- `complete_all_subtasks` — `UPDATE tasks SET is_completed = true WHERE parent_id = $1`
- `reopen_all_subtasks` — `UPDATE tasks SET is_completed = false WHERE parent_id = $1`
- `create_subtask` — создать новую подзадачу по шаблону (только title)

*Recursion Guard:*
- `create_subtask` **запрещён** если `taskInfo.ParentID != nil` — нет подзадач у подзадач

**Файлы**:
- `backend/internal/domain/automation/rule.go` — новые триггеры, условия, действия
- `backend/internal/application/automation/subtask_automation.go` — логика (файл уже существует)
- `frontend/src/features/automations/useRuleBuilder.ts` — TRIGGER_OPTIONS, ACTION_TYPES
- `frontend/src/features/automations/RuleBuilderSections.tsx` — UI новых действий

---

### TASK-B05: Архив задач — секция, батч-сброс, авто-архивация {#task-b05}

✅ **Статус**: Реализовано (Done) · **Сложность**: L · **Бэкенд (миграция) + фронтенд**

**Продуктовая цель**: Быстро убрать выполненные задачи с доски — вручную или автоматически — с возможностью восстановить.

**Принятые решения**:

*Схема БД (новая миграция 000014):*
```sql
ALTER TABLE tasks ADD COLUMN is_archived BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE tasks ADD COLUMN archived_at TIMESTAMPTZ NULL;
CREATE INDEX idx_tasks_workspace_archived ON tasks(workspace_id, is_archived) WHERE is_archived = true;
```

*UI — доступ к архиву:*
- Отдельная секция «**Архив**» в боковом сайдбаре (как Inbox) — иконка `Archive` после `Inbox`
- Показывает счётчик архивных задач
- Внутри: список/канбан архивных задач с фильтром по проекту/доске

*Пакетный сброс (Column Flush):*
- Кнопка в меню колонки (3 точки) → «Сгрузить выполненные в архив»
- `POST /workspaces/{wid}/columns/{id}/archive-completed`
- **Реализация**: одним SQL (`UPDATE tasks SET is_archived = true WHERE column_id = $1 AND is_completed = true`)

*Разархивирование:*
- Кнопка «Восстановить» на карточке в архиве → поповер выбора доски + колонки
- `PATCH /workspaces/{wid}/tasks/{id}/unarchive` с `{ columnId: ..., boardId: ... }`

*Авто-архивация (ADR-0008 Batch Fast Path):*
- Добавить `BatchTaskExecutor` интерфейс в `ports.go`
- Реализовать `ArchiveCompletedTasksInColumn(columnID, olderThanHours)` в репозитории
- В `Worker.processSLATimerRule/processScheduledSweepRule` добавить `canUseBatchPath()` проверку

*Фильтрация на канбан-доске:*
- Все существующие запросы задач добавляют `AND is_archived = false`

**Новые API**:
- `GET /workspaces/{wid}/tasks/archived?project_id=...`
- `POST /workspaces/{wid}/columns/{id}/archive-completed`
- `PATCH /workspaces/{wid}/tasks/{id}/unarchive`

**Файлы**:
- `backend/migrations/000014_archive.up.sql` — новая миграция
- `backend/internal/application/automation/ports.go` — `BatchTaskExecutor`
- `backend/internal/infrastructure/repository/task_repo.go` — реализация batch методов
- `backend/internal/application/automation/worker.go` — `canUseBatchPath()` + fast path
- `backend/internal/domain/events/` — `BatchSweepCompletedEvent`
- `frontend/src/widgets/sidebar/` — секция «Архив»
- `frontend/src/features/archive/ArchiveView.tsx` — **новый** компонент

---

## 🏃 Спринт 0.9.4 — «История изменений задачи»

> **Цель**: Прозрачность: когда и кем изменялась задача. Новая таблица + EventBus listener.

---

### TASK-B06: История изменений задачи в шторке (Activity Log) {#task-b06}

✅ **Статус**: Реализовано (Done) · **Сложность**: L · **Бэкенд (миграция + EventBus) + фронтенд**

**Продуктовая цель**: В шторке задачи видеть хронологию — кто и когда перенёс задачу, сменил приоритет, добавил тег.

**Принятые решения**:

*Схема БД (новая миграция 000015):*
```sql
CREATE TABLE task_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    activity_type VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_task_activities_task_id ON task_activities(task_id, created_at DESC);
```

*Типы событий (`activity_type`):*
- `column_moved` — payload: `{ from: "Название", to: "Название" }`
- `status_changed` — payload: `{ completed: true/false }`
- `priority_changed` — payload: `{ from: "high", to: "medium" }`
- `due_date_set` / `due_date_cleared`
- `tag_added` / `tag_removed` — payload: `{ tag: "название" }`
- `title_changed` — payload: `{ from: "...", to: "..." }`
- `subtask_added` — payload: `{ title: "..." }`

*Запись через EventBus:*
- Новый listener `ActivityLogger` подписывается на `EventTaskUpdated`, `EventTaskMoved`, `EventTaskCompleted`
- `ActivityLogger` живёт в `backend/internal/application/task/` или отдельном `activity/` подпакете

*UI — вкладка «История»:*
- Дополнительная вкладка «История» рядом с «Подзадачи» в шторке (`TaskDrawer.tsx`)
- Хронологическая лента событий с относительным временем («2 минуты назад», «вчера в 14:30»)
- Неактивна по умолчанию (не грузит данные пока не открыта)

**Новый API**: `GET /workspaces/{wid}/tasks/{id}/activities`

**Файлы**:
- `backend/migrations/000015_task_activities.up.sql`
- `backend/internal/application/task/activity_logger.go` — **новый** listener
- `backend/internal/infrastructure/http/tasks_handler.go` — handler GetActivities
- `frontend/src/widgets/task-drawer/TaskActivityTab.tsx` — **новый** компонент

---

## 🏃 Спринт 0.9.5 — «UX Polish: Автоматизации, Менеджер тегов и Профиль»

> **Цель**: Структурирование селекторов автоматизаций, полноэкранный менеджер тегов и настройки профиля.

---

### TASK-B14: UX Конструктора автоматизаций — расписание без жесткой колонки и группировка досок/колонок {#task-b14}

✅ **Статус**: Реализовано (Done) · **Сложность**: M · **Фронтенд + Бэкенд (automation worker)**

**Продуктовая цель**: Устранить путаницу в скоупах правил и сделать выбор целевых досок и колонок структурированным и удобным.

**Принятые решения**:
1. **Триггер «Ежедневное расписание» (`scheduled_sweep`)**:
   - При скоупе правила «На всю доску» убрать жесткую привязку к одной колонке — позволить правилу выполнять сброс/проверку по всем колонкам доски (или опционально указывать колонку).
   - В UI триггера поле колонки становится опциональным при выборе скоупа «На доску» («Все колонки доски» по умолчанию).
2. **Иерархическая группировка в селекторах целевой доски и колонки**:
   - Селектор целевой доски: группировать по Проектам (`Проект: <Название>` → список досок).
   - Селектор целевой колонки: группировать по Проектам и Доскам (`<Проект> / <Доска>` → список колонок).
   - Использовать структурированные выпадающие списки (`<optgroup>` или кастомный поповер-селектор с категориями) в `RuleBuilderModal.tsx`, `MoveTaskModal.tsx` и `RestoreTaskPopover.tsx`.

**Файлы**:
- `frontend/src/features/automations/RuleBuilderModal.tsx`
- `frontend/src/features/automations/RuleBuilderSections.tsx`
- `frontend/src/features/automations/useRuleBuilder.ts`
- `frontend/src/features/kanban-board/MoveTaskModal.tsx`
- `backend/internal/application/automation/worker.go` (поддержка sweep для доски без column_id)

---

### TASK-B15: Полноэкранный Менеджер тегов (View-first в MainSheet) {#task-b15}

✅ **Статус**: Реализовано (Done) · **Сложность**: M · **Фронтенд**

**Продуктовая цель**: Перенести управление тегами воркспейса из тесного сайдбар-поповера в просторный полноценный экран (как «Автоматизации», «Архив» и «Сегодня»).

**Принятые решения**:
- В `MainSheet.tsx` добавить вид `tags` (по клику на пункт «Теги» в сайдбаре или URL `/tags`).
- Создать компонент `frontend/src/features/tag-manager/TagManagerView.tsx`:
  - Полноразмерный заголовок с поисковой строкой и кнопкой «Удалить неиспользуемые теги».
  - Просторная сетка/таблица тегов: название, бейдж цвета, счетчик связанных задач, палитра смены цвета, переименование по двойному клику, кнопка удаления.
  - Состояния: загрузка (Skeleton), пустой список (Empty state), ошибка (Error).
- В `Sidebar.tsx` клик по «Теги» открывает полноэкранный `TagManagerView` вместо модального поповера.

**Файлы**:
- `frontend/src/features/tag-manager/TagManagerView.tsx` (новый компонент)
- `frontend/src/features/tag-manager/index.ts`
- `frontend/src/widgets/main-sheet/MainSheet.tsx`
- `frontend/src/widgets/sidebar/Sidebar.tsx`

---

### TASK-B09: Профиль пользователя — имя и аватар {#task-b09}

⏸️ **Статус**: Отложено пользователем · **Сложность**: M · **Бэкенд (миграция) + фронтенд**

**Продуктовая цель**: Персонализация — пользователь видит своё имя и аватар в интерфейсе.

**Принятые решения**:
- **Скоуп**: только `display_name` и `avatar_url`. Шаблоны досок, 2FA, часовой пояс — это Фаза 8 Roadmap
- **Доступ**: кнопка/аватар в нижней части сайдбара → модалка профиля
- **Аватар**: загрузка через существующую S3-инфраструктуру (MinIO), кроп перед загрузкой

*Схема БД (новая миграция 000016):*
```sql
ALTER TABLE users ADD COLUMN display_name VARCHAR(100) NULL;
ALTER TABLE users ADD COLUMN avatar_url TEXT NULL;
```

**Новые API**:
- `PATCH /users/me` — обновить `display_name`
- `POST /users/me/avatar` — загрузить аватар (multipart, через S3)

**Файлы**:
- `backend/migrations/000016_user_profile.up.sql`
- `backend/internal/application/workspace/` или `user/` — Use Case обновления профиля
- `frontend/src/widgets/sidebar/UserProfileButton.tsx` — обновить отображение имени/аватара
- `frontend/src/features/user-profile/UserProfileModal.tsx` — **новый** компонент

---

## ⏳ Future — Subtasks Automation Suite (после стабилизации)

---

### TASK-B01: Автоматизация подзадач (Subtasks Automation Suite) {#task-b01}

🔲 **Статус**: В бэклоге (будущее) · **Сложность**: L

**Продуктовая цель**: Правила автоматизаций реагируют на жизненный цикл подзадач.

**Принятые решения**:

*Новые триггеры в `rule.go`:*
- `all_subtasks_completed` — все подзадачи завершены
- `subtask_created` — добавлена новая подзадача
- `any_subtask_completed` — выполнена очередная подзадача

*Новые условия:*
- `subtasks_all_completed` (bool)
- `subtasks_count` (число с операторами `>`, `==`, `0`)

*Новые действия:*
- `complete_all_subtasks` — `UPDATE tasks SET is_completed = true WHERE parent_id = $1`
- `reopen_all_subtasks` — `UPDATE tasks SET is_completed = false WHERE parent_id = $1`
- `create_subtask` — создать новую подзадачу по шаблону (только title)

*Recursion Guard:*
- `create_subtask` **запрещён** если `taskInfo.ParentID != nil` — нет подзадач у подзадач

**Файлы**:
- `backend/internal/domain/automation/rule.go` — новые триггеры, условия, действия
- `backend/internal/application/automation/subtask_automation.go` — логика (файл уже существует)
- `frontend/src/features/automations/useRuleBuilder.ts` — TRIGGER_OPTIONS, ACTION_TYPES
- `frontend/src/features/automations/RuleBuilderSections.tsx` — UI новых действий

