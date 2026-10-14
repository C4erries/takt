# Спринт 0.2.2: Стабилизация и подгонка под обновленные конвенции (TanStack Query, Zustand, Testify, Mockery)

* **Даты проведения**: 2026-10-14 — 2026-10-14
* **Статус**: ✅ Завершен (100% Definition of Done)
* **Связанные документы**: [docs/conventions/backend.md](../conventions/backend.md), [docs/conventions/frontend.md](../conventions/frontend.md), [docs/conventions/README.md](../conventions/README.md)

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

* **Бизнес-цель**: Синхронизировать кодовую базу с обновленным сводом конвенций перед переходом к Фазе 2 (Канбан, проекты и автоматизации).
* **Принятые решения**:
  1. **Frontend State Management**:
     - Серверные данные (`tasks`, `inbox`) управляются через **`@tanstack/react-query v5`** с оптимистичными мутациями (`onMutate` -> откат в `onError` -> инвалидация в `onSettled`).
     - Клиентский UI-стейт изолирован в легковесном **`Zustand`** сторе `useTaskStore` (активная задача, модальное окно Quick Capture, навигация).
  2. **Go Test Standards**:
     - Исключены рукописные фейки в пользу автогенерированных моков **`mockery v3`** (`mocks/`).
     - Все тесты Use-Cases переведены на паттерн **Table-Driven Tests** с параллельным выполнением `t.Parallel()`.
     - Применена библиотека ассертов **`testify`** (`require.*` для прерывающих проверок, `assert.*` для данных).

---

## 🎯 2. Выполненные задачи и Затронутые подпакеты

1. **Frontend (`frontend/`)**:
   - Установлены зависимости `@tanstack/react-query` и `zustand`.
   - `features/task-navigation/taskStore.ts`: Zustand стор UI-состояния.
   - `features/task-navigation/useTasksQuery.ts`: кастомные хуки TanStack Query (`useInboxTasks`, `useCreateTaskMutation`, `useToggleTaskMutation`, `useDeleteTaskMutation`).
   - `app/main.tsx` и `app/App.tsx`: корневой `QueryClientProvider`, привязка UI к стору и квери.
2. **Backend Unit-тесты (`backend/internal/application/`)**:
   - `application/task/usecases_test.go`: переписан на Table-Driven с использованием `mocks.MockTaskRepository`, `testify/require` и `testify/assert`.
   - `application/workspace/usecases_test.go`: переписан на Table-Driven с моками `MockUserRepository`, `MockWorkspaceRepository`, `MockTransactor`, `MockPasswordHasher`, `MockTokenGenerator`, `MockInitialTasksCreator`.

---

## 📊 3. Метрики качества и Приемочный контроль

* **Архитектурный линтер (`make check-arch`)**:
  - `PASS (0 ошибок)`: все 5 проверок пройдены, лимиты строк и подпакетов соблюдены.
* **Юнит-тесты Go (`make test -race`)**:
  - `100% PASS` без гонок данных.
* **Сборка фронтенда (`cd frontend && npm run build`)**:
  - `0 ошибок, 0 предупреждений` (Vite production bundle собран за 3.98s).
* **Сквозные автотесты (`make test-e2e`)**:
  - `14/14 PASS (100%)`.

