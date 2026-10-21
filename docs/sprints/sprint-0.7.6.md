# Отчет о спринте 0.7.6: Хотфиксы управления досками/колонками, динамический title, адаптивный фавикон и обновление README

* **Статус**: ✅ Завершен (100% Definition of Done, 0 ошибок `make check-arch`, 0 ошибок `npm run build`, `PASS` всех тестов)
* **Дата закрытия**: 2026-10-20

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

1. **Удаление и переименование досок (Board Management Hotfix)**:
   - Контракт OpenAPI: добавлен `boardItem` с операциями `DELETE` и `PATCH` по пути `/workspaces/{workspace_id}/projects/{project_id}/boards/{board_id}`.
   - Бизнес-правило удаления: запрещено удалять единственную доску проекта (`ErrLastBoard` $\rightarrow$ 400).
   - Безопасность данных: при удалении доски все задачи из её колонок внутри транзакции перемещаются в Inbox (`project_id = NULL, column_id = NULL`), что сохраняет данные пользователя и исключает нарушение внешнего ключа `fk_tasks_column ON DELETE RESTRICT`.
   - Frontend: реализованы `useDeleteBoardMutation` и `useUpdateBoardMutation`, проброшены `onDelete` и `onRename` в `BoardsRail.tsx` $\rightarrow$ `BoardCard.tsx`. При удалении активной доски фокус автоматически переключается на первую доступную доску.
2. **Кнопка удаления колонки (Column Deletion UI)**:
   - Эндпоинт `DELETE /workspaces/{workspace_id}/boards/{board_id}/columns/{column_id}` поддержан с валидацией пустоты колонки (`ErrColumnNotEmpty` $\rightarrow$ 409).
   - Frontend: реализован `useDeleteColumnMutation(workspaceId, boardId)`.
   - В шапке колонки `KanbanColumn.tsx` добавлена кнопка удаления с иконкой мусорной корзины. Если в колонке есть задачи — кнопка деактивирована с подсказкой «Нельзя удалить колонку с задачами. Сначала переместите или удалите задачи.».
3. **Динамический Title страницы**:
   - Формат вкладки: `<Название активной доски> | Takt`.
   - Fallback: если открыты «Входящие» — `Входящие | Takt`, если «Сегодня» — `Сегодня | Takt`, иначе `Takt`.
4. **Адаптивный / Тематический Favicon**:
   - Векторная лаконичная SVG-иконка Takt (ритмический штрих / тактовый акцент).
   - Динамическое обновление фавикона при смене темы оформления в соответствии с акцентным цветом темы (`var(--accent-solid)`).
5. **Полная зачистка подписи «Soft Editorial GTD»**:
   - Удаление из `index.html` (`<title>Takt</title>`), `App.tsx` (витрина компонентов) и `AuthScreen.tsx` (футер авторизации).
6. **Обновление корневого `README.md`**:
   - Новый раздел о темах оформления и кинетическом Ambient-движке (16 палитр, живой океанический риф с SVG-крабом, метель Киото, осенний листопад Tsukimi, звездное небо).
   - Скриншоты тем и актуализация архитектурной схемы.
7. **Глобальный фикс инпутов в темных темах (Fast-Track Hotfix)**:
   - Добавлены дефолты `input, textarea { background-color: var(--bg-surface); color: var(--text-primary); }` в `index.css`.
   - Устранен сбой с `bg-canvas/40` и `bg-canvas/50`, приводивший к белому фону инпутов создания досок и задач.

---

## 📋 2. Реализованный скоуп работ

* `api/openapi/paths/projects.yaml`: операции `DELETE` и `PATCH` для досок.
* `backend/internal/application/project/`: `DeleteBoard`, `UpdateBoard` с сохранением задач в Inbox.
* `backend/internal/infrastructure/http/project_handler.go`: хендлеры удаления и переименования досок.
* `frontend/src/features/kanban-board/KanbanColumn.tsx`: удаление колонок и валидация.
* `frontend/src/widgets/boards-panel/BoardsRail.tsx` & `BoardCard.tsx`: удаление и переименование досок.
* `frontend/src/shared/lib/hooks/useDocumentTitle.ts`: динамический title вкладки.
* `frontend/src/shared/lib/favicon.ts`: реактивный SVG-фавикон под тему.
* `README.md`: витрина тем со скриншотами.

---

## 🎯 3. Definition of Done (DoD) Метрики качества

- [x] Все пункты ТЗ спринта 0.7.6 выполнены.
- [x] `make check-arch` — 0 ошибок (все лимиты строк и чистота слоев соблюдены).
- [x] `npm run build && npm run lint` — 0 ошибок компиляции и типов.
- [x] `make test` — 100% юнит-тестов зеленые.

