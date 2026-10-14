# 🎋 Takt

> **Скоростной таск-трекер и персональная система продуктивности с канбан-досками и фоновыми автоматизациями рабочих процессов.**  
> Выполнен в эстетике **Editorial Tech / Humanist Minimalism** (японская канцелярия Midori MD / Hobonichi + Claude).

---

## 🧭 Документация и Контекстная ОС

Проект управляется через строгую документационную систему (`docs/`):

* 🧭 [AGENTS.md](AGENTS.md) — Главная инструкция для AI-агентов (правила «Шага 0», режим «Штаб», лимиты файлов).
* 🎯 [PROJECT_CONTEXT.md](docs/PROJECT_CONTEXT.md) — Архитектурный и продуктовый паспорт проекта, ER-диаграмма и Карта пакетов.
* 📏 [Инженерные конвенции](docs/conventions/README.md) ([Backend](docs/conventions/backend.md), [Frontend](docs/conventions/frontend.md)) — правила Clean Architecture, Дизайн-паспорт («Soft Editorial GTD»), тестирование (testify, mockery v3).
* 🗺️ [ROADMAP.md](docs/ROADMAP.md) — Стратегическая дорожная карта развития по фазам (Фаза 0 – Фаза 4).
* ⚡ [CURRENT_SPRINT.md](docs/CURRENT_SPRINT.md) — Оперативный центр управления текущим спринтом (Фаза 0: Инженерный фундамент).
* 🏛️ [ADR-0001: Архитектурный стек](docs/decisions/0001-initial-architecture-and-stack.md) — Обоснование выбора Go 1.23+ Echo, sqlc, React, OpenAPI и Pytest E2E.

---

## 🛠️ Технологический стек

* **Backend**: Go 1.23+, роутер Echo v4, pgx/v5, sqlc, транзактор через Context, EventBus.
* **Frontend**: React 18/19, TypeScript, Vite, Tailwind CSS, TanStack Query, Zustand, `@dnd-kit/core`.
* **Дизайн**: Newsreader (Serif) + Geist Sans + JetBrains Mono, палитра Washi (`#F7F5EE`) и Matcha (`#566B44`).
* **API**: OpenAPI 3.0 мультифайловый (`api/openapi/*`), бандлер `@redocly/cli`, кодогенерация `oapi-codegen`.
* **Инфраструктура**: Docker Compose, PostgreSQL 16, Nginx (reverse proxy), Makefile.
* **Тестирование**: Go Unit (`go test -race`), Pytest E2E в `tests/` (`httpx`).
