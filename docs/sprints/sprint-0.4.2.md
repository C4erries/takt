# Спринт 0.4.2: Собственное S3-хранилище (MinIO/S3), Вложения и Медиа к задачам

* **Даты проведения**: 2026-10-15 — 2026-10-15
* **Статус**: ✅ Завершен (100% Definition of Done)
* **Связанные документы**: [docs/CURRENT_SPRINT.md](../CURRENT_SPRINT.md), [docs/decisions/0003-s3-storage-and-attachments-architecture.md](../decisions/0003-s3-storage-and-attachments-architecture.md), [docs/PROJECT_CONTEXT.md](../PROJECT_CONTEXT.md)

---

## 🧠 1. Зафиксированные микро-решения спринта (Preserved Micro-Decisions)

1. **Архитектура загрузки через бэкенд (Multipart Streaming Proxy)**:
   - Эндпоинт `POST /api/v1/workspaces/{workspace_id}/tasks/{task_id}/attachments` (`multipart/form-data`, поле `file`).
   - Бэкенд валидирует права доступа пользователя (членство в воркспейсе), размер и тип файла, стримит данные в MinIO/S3 через официальный AWS SDK v2 (`PutObject`) и создает запись в таблице `attachments`.
   - Решение исключает сложности с CORS и маршрутизацией MinIO-доменов в Docker Compose, обеспечивая надежный централизованный контроль безопасности.
2. **Перманентный Content URL (Защита от протухания ссылок в Markdown)**:
   - Просмотр и скачивание осуществляются по стабильному URL:
     `GET /api/v1/workspaces/{workspace_id}/attachments/{attachment_id}/content` (с заголовками `Content-Type`, `Cache-Control: public, max-age=86400`, `Content-Disposition: inline; filename="..."`).
   - Для скачивания как вложение поддерживается query-параметр `?download=true` (`Content-Disposition: attachment`).
   - Ссылки на картинки в заметках вида `![Скриншот](/api/v1/workspaces/{wid}/attachments/{aid}/content)` никогда не ломаются со временем и не требуют периодической перегенерации подписей.
3. **Лимиты, типы файлов и конфигурация**:
   - Максимальный размер файла настраивается через конфигурацию `MAX_ATTACHMENT_SIZE_BYTES` (по умолчанию 50 МБ = `52428800`).
   - Поддерживаются любые типы файлов (изображения `image/*`, документы `application/pdf`, текстовые файлы `text/*`, архивы `application/zip` и др.).
   - Для картинок отображается превью, для остальных файлов — документ-карточка с форматированным размером.
4. **Жизненный цикл и хранение в S3**:
   - Ключ S3: `workspaces/{workspace_id}/tasks/{task_id}/{uuid}_{sanitized_file_name}`.
   - При удалении задачи метаданные в БД удаляются каскадно (`ON DELETE CASCADE`), а в S3 файлы удаляются через `DeleteTaskAttachments`. При удалении конкретного файла (`DELETE /attachments/{id}`) объект удаляется из S3 и запись удаляется из БД.
5. **Совмещенный UX (Вставка по Ctrl+V + Галерея вложений)**:
   - **Вставка из буфера (`Ctrl+V`)**: В поле Markdown-заметок перехватывается `onPaste`. Если в буфере изображение (`image/*`), оно автоматически загружается на сервер, вставляется Markdown-тег `![имя_файла](content_url)`, а файл добавляется в список вложений задачи.
   - **Секция «Вложения» в TaskInspector**: Под Markdown-заметками размещается блок с галереей прикрепленных файлов (превью картинок, карточки файлов, кнопка удаления, кнопка «+ Прикрепить файл» / скрытый файловый инпут, drag-and-drop).
6. **Инфраструктура Docker**:
   - В `docker-compose.yml` добавлены сервисы `minio` (образ `minio/minio:latest`, порты 9000/9001) и init-контейнер `minio_createbucket` (на базе `minio/mc`) для автоматического создания бакета `takt-attachments`.

---

## 🎯 2. Выполненные задачи и Затронутые подпакеты

1. **Миграции БД и SQL**:
   - `backend/migrations/000007_attachments.up.sql` и `.down.sql`: таблица `attachments`, внешние ключи `ON DELETE CASCADE` к `tasks` и `workspaces`, индексы `idx_attachments_task_id`, `idx_attachments_workspace_id`.
   - `backend/queries/attachments.sql`: запросы `CreateAttachment`, `GetAttachmentByID`, `ListAttachmentsByTaskID`, `DeleteAttachment`, `DeleteAttachmentsByTaskID`.
