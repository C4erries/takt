# Текущий спринт / Оперативный центр управления (`CURRENT_SPRINT.md`)

> Этот файл — **единственная точка правды** о том, какие решения приняты по текущей задаче и что команда и AI-агенты делают прямо сейчас.
> Работа над каждым спринтом строго делится на **Фазу А (Штаб: Принятие решений)** и **Фазу Б (Конвейер: Реализация и Верификация)**.
> Завершенные спринты архивируются в каталог `docs/sprints/sprint-X.Y.Z.md` с записью в реестр `docs/sprints/README.md`.

## Текущий спринт: Спринт 0.3.1b (Технический: адаптация HTTP-сервера под ServerInterface из oapi-codegen) ✅ ЗАВЕРШЕН (Готов к архивации)

---

## 🧠 1. Аналитика и Зафиксированные микро-решения спринта (Режим «Штаб»)

* **Бизнес-цель**: Привести HTTP-слой бэкенда (`internal/infrastructure/http/`) в 100% соответствие с обновленной конвенцией Contract-First из коммита `2497366` (`docs/conventions/backend.md`).
* **Зафиксированные архитектурные микро-решения**:
  1. **Имплементация `generated.ServerInterface`**:
     - Главный HTTP-хендлер/фасад (`Handler` в `internal/infrastructure/http/`) реализует интерфейс `generated.ServerInterface` из `internal/infrastructure/http/generated/api.gen.go`.
     - На уровне компилятора объявлена строгая статическая проверка:
       ```go
       var _ generated.ServerInterface = (*Handler)(nil)
       ```
  2. **Табу на рукописные DTO**:
     - Удалены все рукописные структуры запросов и ответов (`ProjectResponse`, `CreateProjectHTTPRequest`, `TaskResponse`, `RegisterRequest` и т.д.) из файлов хендлеров `auth_handler.go`, `tasks_handler.go`, `project_handler.go`.
     - Все методы принимают и возвращают сгенерированные типы из пакета `generated.*`.
  3. **Стандартная регистрация роутов через oapi-codegen**:
     - В `internal/infrastructure/container/router.go` ручная регистрация эндпоинтов заменена на `generated.RegisterHandlers(apiGroup, handler)`.
     - JWT Auth Middleware монтируется через `AuthMiddlewareWithSkipper` для безопасного пропуска открытых роутов (`/auth/*`, `/health`).
  4. **Сохранение чистоты слоев**:
     - Пакет `generated` изолирован в `internal/infrastructure/http/generated/` и не протекает в `domain/` или `application/`.

---

## 📋 Таблица задач спринта 0.3.1b

| # | Задача | Статус | Приоритет | Ответственный | Заметки и контракты |
|---|--------|--------|-----------|---------------|---------------------|
| 1 | Рефакторинг `internal/infrastructure/http/` под реализацию `generated.ServerInterface` | ✅ Done | Критический | `backend_developer` | Замена рукописных DTO на `generated.*`, имплементация всех методов интерфейса |
| 2 | Перевод регистрации роутов в `container/router.go` на `generated.RegisterHandlers` | ✅ Done | Высокий | `backend_developer` | Подключение Auth Middleware и Echo группы |
| 3 | Обновление HTTP unit-тестов под вызовы `ServerInterface` | ✅ Done | Высокий | `backend_developer` | `auth_test.go`, `tasks_test.go`, `project_test.go` |
| 4 | Приемочный контроль и верификация конвейера | ✅ Done | Критический | `lead_architect` | `make generate && make check-arch && make test && npm run build` (0 ошибок) |

---

## 🎯 Definition of Done (DoD)

- [x] В `backend/internal/infrastructure/http/` реализован интерфейс `generated.ServerInterface` со статической проверкой компилятора `var _ generated.ServerInterface = (*Handler)(nil)`.
- [x] Полностью отсутствуют рукописные DTO запросов/ответов в хендлерах; используются исключительно структуры из `internal/infrastructure/http/generated`.
- [x] Роутинг в `internal/infrastructure/container/router.go` привязан через `generated.RegisterHandlers`.
- [x] `make generate` успешно пересобирает спецификацию OpenAPI и код Echo сервера.
- [x] `make check-arch` — 0 ошибок по всем 5 проверкам.
- [x] `make test` — 100% PASS с флагом `-race`.
- [x] `cd frontend && npm run build` — 0 ошибок сборки.

---

## Лог изменений спринта

| Дата | Что изменилось |
|------|---------------|
| 2026-10-14 | **Спринт 0.3.1b завершен**: HTTP-сервер успешно адаптирован под `generated.ServerInterface`, все рукописные DTO устранены, роутер привязан через `generated.RegisterHandlers`. Полный конвейер верификации пройден со 100% успехом. |
