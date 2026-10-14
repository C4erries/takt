# Текущий спринт / Оперативный центр управления (`CURRENT_SPRINT.md`)

> Этот файл — **единственная точка правды** о том, какие решения приняты по текущей задаче и что команда и AI-агенты делают прямо сейчас.
> Работа над каждым спринтом строго делится на **Фазу А (Штаб: Принятие решений)** и **Фазу Б (Конвейер: Реализация и Верификация)**.
> Завершенные спринты архивируются в каталог `docs/sprints/sprint-X.Y.Z.md` с записью в реестр `docs/sprints/README.md`.

## Текущий спринт: Спринт 0.3.1b (Технический: адаптация HTTP-сервера под ServerInterface из oapi-codegen) 🔵 В РАБОТЕ

---

## 🧠 1. Аналитика и Зафиксированные микро-решения спринта (Режим «Штаб»)

* **Бизнес-цель**: Привести HTTP-слой бэкенда (`internal/infrastructure/http/`) в 100% соответствие с обновленной конвенцией Contract-First из коммита `2497366` (`docs/conventions/backend.md`).
* **Зафиксированные архитектурные микро-решения**:
  1. **Имплементация `generated.ServerInterface`**:
     - Главный HTTP-хендлер/фасад (`Handler` или `Server` в `internal/infrastructure/http/`) обязан реализовывать интерфейс `generated.ServerInterface` из `internal/infrastructure/http/generated/api.gen.go`.
     - На уровне компилятора объявляется строгая статическая проверка:
       ```go
       var _ generated.ServerInterface = (*Handler)(nil)
       ```
  2. **Табу на рукописные DTO**:
     - Удалить все рукописные структуры запросов и ответов (`ProjectResponse`, `CreateProjectHTTPRequest`, `TaskResponse`, `RegisterRequest` и т.д.) из файлов хендлеров `auth_handler.go`, `tasks_handler.go`, `project_handler.go`.
     - Все методы обязаны принимать и возвращать сгенерированные типы из пакета `generated.*` (например, `generated.CreateProjectRequest`, `generated.Project`, `generated.Task`).
  3. **Стандартная регистрация роутов через oapi-codegen**:
     - В `internal/infrastructure/container/router.go` ручная регистрация эндпоинтов (`api.POST("/workspaces/:id/projects", ...)`) заменяется на `generated.RegisterHandlers(apiGroup, handler)`.
     - JWT Auth Middleware монтируется на защищенную группу роутов `/api/v1` или на уровне Echo middleware.
  4. **Сохранение чистоты слоев**:
     - Пакет `generated` изолирован в `internal/infrastructure/http/generated/` и не протекает в `domain/` или `application/`.
     - Преобразования между DTO `generated.*` и DTO application-слоя происходят исключительно внутри `internal/infrastructure/http/`.

---

## 📋 Таблица задач спринта 0.3.1b

| # | Задача | Статус | Приоритет | Ответственный | Заметки и контракты |
|---|--------|--------|-----------|---------------|---------------------|
| 1 | Рефакторинг `internal/infrastructure/http/` под реализацию `generated.ServerInterface` | 🔵 In Progress | Критический | `backend_developer` | Замена рукописных DTO на `generated.*`, имплементация всех методов интерфейса |
| 2 | Перевод регистрации роутов в `container/router.go` на `generated.RegisterHandlers` | 🔲 Planned | Высокий | `backend_developer` | Подключение Auth Middleware и Echo группы |
| 3 | Обновление HTTP unit-тестов под вызовы `ServerInterface` | 🔲 Planned | Высокий | `backend_developer` | `auth_test.go`, `tasks_test.go`, `project_test.go` |
| 4 | Приемочный контроль и верификация конвейера | 🔲 Planned | Критический | `lead_architect` | `make generate && make check-arch && make test && npm run build` |

---

## 🎯 Definition of Done (DoD)

- [ ] В `backend/internal/infrastructure/http/` реализован интерфейс `generated.ServerInterface` со статической проверкой компилятора `var _ generated.ServerInterface = (*Handler)(nil)`.
- [ ] Полностью отсутствуют рукописные DTO запросов/ответов в хендлерах; используются исключительно структуры из `internal/infrastructure/http/generated`.
- [ ] Роутинг в `internal/infrastructure/container/router.go` привязан через `generated.RegisterHandlers`.
- [ ] `make generate` успешно пересобирает спецификацию OpenAPI и код Echo сервера.
- [ ] `make check-arch` — 0 ошибок по всем 5 проверкам.
- [ ] `make test` — 100% PASS с флагом `-race`.
- [ ] `cd frontend && npm run build` — 0 ошибок сборки.

---

## Лог изменений спринта

| Дата | Что изменилось |
|------|---------------|
| 2026-10-14 | **Спринт 0.3.1b открыт**: Технический спринт по адаптации HTTP-сервера под `generated.ServerInterface` и исключению рукописных DTO в соответствии с коммитом `2497366` (`docs/conventions/backend.md`). |
