# Текущий спринт / Оперативный центр управления (`CURRENT_SPRINT.md`)

> Этот файл — **единственная точка правды** о том, какие решения приняты по текущей задаче и что команда и AI-агенты делают прямо сейчас.
> Работа над каждым спринтом строго делится на **Фазу А (Штаб: Принятие решений)** и **Фазу Б (Конвейер: Реализация и Верификация)**.
> Завершенные спринты архивируются в каталог `docs/sprints/sprint-X.Y.Z.md` с записью в реестр `docs/sprints/README.md`.

## Текущий спринт: Спринт 0.9.4 — «История изменений задачи (Activity Log)» 🟡 ФАЗА Б (В работе)

---

## 🧠 1. Аналитика и Зафиксированные микро-решения спринта (Штаб)

* **Бизнес-цель**: Прозрачность жизненного цикла задачи. Пользователь видит в шторке (`TaskDrawer.tsx`) полную хронологическую ленту изменений: кто и когда перенёс задачу в другую колонку, изменил статус/приоритет, назначил дедлайн, добавил тег или подзадачу.

* **Связь с бэклогом и ADR**: [`docs/BACKLOG.md`](BACKLOG.md) (TASK-B06).

* **Принятые микро-решения**:

  1. **Схема БД (миграция 000015_task_activities)**:
     ```sql
     CREATE TABLE task_activities (
         id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
         task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
         user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
         activity_type VARCHAR(64) NOT NULL,
         payload JSONB NOT NULL DEFAULT '{}',
         created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
     );
     CREATE INDEX idx_task_activities_task_id ON task_activities(task_id, created_at DESC);
     ```
     `user_id` может быть `NULL` для системных/автоматических действий.

  2. **Типы событий и payload (`activity_type`)**:
     - `column_moved`: `{ "from_column": "В работе", "to_column": "Готово" }`
     - `status_changed`: `{ "is_completed": true/false }`
     - `priority_changed`: `{ "from_priority": "normal", "to_priority": "high" }`
     - `due_date_set`: `{ "due_date": "2026-10-25T00:00:00Z" }`
     - `due_date_cleared`: `{}`
     - `tag_added`: `{ "tag_name": "Баг", "color_hex": "#ef4444" }`
     - `tag_removed`: `{ "tag_name": "Баг" }`
     - `title_changed`: `{ "from_title": "...", "to_title": "..." }`
     - `subtask_added`: `{ "title": "...", "subtask_id": "..." }`
     - `task_archived`: `{}`
     - `task_unarchived`: `{ "column_name": "..." }`

  3. **EventBus & Асинхронное логирование**:
     - Listener `ActivityLogger` подписывается на события через EventBus:
       `task_moved`, `task_updated`, `task_completed`.
     - Также Use Cases (например, привязка/отвязка тегов, добавление подзадачи) могут логировать напрямую через `ActivityRepository` или публикацию событий.

  4. **API Контракты (OpenAPI)**:
     - `GET /workspaces/{workspace_id}/tasks/{task_id}/activities` -> `[]TaskActivity`
     - Модель `TaskActivity`:
       ```yaml
       TaskActivity:
         type: object
         required: [id, taskId, activityType, createdAt]
         properties:
           id: { type: string, format: uuid }
           taskId: { type: string, format: uuid }
           userId: { type: string, format: uuid, nullable: true }
           userName: { type: string, nullable: true }
           activityType: { type: string }
           payload: { type: object, additionalProperties: true }
           createdAt: { type: string, format: date-time }
       ```

  5. **UI — Вкладка «История» в шторке (`TaskDrawer.tsx`)**:
     - Дополнительная вкладка «История» рядом с «Подзадачи» в шторке задачи.
     - Ленивая загрузка (запрос выполняется только при открытии вкладки).
     - Хронологическая вертикальная лента с иконками типов событий, именем пользователя, красивым описанием и относительным временем (`2 мин назад`, `вчера в 15:40`).

---

## 📋 2. Таблица задач спринта 0.9.4

| # | Задача | Статус | Приоритет | Ответственный | Файлы |
|---|--------|--------|-----------|---------------|-------|
| 1 | **[B06 Step 1] Activities Schema, Queries & OpenAPI**: миграция 000015, sqlc queries, OpenAPI эндпоинт, domain model, make generate | ✅ Готово | Высокий | `backend_developer` | `migrations/`, `queries/activities.sql`, `api/openapi/*`, `internal/domain/task/` |
| 2 | **[B06 Step 2] Activities Listener, Use Cases & HTTP**: listener в EventBus, repo methods, HTTP handler GetTaskActivities, тесты | ✅ Готово | Высокий | `backend_developer` | `application/task/`, `infrastructure/repository/`, `infrastructure/http/` |
| 3 | **[B06 UI] Task Drawer History Tab**: компонент `TaskActivityTab.tsx`, иконки действий, относительное время | ⏳ Ожидает | Высокий | `frontend_developer` | `frontend/src/widgets/task-drawer/` |
| 4 | **[Verification]**: `make check-arch`, `npm run build`, `npm run lint`, `make test`, `docker compose up -d --build`, `pytest tests/`, коммит | ⏳ В очереди | Критический | Lead Orchestrator | — |

---

## 🎯 3. Definition of Done (DoD)

- [x] Выполнена миграция `000015_task_activities` (таблица `task_activities`, индексы).
- [x] Описан OpenAPI эндпоинт `GET /workspaces/{workspace_id}/tasks/{task_id}/activities` и сгенерированы типы.
- [x] Реализованы queries для sqlc (`CreateActivity`, `ListActivitiesByTaskID`).
- [x] Реализован чистый домен `TaskActivity` с валидацией инвариантов.
- [x] Реализован `ActivityLogger` (EventBus listener) и/или методы логирования активности в Use Cases.
- [x] Реализован HTTP-хендлер `GetTaskActivities` с проверкой доступа к воркспейсу.
- [ ] Во фронтенде в `TaskDrawer` добавлена вкладка «История» с человекопонятным рендерингом событий и датами.
- [x] `make check-arch` — 0 ошибок (файлы < 400 строк, папки <= 7 файлов, чистый domain).
- [ ] `npm run build && npm run lint` — 0 ошибок.
- [x] `make test` — все тесты зелёные.
- [ ] `uv run pytest tests/` — 100% PASS.

---

## 🗓️ 4. Очередь следующих спринтов

| Спринт | Задачи | Суть |
|--------|--------|------|
| **0.9.5** | B14, B15, B09 | UX Polish (расписание без жесткой колонки, иерархия проектов/досок/колонок, полноэкранный менеджер тегов) + Профиль (отложено) |
| **Future** | B01 | Subtasks Automation Suite (L-задача) |
