# Текущий спринт / Оперативный центр управления (`CURRENT_SPRINT.md`)

> Этот файл — **единственная точка правды** о том, какие решения приняты по текущей задаче и что команда и AI-агенты делают прямо сейчас.
> Работа над каждым спринтом строго делится на **Фазу А (Штаб: Принятие решений)** и **Фазу Б (Конвейер: Реализация и Верификация)**.
> Завершенные спринты архивируются в каталог `docs/sprints/sprint-X.Y.Z.md` с записью в реестр `docs/sprints/README.md`.

## Текущий спринт: Спринт 0.7.6 (Хотфиксы управления досками/колонками, динамический title, адаптивный фавикон и обновление README) 🟢 ФАЗА Б (Завершено & Верифицировано)

---

## 🧠 1. Аналитика и Зафиксированные микро-решения спринта (Штаб)

* **Бизнес-цель**: Устранить блокеры управления структурой проекта (удаление и переименование досок, удаление колонок), завершить полировку идентичности сайта (динамический title вкладки, живой фавикон под тему, зачистка устаревших подписей) и обновить витрину проекта в README.
* **Ключевые микро-решения**:
  1. **Удаление и переименование досок (Board Management Hotfix)**:
     - Контракт OpenAPI: добавляем `boardItem` с операциями `DELETE` и `PATCH` по пути `/workspaces/{workspace_id}/projects/{project_id}/boards/{board_id}`.
     - Бизнес-правило удаления: запрещено удалять единственную доску проекта (`ErrLastBoard` $\rightarrow$ 400).
     - Безопасность данных: при удалении доски все задачи из её колонок внутри транзакции перемещаются в Inbox (`project_id = NULL, column_id = NULL`), что сохраняет пользовательские данные и исключает нарушение внешнего ключа `fk_tasks_column ON DELETE RESTRICT`.
     - Frontend: реализация `useDeleteBoardMutation` и `useUpdateBoardMutation`, проброс `onDelete` и `onRename` в `BoardsRail.tsx` $\rightarrow$ `BoardCard.tsx`. При удалении активной доски фокус автоматически переключается на первую доступную доску.
  2. **Кнопка удаления колонки (Column Deletion UI)**:
     - Эндпоинт `DELETE /workspaces/{workspace_id}/boards/{board_id}/columns/{column_id}` уже поддержан в бэкенде с валидацией пустоты колонки (`ErrColumnNotEmpty` $\rightarrow$ 409).
     - Frontend: реализуем `useDeleteColumnMutation(workspaceId, boardId)`.
     - В шапке колонки `KanbanColumn.tsx` добавляем кнопку удаления (с иконкой мусорной корзины). Если в колонке есть задачи — кнопка деактивирована с подсказкой «Переместите задачи перед удалением» либо выдается предупреждение.
  3. **Динамический Title страницы**:
     - Формат вкладки: `<Название активной доски> | Takt`.
     - Fallback: если открыты «Входящие» — `Входящие | Takt`, если «Сегодня» — `Сегодня | Takt`, иначе `Takt`.
  4. **Адаптивный / Тематический Favicon**:
     - Векторная лаконичная SVG-иконка Takt (стилизованный ритмический штрих / тактовый акцент).
     - Динамическое обновление фавикона при смене темы оформления в соответствии с акцентным цветом темы (`var(--accent-solid)`).
  5. **Полная зачистка подписи «Soft Editorial GTD»**:
     - Удаление из `index.html` (`<title>Takt</title>`), `App.tsx` (витрина компонентов) и `AuthScreen.tsx` (футер авторизации).
  6. **Обновление корневого `README.md`**:
     - Новый раздел о темах оформления и кинетическом Ambient-движке (16 палитр, живой океанический риф с SVG-крабом, метель Киото, осенний листопад Tsukimi, звездное небо).
     - Скриншоты тем и актуализация архитектурной схемы.

---

## 📋 Таблица задач спринта 0.7.6

| # | Задача | Статус | Приоритет | Ответственный | Заметки и контракты |
|---|--------|--------|-----------|---------------|---------------------|
| 1 | **[OpenAPI & Backend Board API]**: Добавление `boardItem` (DELETE & PATCH) в `paths/projects.yaml`, генерация типов, `project.Service.DeleteBoard/UpdateBoard`, Echo хендлеры и тесты | ✅ Готово | Критический | Lead (Backend Contract) | `paths/projects.yaml`, `service.go`, `project_handler.go` |
| 2 | **[Board Deletion UI]**: `useDeleteBoardMutation`, `useUpdateBoardMutation`, проброс `onDelete`/`onRename` в `BoardsRail.tsx` | ✅ Готово | Критический | `frontend_developer` | `queries.ts`, `BoardsRail.tsx` |
| 3 | **[Column Deletion UI]**: `useDeleteColumnMutation`, кнопка удаления в `KanbanColumn.tsx` | ✅ Готово | Высокий | `frontend_developer` | `columnQueries.ts`, `KanbanColumn.tsx` |
| 4 | **[Dynamic Title & Favicon]**: Title `<Доска> \| Takt`, реактивный SVG-фавикон под палитру темы, зачистка «Soft Editorial GTD» | ✅ Готово | Средний | `frontend_developer` | `useDocumentTitle`, `favicon`, `index.html`, `App.tsx`, `AuthScreen.tsx` |
| 5 | **[README.md & Theme Screenshots]**: Обновление главного `README.md`, снятие скриншотов новых тем | ✅ Готово | Средний | Lead / Chief Designer | `README.md`, `docs/assets/themes/` |
| 6 | **[Verification Pipeline]**: Прогоны `make check-arch`, `npm run build`, `npm run lint`, `make test` | ✅ Готово | Критический | Gatekeeper Pipeline | 0 ошибок |

---

## 🎯 Definition of Done (DoD)

- [x] Удаление доски работает через UI `BoardsRail` с подтверждением и корректным сбросом задач в Inbox.
- [x] Единственную доску проекта удалить нельзя (`canDelete = false` в UI, `ErrLastBoard` в API).
- [x] Переименование доски в `BoardCard` работает и сохраняется на бэкенде.
- [x] В шапке колонки `KanbanColumn` есть кнопка удаления колонки, пустая колонка успешно удаляется.
- [x] Title браузера отображает `<Название доски> | Takt` (или фоллбэки).
- [x] Фавикон реактивно обновляется под активную тему или имеет стильный SVG-вид.
- [x] Строка «Soft Editorial GTD» полностью удалена из UI и исходников.
- [x] Главный `README.md` содержит актуальные темы оформления и скриншоты.
- [x] `make check-arch`, `npm run build`, `npm run lint`, `make test` проходят без ошибок.
