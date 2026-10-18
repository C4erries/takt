# ADR-0004: Спецификация Low-Code DSL правил автоматизации (Trigger-Condition-Action)

* **Дата**: `2026-10-18`
* **Статус**: `Accepted`
* **Инициатор / Автор**: `Pavel (Founder) & Lead AI Architect`
* **Затрагиваемые подсистемы**: `api/openapi/*`, `backend/internal/domain/automation`, `backend/internal/application/automation`, `frontend/src/features/automations`

---

## 🎯 1. Контекст и Проблема (Context)

В рамках Фазы 5 (Спринт 0.6.2) требуется перейти от жестко зашитых пресетов к гибкому пользовательскому Low-Code конструктору («КОГДА $\rightarrow$ ЕСЛИ $\rightarrow$ ТО»). 
Необходимо зафиксировать единый, компактный и строго типизированный формат DSL (Domain Specific Language) для хранения в JSONB, валидации и выполнения на бэкенде и фронтенде.

---

## 📐 2. Синтаксис и Спецификация DSL (Syntax & JSON Schema)

Правило автоматизации описывается JSON-объектом `rule_definition`:

```json
{
  "trigger": {
    "type": "task_entered_column",
    "params": {
      "column_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"
    }
  },
  "conditions": [
    {
      "field": "priority",
      "operator": "equals",
      "value": "urgent"
    }
  ],
  "actions": [
    {
      "type": "move_to_column",
      "params": {
        "target_column_id": "c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f"
      }
    }
  ]
}
```

### 1. Триггеры (`trigger`)

Объект `{ "type": string, "params": object }`.

| Тип триггера (`type`) | Назначение | Параметры (`params`) |
|---|---|---|
| `task_entered_column` | Попадание задачи в колонку (DND, создание) | `{ "column_id": UUID }` |
| `task_completed_in_column` | Завершение задачи в колонке (клик чекбокса) | `{ "column_id": UUID }` |
| `task_field_changed` | Изменение поля задачи (приоритет, срок, статус) | `{ "field": string, "to"?: any }` |
| `column_sla_timer` | Превышение времени нахождения в колонке | `{ "column_id": UUID, "duration_hours": number }` |
| `scheduled_sweep` | Ежедневный запуск по расписанию для колонки | `{ "column_id": UUID, "time_utc"?: string }` |

### 2. Условия (`conditions`)

Массив объектов `{ "field": string, "operator": string, "value": any }`. Все условия объединяются логическим оператором **`AND`** (если массив пуст — действие выполняется безусловно).

* **Поля (`field`)**: `priority`, `is_completed`, `due_date`, `time_in_column_hours`, `tags`.
* **Операторы (`operator`)**:
  - `equals` / `not_equals` — строгое равенство/неравенство (для строк, чисел, boolean).
  - `contains` / `not_contains` — вхождение в список тегов или подстроку.
  - `is_empty` / `is_not_empty` — проверка на заполненность дедлайна или тегов.
  - `greater_than` / `less_than` — числовые и временные сравнения (например, `time_in_column_hours > 24`).

### 3. Действия (`actions`)

Массив объектов `{ "type": string, "params": object }`. Выполняются строго последовательно.

| Тип действия (`type`) | Назначение | Параметры (`params`) |
|---|---|---|
| `move_to_column` | Переместить задачу в указанную колонку | `{ "target_column_id": UUID }` |
| `set_priority` | Установить приоритет задачи | `{ "priority": "urgent" \| "high" \| "medium" \| "low" \| "none" }` |
| `add_tag` | Добавить тег к задаче | `{ "tag_id"?: UUID, "name"?: string }` |
| `remove_tag` | Удалить тег с задачи | `{ "tag_id"?: UUID, "name"?: string }` |
| `set_due_date` | Задать или скорректировать срок | `{ "preset": "today" \| "tomorrow" \| "clear", "offset_days"?: number }` |
| `archive_task` | Архивировать задачу | `{}` |

---

## 🔒 3. Безопасность и Валидация (Constraints & Limits)

1. **Лимит действий**: Не более 5 `actions` на одно правило.
2. **Лимит условий**: Не более 5 `conditions` на одно правило.
3. **Cycle Guard**: Не более 3 каскадных срабатываний правил на одно первичное действие пользователя (защита от циклов A $\rightarrow$ B $\rightarrow$ A).
4. **Валидация JSON-схемы**: При сохранении через REST API (`POST /workspaces/{id}/automations`) бэкенд строго валидирует DTO схемы, исключая неизвестные типы триггеров и действий.
5. **Tenant Isolation**: Все `column_id` и `target_column_id` проверяются на принадлежность текущему проекту и воркспейсу пользователя.

