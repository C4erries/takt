# Спринт 0.4.1: Единые Воркспейсы (Unified Workspaces), Роли Owner/Member и Инвайты

* **Даты проведения**: 2026-10-15 — 2026-10-15
* **Статус**: ✅ Завершен (100% Definition of Done)
* **Связанные документы**: [docs/CURRENT_SPRINT.md](../CURRENT_SPRINT.md), [docs/conventions/backend.md](../conventions/backend.md), [docs/conventions/frontend.md](../conventions/frontend.md)

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

1. **Концепция «Единых Воркспейсов» (Unified Workspaces)**:
   - Воркспейсы универсальны: отсутствует жесткое разделение на `personal` и `shared`. Любой воркспейс может использоваться как индивидуально, так и совместно с командой.
   - Пользователь может иметь несколько воркспейсов («Личное», «Работа», «Сайд-проект» и др.) и переключаться между ними в интерфейсе.
   - При создании любого воркспейса для него автоматически генерируются 3 пресетных правила автоматизаций и стартовые обучающие задачи.
2. **Токен-инвайты (Token Invite Links)**:
   - Таблица `workspace_invites`: `workspace_id`, `invited_by_id`, уникальный 32-байтный hex-токен со сроком действия 7 дней.
   - `POST /api/v1/workspaces/{workspace_id}/invites`: Владелец генерирует или получает активную ссылку приглашения.
   - `GET /api/v1/invites/{token}`: публичный предпросмотр без авторизации (название воркспейса, имя пригласившего).
   - `POST /api/v1/invites/{token}/accept`: принятие инвайта авторизованным пользователем (добавление в `workspace_members` с ролью `member`).
3. **Ролевая модель (RBAC: Owner vs Member)**:
   - **`Owner`**: Полный контроль (переименование воркспейса, удаление, генерация инвайтов, исключение участников, управление автоматизациями).
   - **`Member`**: Полноценная работа с контентом (задачи, проекты, доски, колонки, теги), но запрет на деструктивные действия над воркспейсом и чужими участниками.
   - При удалении/выходе участника созданные им задачи и проекты **остаются в воркспейсе** (не удаляются каскадно), так как принадлежат `workspace_id`.
   - `Member` может в любой момент добровольно покинуть воркспейс (`DELETE /members/{user_id}`).
4. **Переключение активного воркспейса**:
   - Эндпоинт `POST /api/v1/workspaces/{workspace_id}/switch`: проверяет членство пользователя в целевом воркспейсе, генерирует новый валидный JWT токен с `WorkspaceID` и выставляет его в Cookie/Header.
   - В интерфейсе сайдбара интерактивный селектор `WorkspaceSwitcher` в эстетике Washi со списком доступных пространств, чекмарком у активного, кнопкой «+ Новый воркспейс» и «Участники и доступ».

---

## 🎯 2. Выполненные задачи и Затронутые подпакеты

1. **Миграции БД и SQL**:
   - `backend/migrations/000006_workspace_invites.up.sql`: таблица `workspace_invites`, снятие жесткого ограничения `workspaces_type_check`, индексы.
   - `backend/queries/workspaces.sql`: запросы `UpdateWorkspaceName`, `ListWorkspaceMembers`, `DeleteWorkspaceMember`, `CreateWorkspaceInvite`, `GetWorkspaceInviteByToken`, `GetActiveWorkspaceInvite`, `DeleteWorkspaceInvite`.
2. **OpenAPI Контракты и Кодогенерация**:
   - `api/openapi/schemas/workspace.yaml`: схемы `Workspace`, `WorkspaceMember`, `WorkspaceInvite`, `InvitePreview`, `WorkspaceSwitchResponse`.
   - `api/openapi/paths/workspaces.yaml`: эндпоинты `/workspaces`, `/workspaces/{id}`, `/switch`, `/members`, `/invites`, `/invites/{token}`, `/accept`.
   - Сгенерирован обновленный `ServerInterface` в `internal/infrastructure/http/generated/api.gen.go`.
3. **Чистый доменный слой (`internal/domain/workspace`)**:
   - `entity.go`: сущность `WorkspaceInvite` с методом `IsExpired()`, конструктор `NewWorkspace(ownerID, name)` с валидацией инвариантов.
   - `errors.go`: доменные ошибки `ErrInviteNotFound`, `ErrInviteExpired`, `ErrAlreadyMember`, `ErrCannotRemoveOwner`, `ErrInvalidInviteToken`, `ErrInvalidWorkspaceName`.
   - Табличные юнит-тесты инвариантов в `entity_test.go`.
4. **Use Cases и Сервисы (`internal/application/workspace`)**:
   - `workspace_usecases.go`: 9 сценариев (создание, переименование, переключение JWT, список участников, исключение, создание инвайта, предпросмотр, принятие).
   - Табличные тесты `workspace_usecases_test.go` и `invites_usecases_test.go` с моками mockery v3.
5. **Инфраструктура (`internal/infrastructure/`)**:
   - `repository/workspace_repository.go`: методы работы с участниками, инвайтами и обновлением имени воркспейса.
   - `http/middleware.go`: объединение middleware аутентификации и CORS (строго 7 рабочих файлов в пакете).
   - `http/workspaces_handler.go`: хендлеры управления воркспейсами, участниками и инвайтами с проверками безопасности и RBAC.
   - `container/router.go`: публичный роутинг для `GET /api/v1/invites/:token`.
6. **Фронтенд (`frontend/src/`)**:
   - `entities/workspace/`: типы `Workspace`, `WorkspaceMember`, `WorkspaceInvite`, `InvitePreview`, API-клиент и TanStack Query хуки.
   - `features/workspace-switcher/`: интерактивный выпадающий дропдаун `WorkspaceSwitcher` в шапке сайдбара и модальное окно создания `CreateWorkspaceModal`.
   - `features/workspace-members/`: модальное окно участников `WorkspaceMembersModal` (с копированием ссылки приглашения, удалением и выходом) и компонент `AcceptInviteModal` для автоматической обработки инвайт-ссылок.
7. **E2E-тестирование (`tests/e2e/test_workspaces.py`)**:
   - 6 новых сквозных E2E-тестов на Pytest: создание второго воркспейса, автосидирование автоматизаций, переключение `/switch`, переименование, полный цикл инвайта и принятия, проверка RBAC (запрет действий для member, запрет удаления owner).
   - Итог: **31/31 PASS (100%)**.

---

## 📊 3. Метрики качества и Приемочный контроль

* **`make check-arch`**: `PASS (0 ошибок)` по всем 5 проверкам архитектурного линтера.
* **`make test`**: `100% PASS` с флагом `-race` во всех пакетах бэкенда.
* **`make test-e2e`**: `31/31 PASS (100%)` в Docker-окружении (6 новых тестов воркспейсов и инвайтов).
* **`cd frontend && npm run build`**: `0 ошибок, 0 предупреждений` (Vite build собран за 4.31s).

