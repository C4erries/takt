# Спринт 0.2.1: Авторизация, Персональный Workspace и Захват в Inbox (Quick Capture)

* **Даты проведения**: 2026-10-14 — 2026-10-14
* **Статус**: ✅ Завершен (100% Definition of Done)
* **Связанные глобальные ADR**: [ADR-0002](../decisions/0002-jwt-auth-and-tenant-isolation.md)

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

* **Бизнес-цель**: Сквозной Zero-friction онбординг (регистрация $\rightarrow$ автоматическое создание личного пространства «Личное» и наполнение обучающими задачами) и быстрый захват задач во «Входящие» (Inbox) с клавиатурной J/K навигацией в эстетике японской канцелярии («Soft Editorial GTD»).
* **Ключевые продуктовые правила и Edge Cases**:
  - **Гибридный JWT**: Сессионный токен выставляется сервером в `HttpOnly; SameSite=Lax; Path=/` Cookie для браузера (защита от XSS), с поддержкой `Authorization: Bearer <token>` для Pytest автотестов и API-клиентов.
  - **Транзакционный онбординг**: При регистрации пользователя в единой транзакции создаются:
    1. Запись `users` (пароль хешируется через `bcrypt`).
    2. Запись `workspaces` с типом `personal` («Личное»).
    3. Запись `workspace_members` с ролью `owner`.
    4. 3 стартовые обучающие задачи в Inbox («Нажми J и K для клавиатурного перемещения», «Нажми Пробел или кликни круглый чекбокс», «Нажми C в любом месте для Quick Capture»).
  - **Гибридный Quick Capture**: Всегда доступная инлайн-строка быстрого ввода поверх листа Washi + всплывающее модальное окно быстрого сброса по глобальному хоткею `C` (`Cmd/Ctrl+Enter` для сохранения).
  - **Пружинная J/K навигация**: Плавное перемещение фокусной строки с мягкой подсветкой `--accent-surface` (`cubic-bezier(0.22, 1, 0.36, 1)`, 150ms).
  - **Строгая Tenant Isolation (Защита от IDOR)**: Все запросы к задачам принудительно фильтруются по `workspace_id` текущей сессии. Попытка обратиться к чужой задаче возвращает `404 Not Found`.

---

## 🎯 2. Ключевые результаты и Затронутые подпакеты

1. **База данных и SQL (`backend/`)**:
   - Миграция `000002_auth_and_inbox.up.sql`: таблицы `users`, `workspaces`, `workspace_members`, `tasks` с внешними ключами и индексами.
   - Запросы для `sqlc`: `queries/users.sql`, `queries/workspaces.sql`, `queries/tasks.sql`.
2. **Контракты API (`api/openapi/`)**:
   - `paths/auth.yaml` (`/auth/register`, `/auth/login`, `/auth/logout`, `/auth/me`).
   - `paths/tasks.yaml` (`/tasks/inbox` GET/POST, `/tasks/{id}/toggle` PATCH, `/tasks/{id}` DELETE).
   - Схемы `schemas/user.yaml`, `schemas/workspace.yaml`, `schemas/task.yaml`. Сборка `make generate`.
3. **Backend Clean Architecture (Go 1.23+)**:
   - `internal/domain/workspace/` и `internal/domain/task/` (чистые сущности, инварианты, доменные ошибки).
   - `internal/application/workspace/` и `internal/application/task/` (сценарии использования, порты потребителя `ports.go`).
   - `internal/infrastructure/repository/` (`UserRepository`, `WorkspaceRepository`, `TaskRepository` для `pgxpool`).
   - `internal/infrastructure/http/`: `AuthMiddleware` (валидация JWT Cookie/Bearer), `AuthHandler`, `TasksHandler`.
   - `internal/infrastructure/container/`: Composition Root сборка роутов и зависимостей.
4. **Frontend («Soft Editorial GTD»)**:
   - 3-колоночный каркас «Surface-on-Canvas» (Sidebar 240px, Washi лист max 880px, правые Bento-блоки Inspector + MiniCalendar).
   - Лист бумаги Washi с разлиновкой, круговыми чекбоксами 18px и капсульными бейджами приоритета.
   - Инлайн Quick Capture + модалка по хоткею `C`.
   - Пружинная навигация `J/K` и шорткат закрытия задачи `Space/X`.
   - 4 состояния экрана: Loading (Skeleton), ErrorBanner с повтором, EmptyState, Success.
   - Модалка авторизации и регистрации (`AuthModal`).
5. **E2E Тестирование на Python (`tests/`)**:
   - Фикстуры `conftest.py`: регистрация пользователей, HttpOnly Cookie клиент, Bearer клиент, второй тенант.
   - `tests/e2e/test_auth.py`: полный цикл авторизации, онбординг стартовых задач, валидация DTO.
   - `tests/e2e/test_inbox.py`: создание задач, переключение завершения, удаление, изоляция данных IDOR.

---

## 📊 3. Метрики качества, Архитектурный контроль и Безопасность

* **Архитектурный линтер (`make check-arch`)**:
  - `PASS (0 ошибок)`: 5 из 5 проверок пройдены (лимиты строк `< 400`, файлов `<= 7`, чистота домена, отсутствие сырых цветов, строго 3 корневых подпакета в `internal/`).
* **Юнит-тесты Go (`make test -race`)**:
  - `18/18 PASS (100%)`: без гонок данных.
* **Сборка фронтенда (`npm run build`)**:
  - `0 ошибок, 0 предупреждений` (Vite production bundle собран за 3.38s).
* **Сквозные автотесты (`make test-e2e`)**:
  - `14/14 PASS (100%)` за 3.14s.
* **Аудит безопасности (Security Gate)**:
  - Пароли хешируются с солью (`bcrypt.GenerateFromPassword`).
  - JWT токен защищен флагами `HttpOnly; SameSite=Lax`.
  - Строгая изоляция тенантов: тест `test_tenant_isolation_idor_protection` гарантирует `404` при обращении к чужим задачам.

