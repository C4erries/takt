# Спринт 0.9.7 — Automations Suite & Профиль пользователя

**Период**: 25.10.2026
**Статус**: ЗАВЕРШЕН (100% DoD)

---

## 🎯 Цели спринта
1. **TASK-B20**: Действие «Вернуть в предыдущую колонку» (`return_to_previous_column`) с трехуровневой стратегией fallback (предыдущая колонка из `task_activities` -> колонка Inbox текущей доски -> общее Inbox воркспейса).
2. **TASK-B21**: Расширенный триггер статуса выполнения задачи — запуск правил при отмене выполнения задачи (`task_uncompleted`).
3. **TASK-B01**: Набор автоматизаций подзадач (Subtasks Automation Suite) — триггеры `all_subtasks_completed`, `any_subtask_completed`, `subtask_created`, условия `subtasks_all_completed`, `subtasks_count` и действия `complete_all_subtasks`, `reopen_all_subtasks` с защитой от рекурсивных циклов.
4. **TASK-B09**: Профиль пользователя — миграция `000017_user_profile` (`avatar_url`), эндпоинты `PATCH /users/me` и `POST /users/me/avatar` (S3/MinIO), модальное окно профиля в UI и отображение аватара в сайдбаре.

---

## 📦 Выполненные задачи и поставка

### Бэкенд
- **Миграция `000017_user_profile`**: добавление колонки `avatar_url TEXT NULL` в таблицу `users`.
- **OpenAPI & Серверные контракты**:
  - `PATCH /users/me` (схема `UpdateUserProfileRequest` с `displayName`).
  - `POST /users/me/avatar` (multipart-загрузка изображения, валидация формата JPEG/PNG/WebP и размера < 5MB).
  - Поля `displayName` и `avatarUrl` в структуре `User`.
- **Доменный слой**:
  - Методы `UpdateProfile` и `SetAvatar` в сущности `User` (`domain/workspace`).
  - Константы триггеров и действий в `domain/automation` с валидацией инвариантов.
- **Инфраструктура & Evaluator**:
  - `evaluator.go` и `evaluator_helpers.go`: реализация действий `return_to_previous_column`, `complete_all_subtasks`, `reopen_all_subtasks`, условий подзадач.
  - `listener.go`: подписка на `TriggerTaskUncompleted`, `TriggerAllSubtasksCompleted`, `TriggerAnySubtaskCompleted`, `TriggerSubtaskCreated`.
  - `task_repository.go` и `project_repository.go`: методы `ListTaskActivities`, `ColumnExists`, `GetInboxColumn`.
  - `workspaces_handler.go`: хендлеры `UpdateUserProfile` и `UploadUserAvatar` с сохранением в S3.

### Фронтенд
- **Профиль пользователя**:
  - Компонент `UserProfileModal.tsx` с редактированием имени, валидацией аватара и мгновенным предпросмотром.
  - Сайдбар (`Sidebar.tsx`): кликабельная карточка пользователя с аватаром (или инициалами) и открытием модалки профиля.
- **Конструктор автоматизаций**:
  - `RuleBuilderSections.tsx` и `useRuleBuilder.ts`: поддержка триггеров выполнения и подзадач, селектор булевых условий `subtasks_all_completed` и числовых `subtasks_count`, действия `return_to_previous_column` (с плашкой объяснения fallback), `complete_all_subtasks`, `reopen_all_subtasks`.

---

## 🛡️ Верификация и метрики качества
- `make check-arch`: **PASSED** (0 ошибок: лимит строк < 500, папки <= 7 файлов, чистый домен, только семантические токены дизайна).
- `make test`: **100% PASS** (`-race`).
- `npm run lint && npm run build`: **0 ошибок, 0 ворнингов**.
- `docker compose up -d --build && uv run pytest tests/`: **39 passed** (100% E2E PASS).
