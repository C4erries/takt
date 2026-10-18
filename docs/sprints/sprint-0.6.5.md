# Спринт 0.6.5: Умные поля условий Да/Нет, Даты и Приоритеты, Оператор «Один из» (OR/IN), Авто-выполнение и Статус дедлайна

* **Период проведения**: 2026-10-18
* **Статус**: 🟢 Завершен (100% DoD)
* **Ветка**: `main`

---

## 🎯 Цели и Результаты спринта

1. **Контекстные селекторы значений под каждое поле условия (`ConditionValuePicker`)**:
   - **«Завершена» (`is_completed`)**: сырое текстовое поле заменено на две понятные кнопки-пилюли `[ Да (выполнена) ]` и `[ Нет (в работе) ]` (сохраняют булевы значения `true` / `false`).
   - **«Приоритет» (`priority`)**: при обычных операторах — выпадающий список с цветными Washi-маркерами; при операторе «Один из» (`in`) — интерактивные кликабельные чипы мультиселекта (`[✓ Срочный] [✓ Высокий] ...`).
   - **«Статус дедлайна» (`due_date_status`)**: новый виртуальный статус задачи («Просрочен», «Горит (<= 48ч)», «В графике», «Без срока»).
   - **«Дедлайн» (`due_date`)**: быстрые пресеты в 1 клик `[ Сегодня ]`, `[ Завтра ]`, `[ Просрочен ]` плюс нативный календарный выбор даты (`YYYY-MM-DD`).
   - **«Время в колонке» (`time_in_column_hours`)**: пресеты `[ 24 ч ]`, `[ 48 ч ]`, `[ 72 ч ]` и числовой ввод с суффиксом «ч».

2. **Поддержка логики «ИЛИ» (OR) и наборов допустимых значений**:
   - Добавлены операторы **`in` («один из»)** и **`not_in` («ни один из»)** в спецификацию [automation.yaml](file:///home/pavel/projects/takt/api/openapi/schemas/automation.yaml), доменный слой [rule.go](file:///home/pavel/projects/takt/backend/internal/domain/automation/rule.go) и обработчик [evaluator.go](file:///home/pavel/projects/takt/backend/internal/application/automation/evaluator.go).
   - Прямо между строками условий отображается интерактивный бейдж связки: `[ ЕСЛИ ] Условие 1` $\rightarrow$ `[ И ▾ / ИЛИ ▾ ]` $\rightarrow$ `Условие 2` для наглядного управления логикой правила.

3. **Расширенные действия Low-Code**:
   - ✅ **«Отметить выполненной»** (`mark_completed`): автоматическая установка `is_completed = true` (поддержан SQL-запрос `SetTaskCompletion` и вызов в `evaluator.go`).
   - 🔄 **«Снять выполнение»** (`mark_uncompleted`): снятие отметки выполнения задачи (`is_completed = false`).
   - 🗑️ **«Очистить дедлайн»** (`clear_due_date`): сброс даты срока задачи.
   - Для этих действий в конструкторе отображается лаконичный бейдж *«Параметры не требуются»*.

4. **Инфраструктура и кодогенерация**:
   - В [tasks.sql](file:///home/pavel/projects/takt/backend/queries/tasks.sql) добавлен запрос `SetTaskCompletion`.
   - В [ports.go](file:///home/pavel/projects/takt/backend/internal/application/automation/ports.go) расширен интерфейс `TaskModifier`.
   - В [automation_repository.go](file:///home/pavel/projects/takt/backend/internal/infrastructure/repository/automation_repository.go) реализована модификация завершения задачи с проверкой `workspace_id`.

---

## 👥 Команда и Распределение ролей

* **Lead-оркестратор**: Проектирование селекторов значений, согласование логики «Один из» и операторов И/ИЛИ, решение конфликтов docker-образов, Playwright E2E-верификация.
* **`backend_developer`**: Contract-First Шаг 1 и Шаг 2 (OpenAPI enums, sqlc `SetTaskCompletion`, `TaskModifier`, чистый домен, вычисление `due_date_status`, операторов `in`/`not_in` и выполнение новых действий в `evaluator.go` с table-driven тестами).
* **`frontend_developer`**: Реализация `ConditionValuePicker` в `RuleBuilderSections.tsx`, чипы мультиселекта приоритетов, кнопки Да/Нет, пресеты дедлайнов, интерактивная связка И/ИЛИ между строками, бейджи действий.
* **`qa_engineer` & Lead**: Сквозное тестирование в браузере Playwright с валидацией 3 артефактов.

---

## 🧪 Метрики качества и верификация

1. **`make check-arch`**: PASSED (0 ошибок, лимит строк < 400, ровно 7 рабочих файлов в `features/automations/`).
2. **`make test`**: PASSED (100% Go unit-тестов с флагом `-race`).
3. **`npm run build`**: PASSED (0 ошибок компиляции TypeScript и Vite).
4. **Браузерная верификация Playwright**:
   - `sprint_065_1_smart_condition_pickers.png` — кнопки Да/Нет, чипы приоритетов «Один из», статус дедлайна, пресеты дедлайна и связка ИЛИ.
   - `sprint_065_2_new_actions_mark_completed.png` — действия «Отметить выполненной» и «Очистить дедлайн» с бейджем «Параметры не требуются».
   - `sprint_065_3_created_rule_details.png` — сохраненное правило в окне «Подробнее о правиле» с человекочитаемыми условиями.
