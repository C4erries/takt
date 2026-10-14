# Спринт 0.3.1b: Адаптация HTTP-слоя под ServerInterface из oapi-codegen и устранение рукописных DTO

* **Даты проведения**: 2026-10-14 — 2026-10-14
* **Статус**: ✅ Завершен (100% Definition of Done)
* **Связанные документы**: [docs/CURRENT_SPRINT.md](../CURRENT_SPRINT.md), [docs/conventions/backend.md](../conventions/backend.md)

---

## 🧠 1. Зафиксированные микро-решения спринта

1. **Реализация `generated.ServerInterface`**: Главный фасад `Handler` в `internal/infrastructure/http/handler.go` реализует `generated.ServerInterface` с обязательной проверкой компилятора:
   ```go
   var _ generated.ServerInterface = (*Handler)(nil)
   ```
2. **Ликвидация рукописных DTO**: Все структуры запросов/ответов (`CreateProjectHTTPRequest`, `ProjectResponse`, `CreateInboxTaskRequest`, `RegisterRequest` и др.) удалены. Все хендлеры используют строго типы `generated.*` из `internal/infrastructure/http/generated/api.gen.go`.
3. **Автоматическая регистрация маршрутов Contract-First**: Маршрутизация роутера Echo в `internal/infrastructure/container/router.go` привязана через `generated.RegisterHandlers(apiGroup, handler)`.
4. **Auth Middleware с пропуском открытых роутов**: Реализован `AuthMiddlewareWithSkipper` для безопасного пропуска `/auth/login`, `/auth/register`, `/auth/logout`, `/health` при глобальной защите API v1.

---

## 🎯 2. Выполненные задачи и Затронутые файлы

1. `backend/internal/infrastructure/http/handler.go`: проверка `var _ generated.ServerInterface = (*Handler)(nil)`.
2. `backend/internal/infrastructure/http/auth_handler.go`: удаление рукописных DTO, адаптация под сигнатуры `Login`, `Register`, `Logout`, `GetMe`.
3. `backend/internal/infrastructure/http/tasks_handler.go`: адаптация под сигнатуры `GetInboxTasks`, `CreateInboxTask`, `DeleteTask`, `ToggleTaskCompletion`.
4. `backend/internal/infrastructure/http/project_handler.go`: адаптация всех 13 методов управления проектами, досками, колонками и перемещением задач под `openapi_types.UUID` и `generated.*` модели.
5. `backend/internal/infrastructure/http/health.go`: реализация `GetHealth(ctx echo.Context)` с возвратом `generated.Health`.
6. `backend/internal/infrastructure/container/router.go`: подключение через `generated.RegisterHandlers(v1, h)`.
7. `backend/internal/infrastructure/http/*_test.go`: обновление всех unit-тестов.

---

## 📊 3. Метрики качества и Приемочный контроль

* **`make generate`**: Успешная сборка спецификации OpenAPI, генерация `api.gen.go` и `sqlc`.
* **`make check-arch`**: `PASS (0 ошибок)` по всем 5 проверкам.
* **`make test`**: `100% PASS` с флагом `-race` во всех пакетах.
* **`cd frontend && npm run build`**: `0 ошибок, 0 предупреждений`.

