# Спринт 0.6.3: Визуализация приоритетов и дедлайнов, логика ИЛИ и Каталог правил колонок

* **Период проведения**: 2026-10-18
* **Статус**: 🟢 Завершен (100% DoD)
* **Ветка**: `main`

---

## 🎯 Цели и Результаты спринта

1. **Цветовая кодировка карточек задач по приоритету (`border-l-4`)**:
   - Реализована акцентная левая полоса шириной 4px на [KanbanCard.tsx](file:///home/pavel/projects/takt/frontend/src/features/kanban-board/KanbanCard.tsx):
     * `urgent` (Срочный) $\rightarrow$ синий деним (`border-l-accent-denim`).
     * `high` (Высокий) $\rightarrow$ терракотовый / persimmon (`border-l-accent-persimmon`).
     * `medium` (Средний) $\rightarrow$ янтарная солома (`border-l-accent-straw`).
     * `low` (Низкий) $\rightarrow$ мягкий свежий шалфей / бамбуковый зеленый (`border-l-accent-bamboo`, `#40C057`).
     * `none` (Обычный) $\rightarrow$ нейтральный тон Washi.
   - Токен `--accent-bamboo` добавлен в [index.css](file:///home/pavel/projects/takt/frontend/src/index.css) и [tailwind.config.js](file:///home/pavel/projects/takt/frontend/tailwind.config.js).

2. **Заблаговременные умные бейджи дедлайнов (Pre-Deadline Smart Badges)**:
   - 🔴 **Просрочено** $\rightarrow$ предупреждающий бейдж с иконкой `AlertCircle` ⚠️ в оттенке `persimmon` (например, `10 окт.`).
   - 🟠 **Горящий срок $\le 48$ часов** $\rightarrow$ акцентный бейдж с иконкой `Clock` ⏱️ в теплом янтарном оттенке `straw` (`Сегодня`, `Завтра`).
   - ⚪ **Более 2 дней** $\rightarrow$ спокойный нейтральный бейдж дедлайна.

3. **Каталог существующих правил и привязка к колонкам в 1 клик**:
   - В [ColumnAutomationsModal.tsx](file:///home/pavel/projects/takt/frontend/src/features/kanban-board/ColumnAutomationsModal.tsx) добавлена секция «Кастомные правила для этой колонки» с отображением правил, тумблером активности и кнопкой открепления.
   - Добавлен выпадающий список **«+ Прикрепить из каталога»**, позволяющий в 1 клик привязать любое существующее в воркспейсе правило к колонке.
   - В [automations_handler.go](file:///home/pavel/projects/takt/backend/internal/infrastructure/http/automations_handler.go) и [automation_repository.go](file:///home/pavel/projects/takt/backend/internal/infrastructure/repository/automation_repository.go) поддержаны привязка (`columnId`) и открепление (`columnId: null`, `ClearColumnID`).

4. **Логика условий «ИЛИ» (OR) в Low-Code движке и UI**:
   - В OpenAPI манифест [automation.yaml](file:///home/pavel/projects/takt/api/openapi/schemas/automation.yaml) и домен [rule.go](file:///home/pavel/projects/takt/backend/internal/domain/automation/rule.go) добавлено поле `logicalOperator: "and" | "or"`.
   - В [evaluator.go](file:///home/pavel/projects/takt/backend/internal/application/automation/evaluator.go) поддержано короткое замыкание при логике `or`.
   - В [RuleBuilderSections.tsx](file:///home/pavel/projects/takt/frontend/src/features/automations/RuleBuilderSections.tsx) добавлен сегментированный переключатель «Все (И)» / «Хотя бы одно (ИЛИ)».

5. **Зачистка технического ИИ-жаргона**:
   - Удалены все технические термины (`EventBus`, `Cycle Guard`, `глубина каскада`, `в стиле японской канцелярии Washi`).
   - Формулировки заменены на лаконичные продуктовые описания (*"Правила срабатывают мгновенно и защищены от зацикливания"*).

---

## 👥 Команда и Распределение ролей

* **Lead-оркестратор**: Архитектура контрактов, согласование палитры приоритетов, решение конфликтов `twMerge`, браузерная Playwright-приемка.
* **`backend_developer`**: Contract-First Шаг 1 и Шаг 2 (поддержка `logicalOperator` в OpenAPI, домене и `evaluator.go`, привязка/отвязка колонок в sqlc/handler, Go unit-тесты).
* **`frontend_developer`**: Реализация `border-l-4` и смарт-бейджей в `KanbanCard`, секция каталога в `ColumnAutomationsModal`, переключатель И/ИЛИ в `RuleBuilderSections`, удаление ИИ-жаргона.
* **`qa_engineer` & Lead**: Сквозное тестирование в браузере Playwright с валидацией 4 скриншотов.

---

## 🧪 Метрики качества и верификация

1. **`make check-arch`**: PASSED (0 ошибок).
2. **`make test`**: PASSED (100% Go unit-тестов с флагом `-race`).
3. **`npm run build`**: PASSED (0 ошибок компиляции TypeScript и Vite).
4. **Браузерная верификация Playwright**:
   - `sprint_063_1_priority_borders.png` — карточки с акцентными цветными полосами Denim, Terracotta, Straw, Sage Green.
   - `sprint_063_2_pre_deadline_badges.png` — умные бейджи дедлайнов (просрочено ⚠️ и горящий срок ⏱️ Сегодня).
   - `sprint_063_3_column_catalog_binding.png` — секция кастомных правил колонки и привязка из каталога.
   - `sprint_063_4_or_logic_builder.png` — конструктор с активным переключателем «Хотя бы одно (ИЛИ)».

