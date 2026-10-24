# Спринт 0.9.4: «История изменений задачи (Activity Log)»

* **Даты проведения**: 2026-10-24
* **Статус**: ✅ Завершен (100% Definition of Done)
* **Связанные задачи**: TASK-B06

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

* **Бизнес-цель**: Прозрачность жизненного цикла задачи. Пользователь видит в шторке (`TaskDrawer.tsx`) полную хронологическую ленту изменений: кто и когда перенёс задачу в другую колонку, изменил статус/приоритет, назначил дедлайн, добавил тег или подзадачу.
* **Ключевые продуктовые правила и Edge Cases**:
  - `task_activities`: поле `user_id` может быть `NULL` для системных действий и автоматизаций. При отображении автор указывается как «Система» или «Автоматизация» с соответствующей иконкой `Bot`.
  - Типы событий (`activity_type`): `column_moved`, `status_changed`, `priority_changed`, `due_date_set`, `due_date_cleared`, `tag_added`, `tag_removed`, `title_changed`, `subtask_added`, `task_archived`, `task_unarchived`.
  - Ленивая загрузка: данные истории запрашиваются только при переключении пользователя на вкладку «История» в шторке задачи (`enabled: activeTab === 'history'`).
  - Относительное время («только что», «5 мин назад», «вчера в 15:30») с абсолютной всплывающей подсказкой даты.
* **Локальные технические решения**:
  - EventBus listener `ActivityLogger` подписывается на события `EventTaskMoved`, `EventTaskCompleted`, `EventTaskUpdated` и фиксирует записи активности асинхронно без блокировки основного потока.
  - Прямое логирование в Use Cases для специализированных действий (`AttachTagToTask`, `DetachTagFromTask`, `CreateSubtask`).
  - Контроль лимитов архитектурного линтера: рефакторинг хелперов маппинга в `helpers.go` и `handler.go` сохранил все файлы < 400 строк при лимите пакетов <= 7 файлов.

---

## 🎯 2. Ключевые результаты и Затронутые подпакеты

1. **Модель данных и Доменные подпакеты (`backend/internal/domain/*`)**:
   - Миграция `000015_task_activities.up.sql` / `down.sql`: таблица `task_activities`, индекс `idx_task_activities_task_id`.
   - Запросы sqlc в `backend/queries/activities.sql`: `CreateTaskActivity`, `ListTaskActivitiesByTaskID` с `LEFT JOIN users` для получения имени/email автора.
   - Чистый домен `backend/internal/domain/task/activity.go`: сущность `TaskActivity`, константы типов, валидация инвариантов и конструктор `NewTaskActivity`.
   - Unit-тесты домена в `activity_test.go`.

2. **Контракты и Сценарии использования (`backend/internal/application/*` & `api.yaml`)**:
   - OpenAPI: эндпоинт `GET /workspaces/{workspace_id}/tasks/{task_id}/activities` и схема `TaskActivity`.
   - Порты: `ActivityRepository` методы `CreateActivity` и `ListActivitiesByTaskID`.
   - `backend/internal/application/task/activity_logger.go`: `ActivityLogger` слушатель EventBus и Use Case `ListTaskActivities`.
   - Репозиторий `internal/infrastructure/repository/task_repository.go`.
   - HTTP-хендлер `internal/infrastructure/http/tasks_handler.go` (`ListTaskActivities`).
   - DI связывание в `internal/infrastructure/container/services.go`.

3. **Пользовательский интерфейс (`frontend/src/widgets/task-drawer/*`)**:
   - `TaskActivityTab.tsx`: вертикальная хронологическая лента с 4 состояниями (Skeleton, Empty, Error, Success), иконками Lucide и относительным временем.
   - `TaskDrawer.tsx`: переключатель вкладок «Детали» / «История».
   - `activityQueries.ts`: React Query хук `useTaskActivitiesQuery`.

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
  - Проверена изоляция тенанта (`workspace_id`), принадлежность задачи воркспейсу, отсутствие IDOR и утечек секретов.

