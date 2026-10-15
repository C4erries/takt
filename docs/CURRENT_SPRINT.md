# Текущий спринт / Оперативный центр управления (`CURRENT_SPRINT.md`)

> Этот файл — **единственная точка правды** о том, какие решения приняты по текущей задаче и что команда и AI-агенты делают прямо сейчас.
> Работа над каждым спринтом строго делится на **Фазу А (Штаб: Принятие решений)** и **Фазу Б (Конвейер: Реализация и Верификация)**.
> Завершенные спринты архивируются в каталог `docs/sprints/sprint-X.Y.Z.md` с записью в реестр `docs/sprints/README.md`.

## Текущий спринт: Спринт 0.4.2 (Фаза 3: Собственное S3-хранилище MinIO/S3, Вложения и Медиа к задачам) ✅ ЗАВЕРШЕН (Готов к архивации)

---

## 🧠 1. Аналитика и Зафиксированные микро-решения спринта (Режим «Штаб»)

* **Бизнес-цель**: Реализовать файловое хранилище на базе S3 (MinIO в локальном dev-окружении, любой S3-провайдер в production), позволяющее прикреплять вложения (файлы, документы) к задачам, отображать их в блоке «Вложения» в TaskInspector и мгновенно вставлять скриншоты из буфера обмена (`Ctrl+V`) прямо в текст Markdown-заметок с отображением превью.
* **Зафиксированные архитектурные и продуктовые микро-решения**:
  1. **Архитектура загрузки (Бэкенд-прокси / Multipart)**:
     - Загрузка выполняется через бэкенд: `POST /api/v1/workspaces/{workspace_id}/tasks/{task_id}/attachments` (`multipart/form-data`, поле `file`).
     - Бэкенд валидирует права доступа пользователя (членство в воркспейсе), размер и тип файла, стримит данные в MinIO/S3 через AWS SDK v2 (`PutObject`) и создает запись в таблице `attachments`.
  2. **Перманентный Content URL (Защита от протухания ссылок в Markdown)**:
     - Просмотр и скачивание осуществляются по стабильному URL:
       `GET /api/v1/workspaces/{workspace_id}/attachments/{attachment_id}/content` (с заголовками `Content-Type`, `Cache-Control: public, max-age=86400`, `Content-Disposition: inline; filename="..."`).
     - Для скачивания как вложение поддерживается query-параметр `?download=true` (`Content-Disposition: attachment`).
  3. **Лимиты, типы файлов и конфигурация**:
     - Лимит размера настраивается через конфигурацию: `MAX_ATTACHMENT_SIZE_BYTES` (по умолчанию **50 МБ** = `52428800`).
     - Поддерживаются любые типы файлов (изображения `image/*`, документы `application/pdf`, текстовые файлы `text/*`, архивы `application/zip` и др.).
  4. **Жизненный цикл и хранение в S3**:
     - Ключ S3: `workspaces/{workspace_id}/tasks/{task_id}/{uuid}_{sanitized_file_name}`.
     - При удалении задачи метаданные в БД удаляются каскадно (`ON DELETE CASCADE`), а в S3 файлы удаляются через сценарий `DeleteTaskAttachments`. При удалении конкретного файла (`DELETE /attachments/{id}`) объект удаляется из S3 и запись удаляется из БД.
  5. **UX и интерфейс (Совмещенный режим)**:
     - **Вставка из буфера (`Ctrl+V`)**: В поле Markdown-заметок перехватывается `onPaste`. Если в буфере изображение (`image/*`), оно загружается на сервер, вставляется Markdown-тег `![имя_файла](content_url)`, а файл также автоматически появляется в списке вложений задачи.
     - **Секция «Вложения» в TaskInspector**: Под Markdown-заметками размещается блок с галереей прикрепленных файлов (превью для картинок, карточки для файлов, кнопка удаления, кнопка «+ Прикрепить файл» / скрытый файловый инпут, drag-and-drop).
  6. **Инфраструктура Docker**:
     - В `docker-compose.yml` добавлены сервисы `minio` и `minio_createbucket` (на базе `minio/mc`) для автосоздания бакета `takt-attachments`.

---

## 📋 Таблица задач спринта 0.4.2

