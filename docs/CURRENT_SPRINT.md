# Текущий спринт / Оперативный центр управления (`CURRENT_SPRINT.md`)

> Этот файл — **единственная точка правды** о том, какие решения приняты по текущей задаче и что команда и AI-агенты делают прямо сейчас.
> Работа над каждым спринтом строго делится на **Фазу А (Штаб: Принятие решений)** и **Фазу Б (Конвейер: Реализация и Верификация)**.
> Завершенные спринты архивируются в каталог `docs/sprints/sprint-X.Y.Z.md` с записью в реестр `docs/sprints/README.md`.

## Текущий спринт: Спринт 0.9.2 — «Core Features: Дублирование и Скоупы» 🟡 ФАЗА А (Готово к запуску)

---

## 🧠 1. Аналитика и Зафиксированные микро-решения спринта (Штаб)

* **Бизнес-цель**: Добавить возможность глубокого дублирования задач с подзадачами, тегами и вложениями, а также починить фильтрацию области действия автоматизаций по доске в фоновом слушателе событий.

* **Связь с бэклогом**: Подробные спецификации задач — в [`docs/BACKLOG.md`](BACKLOG.md) (TASK-B03, TASK-B08).

* **Принятые микро-решения**:

  1. **TASK-B03 — Глубокое копирование задачи (Deep Duplicate)**:
     - Новый API: `POST /workspaces/{wid}/tasks/{task_id}/duplicate` (возвращает скопированную `Task`).
     - Что копируется:
       - Название с префиксом `Копия — [Оригинальное название]`.
       - Описание (`description_markdown`), приоритет, проект, колонка, позиция (`original.position + 0.5`).
       - Теги (`task_tags`).
       - Подзадачи (все дочерние задачи с `parent_id = original.id` копируются с новыми UUID и `parent_id = new_task.id`).
       - Вложения (`attachments`): shallow copy (новые записи в `attachments` с теми же `s3_key`, S3 не дублируется).
     - Что НЕ копируется: `due_date`, статус выполнения (`is_completed = false`, `completed_at = null`), история изменений.
     - Операция выполняется в единой транзакции базы данных.
     - UI: пункт «Дублировать задачу» в выпадающем меню карточки задачи (`KanbanCard.tsx`, и в шторке `TaskDrawer` при наличии меню).

  2. **TASK-B08 — Скоуп автоматизаций по доске в EventListener & UI**:
     - В `backend/internal/application/automation/listener.go`:
       - В обработчиках событий задач (`HandleTaskCreated`, `HandleTaskEnteredColumn`, `HandleTaskCompleted`, `HandleTaskFieldChanged`) добавить фильтрацию по `board_id`:
         ```go
         if rule.BoardID != nil && (taskInfo.ProjectID == nil || *rule.BoardID != *taskInfo.ProjectID) {
             continue
         }
         ```
     - В `frontend/src/features/automations/AutomationsHub.tsx`:
       - Обновить категории фильтров: вместо `'all' / 'column' / 'global'` сделать `'all' / 'board' / 'column'` с понятными лейблами «Все», «На доску», «На колонку».
     - В `frontend/src/features/automations/RuleBuilderModal.tsx` и `RuleBuilderSections.tsx`:
       - Добавить выбор «Область применения» (Вся доска / Конкретная колонка / Глобально) с сохранением соответствующего `board_id` / `column_id`.

---

## 📋 2. Таблица задач спринта 0.9.2

| # | Задача | Статус | Приоритет | Ответственный | Файлы |
|---|--------|--------|-----------|---------------|-------|
| 1 | **[B03 Step 1] Duplicate Task Contracts & SQL**: OpenAPI эндпоинт, queries sqlc, make generate, чистый домен | ✅ Готово | Высокий | `backend_developer` | `api/openapi/*`, `queries/tasks.sql` |
| 2 | **[B03 Step 2] Duplicate Task Application & Infra**: Use Case `DuplicateTask`, транзакция, http хендлер, тесты | ✅ Готово | Высокий | `backend_developer` | `application/task/`, `infrastructure/http/` |
| 3 | **[B08 Backend] Automation Board Scope Filter**: фильтрация по `board_id` в `listener.go` + unit-тесты | ✅ Готово | Высокий | `backend_developer` | `application/automation/listener.go` |
| 4 | **[B03 UI] Duplicate Task Action**: пункт «Дублировать задачу» в меню карточки | ⏳ Ожидает | Средний | `frontend_developer` | `KanbanCard.tsx` |
| 5 | **[B08 UI] Automations Scope Selector & Filters**: фильтры в Хабе и область применения в модалке | ⏳ Ожидает | Средний | `frontend_developer` | `AutomationsHub.tsx`, `RuleBuilderModal.tsx` |
| 6 | **[Verification]**: `make check-arch`, `npm run build`, `npm run lint`, `make test`, коммит | ⏳ В очереди | Критический | Lead Orchestrator | — |

---

## 🎯 3. Definition of Done (DoD)

- [x] API `POST /workspaces/{wid}/tasks/{task_id}/duplicate` создаёт полную копию задачи с подзадачами, тегами и вложениями под префиксом «Копия — » в транзакции.
- [ ] В контекстном меню карточки задачи на канбан-доске есть пункт «Дублировать», при клике задача копируется и появляется на доске.
- [x] Правила автоматизации с `board_id` корректно применяются только к задачам заданной доски и игнорируют другие доски.
- [ ] В Хабе автоматизаций доступны фильтры «Все», «На доску», «На колонку», а в конструкторе правил есть выбор области применения.
- [ ] `make check-arch` — 0 ошибок.
- [ ] `npm run build && npm run lint` — 0 ошибок.
- [ ] `make test` — все тесты зелёные.

---

## 🗓️ 4. Очередь следующих спринтов

| Спринт | Задачи | Суть |
|--------|--------|------|
| **0.9.3** | B05 | Архив задач (миграция + Batch SQL ADR-0008) |
| **0.9.4** | B06 | История изменений задачи (новая таблица + EventBus listener) |
| **0.9.5** | B09 | Профиль пользователя: имя + аватар |
| **Future** | B01 | Subtasks Automation Suite (L-задача) |

