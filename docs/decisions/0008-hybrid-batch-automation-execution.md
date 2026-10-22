# ADR-0008: Гибридная стратегия исполнения автоматизаций: Batch SQL vs General Loop

* **Дата**: `2026-10-22`
* **Статус**: `Accepted`
* **Инициатор / Автор**: `Pavel (Founder) & Lead AI Architect`
* **Затрагиваемые подсистемы**: `backend/internal/application/automation`, `backend/internal/infrastructure/repository`, `backend/internal/domain/automation`

---

## 🎯 1. Контекст и Проблема (Context & Problem Statement)

Фоновый воркер (`Worker`) и обработчики событий (`EventListener`) движка автоматизаций обрабатывают два класса событий принципиально разной природы:

1. **Point-событие** — событие сгенерировано конкретной задачей (пользователь перетащил карточку, нажал чекбокс). Контекст: одна задача, один `TaskID` в payload.

2. **Sweep-событие** — событие не привязано к конкретной задаче, а обходит группу задач по расписанию или таймеру. Примеры: `column_sla_timer`, `scheduled_sweep`, `archive_task by timer`. Контекст: колонка/доска/воркспейс, подходящих задач может быть от 0 до 100 000+.

### Текущая проблема (N+1 запрос)

Текущий `Worker` для sweep-событий работает по паттерну **N+1**:

```
processSLATimerRule(rule):
    tasks := taskReader.ListTasksByColumn(columnID)   ← SELECT *
    for each task:
        evaluator.ExecuteRule(rule, task)
            → taskModifier.ArchiveTask(task.ID)       ← UPDATE × N
```

При 1000 выполненных задач в колонке — 1001 запрос к PostgreSQL за один тик воркера. При нескольких правилах и воркспейсах — кратно хуже.

---

## 🔍 2. Рассмотренные варианты (Considered Options)

### Вариант 1: Текущий — General Loop (Go итерирует задачи, per-task UPDATE)

* **Суть**: `SELECT * FROM tasks WHERE column_id = ?` → Go-цикл → `executeAction(task)` → `UPDATE tasks SET ... WHERE id = ?` × N.
* **Плюсы (+)**:
  - Гибкость: условия из DSL (приоритет, теги, `due_date_status`) вычисляются в Go.
  - Каждый `executeAction` может файрить события в `EventBus` (история, уведомления).
  - Минимум нового кода — работает «из коробки» для любого нового действия.
* **Минусы (-)**:
  - N+1 запросов к БД — критично при больших колонках.
  - Задачи путешествуют по сети туда-обратно без пользы (SELECT → Go → UPDATE).
  - Не масштабируется: при 10k задач в колонке тик воркера занимает секунды.

### Вариант 2: Pure SQL / pg_cron (вся логика в PostgreSQL)

* **Суть**: `pg_cron` в PostgreSQL запускает scheduled job напрямую в БД.
* **Плюсы (+)**:
  - Бэкенд вообще не задействован; максимальная эффективность.
  - Один `UPDATE ... WHERE ...` — атомично и быстро.
* **Минусы (-)**:
  - Требует расширение `pg_cron` — внешняя зависимость, не у всех хостеров доступна.
  - Логика правил «размазана» между Go и SQL — сложно поддерживать.
  - Нет доступа к `EventBus` → нет истории изменений, нет агрегированных событий.
  - Правила хранятся в Go-коде, а выполняются в SQL — нарушение принципа единственного источника истины.

### Вариант 3 (Выбран): Гибридный — Batch SQL Fast Path + General Loop Fallback

* **Суть**: Worker и EventListener определяют тип исполнения по характеристикам правила. Для «простых» sweep-паттернов (одно действие, транслируемые условия) — один батч-запрос. Для сложных правил (несколько действий, сложные условия) — текущий Go-цикл.

---

## 🏆 3. Принятое решение и Обоснование (Decision Outcome)

**Выбран Вариант 3**: Гибридное исполнение автоматизаций с **Batch SQL Fast Path** для sweep-операций с однотипным действием и **General Loop** как fallback для сложных условий и комбинированных действий.

**Обоснование**: Для наиболее частых и потенциально самых «тяжёлых» сценариев (архивация выполненных задач по таймеру, массовое проставление приоритета) батч-запрос даёт O(1) обращений к БД вместо O(N). При этом General Loop сохраняется для гибкости: сложные DSL-условия остаются в Go где есть полный `TaskInfo` контекст.

---

## 📐 4. Спецификация Гибридного Исполнения

### 4.1 Новый интерфейс `BatchTaskExecutor`

Добавляется в `backend/internal/application/automation/ports.go` рядом с `TaskModifier`:

