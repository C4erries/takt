# Текущий спринт / Оперативный центр управления (`CURRENT_SPRINT.md`)

> Этот файл — **единственная точка правды** о том, какие решения приняты по текущей задаче и что команда и AI-агенты делают прямо сейчас.
> Работа над каждым спринтом строго делится на **Фазу А (Штаб: Принятие решений)** и **Фазу Б (Конвейер: Реализация и Верификация)**.
> Завершенные спринты архивируются в каталог `docs/sprints/sprint-X.Y.Z.md` с записью в реестр `docs/sprints/README.md`.

## Текущий спринт: Спринт 0.9.7 — «Automations Suite & Профиль пользователя» ✅ ЗАВЕРШЕН

---

## 🧠 1. Аналитика и Зафиксированные микро-решения спринта (Штаб)

* **Бизнес-цель**: Реализация продвинутого комплекса правил автоматизации (возврат в предыдущую колонку, триггеры смены статуса выполнения задачи, автоматизация каскада подзадач) и персонализация пользователя (редактирование имени и загрузка аватара в S3-хранилище).

* **Связь с бэклогом**: [`docs/BACKLOG.md`](BACKLOG.md) (TASK-B20, TASK-B21, TASK-B01, TASK-B09).

* **Принятые микро-решения**:

  1. **TASK-B20 (Возврат в предыдущую колонку при отмене выполнения)**:
     - Новое действие в DSL автоматизаций: `ActionReturnToPreviousColumn = "return_to_previous_column"`.
     - Алгоритм: из таблицы `task_activities` считывается последнее событие перемещения задачи между колонками (`activity_type = 'column_moved'`), откуда извлекается `from_column_id`.
     - **Fallback стратегия (утверждена пользователем)**:
       1) Если колонка `from_column_id` существует и не удалена — задача перемещается туда.
       2) Если колонка не найдена или перемещений не было — перемещение в колонку Inbox текущей доски (`is_inbox = true` на доске задачи).
       3) Если у доски нет колонки Inbox — сброс в общие Входящие воркспейса (`column_id = NULL`).
     - Поддержка в UI конструктора правил (`RuleBuilderSections.tsx`).

  2. **TASK-B21 (Расширенный триггер статуса выполнения задачи)**:
     - Новые/уточненные триггеры в DSL:
       - `TriggerTaskCompleted = "task_completed"` (по умолчанию).
       - `TriggerTaskUncompleted = "task_uncompleted"` (снятие отметки выполнения).
     - В конструкторе UI: триггер «Смена статуса выполнения» с подвыбором статуса (`Выполнена` / `Отменено выполнение`).
     - В `EventListener`: подписка на события смены `is_completed` с проверкой флага и вызовом правил.

  3. **TASK-B01 (Автоматизация подзадач / Subtasks Automation Suite)**:
     - Новые триггеры: `TriggerAllSubtasksCompleted = "all_subtasks_completed"`, `TriggerAnySubtaskCompleted = "any_subtask_completed"`, `TriggerSubtaskCreated = "subtask_created"`.
     - Новые условия: `subtasks_all_completed` (bool), `subtasks_count` (int, операторы `>`, `==`, `0`).
     - Новые действия: `complete_all_subtasks`, `reopen_all_subtasks`.
     - Защита от бесконечных циклов (`Cycle Guard`) и запрет создания подзадач глубже 1 уровня.

  4. **TASK-B09 (Профиль пользователя: имя и аватар)**:
     - Миграция `000017_user_profile.up.sql`: добавление полей `display_name VARCHAR(100)` и `avatar_url TEXT` в таблицу `users`.
     - OpenAPI контракт:
       - `PATCH /users/me`: обновление `displayName`.
       - `POST /users/me/avatar`: загрузка изображения (multipart/form-data), валидация размера (< 5MB) и MIME-типа (image/jpeg, png, webp). Сохранение в MinIO S3 и генерация постоянного URL.
       - Поля `displayName` и `avatarUrl` в объекте `User`.
     - Фронтенд: компонент `UserProfileModal` при клике на плашку пользователя в сайдбаре (`Sidebar`), превью аватара, загрузка файла с обрезкой/индикатором и редактирование имени.

---

## 📋 2. Таблица задач спринта 0.9.7

| # | Задача | Статус | Приоритет | Ответственный | Файлы |
|---|--------|--------|-----------|---------------|-------|
| 1 | **[Step 1 Backend] Schema, OpenAPI & Domain**: миграция 000017 (user profile), queries `users.sql`, `api/openapi/*` (users/me, avatar, triggers/actions), domain models & invariant tests | ✅ Завершено | Высокий | `backend_developer` | `migrations/`, `queries/`, `api/openapi/`, `internal/domain/automation/`, `internal/domain/workspace/` |
| 2 | **[Step 2 Backend] Automation Suite Evaluator & User Profile**: реализация `return_to_previous_column` с fallback, subtask triggers/actions, upload avatar to S3, HTTP handlers, тесты | ✅ Завершено | Высокий | `backend_developer` | `application/automation/`, `application/workspace/`, `infrastructure/repository/`, `infrastructure/http/` |
| 3 | **[Step 2 Frontend] Automations UI & User Profile Modal**: поддержка новых триггеров и действий в RuleBuilder, UserProfileModal в сайдбаре, загрузка аватара и отображение в UI | ✅ Завершено | Высокий | `frontend_developer` | `RuleBuilderSections.tsx`, `useRuleBuilder.ts`, `UserProfileModal.tsx`, `Sidebar.tsx` |
| 4 | **[Verification]**: `make check-arch`, `make test`, `npm run build`, `docker compose up -d --build`, `pytest tests/`, коммит | ✅ Завершено | Критический | Lead Orchestrator | — |

---

## 🎯 3. Definition of Done (DoD)

- [x] Создана и применена миграция `000017_user_profile` (`display_name`, `avatar_url`).
- [x] В OpenAPI описаны эндпоинты `PATCH /users/me` и `POST /users/me/avatar`.
- [x] Действие `return_to_previous_column` корректно находит прошлую колонку по `task_activities`, а при отсутствии откатывается на `is_inbox = true` доски или общее Inbox воркспейса.
- [x] Поддерживаются триггеры `task_uncompleted`, `all_subtasks_completed`, `any_subtask_completed`, `subtask_created`.
- [x] Поддерживаются действия `complete_all_subtasks`, `reopen_all_subtasks`.
- [x] Пользователь может изменить имя и загрузить аватарку в модальном окне профиля.
- [x] `make check-arch` — 0 ошибок (файлы < 500 строк, папки <= 7 файлов, семантические токены).
- [x] `npm run build && npm run lint` — 0 ошибок.
- [x] `make test` — все тесты зелёные.
- [x] `uv run pytest tests/` — 100% PASS.
