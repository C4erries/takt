# Спринт 0.3.1: Проекты, Колонки и Канбан-доска (DND, Fractional Indexing, ViewToggle)

* **Даты проведения**: 2026-10-14 — 2026-10-14
* **Статус**: ✅ Завершен (100% Definition of Done)
* **Связанные документы**: [docs/CURRENT_SPRINT.md](../CURRENT_SPRINT.md), [docs/conventions/backend.md](../conventions/backend.md), [docs/conventions/frontend.md](../conventions/frontend.md)

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

1. **Дефолтные GTD-колонки**: При создании проекта транзакционно генерируется базовая доска «Основная» с 3 GTD-колонками: «К выполнению» (`todo`, 1000.0), «В работе» (`in_progress`, 2000.0), «Готово» (`done`, 3000.0).
2. **Fractional Indexing ($O(1)$ DND)**: Поле `position` задач и колонок переведено на `DOUBLE PRECISION`. Перемещение карточки между $A$ и $B$ вычисляется как $(pos_A + pos_B) / 2$, что дает точечный $O(1)$ UPDATE одной строки без каскадных блокировок соседних задач в PostgreSQL.
3. **Защитные барьеры при удалении (Safety Guard)**:
   - Запрещено удалять колонку, пока в ней находятся задачи (`409 Conflict` / `ErrColumnNotEmpty`);
   - При удалении проекта все задачи безопасно разлинковываются (`project_id = NULL, column_id = NULL`) и возвращаются в Inbox.
4. **Тактильный интерфейс Washi/Matcha**:
   - Канбан-доска на `@dnd-kit/core` со стилизацией колонок Washi, плавающим `DragOverlay` и круговыми чекбоксами 18px;
   - Переключатель видов `ViewToggle`: Канбан-доска (⊞) vs плоский список (☰).

---

## 🎯 2. Выполненные задачи и Затронутые подпакеты

1. **Миграции БД и SQL**:
   - `backend/migrations/000003_projects_and_boards.up.sql`: таблицы `projects`, `boards`, `columns`, изменение `position` на `DOUBLE PRECISION`.
   - `backend/queries/projects.sql`: sqlc-запросы CRUD, `CountTasksInColumn`, `MoveTaskToColumnAndPosition`, `ResetProjectTasksToInbox`.
2. **OpenAPI Контракты и Кодогенерация**:
   - `api/openapi/paths/projects.yaml`, `api/openapi/schemas/project.yaml`.
   - Сборка спецификации через `@redocly/cli bundle` $\rightarrow$ `api/dist/openapi.json`.
   - Генерация Echo-сервера и моделей в `internal/infrastructure/http/generated/api.gen.go` через `oapi-codegen`.
3. **Чистый доменный слой (`internal/domain/project`)**:
   - `project.go`, `board.go`, `column.go`, `errors.go`, `project_test.go`, `column_test.go`.
4. **Слой Use Cases (`internal/application/project`)**:
   - `ports.go`, `dto.go`, `service.go`, `service_test.go` (table-driven с `testify` и `mockery v3`).
5. **Инфраструктурный слой (`internal/infrastructure/`)**:
   - Репозитории: `repository/project_repository.go`.
   - HTTP хендлеры: `http/project_handler.go`, `http/project_test.go`.
   - Composition Root: `container/app.go`, `repos.go`, `services.go`, `router.go`.
6. **Фронтенд (`frontend/src/`)**:
   - Сущность `entities/project`: типы, API клиент и хуки TanStack Query v5 с оптимистичными обновлениями.
   - Фича `features/kanban-board`: `KanbanBoard`, `KanbanColumn`, `KanbanCard`, `ViewToggle`, хук `useKanbanDnd`.
   - Виджеты: навигация по проектам в `widgets/sidebar`, переключение видов Inbox / Project в `widgets/main-sheet`.

---

## 📊 3. Метрики качества и Приемочный контроль

* **`make check-arch`**: `PASS (0 ошибок)` по всем 5 уровням проверки (лимит строк < 400, число файлов <= 7, чистота домена, отсутствие сырых цветов Tailwind, структура internal).
* **`make test`**: `100% PASS` с флагом `-race` во всех подпакетах бэкенда.
* **`cd frontend && npm run build`**: `0 ошибок, 0 предупреждений` (Vite build собран успешно).

