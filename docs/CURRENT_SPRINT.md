# Текущий спринт / Оперативный центр управления (`CURRENT_SPRINT.md`)

> Этот файл — **единственная точка правды** о том, какие решения приняты по текущей задаче и что команда и AI-агенты делают прямо сейчас.
> Работа над каждым спринтом строго делится на **Фазу А (Штаб: Принятие решений)** и **Фазу Б (Конвейер: Реализация и Верификация)**.
> Завершенные спринты архивируются в каталог `docs/sprints/sprint-X.Y.Z.md` с записью в реестр `docs/sprints/README.md`.

## Текущий спринт: Спринт 0.9.6 — «UI & UX Polish: Заметки, Навигация, Метрики и Скоупы тегов» 🟡 ФАЗА Б (В работе)

---

## 🧠 1. Аналитика и Зафиксированные микро-решения спринта (Штаб)

* **Бизнес-цель**: Закрытие всех накопившихся UX-раздражителей (редактирование заметок без прыжков, компактные иконки метрик досок, надёжная навигация тегов) и внедрение проектных/глобальных тегов (БД + API + UI).

* **Связь с бэклогом**: [`docs/BACKLOG.md`](BACKLOG.md) (TASK-B16, TASK-B17, TASK-B18, TASK-B19, TASK-B22, TASK-B23).

* **Принятые микро-решения**:

  1. **TASK-B16 (Редактирование описания по кнопке)**:
     - `MarkdownNotes.tsx`: убрать `onClick` со всего контейнера превью. Текст можно свободно выделять и копировать.
     - Добавить кнопку-карандаш (`Pencil`) в шапку блока. Клик по кнопке (или клик по пустому плейсхолдеру) переключает в режим редактирования.

  2. **TASK-B17 (Иконки Columns и Activity на карточках досок)**:
     - `BoardCard.tsx`: заменить текст `N колонки · M в работе` на векторные иконки Lucide (`Columns3` и `Activity`) с цветом `text-muted`.

  3. **TASK-B18 (Надежная навигация и компактность тегов)**:
     - `Sidebar.tsx`: перенести «Теги» в верхний блок навигации (рядом с Архивом, Входящими и Сегодня) со счетчиком тегов. При клике сбрасывать `selectedProjectId` и `activeBoardId`.
     - `MainSheet.tsx`: исправить условие `isProjectNav` (не должно перекрывать явные роуты `'tags'`, `'archive'`, `'today'`, `'inbox'`, `'automations'`).
     - `TagManagerView.tsx`: выравнивание CSS Grid `content-start` и `auto-rows-min`, фиксированная компактная высота карточек.

  4. **TASK-B19 (Загрузка досок всех проектов в конструкторе)**:
     - `AutomationsHub.tsx`: вместо чтения кэша React Query гарантированно запрашивать доски всех проектов воркспейса (чтобы проекты вроде «GTD» всегда отображались). Очистить термины «SLA» в текстах описания.

  5. **TASK-B22 & B23 (Скоупы тегов: Workspace vs Project)**:
     - Миграция `000016_tag_scopes.up.sql`:
       ```sql
       ALTER TABLE tags ADD COLUMN project_id UUID NULL REFERENCES projects(id) ON DELETE CASCADE;
       DROP INDEX IF EXISTS idx_tags_workspace_name;
       CREATE UNIQUE INDEX idx_tags_workspace_project_name 
         ON tags(workspace_id, COALESCE(project_id, '00000000-0000-0000-0000-000000000000'), LOWER(name));
       CREATE INDEX idx_tags_project_id ON tags(project_id) WHERE project_id IS NOT NULL;
       ```
     - OpenAPI контракт: поле `projectId: UUID | null` в схеме `Tag` и `CreateTagRequest`.
     - `GET /workspaces/{wid}/tags?project_id=...`: возвращает глобальные теги воркспейса + теги указанного проекта.
     - `TagManagerView.tsx`: фильтр «Все теги» / «Глобальные» / «Теги проекта», бейдж проекта на карточке тега.
     - Автокомплит тегов задачи: предлагает только глобальные теги воркспейса + теги проекта задачи.

---

## 📋 2. Таблица задач спринта 0.9.6

| # | Задача | Статус | Приоритет | Ответственный | Файлы |
|---|--------|--------|-----------|---------------|-------|
| 1 | **[B22 Step 1] Tag Scopes Schema & OpenAPI**: миграция 000016, sqlc queries, openapi.yaml, make generate, чистый домен `Tag` | 🔄 В работе | Высокий | `backend_developer` | `migrations/`, `queries/tasks.sql`, `api/openapi/*`, `internal/domain/task/` |
| 2 | **[B22 Step 2] Tag Scopes Application & HTTP**: use cases, repo methods, http handler с поддержкой `project_id`, тесты | ⏳ Ожидает | Высокий | `backend_developer` | `application/task/`, `infrastructure/repository/`, `infrastructure/http/` |
| 3 | **[B16-B19 UI Polish]**: кнопка-карандаш в MarkdownNotes, иконки в BoardCard, навигация сайдбара и компактность TagManagerView, доски в AutomationsHub | 🔄 В работе | Высокий | `frontend_developer` | `MarkdownNotes.tsx`, `BoardCard.tsx`, `Sidebar.tsx`, `MainSheet.tsx`, `AutomationsHub.tsx`, `TagManagerView.tsx` |
| 4 | **[B23 UI] Tag Scopes UI**: фильтрация по проектам в TagManagerView, бейджи скоупа, автокомплит тегов по проекту | ⏳ Ожидает | Средний | `frontend_developer` | `TagManagerView.tsx`, `TaskTagSection.tsx` |
| 5 | **[Verification]**: `make check-arch`, `npm run build`, `npm run lint`, `make test`, `docker compose up -d --build`, `pytest tests/`, коммит | ⏳ В очереди | Критический | Lead Orchestrator | — |

---

## 🎯 3. Definition of Done (DoD)

- [ ] В `MarkdownNotes` текст свободно выделяется и копируется; редактирование открывается по кнопке-карандашу или по клику на пустое описание.
- [ ] На карточках досок в `BoardsRail` отображаются иконки `<Columns3 />` и `<Activity />` вместо текста.
- [ ] Пункт «Теги» перенесен в верхний блок сайдбара со счетчиком; страница тегов надежно открывается с любой доски проекта.
- [ ] Карточки тегов в `TagManagerView` имеют компактную высоту и не растягиваются на весь экран.
- [ ] В конструкторе правил гарантированно загружаются и отображаются доски всех проектов воркспейса (включая GTD).
- [ ] Выполнена миграция `000016_tag_scopes` (поле `project_id` и уникальный составной индекс).
- [ ] API поддерживает получение и создание тегов с привязкой к проекту (`project_id`).
- [ ] В интерфейсе тегов отображаются проектные бейджи и фильтр по проектам; в карточке задачи предлагаются релевантные теги.
- [ ] `make check-arch` — 0 ошибок (файлы < 400 строк, папки <= 7 файлов).
- [ ] `npm run build && npm run lint` — 0 ошибок.
- [ ] `make test` — все тесты зелёные.
- [ ] `uv run pytest tests/` — 100% PASS.

---

## 🗓️ 4. Очередь следующих спринтов

| Спринт | Задачи | Суть |
|--------|--------|------|
| **0.9.7** | B20, B21, B01, B09 | Automations Suite (возврат в колонку, триггеры выполнения, подзадачи) & Профиль пользователя |
