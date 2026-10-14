# Текущий спринт / Оперативный центр управления (`CURRENT_SPRINT.md`)

> Этот файл — **единственная точка правды** о том, какие решения приняты по текущей задаче и что команда и AI-агенты делают прямо сейчас.
> Работа над каждым спринтом строго делится на **Фазу А (Штаб: Принятие решений)** и **Фазу Б (Конвейер: Реализация и Верификация)**.
> Завершенные спринты архивируются в каталог `docs/sprints/sprint-X.Y.Z.md` с записью в реестр `docs/sprints/README.md`.

## Текущий спринт: Спринт 0.2.2 (Этап 0.2: Стабилизация и подгонка под обновленные конвенции) ✅ ЗАВЕРШЕН (Готов к архивации)

---

## 🧠 1. Аналитика и Зафиксированные микро-решения спринта

* **Бизнес-цель**: Синхронизировать кодовую базу с обновленным сводом конвенций (`docs/conventions/backend.md`, `docs/conventions/frontend.md`, `docs/conventions/README.md`) перед переходом к Фазе 1.
* **Принятые микро-решения**:
  1. **Backend Tests**: Перевод всех тестов Use-Cases на паттерн **Table-Driven Tests** с библиотекой **`testify`** (`require.*` для прерывающих условий, `assert.*` для данных) и автогенерируемыми моками **`mockery v3`** (`mocks/`) вместо ad-hoc фейков.
  2. **Frontend State**: Установка и внедрение **`TanStack Query v5`** (с оптимистичными обновлениями для мутаций) и **`Zustand`** для легковесного локального UI-состояния (навигация, активный фокус).
  3. **Архитектурный линтер**: Гарантировать корректное игнорирование автогенерируемых моков `mock_*.go` в `scripts/check_arch.sh`.

---

## 📋 Таблица задач микроспринта 0.2.2

| # | Задача | Статус | Приоритет | Ответственный | Заметки и контракты |
|---|--------|--------|-----------|---------------|---------------------|
| 1 | Установка и интеграция TanStack Query v5 и Zustand | ✅ Done | Критический | `frontend_developer` | `@tanstack/react-query`, `zustand`, QueryClientProvider, useMutation |
| 2 | Рефакторинг Go unit-тестов на Table-Driven + Testify + Mockery | ✅ Done | Критический | `backend_developer` | `application/task/usecases_test.go`, `application/workspace/usecases_test.go` |
| 3 | Валидация качества, сборки и E2E тестов | ✅ Done | Критический | `lead_architect` / `qa_engineer` | `make check-arch && make test && make build && make test-e2e` |

---

## 🎯 Definition of Done (DoD)

- [x] Все Use-Cases тесты Go используют табличный паттерн, `testify/require`, `testify/assert` и `mockery` моки.
- [x] Во фронтенде установлены `@tanstack/react-query` и `zustand`; серверные запросы кэшируются через TanStack Query.
- [x] `make check-arch` — 0 ошибок (все 5 уровней пройдены).
- [x] `make test` — 100% PASS с флагом `-race`.
- [x] `cd frontend && npm run build` — 0 ошибок, 0 ворнингов.
- [x] `make test-e2e` — 100% PASS (14/14 тестов зеленые).

---

## Лог изменений спринта

| Дата | Что изменилось |
|------|---------------|
| 2026-10-14 | **Микроспринт 0.2.2**: инициализирован и завершен. Внедрены `@tanstack/react-query` v5 и `zustand`, Go тесты переведены на Table-Driven + Testify + Mockery v3. Конвейер верификации `make check-arch`, `make test`, `make build`, `make test-e2e` пройден на 100%. |
