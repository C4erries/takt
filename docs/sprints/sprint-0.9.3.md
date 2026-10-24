# Спринт 0.9.3: «Архив задач (Batch SQL ADR-0008)»

* **Даты проведения**: 2026-10-24
* **Статус**: ✅ Завершен (100% Definition of Done)
* **Связанные глобальные ADR**: [ADR-0008: Batch SQL Fast Path for Sweepers](../decisions/0008-batch-sql-fast-path-for-sweepers.md)

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

* **Бизнес-цель**: Реализация функционала архивации завершенных задач для сохранения чистоты канбан-доски с возможностью просмотра и разархивации задач обратно на доску.
* **Ключевые продуктовые правила и Edge Cases**:
  - Архивация не удаляет задачи физически, а выставляет `is_archived = true`, `archived_at = NOW()` и сохраняет `column_id` (исходная колонка).
  - При разархивации пользователь выбирает целевую доску и колонку (по умолчанию исходная колонка, если доступна). Задача перемещается в выбранную колонку в конец списка (`position = max + 1000`).
  - Активные выборки задач (`ListTasksByProject`, `ListTasksByColumn`, `ListTodayTasks`, `GetInboxTasks`, `CountActiveTasksByBoard`) исключают архивированные задачи (`is_archived = false`).
  - Быстрое действие в заголовке колонки: «Сгрузить в архив» с confirmation popover, отображающим количество выполненных задач для архивации.
* **Локальные технические решения**:
  - **ADR-0008 Batch SQL Fast Path**: Для периодических правил автоматизации с действием `ActionArchiveTask` и условием `is_completed == true` воркер запускает единый batch SQL-запрос `ArchiveCompletedTasksBatchSQL` минуя N+1 обход, после чего генерирует `BatchSweepCompletedEvent`.
  - Отдельный экран `ArchiveView.tsx` с фильтрацией по проекту/доске, поиском и постраничной загрузкой.
  - Счетчик архивированных задач в боковом меню (`Sidebar.tsx`) с бейджем.

---

## 🎯 2. Ключевые результаты и Затронутые подпакеты

1. **Модель данных и Доменные подпакеты (`backend/internal/domain/*`)**:
   - Миграция `000014_archive.up.sql` / `down.sql`: поля `is_archived`, `archived_at`, индекс `idx_tasks_archived`.
   - Обновлены sql-запросы `tasks.sql` и `projects.sql` с фильтрацией `AND is_archived = false`.
   - Запросы для архива: `ListArchivedTasks`, `CountArchivedTasks`, `ArchiveCompletedTasksInColumn`, `UnarchiveTask`, `ArchiveCompletedTasksBatchSQL`.
   - Доменная сущность `Task`: методы `Archive()` и `Unarchive()`.
   - Событие `BatchSweepCompletedEvent` для уведомления о завершении пакетной архивации.

2. **Контракты и Сценарии использования (`backend/internal/application/*` & `api.yaml`)**:
   - OpenAPI: `GET /workspaces/{workspace_id}/tasks/archived`, `POST /workspaces/{workspace_id}/columns/{column_id}/archive-completed`, `PATCH /workspaces/{workspace_id}/tasks/{task_id}/unarchive`.
   - Порт `BatchTaskExecutor` в `application/automation/ports.go`.
   - Fast Path `CanUseBatchPath` в `application/automation/worker.go`.
   - Use Cases в `application/task/usecases_archive.go`: `ListArchivedTasks`, `ArchiveCompletedTasksInColumn`, `UnarchiveTask`.
   - Реализация репозиториев в `infrastructure/repository/task_repository.go`.
   - HTTP-хендлеры в `infrastructure/http/tasks_handler.go` и `project_handler.go`.
   - Связывание зависимостей в `infrastructure/container/services.go`.

3. **Пользовательский интерфейс (`frontend/src/features/*` & `shared/components/ui/*`)**:
   - Фича `features/archive/`: `ArchiveView.tsx`, `RestoreTaskPopover.tsx`, `api.ts`.
   - Пункт в `Sidebar.tsx` («Архив») со счетчиком архивированных задач воркспейса.
   - Меню колонки канбана `KanbanColumn.tsx` с действием «Сгрузить в архив» и всплывающим подтверждением.
   - Интеграция роутинга в `MainSheet.tsx`.

---

## 📊 3. Метрики качества, Архитектурный контроль и Безопасность

* **Архитектурный линтер (`make check-arch`)**:
  - `PASS (0 ошибок)`: лимит строк (< 400), лимит файлов (<= 7 в подпакетах), чистота Domain, структура `backend/internal` и отсутствие сырых цветов подтверждены.
* **Юнит-тесты бэкенда (`make test -race`)**:
  - `100% PASS` без гонок данных.
* **Сборка и линтер фронтенда (`npm run lint && npm run build`)**:
  - `0 ошибок, 0 предупреждений`, Vite сборка успешна.
* **Сквозные автотесты (`uv run pytest tests/`)**:
  - `39 из 39 тестов пройдены успешно (100% PASS)`.
* **Аудит безопасности (Security Gate)**:
  - Проверена изоляция тенанта (`workspace_id`), валидация ID и проверка прав доступа во всех эндпоинтах.

