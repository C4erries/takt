# Спринт 0.9.9 — Activity Audit, Event Actors & Automations Reliability

**Период**: 25.10.2026  
**Статус**: ЗАВЕРШЕН (100% DoD)  
**Веха**: Финализация MVP v1.0.0

---

## 🎯 Цели спринта

1. **TASK-B29**: Проброс `ActorID` (ID пользователя) через доменные события (`TaskMovedEvent`, `TaskCreatedEvent`, `TaskUpdatedEvent`, `TaskCompletedEvent`) и запись автора в аудит-лог активности (`task_activities.user_id`).
2. **TASK-B30**: Устранение критического дефекта в `Event Listener` автоматизаций (баг сравнения `rule.BoardID` с `taskInfo.ProjectID`), проверка принадлежности колонки к доске, восстановление действия «Возврат в предыдущую колонку» и триггера «Отметка выполнения снята».
3. **TASK-B31**: Читаемое форматирование идентификатора автоматизации (`⚡ Название правила (#A-01)`) в истории активности задач.
4. **TASK-B32**: Изоляция модального окна профиля пользователя через React Portal `createPortal(..., document.body)` для устранения сжатия модалки в рамках CSS-контекста `translate-x-0` сайдбара.
5. **TASK-B33**: Выделение отдельной категории «Глобальные» в Хабе автоматизаций (`AutomationsHub.tsx`).
6. **TASK-B34**: Корректное отображение реального имени пользователя в таймлайне активности вместо безликого «Система», а также фиксация событий снятия выполнения («Снята отметка о выполнении»).
7. **HOTFIX Avatar**: Реализация HTTP-эндпоинта `GET /api/v1/users/me/avatar/*` со стримингом бинарного файла аватара из S3/MinIO.

---

## 📦 Выполненные задачи и поставка

### Бэкенд
- **Доменные события (`internal/domain/events/event.go`)**:
  - В события добавлено поле `ActorID *uuid.UUID`.
- **Контекст аутентификации и HTTP-хендлеры**:
  - Echo middleware сохраняет `userID` в `context.Context` по строго типизированному и строковому ключам.
  - HTTP-хендлеры задач, проектов и воркспейсов передают `ActorID` в Use Cases и события.
- **Аудит активности (`internal/application/task/activity_logger.go`)**:
  - Использует `e.ActorID` из событий для сохранения реального `user_id`.
  - Добавлена фиксация снятия выполнения задачи (`is_completed: false`) как `status_changed`.
- **Движок автоматизаций (`internal/application/automation/`)**:
  - В `listener.go` исправлена проверка скоупа: для правил с `BoardID` проверяется принадлежность колонки к доске через `ListColumnIDsByBoard`.
  - В `HandleTaskUpdated` добавлен триггер `TriggerTaskUncompleted`.
  - В `evaluator.go` действие `executeReturnToPreviousColumn` находит предыдущую колонку в аудит-логе с цепочкой fallback (предыдущая колонка $\rightarrow$ inbox доски $\rightarrow$ первая todo $\rightarrow$ inbox воркспейса).
  - Тесты разбиты на `listener_test.go` и `listener_subtasks_test.go` (< 400 строк).
- **Стриминг аватарок пользователя**:
  - `AvatarStorage` дополнен методом `GetObject(ctx, key) (io.ReadCloser, string, error)`.
  - `workspaces_handler.go` дополнен методом `GetUserAvatar` со стримингом контента из MinIO/S3.
  - В роутере зарегистрирован маршрут `GET/HEAD /api/v1/users/me/avatar/*` со снятием обязательной JWT-аутентификации для прямых ссылок `<img>`.

### Фронтенд
- **`Modal.tsx`**: рендеринг через React Portal (`createPortal(..., document.body)`) с SSR-защитой.
- **`AutomationsHub.tsx`**: добавлены фильтр-вкладка «Глобальные» и фильтрация правил.
- **`TaskActivityTab.tsx`**: отображение реального имени пользователя (`Pavel • только что`), бейджа `⚡ Правило (#A-01)`, понятного текста «Снята отметка о выполнении».
- **Темы оформления (`themes.ts`)**: удалены культурные привязки, внедрены нейтральные имена: `Winter Snow`, `Autumn Fall`, `Celestial Atlas`, `Frost Pine`, `Deep Ocean`.

---

## 🛡️ Верификация и метрики качества
- `make check-arch`: **PASSED** (0 ошибок).
- `make test`: **100% PASS** (`-race`).
- `npm run lint && npm run build`: **0 ошибок, 0 ворнингов**.
- `docker compose up -d --build && uv run pytest tests/`: **39 passed** (100% E2E PASS).
- `GET /api/v1/users/me/avatar/*`: **HTTP 200 OK, image/png**.
