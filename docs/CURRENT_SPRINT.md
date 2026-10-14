# Текущий спринт / Оперативный центр управления (`CURRENT_SPRINT.md`)

> Этот файл — **единственная точка правды** о том, какие решения приняты по текущей задаче и что команда и AI-агенты делают прямо сейчас.
> Работа над каждым спринтом строго делится на **Фазу А (Штаб: Принятие решений)** и **Фазу Б (Конвейер: Реализация и Верификация)**.
> Завершенные спринты архивируются в каталог `docs/sprints/sprint-X.Y.Z.md` с записью в реестр `docs/sprints/README.md`.

## Текущий спринт: Спринт 0.2.1 (Этап 0.2: Авторизация, Персональный Workspace и Захват в Inbox) 🔵 В РАБОТЕ (Фаза Б: Конвейер)

---

## 🧠 ФАЗА А: ШТАБ — Аналитика и Принятые решения (Роль: CEO, CTO & Product Architect)

### 1. Бизнес-обоснование (CEO View)
* **Проблема**: Фундамент и UI-Kit готовы, но сервис пока статичен — пользователь не может сохранить свои персональные задачи, войти в свою учетную запись и мгновенно сбросить мысли в Inbox.
* **Решение**: Реализовать сквозной цикл Zero-friction онбординга: регистрация пользователя $\rightarrow$ автоматическое создание личного пространства («Личное») и наполнение обучающими задачами $\rightarrow$ быстрый захват задач в Inbox (по хоткею `C` или инлайн-строке) с пружинной J/K навигацией.
* **Бизнес-эффект / Критерий успеха**: Пользователь регистрируется за 5 секунд, мгновенно попадает в наполненный тактильный интерфейс листа Washi, добавляет задачи и отмечает их выполнение с клавиатуры.

### 2. Принятые микро-решения спринта (Micro-Decisions)
* **Стратегия JWT**: Гибридный сессионный токен. Выставляется сервером в `HttpOnly; SameSite=Lax; Path=/` Cookie для Web SPA, а также поддерживается `Authorization: Bearer <token>` для автотестов и API-клиентов ([ADR-0002](decisions/0002-jwt-auth-and-tenant-isolation.md)).
* **Онбординг**: При регистрации автоматически и атомарно в транзакции создается `personal` Workspace («Личное») и 3 стартовые обучающие задачи в Inbox:
  1. *«Нажми J и K для клавиатурного перемещения по строкам»*
  2. *«Нажми Пробел или кликни круглый чекбокс, чтобы завершить задачу»*
  3. *«Нажми C в любом месте экрана для мгновенного сброса мысли (Quick Capture)»*
* **Формат Quick Capture**: Гибридный режим. Постоянная инлайн-строка ввода поверх разлинованного листа задач + быстрое модальное окно захвата по нажатию горячей клавиши `C`.
* **Архитектурный стандарт**: Строго 3 подпакета в `backend/internal/` (`domain`, `application`, `infrastructure`). Вся HTTP-доставка в `infrastructure/http/`, репозитории в `infrastructure/repository/`, сборка в `infrastructure/container/`.
* **Tenant Isolation**: Все запросы к задачам принудительно фильтруются по `workspace_id` текущей сессии с верификацией прав в базе данных.

### 3. Архитектура изменений и затрагиваемые директории
* **Контракты API**: `api/openapi/paths/auth.yaml`, `api/openapi/paths/tasks.yaml`, `api/openapi/schemas/user.yaml`, `api/openapi/schemas/task.yaml`.
* **БД & Миграции**: `backend/migrations/000002_auth_and_inbox.up.sql`, `backend/queries/*.sql`.
* **Backend**: `internal/domain/workspace/`, `internal/domain/task/`, `internal/application/workspace/`, `internal/application/task/`, `internal/infrastructure/repository/`, `internal/infrastructure/http/`.
* **Frontend**: `frontend/src/features/quick-capture/`, `frontend/src/features/task-navigation/`, `frontend/src/widgets/sidebar/`, `frontend/src/widgets/main-sheet/`.
* **Tests**: `tests/e2e/test_auth.py`, `tests/e2e/test_inbox.py`.

---

## 📋 ФАЗА Б: КОНВЕЙЕР — Таблица задач спринта

| # | Задача | Статус | Приоритет | Ответственный / Субагент | Заметки и контракты |
|---|--------|--------|-----------|---------------------------|---------------------|
| 1 | Миграция БД (users, workspaces, tasks) и sqlc-запросы | ✅ Done | Критический | `backend_developer` | `000002_auth_and_inbox.up.sql`, `queries/*.sql` |
| 2 | OpenAPI спецификация Auth & Inbox Tasks контрактов | ✅ Done | Критический | `backend_developer` | `paths/auth.yaml`, `paths/tasks.yaml`, `make generate` |
| 3 | Domain & Application сценарии (User, Workspace, Task, JWT) | ✅ Done | Критический | `backend_developer` | `internal/domain/*`, `internal/application/*`, `ports.go` |
| 4 | HTTP-хендлеры, Auth Middleware и Wiring в container | ✅ Done | Критический | `backend_developer` | `infrastructure/http/*`, `infrastructure/container/*` |
| 5 | Фронтенд 3-колоночный каркас, лист Washi и Quick Capture | 🔵 In Progress | Критический | `frontend_developer` | `widgets/sidebar`, `widgets/main-sheet`, `features/quick-capture` |
| 6 | Пружинная J/K навигация и закрытие задач по шорткату | 🔵 In Progress | Высокий | `frontend_developer` | `features/task-navigation`, хоткеи J/K/Space |
| 7 | Pytest E2E тесты (Auth, Onboarding, Inbox CRUD, Tenant Isolation) | 🔵 In Progress | Критический | `qa_engineer` | `tests/e2e/test_auth.py`, `tests/e2e/test_inbox.py` |

> **Статусы задач**: `🔲 Planned` (Запланировано) · `🔵 In Progress` (В работе) · `🛑 Blocked` (Блокер) · `✅ Done` (Завершено)

---

## 🎯 Definition of Done (DoD) Спринта

- [ ] **Архитектурный линтер (`make check-arch`) проходит с 0 ошибок**: файлы `< 400 строк`, в пакетах `<= 7 файлов`, чистые слои DDD, 3 корневых подпакета в `internal/`, нет сырых цветов Tailwind.
- [ ] Пользователь может зарегистрироваться и войти по email/паролю, сессионный JWT выставляется в HTTP-only cookie.
- [ ] При регистрации автоматически создается Personal Workspace и 3 стартовые обучающие задачи в Inbox.
- [ ] Задачи создаются в Inbox через инлайн-поле и модальное окно по хоткею `C`.
- [ ] Работает J/K навигация по задачам и завершение задачи круглым чекбоксом.
- [ ] Полная мультитенантная изоляция: пользователь не может прочитать или изменить чужие задачи.
- [ ] `make test` и `make test-e2e` проходят со 100% результатом (PASS).

---

## Лог изменений спринта

| Дата | Что изменилось |
|------|---------------|
| 2026-10-14 | **Фаза А (Штаб)**: согласован Блиц-опрос сооснователей, принят [ADR-0002](decisions/0002-jwt-auth-and-tenant-isolation.md), сформирован скоуп Спринта 0.2.1 (Этап 0.2). |
| 2026-10-14 | **Фаза Б (Конвейер)**: запущен субагент `backend_developer` на Задачи 1–4 (Contract-First: миграции, OpenAPI, domain/app/infra). |
