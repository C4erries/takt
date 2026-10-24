# Спринт `0.9.1`: UX Polish: Создание задач, Подзадачи, Теги

* **Даты проведения**: 2026-10-24
* **Статус**: ✅ Завершен (100% Definition of Done)
* **Связанные задачи**: TASK-B12, TASK-B02, TASK-B13

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

* **Бизнес-цель**: Три связанных UX-улучшения вокруг создания задач, эргономики подзадач и централизованного управления тегами воркспейса.
* **Ключевые продуктовые правила и UI решения**:
  1. **TASK-B12 — Хоткей заметок при создании задачи**:
     - В форме инлайн-создания (`KanbanColumn.tsx`): однострочный input заголовка по умолчанию.
     - `Shift+Enter` в поле заголовка плавно раскрывает `textarea` заметок/описания с кинетической анимацией (`motion/react` `AnimatePresence`).
     - Фокус автоматически переходит в `textarea`.
     - `Ctrl+Enter` (или `Cmd+Enter`) из любого поля создает задачу с названием и описанием (`descriptionMarkdown`).
     - Обычный `Enter` в input заголовка создает задачу только с названием (если textarea была раскрыта — сворачивается и очищается).
  2. **TASK-B02 — Эргономика подзадач + привязка существующей задачи**:
     - В `TaskSubtasksSection.tsx` увеличена высота инпута: `min-h-[34px] py-1.5`.
     - Кнопка-иконка `Link2` рядом с кнопкой добавления подзадачи открывает поповер `AttachTaskPopover.tsx`.
     - Поиск через `GET /workspaces/{id}/search?q=...` с дебаунсом 200 мс.
     - Фильтрация: исключены текущая задача (`task.id !== currentTask.id`), задачи с `parentId != null` (нет цепочек подзадач подзадачи) и задачи, у которых уже есть свои подзадачи (запрет 2-го уровня вложенности).
     - Привязка задачи выполняется через `PATCH /workspaces/{wid}/tasks/{id}` с `{ parentId: currentTask.id }` с инвалидацией кэша TanStack Query.
  3. **TASK-B13 — Централизованный менеджер тегов воркспейса**:
     - Точка входа: кнопка с иконкой `Tag` в сайдбаре (`Sidebar.tsx`) в блоке быстрых инструментов.
     - Новые API-контракты:
       - `GET /workspaces/{wid}/tags/metrics` — теги с числом привязанных задач `task_count`.
       - `PATCH /workspaces/{wid}/tags/{id}` — обновление имени/цвета тега.
       - `DELETE /workspaces/{wid}/tags/{id}` — удаление тега (с каскадной отвязкой от задач).
       - `DELETE /workspaces/{wid}/tags/unused` — удаление всех тегов с `task_count = 0`.
     - Модалка `TagManagerModal.tsx`: фильтрация в реальном времени, 12-цветная палитра `TAG_PALETTE`, inline-переименование, подтверждение удаления («Тег будет отвязан от X задач») и кнопка «Удалить неиспользуемые».

---

## 🎯 2. Ключевые результаты и Затронутые подпакеты

1. **Контракты API (`api/openapi/`)**:
   - `schemas/tag.yaml`: добавлена схема `TagWithMetrics`.
   - `paths/tags.yaml`: добавлены операции `tagsMetrics` и `unusedTags`.
   - `openapi.yaml`: пути `/tags/metrics` и `/tags/unused` зарегистрированы до `/tags/{tag_id}`.
2. **База данных и SQL (`backend/queries/` & `internal/infrastructure/db/sqlc/`)**:
   - `tags.sql`: добавлены запросы `ListWorkspaceTagsWithMetrics` и `DeleteUnusedTags`.
   - sqlc: сгенерированы методы в `tags.sql.go` и `querier.go`.
3. **Чистый домен и Application (`backend/internal/domain/task/` & `internal/application/task/`)**:
   - `domain/task/tag.go`: модель `TagWithMetrics` и инварианты `NewTagWithMetrics`.
   - `application/task/ports.go`: добавлены методы в `TagRepository` и `UseCases`.
   - `application/task/tags_usecases.go`: сценарии `ListWorkspaceTagsWithMetrics` и `DeleteUnusedTags` с Tenant Access Check.
   - `application/task/usecases_tags_metrics_test.go`: 100% покрытие use cases тестами.
4. **Инфраструктура бэкенда (`backend/internal/infrastructure/`)**:
   - `repository/task_repository.go`: реализация методов работы с тегами и метриками.
   - `http/tasks_handler.go`: обработчики `ListWorkspaceTagsMetrics` и `DeleteUnusedWorkspaceTags`.
   - `http/tags_test.go`: юнит-тесты Echo-хендлеров.
5. **Пользовательский интерфейс (`frontend/src/`)**:
   - `features/kanban-board/KanbanColumn.tsx`: инлайн-создание с `Shift+Enter` описанием и `Ctrl+Enter` сабмитом.
   - `widgets/task-drawer/AttachTaskPopover.tsx`: поиск и привязка существующих задач.
   - `widgets/task-drawer/TaskSubtasksSection.tsx`: эргономичный инпут и кнопка привязки задачи.
   - `features/tag-manager/`: модуль `TagManagerModal.tsx` и TanStack Query хуки.
   - `widgets/sidebar/Sidebar.tsx`: интеграция точки входа «Теги».

---

## 📊 3. Метрики качества, Архитектурный контроль и Безопасность

* **Архитектурный линтер (`make check-arch`)**:
  - `PASS (0 ошибок)`: лимит строк (< 400 строк), лимит файлов (<= 7 в пакетах), слои DDD и чистота токенов подтверждены.
* **Юнит-тесты бэкенда (`make test`)**:
  - `100% PASS` с флагом `-race`.
* **Сборка и линтер фронтенда (`npm run lint && npm run build`)**:
  - `0 ошибок, 0 ворнингов`.
* **Сквозные автотесты (`tests/` uv run pytest)**:
  - `39 из 39 тестов пройдены успешно (100% PASS)`.
* **Docker Compose**:
  - Все сервисы пересобраны и запущены в контейнерах, `backend` слушает `:8080` без ошибок.
