# Спринт `0.9.2`: Core Features: Дублирование и Скоупы

* **Даты проведения**: 2026-10-24
* **Статус**: ✅ Завершен (100% Definition of Done)
* **Связанные задачи**: TASK-B03, TASK-B08

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

* **Бизнес-цель**: Добавить возможность глубокого дублирования задач со всеми подзадачами, тегами и вложениями, а также обеспечить точную фильтрацию области действия автоматизаций по доске в фоновом слушателе событий.
* **Ключевые продуктовые правила и UI решения**:
  1. **TASK-B03 — Глубокое копирование задачи (Deep Duplicate)**:
     - Новый API: `POST /workspaces/{wid}/tasks/{task_id}/duplicate` (возвращает 201 Created с полной моделью `Task`).
     - Что копируется:
       - Название с префиксом `Копия — [Оригинальное название]`.
       - Описание (`description_markdown`), приоритет, проект, колонка, позиция (`original.position + 1`).
       - Теги (`task_tags`).
       - Подзадачи (все дочерние задачи с `parent_id = original.id` копируются с новыми UUID, сохраняя порядок и теги подзадач).
       - Вложения (`attachments`): shallow copy (новые записи в БД с теми же `s3_key`, файлы в S3 не дублируются).
     - Что НЕ копируется: `due_date`, статус выполнения (`is_completed = false`, `completed_at = null`), история изменений, порядок фокуса (`focus_order = null`).
     - Выполняется строго в рамках контекстной транзакции (`transactor.WithinTx`).
     - UI: в карточке задачи на канбан-доске (`KanbanCard.tsx`) добавлено контекстное меню «3 точки» с пунктом «Дублировать» (иконка `Copy`).
  2. **TASK-B08 — Скоуп автоматизаций по доске в EventListener & UI**:
     - В `listener.go` в методе `evaluateAndExecute` добавлена проверка области применения правила:
       ```go
       if rule.BoardID != nil && (taskInfo.ProjectID == nil || *rule.BoardID != *taskInfo.ProjectID) {
           return
       }
       ```
     - В `AutomationsHub.tsx`: заменены вкладки фильтрации на `'all' | 'board' | 'column'` с аккуратными лейблами «Все», «На доску», «На колонку» и счётчиками правил.
     - В `RuleBuilderModal.tsx` и `useRuleBuilder.ts`: добавлен выбор «Область применения» (Вся доска / Конкретная колонка / Глобально) с сохранением эффективных `boardId` и `columnId`.

---

## 🎯 2. Ключевые результаты и Затронутые подпакеты

1. **Контракты API (`api/openapi/`)**:
   - `paths/tasks.yaml`: добавлена операция `duplicateTask` (`POST /workspaces/{workspace_id}/tasks/{task_id}/duplicate`).
   - `openapi.yaml`: зарегистрирован путь `/workspaces/{workspace_id}/tasks/{task_id}/duplicate`.
2. **База данных и SQL (`backend/queries/` & `internal/infrastructure/db/sqlc/`)**:
   - `tasks.sql`: добавлен запрос `DuplicateTaskRow`.
   - sqlc: сгенерированы методы в `tasks.sql.go` и `querier.go`.
3. **Чистый домен и Application (`backend/internal/domain/task/` & `internal/application/`)**:
   - `domain/task/task.go`: доменный метод `Duplicate` с тестами инвариантов.
   - `application/task/ports.go`: методы `DuplicateTask` в `UseCases` и интерфейсы `Transactor`, `AttachmentRepository`.
   - `application/task/usecases_duplicate.go`: реализация транзакционного сценария глубокого дублирования.
   - `application/automation/listener.go`: фильтрация по `rule.BoardID` с тестом `TestEventListener_BoardScopeFilter`.
4. **Инфраструктура бэкенда (`backend/internal/infrastructure/`)**:
   - `container/services.go`: связывание сервиса задач с транзактором и репозиторием вложений.
   - `http/tasks_handler.go`: хендлер `DuplicateTask` (201 Created).
   - `http/tasks_duplicate_test.go`: table-driven тесты хендлера дублирования.
5. **Пользовательский интерфейс (`frontend/src/`)**:
   - `entities/task/`: метод `duplicateTask` в `api.ts` и хук `useDuplicateTaskMutation` в `queries.ts`.
   - `features/kanban-board/KanbanCard.tsx`: контекстное меню «3 точки» с вызовом дублирования и перемещения.
   - `features/automations/AutomationsHub.tsx`: фильтры «Все», «На доску», «На колонку».
   - `features/automations/RuleBuilderModal.tsx` & `useRuleBuilder.ts`: выбор «Область применения».

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

