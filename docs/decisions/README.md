# Архитектурные решения (Architecture Decision Records, ADR)

> Этот каталог содержит **хронологию всех ключевых глобальных технических и архитектурных решений**, принятых в проекте **Takt**.
>
> **Правило экономии контекста для AI-агентов (Шаг 0)**:
> Агентам **ЗАПРЕЩЕНО** читать все файлы в этой папке подряд. На Шаге 0 агент читает **только Индексный реестр ниже** и открывает полный файл конкретного ADR только тогда, когда задача напрямую затрагивает эту подсистему.

---

## 🗂️ Индексный реестр архитектурных решений (ADR Index Registry)

| ID | Название и ссылка на файл | Дата | Статус | Краткая суть решения (в 1 предложение) |
|---|---|---|---|---|
| **ADR-0001** | [0001-initial-architecture-and-stack.md](0001-initial-architecture-and-stack.md) | 2026-10-14 | `Accepted` | Выбор стека Go 1.23+ (Echo, sqlc, pgx v5, Transactor) + React/TS + PostgreSQL 16 + Multi-file OpenAPI + Pytest E2E + Nginx |
| **ADR-0002** | [0002-jwt-auth-and-tenant-isolation.md](0002-jwt-auth-and-tenant-isolation.md) | 2026-10-14 | `Accepted` | Гибридный JWT (HTTP-only Cookie + Bearer Fallback), автосоздание Personal Workspace и Tenant Isolation |
| **ADR-0003** | [0003-s3-storage-and-attachments-architecture.md](0003-s3-storage-and-attachments-architecture.md) | 2026-10-15 | `Accepted` | S3-хранилище (MinIO dev), бэкенд-прокси загрузки multipart, перманентные Content URLs для Markdown и лимит 50 МБ |
| **ADR-0004** | [0004-low-code-automations-dsl.md](0004-low-code-automations-dsl.md) | 2026-10-18 | `Accepted` | Спецификация Low-Code DSL правил автоматизации (Trigger-Condition-Action) в JSONB, типизация и валидация |
| **ADR-0005** | [0005-frontend-motion-and-kinetics-engine.md](0005-frontend-motion-and-kinetics-engine.md) | 2026-10-19 | `Accepted` | Движок кинетики на motion/react 12+, пружинная физика, пресеты, хореография чекбокса и A11y Dampening |
| **ADR-0006** | [0006-editorial-tech-themes-system.md](0006-editorial-tech-themes-system.md) | 2026-10-19 | `Accepted` | Система 12 тем Editorial Tech, переключатель палитр, двухуровневый фон и FSD-рефакторинг |
| **ADR-0007** | [0007-animated-themes-and-interactive-ambient-engine.md](0007-animated-themes-and-interactive-ambient-engine.md) | 2026-10-20 | `Accepted` | Анимированные и интерактивные темы (Snow, Tsukimi, Celestial), Canvas 2D Engine и контроль интенсивности |
| **ADR-0008** | [0008-hybrid-batch-automation-execution.md](0008-hybrid-batch-automation-execution.md) | 2026-10-22 | `Accepted` | Гибридное исполнение sweep-автоматизаций: Batch SQL Fast Path (1 запрос) для простых действий + General Loop fallback для сложных DSL-условий |

---

## 📌 Когда нужно создавать глобальный ADR, а когда — Микро-решение в спринте?

Чтобы не разводить бюрократию, мы разделяем решения на два уровня:
1. **Локальные решения фичи (Micro-Decisions)** -> записываются прямо в `docs/CURRENT_SPRINT.md` (а затем уходят в архив спринта `docs/sprints/`). Сюда относятся правила валидации конкретной формы, поведение кнопки, локальные лимиты.
2. **Глобальные архитектурные решения (ADR в этом каталоге)** -> создаются **только** если решение влияет на весь проект вдолгую:
   * Затрагивает общую структуру базы данных, модель хранения или мультитенантность;
   * Меняет или добавляет ключевые библиотеки, фреймворки или внешние сервисы;
   * Определяет протоколы взаимодействия (REST, схема авторизации, стратегия кэширования и рейт-лимитов);
   * Устанавливает или меняет общепроектные стандарты архитектуры или дизайн-системы.

---

## 📝 Именование файлов и Статусы ADR

Файлы нумеруются последовательно четырьмя цифрами по шаблону `template.md`:
`XXXX-kebab-case-title.md`

### Статусы ADR:
* **Proposed** (Предложено): находится на обсуждении в режиме «Штаб» с Product Owner / CTO.
* **Accepted** (Принято): утверждено и обязательно к соблюдению всеми агентами и разработчиками.
* **Rejected** (Отклонено): рассмотрено, но отклонено (с сохранением файла для фиксации причин, почему так делать не надо).
* **Superseded by [ADR-YYYY]** (Заменено): устарело при эволюции/пивоте проекта в пользу более нового решения `ADR-YYYY`.
