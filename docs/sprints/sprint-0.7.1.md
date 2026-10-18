# Спринт 0.7.1: UI/UX Elevation: Умная система тегов, приоритет «В ожидании», селекторы и панорамирование колесиком мыши

* **Период проведения**: 2026-10-19
* **Статус**: 🟢 Завершен (100% DoD)
* **Ветка**: `main`

---

## 🎯 Цели и Результаты спринта

1. **Багфикс надежности тегов (пакетная подгрузка тегов задач)**:
   - В [tags.sql](file:///home/pavel/projects/takt/backend/queries/tags.sql) добавлен пакетный SQL-запрос `ListTagsForTasks` (`WHERE tt.task_id = ANY($1::uuid[])`), выполнен `make generate`.
   - В [helpers.go](file:///home/pavel/projects/takt/backend/internal/infrastructure/repository/helpers.go) и [project_repository.go](file:///home/pavel/projects/takt/backend/internal/infrastructure/repository/project_repository.go) методы `ListTasksByProject` и `ListTasksByColumn` дополнены вызовом `loadTagsForTasks(ctx, q, res)`: теги задач теперь загружаются и привязываются ко всем карточкам проекта, устранена проблема затирания тегов при обновлении карточки.

2. **Багфикс удаления дедлайна (гарантированное стирание в БД)**:
   - В [tasks_handler.go](file:///home/pavel/projects/takt/backend/internal/infrastructure/http/tasks_handler.go) в методе `UpdateTask` добавлено чтение сырого тела запроса: при передаче `{"dueDate": null}` выставляется `input.ClearDueDate = true`.
   - Дата дедлайна теперь окончательно стирается в PostgreSQL и не возвращается при перезагрузке страницы.

3. **Умный автокомплит тегов и 6-цветная пастельная палитра (без `#`)**:
   - Выделен компонент [TaskTagSection.tsx](file:///home/pavel/projects/takt/frontend/src/widgets/task-drawer/TaskTagSection.tsx) (269 строк, лимиты файлов `<= 7` строго соблюдены).
   - Живой поиск по существующим тегам воркспейса при вводе в инпут карточки задачи с фильтрацией на лету и быстрым прикреплением в 1 клик.
   - Мини-палитра из 6 пастельных оттенков (Деним `#4A6B82`, Хурма `#D97746`, Матча `#566B44`, Солома `#B8943F`, Слива `#8A4F7D`, Суми `#5C5C5C`) с возможностью создания нового тега.
   - Символ `#` удален из всех карточек Канбана, шторки задачи, фильтров и бейджей — лаконичные чистые капсулы с цветной индикаторной точкой.

4. **Ревизия приоритетов: замена «Срочно» $\rightarrow$ «В ожидании» (`pending`)**:
   - Выполнена SQL-миграция [000010_priority_pending.up.sql](file:///home/pavel/projects/takt/backend/migrations/000010_priority_pending.up.sql): обновлен CHECK constraint, сконвертированы существующие задачи `urgent` $\rightarrow$ `pending`.
   - В OpenAPI схемах [task.yaml](file:///home/pavel/projects/takt/api/openapi/schemas/task.yaml), [project.yaml](file:///home/pavel/projects/takt/api/openapi/schemas/project.yaml) и [automation.yaml](file:///home/pavel/projects/takt/api/openapi/schemas/automation.yaml) добавлен `pending`.
   - В интерфейсе селектора приоритетов: активная кнопка подсвечивается **собственным цветом** приоритета (синий деним, терракотовый, соломенный, бамбуковый) вместо единого зеленого.

5. **Панорамирование Канбан-доски колесиком мыши (Middle Mouse Click Pan)**:
   - В [KanbanBoard.tsx](file:///home/pavel/projects/takt/frontend/src/features/kanban-board/KanbanBoard.tsx) прямо на контейнере со скроллом реализован захват клика колесика (`button === 1`), режим панорамирования с курсором `cursor-grabbing` и плавным drag-scroll (`scrollLeft`).

---

## 👥 Команда и Распределение ролей

* **Lead-оркестратор**: Проектирование микро-решений D-071-1..4, гейткипер-верификация контрактов, применение SQL-миграции, пересборка backend-контейнера и E2E Playwright верификация.
* **`backend_developer`**: Шаг 1 (OpenAPI схемы, миграция 000010, домен `internal/domain/task/`), Шаг 2 (пакетный запрос `ListTagsForTasks`, `loadTagsForTasks` в `project_repository.go`, `ClearDueDate` в `tasks_handler.go` и unit-тесты).
* **`frontend_developer`**: Компонент `TaskTagSection.tsx`, автокомплит и палитра цветов тегов, удаление символа `#`, приоритет `pending` с self-colored кнопками, drag-to-scroll колесиком мыши в `KanbanBoard.tsx`.
* **`qa_engineer` & Lead**: Независимое тестирование Playwright в браузере Chromium с фиксацией 4 артефактов.

---

## 🧪 Метрики качества и верификация

1. **`make check-arch`**: PASSED (0 ошибок, лимит строк < 400 во всех файлах, число файлов в пакетах <= 7).
2. **`make test`**: PASSED (100% Go unit-тестов с флагом `-race`).
3. **`npm run build`**: PASSED (0 ошибок компиляции TypeScript и Vite).
4. **Браузерная верификация Playwright**:
   - `sprint_071_1_tag_autocomplete_and_palette.png` — живой поиск тегов, 6-цветная палитра и активный «В ожидании».
   - `sprint_071_2_priority_self_colored_pending.png` — созданный тег без `#`, очищенный срок (`+ Срок`) и self-colored селекторы.
   - `sprint_071_3_cleared_due_date_and_clean_tags.png` — персистентность тега и очищенного дедлайна после перезагрузки страницы (F5).
   - `sprint_071_4_middle_click_panned_board.png` — плавное горизонтальное панорамирование Канбан-доски колесиком мыши.

