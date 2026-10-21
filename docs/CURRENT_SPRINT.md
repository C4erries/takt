# Текущий спринт / Оперативный центр управления (`CURRENT_SPRINT.md`)

> Этот файл — **единственная точка правды** о том, какие решения приняты по текущей задаче и что команда и AI-агенты делают прямо сейчас.
> Работа над каждым спринтом строго делится на **Фазу А (Штаб: Принятие решений)** и **Фазу Б (Конвейер: Реализация и Верификация)**.
> Завершенные спринты архивируются в каталог `docs/sprints/sprint-X.Y.Z.md` с записью в реестр `docs/sprints/README.md`.

## Текущий спринт: Спринт 0.8.0b (Мигратор в Docker Compose, умный перенос задач Linear-style, богатая палитра тегов, скрытие номеров задач и Error Observability) 🟢 ФАЗА Б (Завершено & Верифицировано)

---

## 🧠 1. Аналитика и Зафиксированные микро-решения спринта (Штаб)

* **Бизнес-цель**: Полностью автоматизировать накат миграций через отдельный сервис `migrator` в Docker Compose, внедрить прозрачное логирование ошибок в бэкенде, предоставить мощный диалог перемещения задачи между любыми досками и колонками с поиском, расширить палитру тегов до 12 цветов с возможностью редактирования и убрать визуальный шум от номеров задач по умолчанию.
* **Принятые микро-решения**:
  1. **Сервис `migrator` в `docker-compose.yml`**:
     - Добавляем сервис `migrator` на базе официального образа `migrate/migrate:v4.17.0`.
     - Монтируем папку `backend/migrations` в `/migrations`.
     - Команда: `-path=/migrations -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable" up`.
     - Сервис `backend` настраивается с зависимостью:
       ```yaml
       depends_on:
         postgres: { condition: service_healthy }
         migrator: { condition: service_completed_successfully }
       ```
     - Гарантирует автоматический накат всех миграций ДО старта бэкенда при каждом запуске!
  2. **Error Observability в бэкенде**:
     - В `tasks_handler.go`, `project_handler.go` и `middleware.go`: при `status >= 500` выводить подробный `log.Printf("[ERROR 500] %s %s: %+v", c.Request().Method, c.Request().RequestURI, err)` в консоль контейнера.
     - Для клиента сохраняем чистый безопасный ответ `"Internal server error"`, но в `docker compose logs backend` видна полная причина сбойного запроса.
  3. **Масштабируемый диалог перемещения задачи (Linear/WEEEK-style Move Modal)**:
     - Создание компонента `MoveTaskModal.tsx`:
       - Строка быстрого поиска по названию досок и колонок проекта.
       - Двухуровневый аккордеон/дерево: раскрывающиеся доски с колонками.
       - Ограничение по высоте `max-h-80 overflow-y-auto` — окно никогда не вылезает за пределы экрана даже при 50+ досках.
       - Клик по колонке целевой доски мгновенно вызывает `moveTaskMutation` (или `updateTaskMutation`) с обновлением `board_id` и `column_id`.
  4. **Расширенная палитра тегов (12 цветов) и редактирование цвета тега**:
     - В `TagSection.tsx` и `TagCapsule.tsx`:
       - Добавление всплывающего поповера смены цвета тега при клике на тег в шторке.
       - Палитра из 12 оттенков: Persimmon, Terracotta, Amber, Bamboo, Emerald, Ocean, Denim, Indigo, Lavender, Plum, Cocoa, Sand.
       - Мутация обновления цвета тега `updateTagColor` (или `PATCH /tags/{id}`).
  5. **Скрытие номеров задач по умолчанию**:
     - В `KanbanCard.tsx` показ бейджа `#86` зависит от флага `showTaskNumbers`.
     - По умолчанию `showTaskNumbers = false` (номер скрыт).
     - В панель вида/фильтрации (`ViewToggle.tsx` или настройки доски) добавляется аккуратный тумблер `Номера задач`.

---

## 📋 Таблица задач спринта 0.8.0b

| # | Задача | Статус | Приоритет | Ответственный | Заметки и контракты |
|---|--------|--------|-----------|---------------|---------------------|
| 1 | **[Docker Compose Migrator & Error Logging]**: Добавление сервиса `migrator` в `docker-compose.yml`, `depends_on: service_completed_successfully`, dev-логирование 500 в бэкенде | 🔄 В работе | Критический | `backend_developer` | `docker-compose.yml`, `tasks_handler.go`, `middleware.go` |
| 2 | **[Tag Color Editing & 12-Color Palette]**: Расширенная палитра из 12 токенов, редактирование цвета тега в шторке с сохранением в API | ⏳ Ожидает | Высокий | `frontend_developer` | `TaskTagSection.tsx`, `tags.ts` |
| 3 | **[Move Task Modal with Search]**: Диалог перемещения задачи между любыми досками и колонками с поиском и скроллом `max-h-80` | ⏳ Ожидает | Высокий | `frontend_developer` | `MoveTaskModal.tsx`, `KanbanCard.tsx` |
| 4 | **[Task ID Visibility Toggle]**: Скрытие `#ID` по умолчанию, тумблер в настройках вида доски | ⏳ Ожидает | Средний | `frontend_developer` | `KanbanCard.tsx`, `ViewToggle.tsx` |
| 5 | **[Verification Pipeline & Auto-Commit]**: `make check-arch`, `npm run build`, `npm run lint`, `make test`, авто-коммит | ⏳ В очереди | Критический | Lead Orchestrator | Gatekeeper Pipeline |

---

## 🎯 Definition of Done (DoD)

- [ ] В `docker-compose.yml` есть отдельный сервис `migrator`, миграции накатываются автоматически перед запуском бэкенда.
- [ ] Ошибки 500 подробно логируются в консоль бэкенда с указанием метода, URL и первопричины.
- [ ] Пользователь может легко перенести задачу на любую доску и в любую колонку через модалку с поиском.
- [ ] Цвет любого тега можно изменить в шторке задачи, выбрав из 12 стильных оттенков.
- [ ] Номера задач `#ID` скрыты по умолчанию на карточках канбана и включаются только при активации тумблера.
- [ ] Все тесты `make check-arch`, `npm run build`, `npm run lint`, `make test` проходят с 0 ошибок.