| # | Задача | Статус | Приоритет | Ответственный | Заметки и контракты |
|---|--------|--------|-----------|---------------|---------------------|
| 1 | **[Decisions]**: Блиц-опрос и фиксация архитектурных микро-решений S3 | ✅ Done | Критический | `lead_architect` & User | Решения: бэкенд-прокси multipart, Content URL, 50MB, Ctrl+V + галерея, ADR-0003 |
| 2 | **[Backend Step 1]**: Миграция 0007, sqlc-запросы, OpenAPI схемы, генерация кода, чистый домен `internal/domain/attachment` | ✅ Done | Критический | `backend_developer` | `attachments.sql`, `schemas/attachment.yaml`, `paths/attachments.yaml`, `make generate` |
| 3 | **[Gatekeeper Checkpoint 1]**: Проверка контрактов и чистоты домена | ✅ Done | Критический | `lead_architect` | `make check-arch && make generate && make test` (PASS) |
| 4 | **[Backend Step 2]**: MinIO в compose, S3-адаптер (`aws-sdk-go-v2`), Use Cases `application/attachment`, Echo-хендлеры, связывание в `container` | ✅ Done | Высокий | `backend_developer` | `infrastructure/storage/s3`, `attachments_handler.go`, unit & repo tests |
| 5 | **[Frontend]**: Интеграция вложений в TaskInspector, Drag-and-drop, загрузка по `Ctrl+V` в MarkdownNotes, превью | ✅ Done | Высокий | `frontend_developer` | `entities/attachment`, `widgets/task-inspector`, Washi & Matcha UI-Kit |
| 6 | **[QA & Verification]**: Pytest E2E сквозные тесты S3 загрузки, скачивания контента, каскадного удаления и изоляции | ✅ Done | Критический | `qa_engineer` | `tests/e2e/test_attachments.py`, `make check-arch`, `make test`, `make test-e2e` (36/36 PASS) |

---

## 🎯 Definition of Done (DoD)

- [x] Сервис MinIO добавлен в `docker-compose.yml` с автоматическим созданием бакета `takt-attachments`.
- [x] Конфигурация бэкенда поддерживает S3 (`S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_USE_SSL`, `MAX_ATTACHMENT_SIZE_BYTES=52428800`).
- [x] Создана миграция `000007_attachments.up.sql` с внешними ключами и индексами.
- [x] Описаны OpenAPI контракты (`paths/attachments.yaml`, `schemas/attachment.yaml`), кодогенерация `make generate` проходит без ошибок.
- [x] Чистый домен `internal/domain/attachment` покрыт табличными тестами инвариантов.
- [x] Реализован адаптер хранилища S3 в `internal/infrastructure/storage/s3` с использованием `aws-sdk-go-v2`.
- [x] Реализованы сценарии: загрузка вложения (`multipart`), получение списка, отдача бинарного контента с авторизацией, удаление.
- [x] Во фронтенде в [TaskInspector](file:///home/pavel/projects/takt/frontend/src/widgets/task-inspector/TaskInspector.tsx) добавлена галерея вложений и перехват `Ctrl+V` для вставки картинок в [MarkdownNotes](file:///home/pavel/projects/takt/frontend/src/widgets/task-inspector/MarkdownNotes.tsx).
- [x] `make check-arch` — 0 ошибок по всем 5 проверкам.
- [x] `make test` — 100% PASS с флагом `-race`.
- [x] `cd frontend && npm run build` — 0 ошибок сборки.
- [x] Pytest E2E-тесты проверяют полный жизненный цикл вложений и Tenant Isolation (36/36 PASS).

---

## Лог изменений спринта

| Дата | Что изменилось |
|------|---------------|
| 2026-10-15 | **Спринт 0.4.2 успешно завершен**: Зафиксирован [ADR-0003](file:///home/pavel/projects/takt/docs/decisions/0003-s3-storage-and-attachments-architecture.md). Интегрирован MinIO в `docker-compose.yml`, создана миграция `000007_attachments.up.sql`, генерация sqlc и OpenAPI контрактов, чистый домен `internal/domain/attachment`, S3 storage adapter на `aws-sdk-go-v2`, Use Cases в `internal/application/attachment`, HTTP-хендлеры в `internal/infrastructure/http`. На фронтенде реализована галерея вложений [TaskAttachments.tsx](file:///home/pavel/projects/takt/frontend/src/widgets/task-inspector/TaskAttachments.tsx) с Drag-and-drop и автозагрузка скриншотов по `Ctrl+V` в [MarkdownNotes.tsx](file:///home/pavel/projects/takt/frontend/src/widgets/task-inspector/MarkdownNotes.tsx). Полный E2E тест-сьют на Pytest проходит со 100% успехом (36/36 тестов). |