2. **OpenAPI Контракты и Кодогенерация**:
   - `api/openapi/schemas/attachment.yaml`: схема `Attachment` (`id`, `workspaceId`, `taskId`, `fileName`, `fileSize`, `contentType`, `contentUrl`, `createdAt`).
   - `api/openapi/paths/attachments.yaml`: эндпоинты загрузки, листинга, контента и удаления вложений.
   - Сгенерирован обновленный `ServerInterface` в `internal/infrastructure/http/generated/api.gen.go`.
3. **Чистый доменный слой (`internal/domain/attachment`)**:
   - `attachment.go`: модель `Attachment`, конструктор `NewAttachment`, валидация непустого имени, проверка лимита размера (`maxSize`), санитайзинг имен файлов (`SanitizeFileName`) с защитой от Path Traversal, генерация S3-ключа.
   - Табличные unit-тесты инвариантов в `attachment_test.go`.
4. **Инфраструктура S3-хранилища (`internal/infrastructure/storage/s3`)**:
   - `s3.go`: адаптер на официальном `aws-sdk-go-v2/service/s3` (`PutObject`, `GetObject`, `DeleteObject`).
   - `s3_test.go`: юнит-тесты адаптера с моками S3-клиента.
5. **Use Cases и Сервисы (`internal/application/attachment`)**:
   - `ports.go`: интерфейсы `AttachmentRepository`, `StorageService`, `TaskReader`.
   - `service.go`: сценарии `UploadAttachment`, `ListTaskAttachments`, `GetAttachmentContent`, `DeleteAttachment`.
   - Тесты сценариев в `service_test.go` и `service_content_test.go`.
6. **HTTP-адаптер и Репозиторий (`internal/infrastructure/`)**:
   - `repository/attachment_repository.go`: реализация репозитория на базе `sqlc`.
   - `http/attachments_handler.go`: Echo-хендлеры загрузки multipart, отдача стриминга с `Cache-Control` и `Content-Disposition`, удаление с проверкой прав Tenant Isolation.
   - Тесты в `http/attachments_handler_test.go`.
   - Консолидация хендлеров для строгого соблюдения `<= 7 рабочих файлов` в пакете `http/`.
   - Связывание зависимостей в `container/` (`repos.go`, `services.go`, `app.go`).
7. **Фронтенд (`frontend/src/`)**:
   - `shared/api/`: типы `Attachment` и методы API `uploadTaskAttachment`, `listTaskAttachments`, `deleteAttachment`.
   - `entities/attachment/`: React Query хуки `useTaskAttachmentsQuery`, `useUploadAttachmentMutation`, `useDeleteAttachmentMutation`.
   - `widgets/task-inspector/TaskAttachments.tsx`: галерея вложений с превью картинок, карточками файлов, кнопкой загрузки, скачиванием, удалением и Drag-and-drop.
   - `widgets/task-inspector/MarkdownNotes.tsx`: перехват `onPaste` для скриншотов из буфера `Ctrl+V`, автозагрузка и вставка `![name](contentUrl)`, рендеринг превью картинок в режиме просмотра.
8. **E2E-тестирование (`tests/e2e/test_attachments.py`)**:
   - 5 новых сквозных тестов полного цикла: загрузка PNG через multipart, проверка полей схемы `Attachment`, листинг, отдача бинарного контента (inline и attachment), удаление (204 -> 404), каскадное удаление при удалении задачи, негативные проверки Tenant Isolation (403/404 для чужого воркспейса/задачи, 401 для неавторизованных) и валидация 0-байтовых файлов.
   - Итог: **36/36 PASS (100%)**.

---

## 📊 3. Метрики качества и Приемочный контроль

* **`make check-arch`**: `PASS (0 ошибок)` по всем 5 проверкам архитектурного линтера (лимит строк < 400, файлов в пакетах <= 7, чистый домен, отсутствие сырых Tailwind-цветов, строгие 3 подпакета `internal/`).
* **`make test`**: `100% PASS` с флагом `-race` во всех 12 пакетах бэкенда.
* **`make test-e2e`**: `36/36 PASS (100%)` в Pytest E2E сьюте (5 новых тестов S3 и вложений).
* **`cd frontend && npm run build`**: `0 ошибок, 0 предупреждений` (Vite build собран за 3.73s).

