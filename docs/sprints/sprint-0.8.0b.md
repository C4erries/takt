# Отчет о спринте 0.8.0b: Docker Compose Migrator, MoveTaskModal с поиском, 12-цветовая палитра тегов, скрытие номеров задач и Error Observability

* **Статус**: ✅ Завершен (100% Definition of Done, 0 ошибок `make check-arch`, 0 ошибок `npm run build`, `PASS` всех тестов)
* **Дата закрытия**: 2026-10-21

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

1. **Автоматический сервис `migrator` в `docker-compose.yml`**:
   - Добавлен сервис `migrator` на базе официального образа `migrate/migrate:v4.17.0`.
   - Монтирует `backend/migrations` и выполняет команду `up` перед запуском приложения.
   - В сервисе `backend` настроен `depends_on: { migrator: { condition: service_completed_successfully } }`.
   - Это гарантирует, что все SQL-миграции автоматически применяются к базе данных при каждом запуске проекта, исключая падения из-за отсутствующих колонок/таблиц.
2. **Error Observability в бэкенде**:
   - В `middleware.go`, `tasks_handler.go`, `project_handler.go` при возврате ошибки с кодом $\ge 500$ добавлен детальный вывод `log.Printf("[ERROR 500] %s %s: %+v", ...)` в консоль контейнера.
   - Для внешнего клиента ответ остается безопасным (`"Internal server error"`), а в консоли Docker сразу видна точная первопричина сбоя.
3. **Модалка перемещения задачи (`MoveTaskModal.tsx`, Linear/WEEEK-style)**:
   - В контекстном меню задачи (3 точки) пункт «Переместить на доску» открывает масштабируемый диалог перемещения.
   - Поле поиска (автофокус) фильтрует доски и колонки в реальном времени.
   - Сгруппированный список досок и их колонок с ограничением высоты `max-h-72 overflow-y-auto` и стилизованным скроллбаром: интерфейс идеально работает даже при 50+ досках и колонках и никогда не вылезает за пределы экрана.
   - Текущая колонка отмечена бейджем `(текущая)` и деактивирована.
   - Клик по колонке мгновенно перемещает задачу в выбранную доску и колонку.
4. **Расширенная палитра тегов (12 оттенков) и редактирование цвета**:
   - В `tags.ts` палитра расширена до 12 премиальных оттенков GTD/Linear: Persimmon, Terracotta, Amber, Bamboo, Emerald, Ocean, Denim, Indigo, Lavender, Plum, Cocoa, Sand.
   - Реализован эндпоинт `PATCH /workspaces/{id}/tags/{id}` и usecase `UpdateTag`, позволяющий персистентно менять цвет и название существующего тега.
   - В шторке задачи (`TaskTagSection.tsx`) клик по прикрепленному тегу открывает компактный поповер палитры: выбор нового цвета сразу обновляет тег на всех карточках воркспейса.
5. **Скрытие номеров задач по умолчанию + тумблер видимости**:
   - На карточках канбана бейдж номера `#ID` скрыт по умолчанию.
   - В панель переключения видов (`ViewToggle.tsx`) добавлен тумблер с иконкой `Hash` и бейджем `#ID` для опционального включения номеров задач (настройка сохраняется в `localStorage`).

---

## 📋 2. Реализованный скоуп файлов

* `docker-compose.yml`: сервис `migrator` и зависимость бэкенда.
* `backend/queries/tags.sql`: запрос `UpdateTag`.
* `api/openapi/paths/tags.yaml`, `schemas/tag.yaml`, `openapi.yaml`: спецификация `PATCH /tags/{id}`.
* `backend/internal/domain/task/tag.go`, `tag_test.go`: валидация цвета HEX (`SetColorHex`).
* `backend/internal/application/task/dto.go`, `ports.go`, `usecases.go`, `tags_usecases.go`, `usecases_update_tag_test.go`: Use Cases обновления тега.
* `backend/internal/infrastructure/repository/task_repository.go`: реализация `UpdateTag`.
* `backend/internal/infrastructure/http/middleware.go`, `tasks_handler.go`, `project_handler.go`, `tags_test.go`: логирование 500 ошибок и хендлер тегов.
* `frontend/src/entities/task/tags.ts`: 12-цветовая палитра.
* `frontend/src/entities/task/api.ts`, `queries.ts`: API и мутация `useUpdateTagMutation`.
* `frontend/src/entities/task/taskNumber.ts`: реактивный хук `useShowTaskNumbers` (`localStorage`).
* `frontend/src/features/kanban-board/MoveTaskModal.tsx`: модалка перемещения с поиском.
* `frontend/src/features/kanban-board/KanbanCard.tsx`: скрытие номера по умолчанию, вызов модалки.
* `frontend/src/features/kanban-board/ViewToggle.tsx`: тумблер отображения номеров задач.
* `frontend/src/widgets/task-drawer/TaskTagSection.tsx`: поповер смены цвета тега.
* `frontend/src/features/kanban-board/automations/ColumnAutomationsModal.tsx`: декомпозиция пакета для соблюдения лимита <= 7 файлов.

---

## 🎯 3. Definition of Done (DoD) Метрики качества

- [x] Все пункты ТЗ спринта 0.8.0b выполнены.
- [x] `make check-arch` — 0 ошибок (все лимиты строк и чистота слоев соблюдены).
- [x] `npm run build && npm run lint` — 0 ошибок компиляции и типов (Vite собран за 4.89s).
- [x] `make test` — 100% юнит-тестов зеленые.

