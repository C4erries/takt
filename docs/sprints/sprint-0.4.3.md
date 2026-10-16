# Спринт 0.4.3: Расширенная организация, Мультифильтры и Быстрый поиск Cmd+K

* **Даты проведения**: 2026-10-15 — 2026-10-15
* **Статус**: ✅ Завершен (100% Definition of Done)
* **Связанные документы**: [docs/CURRENT_SPRINT.md](../CURRENT_SPRINT.md), [docs/PROJECT_CONTEXT.md](../PROJECT_CONTEXT.md), [docs/ROADMAP.md](../ROADMAP.md)

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

1. **Эндпоинт поиска и алгоритм (Fast Substring ILIKE)**:
   - Создан эндпоинт `GET /api/v1/workspaces/{workspace_id}/search?q=...`
   - Алгоритм: регистронезависимый поиск `ILIKE '%...%'` по `title` и `description_markdown` для задач, и по `name` и `description` для проектов.
   - Ответ `SearchResponse`: `{ "tasks": [ ... ], "projects": [ ... ] }`.
   - Строгая Tenant Isolation: поиск ограничен исключительно текущим активным `workspace_id` и проверяет членство пользователя. Лимит выдачи: до 20 задач и до 10 проектов.
2. **Мультифильтрация на доске и в списках (Board Filter Bar)**:
   - **Теги**: мультивыбор с логикой **`AND`** (отображаются только те задачи, которые содержат **все** выбранные теги).
   - **Приоритеты**: мультивыбор пилюль (`urgent`, `high`, `medium`, `low`, `none`).
   - **Дедлайн**: быстрые фильтры («Все», «Просрочено», «Сегодня», «На этой неделе», «Без срока»).
   - **Статус**: тумблер «Скрыть выполненные».
   - Кнопка «Сбросить все фильтры», появляющаяся при наличии активных критериев.
3. **Сортировка отображения (View Sorting)**:
   - Выпадающий список сортировки: «Вручную (DND)» (по умолчанию), «По дедлайну», «По приоритету», «По дате создания».
   - Сортировка выполняется на клиенте для мгновенного отклика (0ms) без мутации перманентных координат `position` в БД, сохраняя целостность ручного порядка перетаскивания.
4. **Палитра команд `Cmd+K` / `Ctrl+K` (Search & Command Palette)**:
   - Глобальный хоткей `Cmd+K` (macOS) / `Ctrl+K` (Linux/Windows).
   - При пустом поле поиска: быстрые переходы («Входящие», «Фокус Сегодня», «+ Создать задачу», список проектов).
   - При вводе поискового запроса: группировка результатов («Задачи», «Проекты»), навигация клавишами `↑` / `↓`, выбор по `Enter` (открывает задачу в TaskInspector или переходит в проект), закрытие по `Esc`.
   - Стилистика: Washi Canvas & Matcha (`bg-surface`, `border-subtle`, пружинная анимация появления).

---

## 🎯 2. Выполненные задачи и Затронутые подпакеты

1. **SQL и запросы БД**:
   - `backend/queries/search.sql`: запросы `SearchTasksInWorkspace` и `SearchProjectsInWorkspace` с защитой Tenant Isolation.
   - Сгенерирован `search.sql.go` через `sqlc generate`.
2. **OpenAPI Контракты и Кодогенерация**:
   - `api/openapi/schemas/search.yaml`: схема `SearchResponse`.
   - `api/openapi/paths/search.yaml`: путь `GET /workspaces/{workspace_id}/search`.
   - Сгенерирован обновленный `ServerInterface` в `internal/infrastructure/http/generated/api.gen.go`.
3. **Application Layer (`internal/application/search`)**:
   - `ports.go`: интерфейсы `TaskSearcher`, `ProjectSearcher`, `WorkspaceMembershipChecker` и DTO `SearchResult`.
   - `service.go`: Use Case `Search` с валидацией, обрезкой пробелов и проверкой членства в воркспейсе.
   - Юнит-тесты в `service_test.go` с моками.
4. **Репозитории и HTTP-слой (`internal/infrastructure/`)**:
   - `repository/task_repository.go`: метод `SearchTasks`.
   - `repository/project_repository.go`: метод `SearchProjects`.
   - `http/handler.go`: хендлер `SearchWorkspace` (сохранено строго 7 файлов в пакете `http/`).
   - Тесты эндпоинта поиска в `http/search_test.go`.
   - Регистрация в `container/` (`services.go`, `app.go`).
5. **Фронтенд (`frontend/src/`)**:
   - `shared/api/`: метод `searchWorkspace` и типизация `SearchResponse`.
   - `features/command-palette/`: компонент `CommandPaletteModal` и хук `useCommandPalette` (глобальный хоткей `Cmd+K`, debounce 250ms, быстрые действия, стрелочная навигация).
   - `features/board-filters/`: компонент `BoardFilterBar` и утилиты `filterAndSortTasks` (логика `AND` для тегов, приоритеты, фильтрация по срокам, тумблер скрытия выполненных, селектор View Sorting).
   - Интеграция в `widgets/sidebar/Sidebar.tsx`, `app/App.tsx`, `widgets/main-sheet/ProjectView.tsx` и `features/kanban-board/KanbanBoard.tsx`.
6. **E2E-тестирование (`tests/e2e/test_search.py`)**:
   - 3 новых сквозных теста: поиск задач и проектов по подстроке и описанию, регистронезависимость, изоляция чужих воркспейсов, валидация 400 при отсутствии `q` и 401 для неавторизованных.
   - Итог: **39/39 PASS (100%)**.

---

## 📊 3. Метрики качества и Приемочный контроль

* **`make check-arch`**: `PASS (0 ошибок)` по всем 5 проверкам архитектурного линтера.
* **`make test`**: `100% PASS` с флагом `-race` во всех пакетах бэкенда.
* **`make test-e2e`**: `39/39 PASS (100%)` в Pytest E2E сьюте.
* **`cd frontend && npm run build`**: `0 ошибок, 0 предупреждений` (Vite build собран за 3.91s).

