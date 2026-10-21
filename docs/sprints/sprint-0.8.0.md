# Отчет о спринте 0.8.0: Механизм подзадач WEEEK-style (Иерархия, прогресс, чеклист на карточке, хлебные крошки и автоматизация)

* **Статус**: ✅ Завершен (100% Definition of Done, 0 ошибок `make check-arch`, 0 ошибок `npm run build`, `PASS` всех тестов)
* **Дата закрытия**: 2026-10-21

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

1. **Рекурсивная модель данных (Recursive First-Class Tasks)**:
   - В таблицу `tasks` добавлено поле `parent_id UUID NULL REFERENCES tasks(id) ON DELETE CASCADE`.
   - Созданы индексы `idx_tasks_parent_id` (частичный) и `idx_tasks_workspace_parent`.
   - Подзадача является полноправной задачей системы со своими тегами, сроком, приоритетом и вложениями. При создании подзадача наследует `project_id` и `column_id` родителя.
   - Запрещено самовложение задачи (`ErrSelfParenting`).
2. **Чистота Канбан-доски (Zero Duplication on Board)**:
   - В SQL-запросах верхнего уровня (`ListTasksByProject`, `ListTasksByColumn`, `GetInboxTasks`, `GetTodayTasks`) добавлено строгое условие `WHERE parent_id IS NULL`.
   - Подзадачи не засоряют колонки отдельными дублирующими карточками, а визуализируются внутри карточки родительской задачи.
3. **Пакетный подсчет счетчиков прогресса (Subtask Counters)**:
   - Добавлен запрос `GetSubtaskCountsByParentIds` (возвращает `total_count` и `completed_count` за один эффективный SQL-запрос через `GROUP BY parent_id`).
   - В схемы OpenAPI `Task` и Go DTO добавлены поля `subtasksCount` и `subtasksCompletedCount`.
4. **WEEEK-Style UI/UX Канбан-карточки (`KanbanCard.tsx`, скриншот 1)**:
   - Отображение короткого номера `#86`.
   - При наличии подзадач выводится иконка ветвления (`GitFork`), горизонтальный прогресс-бар (`bg-canvas` с заполнением `bg-accent-solid`) и счетчик `X/Y`.
   - Раскрываемый список чеклиста подзадач прямо на карточке: клик по круглому чекбоксу выполняет мгновенный тоггл без раскрытия шторки родителя (`e.stopPropagation()`).
5. **WEEEK-Style UI/UX Шторки задачи (`TaskDrawer.tsx`, скриншоты 2 и 3)**:
   - Быстрый индикатор `⏱ X/Y` под заголовком.
   - Секция `⌄ Подзадачи X/Y` (`TaskSubtasksSection.tsx`) со списком подзадач, чекбоксами и инлайн-инпутом создания `+ Новая задача... (Enter)`.
   - При открытии подзадачи: в хедере отображаются хлебные крошки `#parent > #child` и кликабельная кнопка возврата к родителю `⮤ <Название родительской задачи>` с иконкой `CornerUpLeft`.
6. **Автоматизация завершения родительской задачи (Rule `auto_complete_parent_on_subtasks_done`)**:
   - По умолчанию **выключена** (как запрошено пользователем).
   - При включении в колонке/проекте: когда завершается последняя подзадача (все дочерние задачи выполнены), EventBus автоматически переводит родительскую задачу в `is_completed = true`.

---

## 📋 2. Реализованный скоуп файлов

* `backend/migrations/000013_subtasks.up.sql` & `.down.sql`
* `backend/queries/tasks.sql`, `backend/queries/projects.sql`
* `api/openapi/schemas/task.yaml`, `api/openapi/paths/tasks.yaml`, `api/openapi/openapi.yaml`
* `backend/internal/domain/task/task.go`, `repository.go`, `errors.go`, `task_test.go`
* `backend/internal/domain/events/event.go`
* `backend/internal/application/task/dto.go`, `usecases.go`, `tags_usecases.go`, `usecases_subtasks_test.go`
* `backend/internal/application/automation/ports.go`, `listener.go`, `subtask_automation.go`, `subtask_automation_test.go`
* `backend/internal/infrastructure/repository/task_repository.go`, `project_repository.go`, `automation_repository.go`, `helpers.go`
* `backend/internal/infrastructure/http/tasks_handler.go`, `tasks_test.go`, `subtasks_test.go`
* `frontend/src/entities/task/types.ts`, `api.ts`, `queries.ts`, `taskNumber.ts`, `index.ts`
* `frontend/src/shared/api/types.ts`, `shared/api/apiClient.ts`
* `frontend/src/features/kanban-board/KanbanCard.tsx`
* `frontend/src/widgets/task-drawer/TaskDrawer.tsx`, `TaskSubtasksSection.tsx`, `types.ts`, `index.ts`
* `frontend/src/app/App.tsx`

---

## 🎯 3. Definition of Done (DoD) Метрики качества

- [x] Все пункты ТЗ спринта 0.8.0 выполнены.
- [x] `make check-arch` — 0 ошибок (все лимиты строк и чистота слоев соблюдены).
- [x] `npm run build && npm run lint` — 0 ошибок компиляции и типов (Vite собран за 5.09s).
- [x] `make test` — 100% юнит-тестов зеленые (включая тесты автоматизации авто-завершения родителя).
