# Отчет о спринте 0.7.7: Переименование колонок, персистентный порядок досок и живой стохастический ветер

* **Статус**: ✅ Завершен (100% Definition of Done, 0 ошибок `make check-arch`, 0 ошибок `npm run build`, `PASS` всех тестов)
* **Дата закрытия**: 2026-10-21

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

1. **Переименование колонок (Column Renaming)**:
   - В `KanbanColumn.tsx` поддержан двойной клик по названию колонки и пункт «Переименовать» в контекстном меню 3 точек.
   - Полноценный инлайн-инпут со стилями дизайн-паспорта (`bg-surface`, `border-subtle`, `text-primary`).
   - Сохранение по `Enter` или `blur`, отмена по `Escape`. При попытке сохранить пустое название выполняется откат. Вызывается `useUpdateColumnMutation`.
2. **Персистентный порядок досок (Board Reordering via DND & DB)**:
   - Создана миграция `backend/migrations/000012_board_positions.up.sql` (`position DOUBLE PRECISION NOT NULL DEFAULT 1000.0`, индекс `idx_boards_project_position`).
   - `backend/queries/projects.sql`: запросы `ListBoardsByProject`, `CreateBoard`, `UpdateBoard` обновлены с поддержкой `position` и сортировкой `ORDER BY position ASC, created_at ASC`.
   - OpenAPI: схема `Board` расширена обязательным `position`, а `boardItem` (PATCH) — опциональным `position`.
   - Домен и Use Cases Go обновлены: `Board.Position`, `Board.UpdatePosition`, `UpdateBoardInput`, маппинг в репозитории и `project_handler.go`.
   - Frontend: `BoardsRail.tsx` оснащен `@dnd-kit/core` и `@dnd-kit/sortable` (verticalListSortingStrategy, PointerSensor с distance: 5). При drop пересчитывается Fractional Indexing позиция и отправляется в `useUpdateBoardMutation`.
3. **Живой стохастический ветер в хвойной теме (`frost-pine`)**:
   - В `frontend/src/shared/lib/ambient/presets/snowPreset.ts` детерминированная гармоника заменена на многослойную стохастическую модель:
     - Медленно блуждающий базовый дрейф ветра (`drift`).
     - Случайные нерегулярные порывы (`gustTarget`, `gustCurrent`) со случайной длительностью и интервалами затишья.
     - Микро-турбулентность для естественного кружения хвоинок (`needles`) и снежинок.

---

## 📋 2. Реализованный скоуп файлов

* `backend/migrations/000012_board_positions.up.sql` & `.down.sql`
* `backend/queries/projects.sql`
* `api/openapi/schemas/project.yaml`, `api/openapi/paths/projects.yaml`
* `backend/internal/domain/project/board.go`, `project_test.go`
* `backend/internal/application/project/dto.go`, `ports.go`, `service_board.go`, `service_board_test.go`
* `backend/internal/infrastructure/repository/project_repository.go`, `helpers.go`
* `backend/internal/infrastructure/http/project_handler.go`, `project_test.go`
* `frontend/src/entities/project/api.ts`, `queries.ts`, `types.ts`
* `frontend/src/features/kanban-board/KanbanColumn.tsx`
* `frontend/src/widgets/boards-panel/BoardsRail.tsx`, `BoardCard.tsx`
* `frontend/src/shared/lib/ambient/presets/snowPreset.ts`

---

## 🎯 3. Definition of Done (DoD) Метрики качества

- [x] Все пункты ТЗ спринта 0.7.7 выполнены.
- [x] `make check-arch` — 0 ошибок (все лимиты строк и чистота слоев соблюдены).
- [x] `npm run build && npm run lint` — 0 ошибок компиляции и типов.
- [x] `make test` — 100% юнит-тестов зеленые.

