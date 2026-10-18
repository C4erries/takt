# Спринт 0.6.2: Визуальный Low-Code конструктор правил, таймеры и расписания

* **Период проведения**: 2026-10-18
* **Статус**: 🟢 Завершен (100% DoD)
* **Ветка**: `main`

---

## 🎯 Цели и Результаты спринта

1. **Архитектурный фундамент и Low-Code DSL ([ADR-0004](file:///home/pavel/projects/takt/docs/decisions/0004-low-code-automations-dsl.md))**:
   - Миграция БД [000009_low_code_automations_dsl.up.sql](file:///home/pavel/projects/takt/backend/migrations/000009_low_code_automations_dsl.up.sql): добавлено поле `rule_definition JSONB NOT NULL DEFAULT '{}'::jsonb`.
   - В OpenAPI манифесте [automation.yaml](file:///home/pavel/projects/takt/api/openapi/schemas/automation.yaml) зафиксированы схемы `RuleDefinition`, `Trigger`, `Condition`, `Action`, обновлены DTO создания/обновления.
   - В чистом доменном слое [rule.go](file:///home/pavel/projects/takt/backend/internal/domain/automation/rule.go) реализована типизация DSL, валидация инвариантов (`ValidateRuleDefinition`) с лимитами (до 5 условий, до 5 действий) и табличные тесты в [rule_test.go](file:///home/pavel/projects/takt/backend/internal/domain/automation/rule_test.go).

2. **Движок исполнения правил и фоновый тикер-воркер (Backend)**:
   - В [evaluator.go](file:///home/pavel/projects/takt/backend/internal/application/automation/evaluator.go) реализован `RuleEvaluator`: вычисление условий (логика AND, операторы `equals`, `not_equals`, `contains`, `not_contains`, `is_empty`, `is_not_empty`, `greater_than`, `less_than`) и последовательное применение действий (`move_to_column`, `set_priority`, `add_tag`, `remove_tag`, `set_due_date`, `archive_task`).
   - В [listener.go](file:///home/pavel/projects/takt/backend/internal/application/automation/listener.go) интегрирована реактивная диспетчеризация событий с защитой от циклов (`Cycle Guard`, глубина цепочки $\le 3$).
   - В [worker.go](file:///home/pavel/projects/takt/backend/internal/application/automation/worker.go) реализован фоновый планировщик/тикер для проверки Aging SLA таймеров (`column_sla_timer`) и ежедневных расписаний (`scheduled_sweep`).
   - В [automations_handler.go](file:///home/pavel/projects/takt/backend/internal/infrastructure/http/automations_handler.go) поддержаны CRUD-операции над правилами с валидацией схемы и маппингом.

3. **Sentence Builder и Хаб автоматизаций (Frontend UI)**:
   - В [RuleBuilderModal.tsx](file:///home/pavel/projects/takt/frontend/src/features/automations/RuleBuilderModal.tsx) и [RuleBuilderSections.tsx](file:///home/pavel/projects/takt/frontend/src/features/automations/RuleBuilderSections.tsx) реализован карточный конструктор «КОГДА $\rightarrow$ ЕСЛИ $\rightarrow$ ТО» с автогенерацией понятных названий правил.
   - В [AutomationsHub.tsx](file:///home/pavel/projects/takt/frontend/src/features/automations/AutomationsHub.tsx) создан центральный экран воркспейса (фильтрация по колонкам/глобальным, статус-бейдж, баннер Cycle Guard, тумблеры Вкл/Выкл, удаление правил).
   - В левом сайдбаре [Sidebar.tsx](file:///home/pavel/projects/takt/frontend/src/widgets/sidebar/Sidebar.tsx) добавлен пункт навигации «⚡ Автоматизации» с URL-роутингом `/automations` и бейджем количества активных правил.
   - В [ColumnAutomationsModal.tsx](file:///home/pavel/projects/takt/frontend/src/features/kanban-board/ColumnAutomationsModal.tsx) добавлена кнопка «Собрать своё правило» с предзаполнением текущей колонки в триггере.

---

## 👥 Команда и Распределение ролей

* **Lead-оркестратор**: Проектирование ADR-0004, координация Contract-First фаз, независимая приемка, фикс резолва Nginx и E2E в браузере.
* **`backend_developer`**: Contract-First Шаг 1 (миграция, OpenAPI, sqlc, домен) и Шаг 2 (RuleEvaluator, тикер-воркер, Use Cases, хендлеры, тесты).
* **`frontend_developer`**: Фича `src/features/automations/` (Sentence Builder, AutomationsHub, AutomationCard), интеграция в сайдбар и модалку колонки.
* **`qa_engineer` & `lead_architect`**: Сквозная верификация Playwright в браузере с сохранением 4 визуальных артефактов.

---

## 🧪 Метрики качества и верификация

1. **`make check-arch`**: PASSED (0 ошибок по всем 5 проверкам архитектурного линтера).
2. **`npm run build`**: PASSED (0 ошибок сборки TypeScript и Vite).
3. **`make test`**: PASSED (100% Go unit-тестов с флагом `-race`).
4. **Браузерная верификация Playwright**:
   - `sprint_062_1_automations_hub.png` — экран Хаба автоматизаций с бейджем активных правил и фильтрами.
   - `sprint_062_2_rule_builder_modal.png` — интерфейс блочного конструктора правил «КОГДА $\rightarrow$ ЕСЛИ $\rightarrow$ ТО».
   - `sprint_062_3_rule_card_in_hub.png` — карточка созданного Low-Code правила с цепочкой чипов и тумблером.
   - `sprint_062_4_column_rule_builder.png` — вызов конструктора из контекстного меню колонки с авто-подстановкой триггера.
