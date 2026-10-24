# Текущий спринт / Оперативный центр управления (`CURRENT_SPRINT.md`)

> Этот файл — **единственная точка правды** о том, какие решения приняты по текущей задаче и что команда и AI-агенты делают прямо сейчас.
> Работа над каждым спринтом строго делится на **Фазу А (Штаб: Принятие решений)** и **Фазу Б (Конвейер: Реализация и Верификация)**.
> Завершенные спринты архивируются в каталог `docs/sprints/sprint-X.Y.Z.md` с записью в реестр `docs/sprints/README.md`.

## Текущий спринт: Спринт 0.9.3 — «Архив задач (Batch SQL ADR-0008)» 🟡 ФАЗА А (Готово к запуску)

---

## 🧠 1. Аналитика и Зафиксированные микро-решения спринта (Штаб)

* **Бизнес-цель**: Разгрузка досок от выполненных задач без потери истории с возможностью ручного батч-сброса, авто-архивации по таймеру (ADR-0008 Batch SQL Fast Path) и восстановления из архива.

* **Связь с бэклогом и ADR**: [`docs/BACKLOG.md`](BACKLOG.md) (TASK-B05), [`docs/decisions/0008-hybrid-batch-automation-execution.md`](decisions/0008-hybrid-batch-automation-execution.md).

* **Принятые микро-решения**:

  1. **Схема БД (миграция 000014_archive)**:
     ```sql
     ALTER TABLE tasks ADD COLUMN is_archived BOOLEAN NOT NULL DEFAULT FALSE;
     ALTER TABLE tasks ADD COLUMN archived_at TIMESTAMPTZ NULL;
     CREATE INDEX idx_tasks_workspace_archived ON tasks(workspace_id, is_archived) WHERE is_archived = true;
     ```
     Все регулярные запросы выборки задач на досках (`ListTasksByProject`, `ListTasksByColumn` и т.д.) дополняются условием `AND is_archived = false`.

  2. **API Архива**:
     - `GET /workspaces/{wid}/tasks/archived?project_id=...` — список архивированных задач.
     - `POST /workspaces/{wid}/columns/{id}/archive-completed` — батч-сброс выполненных задач колонки в архив одним SQL (`UPDATE tasks SET is_archived = true, archived_at = NOW() WHERE column_id = $1 AND is_completed = true AND is_archived = false`).
     - `PATCH /workspaces/{wid}/tasks/{id}/unarchive` с `{ columnId: string, boardId?: string }` — разархивирование задачи с возвращением в выбранную колонку (`is_archived = false`, `archived_at = null`).

  3. **ADR-0008 Batch Fast Path в Automations Worker**:
     - Интерфейс `BatchTaskExecutor` в `backend/internal/application/automation/ports.go`:
       `ArchiveCompletedTasksInColumn(ctx context.Context, workspaceID, columnID uuid.UUID, olderThanHours float64) (int64, error)`.
     - Метод `canUseBatchPath(rule)` в воркере: если действие `archive_task`, conditions только `is_completed = true` — использовать быстрый SQL вместо N+1 цикла.
     - Событие `BatchSweepCompletedEvent` в `domain/events/` для агрегированного уведомления.

  4. **UI Архива**:
     - Доступ: пункт «Архив» в боковом сайдбаре (`Sidebar.tsx`, иконка `Archive`, счётчик архивированных задач).
     - Полноценный экран/вид `frontend/src/features/archive/ArchiveView.tsx`: список архивных задач, фильтр по проекту, дата архивации, кнопка «Восстановить» (поповер выбора доски и колонки).
     - В меню колонки (3 точки) на доске: кнопка «Сгрузить выполненные в архив» с подтверждением.

---

## 📋 2. Таблица задач спринта 0.9.3

| # | Задача | Статус | Приоритет | Ответственный | Файлы |
|---|--------|--------|-----------|---------------|-------|
| 1 | **[B05 Step 1] Archive Schema & Contracts**: миграция 000014, sqlc queries, OpenAPI эндпоинты, make generate, чистый домен | ✅ Готово | Высокий | `backend_developer` | `migrations/`, `queries/tasks.sql`, `api/openapi/*` |
| 2 | **[B05 Step 2] Archive Application, Fast Path & HTTP**: use cases, BatchTaskExecutor, worker fast path, handlers, тесты | ✅ Готово | Высокий | `backend_developer` | `application/task/`, `application/automation/`, `infrastructure/http/` |
| 3 | **[B05 UI] Archive Navigation & Column Flush**: иконка в сайдбаре, кнопка сброса выполненных в колонке | ⏳ Ожидает | Средний | `frontend_developer` | `Sidebar.tsx`, `KanbanColumn.tsx` |
| 4 | **[B05 UI] Archive View & Restore Modal**: экран архива `ArchiveView.tsx`, разархивирование в выбранную колонку | ⏳ Ожидает | Высокий | `frontend_developer` | `ArchiveView.tsx`, `RestoreTaskPopover.tsx` |
| 5 | **[Verification]**: `make check-arch`, `npm run build`, `npm run lint`, `make test`, коммит | ⏳ В очереди | Критический | Lead Orchestrator | — |

---

## 🎯 3. Definition of Done (DoD)

- [x] Выполнена миграция `000014_archive` (`is_archived`, `archived_at`, индекс).
- [x] Обычные списки задач доски и Inbox исключают архивные задачи (`AND is_archived = false`).
- [x] Работают API архива: `GET /tasks/archived`, `POST /columns/{id}/archive-completed`, `PATCH /tasks/{id}/unarchive`.
- [x] Реализован ADR-0008 Batch SQL Fast Path в воркере автоматизаций (`ArchiveCompletedTasksInColumn` выполняется 1 SQL-запросом).
- [ ] В сайдбаре есть секция «Архив» со счётчиком; по клику открывается `ArchiveView`.
- [ ] В меню колонки доски доступно действие «Сгрузить выполненные в архив».
- [ ] На карточке архивной задачи есть действие «Восстановить» с выбором целевой колонки.
- [ ] `make check-arch` — 0 ошибок.
- [ ] `npm run build && npm run lint` — 0 ошибок.
- [ ] `make test` — все тесты зелёные.

---

## 🗓️ 4. Очередь следующих спринтов

| Спринт | Задачи | Суть |
|--------|--------|------|
| **0.9.4** | B06 | История изменений задачи (новая таблица + EventBus listener) |
| **0.9.5** | B09 | Профиль пользователя: имя + аватар |
| **Future** | B01 | Subtasks Automation Suite (L-задача) |