```go
// BatchTaskExecutor выполняет действия над группой задач одним SQL-запросом.
// Используется Worker-ом как "fast path" для sweep-автоматизаций.
type BatchTaskExecutor interface {
    // ArchiveCompletedTasksInColumn помечает is_archived = true для всех
    // выполненных задач колонки, обновлённых раньше olderThanHours назад.
    // Возвращает количество затронутых строк.
    ArchiveCompletedTasksInColumn(
        ctx context.Context,
        workspaceID, columnID uuid.UUID,
        olderThanHours float64,
    ) (int64, error)

    // MoveCompletedTasksToColumn перемещает выполненные задачи из колонки
    // в targetColumnID одним UPDATE.
    MoveCompletedTasksToColumn(
        ctx context.Context,
        workspaceID, sourceColumnID, targetColumnID uuid.UUID,
    ) (int64, error)

    // В будущем: SetPriorityBatch, AddTagBatch, ...
}
```

### 4.2 Критерии выбора Fast Path

Worker вызывает `canUseBatchPath(rule)` — функцию-предикат:

```go
// canUseBatchPath возвращает true если правило можно исполнить батч-запросом.
// Условия Fast Path:
//   1. Ровно одно действие в rule.Actions.
//   2. Тип действия входит в список BatchCapableActions.
//   3. Conditions пустой или содержит только is_completed = true.
func canUseBatchPath(rule *domainAutomation.RuleDefinition) bool {
    if len(rule.Actions) != 1 {
        return false
    }
    if !BatchCapableActions[rule.Actions[0].Type] {
        return false
    }
    for _, c := range rule.Conditions {
        if c.Field != domainAutomation.ConditionFieldIsCompleted {
            return false // сложное условие → General Loop
        }
    }
    return true
}

// BatchCapableActions — действия, для которых реализован Batch SQL.
var BatchCapableActions = map[string]bool{
    domainAutomation.ActionArchiveTask:   true,
    domainAutomation.ActionMoveToColumn:  true,
}
```

### 4.3 Обновлённый поток Worker для sweep-правил

```
Worker.processSLATimerRule(rule):
  │
  ├─ canUseBatchPath(rule) == true:
  │   └─ batchExecutor.ArchiveCompletedTasksInColumn(columnID, hours)
  │       → SQL: UPDATE tasks SET is_archived = true, archived_at = NOW()
  │               WHERE column_id = $1
  │                 AND is_completed = true
  │                 AND updated_at < NOW() - ($2 || ' hours')::interval
  │                 AND is_archived = false
  │       → Публикует 1 агрегированное событие BatchSweepCompleted{count}
  │
  └─ canUseBatchPath(rule) == false:
      └─ ListTasksByColumn() → for each task → ExecuteRule()  ← текущий General Loop
```

### 4.4 EventBus при Batch-исполнении

При батч-операции Fire-per-task отсутствует, но публикуется **агрегированное событие**:

```go
type BatchSweepCompletedEvent struct {
    WorkspaceID uuid.UUID
    ColumnID    uuid.UUID
    Action      string // "archive_task" | "move_to_column"
    AffectedCount int64
}
```

Это событие используется:
- Для принудительного инвалидирования кэша фронтенда (WebSocket или poll).
- Для будущей системы истории изменений (TASK-B06) — одна запись `batch_archived` в `task_activities` вместо N строк.

---

## 📈 5. Последствия и План внедрения (Consequences)

### Положительные последствия:
- Архивация 1000 задач = 1 SQL-запрос вместо 1001.
- Тик воркера при большом числе задач занимает миллисекунды, а не секунды.
- Интерфейс `BatchTaskExecutor` расширяем — новые батч-действия добавляются без изменения Worker-логики.
- General Loop сохраняется в неизменном виде — нет регрессий для существующих правил.

### Компромиссы:
- `BatchTaskExecutor` — дополнительный интерфейс в `ports.go`, требует реализации в `infrastructure/repository`.
- При batch-исполнении нет доступа к индивидуальным `TaskID` — история изменений пишется агрегированно.
- Условия batch-пути ограничены: только `is_completed = true`. Более сложные условия по-прежнему требуют General Loop.

### Безопасность:
- Все batch-запросы фильтруются по `workspace_id` — Tenant Isolation гарантирован на уровне SQL.
- `olderThanHours` параметризован через `$2` — SQL injection невозможен.

### План внедрения:
1. Добавить интерфейс `BatchTaskExecutor` в `backend/internal/application/automation/ports.go`.
2. Реализовать методы в `backend/internal/infrastructure/repository/task_repo.go` (или выделить `automation_batch_repo.go`).
3. Добавить `canUseBatchPath()` и обновить `Worker.processSLATimerRule()` и `processScheduledSweepRule()`.
4. Добавить `BatchSweepCompletedEvent` в `backend/internal/domain/events/`.
5. Обновить строку ADR-0008 в индексной таблице `docs/decisions/README.md`.
6. Покрыть unit-тестами: `canUseBatchPath` (table-driven), mock `BatchTaskExecutor`.

