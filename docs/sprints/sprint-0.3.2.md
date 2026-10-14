# Спринт 0.3.2: Правая Bento-панель (Task Inspector, Матричный календарь) и Фокус «Сегодня»

* **Даты проведения**: 2026-10-14 — 2026-10-14
* **Статус**: ✅ Завершен (100% Definition of Done)
* **Связанные документы**: [docs/CURRENT_SPRINT.md](../CURRENT_SPRINT.md), [docs/conventions/backend.md](../conventions/backend.md), [docs/conventions/frontend.md](../conventions/frontend.md)

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

1. **Модель тегов контекста**: Таблицы `tags` и `task_tags`. Единый справочник тегов воркспейса (`@дом`, `@работа`, `20 мин`, `focus`) с капсульными бейджами и быстрым добавлением/удалением.
2. **Гибридный GTD-отбор в «Сегодня» (`Today`)**:
   - Автоматически отбираются задачи, где `due_date <= CURRENT_DATE` (включая просроченные с акцентным бейджем) ИЛИ `focus_order IS NOT NULL` (задачи, явно перенесенные пользователем в фокус дня);
   - Сортировка: `focus_order ASC NULLS LAST, due_date ASC, priority DESC`.
3. **Контекстно-зависимый клик в Матричном календаре**:
   - Если открыт инспектор задачи — клик по ячейке даты мгновенно назначает/обновляет `due_date` этой задачи;
   - Если инспектор закрыт — фильтрует список задач по выбранной дате;
   - Точечные маркеры плотности дедлайнов задач под числами.
4. **Editorial Split-Toggle Markdown Inspector**:
   - В покое: изящный типографичный рендер заметки (шрифты Newsreader / Geist Sans);
   - По клику: бесшовный инлайн textarea-редактор с мгновенным autosave (дебаунс 400мс) и поддержкой `Escape` / клик вне поля;
   - Мгновенные мутации для селекторов приоритета (`none`, `low`, `medium`, `high`, `urgent`), проекта, колонки и дедлайна.

---

## 🎯 2. Выполненные задачи и Затронутые подпакеты

1. **Миграции БД и SQL**:
   - `backend/migrations/000004_tags_and_focus.up.sql`: таблицы `tags`, `task_tags`, индексы.
   - `backend/queries/tags.sql`: запросы для CRUD тегов и связей с задачами.
   - `backend/queries/tasks.sql`: запросы `UpdateTask`, `GetTodayTasks`, `SetTaskFocus`.
2. **OpenAPI Контракты и Кодогенерация**:
   - `api/openapi/schemas/tag.yaml`, `api/openapi/paths/tags.yaml`.
   - Расширение `api/openapi/schemas/task.yaml` и `api/openapi/paths/tasks.yaml` (`/tasks/today`, `PATCH /tasks/{id}`, `POST /tasks/{id}/focus`).
   - Кодогенерация `make generate` (`oapi-codegen ServerInterface` + `sqlc`).
3. **Чистый доменный слой (`internal/domain/task`)**:
   - `tag.go`, `tag_test.go`, расширение сущности `Task` методами `AddTag`, `RemoveTag`, `SetFocusOrder`.
4. **Use Cases (`internal/application/task`)**:
   - `ports.go`, `dto.go`, `usecases.go`, `usecases_focus_test.go`, `usecases_tags_test.go` с моками mockery v3.
5. **Инфраструктура (`internal/infrastructure/`)**:
   - `repository/task_repository.go`: методы тегов и обновлений задач.
   - `http/tasks_handler.go`: реализация методов `ServerInterface` (`GetTodayTasks`, `UpdateTask`, `SetTaskFocus`, `ListWorkspaceTags`, `CreateWorkspaceTag`, `DeleteWorkspaceTag`, `AddTagToTask`, `RemoveTagFromTask`).
   - `http/tasks_test.go`, `http/tags_test.go`.
6. **Фронтенд (`frontend/src/`)**:
   - `entities/task/`: типы `Tag`, API методы и хуки TanStack Query v5 с оптимистичными мутациями.
   - `widgets/task-inspector/`: `TaskInspector.tsx` и `MarkdownNotes.tsx` (Editorial Split-Toggle).
   - `widgets/mini-calendar/`: `MiniCalendar.tsx` с разделителями Tatami, плотностью дедлайнов и контекстным кликом.
   - `widgets/main-sheet/`: `TodayView.tsx` со специализированным гибридным GTD-отбором и J/K навигацией.
   - `widgets/sidebar/`: отображение счетчика задач на сегодня и навигация.
7. **E2E-тестирование (`tests/e2e/`)**:
   - `tests/e2e/test_projects_and_bento.py`: 100% покрытие жизненного цикла проектов, колонок, DND, инспектора, тегов и фокуса дня на Pytest.

---

## 📊 3. Метрики качества и Приемочный контроль

* **`make check-arch`**: `PASS (0 ошибок)` по всем 5 проверкам архитектурного линтера.
* **`make test`**: `100% PASS` с флагом `-race` во всех пакетах бэкенда.
* **`make test-e2e`**: `17/17 PASS (100%)` в Docker-окружении.
* **`cd frontend && npm run build`**: `0 ошибок, 0 предупреждений` (Vite build собран за 3.41s).

