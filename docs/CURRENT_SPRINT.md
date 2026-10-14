# Текущий спринт / Оперативный центр управления (`CURRENT_SPRINT.md`)

> Этот файл — **единственная точка правды** о том, какие решения приняты по текущей задаче и что команда и AI-агенты делают прямо сейчас.
> Работа над каждым спринтом строго делится на **Фазу А (Штаб: Принятие решений)** и **Фазу Б (Конвейер: Реализация и Верификация)**.
> Завершенные спринты архивируются в каталог `docs/sprints/sprint-X.Y.Z.md` с записью в реестр `docs/sprints/README.md`.

## Текущий спринт: Спринт 0.3.2 (Фаза 2 / Этап 0.3: Bento-панель инспектора, матричный календарь и фокус «Сегодня») ✅ ЗАВЕРШЕН (Готов к архивации)

---

## 🧠 1. Аналитика и Зафиксированные микро-решения спринта (Режим «Штаб»)

* **Бизнес-цель**: Реализовать правую Bento-панель (детальный инспектор задачи с Markdown-заметками, тегами контекста и свойствами + матричный календарь дедлайнов с точками плотности) и специализированный вид «Сегодня / Focus of the Day».
* **Зафиксированные продуктовые и архитектурные микро-решения**:
  1. **Модель тегов контекста (Workspace Tags)**:
     - Таблицы `tags` (id, workspace_id, name, color_hex, created_at) и связующая `task_tags` (task_id, tag_id).
     - Единый справочник тегов воркспейса с капсульными бейджами в палитре японской канцелярии (`@дом`, `@работа`, `20 мин`, `focus`).
     - Быстрое назначение и удаление тегов прямо из инспектора задачи.
  2. **Гибридный GTD-отбор в фокус «Сегодня» (`Today`)**:
     - В вид «Сегодня» автоматически попадают задачи, удовлетворяющие условию: `due_date <= CURRENT_DATE` (включая просроченные с акцентным бейджем) ИЛИ `focus_order IS NOT NULL` (задачи, явно перенесенные пользователем в фокус дня).
     - Сортировка: `focus_order ASC NULLS LAST, due_date ASC, priority DESC`.
     - Кнопка/действие «В фокус дня» / «Убрать из фокуса» (`focus_order`).
  3. **Интерактивный Матричный Календарь (Bento Calendar)**:
     - Сетка месяца с 1px тонкими разделителями Tatami (`border-subtle`).
     - Индикаторы дедлайнов (точки плотности задач под числами).
     - **Контекстно-зависимый клик**: если открыт инспектор задачи — клик по ячейке даты мгновенно назначает/меняет `due_date` этой задачи; если инспектор закрыт — фильтрует задачи в центральном листе по выбранной дате.
  4. **Editorial Split-Toggle Markdown Inspector**:
     - В режиме покоя: изящный типографичный рендер Markdown-заметки (шрифты Newsreader / Geist Sans).
     - По клику: бесшовный инлайн textarea-редактор с мгновенным autosave (дебаунс 400мс) и поддержкой клавиши `Escape` для завершения редактирования.
     - Мгновенные мутации при выборе приоритета (`none`, `low`, `medium`, `high`, `urgent`), проекта и колонки.
  5. **Двухфазная поставка Contract-First (oapi-codegen ServerInterface)**:
     - Шаг 1: Миграция `000004_tags_and_focus.up.sql`, sqlc-запросы, расширение OpenAPI-спецификации (`schemas/tag.yaml`, эндпоинты в `paths/tasks.yaml` и `paths/tags.yaml`), `make generate` (`ServerInterface`).
     - Шаг 2: Use Cases (`application/task` и `application/tag`), репозитории, HTTP-хендлеры, связывание в `container/` и тесты.
     - Параллельный фронтенд: `TaskInspector`, `MiniCalendar`, вид `TodayView` с J/K навигацией.

---

## 📋 Таблица задач спринта 0.3.2

| # | Задача | Статус | Приоритет | Ответственный | Заметки и контракты |
|---|--------|--------|-----------|---------------|---------------------|
| 1 | **[Backend Step 1]**: Миграция 0004 (tags, task_tags), sqlc-запросы, OpenAPI контракты и кодогенерация | ✅ Done | Критический | `backend_developer` | `tags`, `task_tags`, эндпоинты Today & Tags, `make generate` |
| 2 | **[Gatekeeper Checkpoint 1]**: Проверка контрактов и кодогенерации | ✅ Done | Критический | `lead_architect` | `make check-arch && make generate && make test` (0 ошибок) |
| 3 | **[Backend Step 2]**: Use-Cases тегов и задач, HTTP-хендлеры (ServerInterface), репозитории и тесты | ✅ Done | Высокий | `backend_developer` | `application/task`, `application/tag`, `infrastructure/http`, `infrastructure/repository` |
| 4 | **[Frontend]**: Правая Bento-панель (TaskInspector, MiniCalendar) и вид «Сегодня» (TodayView) | ✅ Done | Высокий | `frontend_developer` | `TaskInspector.tsx`, `MiniCalendar.tsx`, `TodayView.tsx`, TanStack Query |
| 5 | **[QA & Verification]**: E2E тесты инспектора, календаря, тегов и фокуса дня на Pytest | ✅ Done | Критический | `qa_engineer` | `tests/e2e/test_projects_and_bento.py`, `make check-arch`, `make test`, `make test-e2e` (17/17 PASS) |

---

## 🎯 Definition of Done (DoD)

- [x] Создана миграция `000004_tags_and_focus.up.sql` с таблицами `tags`, `task_tags` и индексами.
- [x] Описаны контракты OpenAPI для тегов (`/tags`), обновления задачи (`PATCH /tasks/{id}`), выборки `Today` (`GET /tasks/today`) и добавления в фокус (`POST /tasks/{id}/focus`).
- [x] HTTP-хендлеры строго реализуют сгенерированный `generated.ServerInterface` без рукописных DTO.
- [x] Разработан `TaskInspector` с Editorial Markdown split-toggle, выбором дедлайна, приоритета, проекта/колонки и тегами.
- [x] Разработан `MiniCalendar` с точками плотности задач и контекстным кликом (установка дедлайна или фильтрация).
- [x] Реализован вид «Сегодня / Focus of the Day» с гибридным отбором (`due_date <= today` или `focus_order IS NOT NULL`).
- [x] `make check-arch` — 0 ошибок (все 5 проверок пройдены).
- [x] `make test` — 100% PASS с флагом `-race`.
- [x] `cd frontend && npm run build` — 0 ошибок сборки.
- [x] Сквозные тесты Pytest `make test-e2e` покрывают жизненный цикл тегов, инспектора и фокуса «Сегодня» (17/17 PASS).

---

## Лог изменений спринта

| Дата | Что изменилось |
|------|---------------|
| 2026-10-14 | **Спринт 0.3.2 завершен**: Реализована правая Bento-панель (`TaskInspector` с Markdown split-toggle и тегами, матричный календарь `MiniCalendar` с плотностью дедлайнов и контекстным кликом) и специализированный вид «Сегодня» (`TodayView` с гибридным GTD-отбором). Написаны E2E-тесты на Pytest (17/17 PASS). Полная верификация пройдена на 100%. |
