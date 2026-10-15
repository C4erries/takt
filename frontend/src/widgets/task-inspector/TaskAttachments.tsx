import React, { useRef, useState } from 'react';
import {
  Paperclip,
  Download,
  Trash2,
  FileText,
  Loader2,
  Plus,
} from 'lucide-react';
import {
  useTaskAttachmentsQuery,
  useUploadAttachmentMutation,
  useDeleteAttachmentMutation,
} from '@/entities/attachment';
import type { Attachment } from '@/entities/attachment';
import { Button } from '@/shared/components/ui';

export interface TaskAttachmentsProps {
  workspaceId?: string;
  taskId: string;
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} Б`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} КБ`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} МБ`;
}

export const TaskAttachments: React.FC<TaskAttachmentsProps> = ({
  workspaceId,
  taskId,
}) => {
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const [isDragOver, setIsDragOver] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);

  const {
    data: attachments = [],
    isLoading,
    isError,
  } = useTaskAttachmentsQuery(workspaceId, taskId);

  const uploadMutation = useUploadAttachmentMutation(workspaceId, taskId);
  const deleteMutation = useDeleteAttachmentMutation(workspaceId, taskId);

  const handleFiles = async (files: FileList | File[]) => {
    setUploadError(null);
    const fileList = Array.from(files);
    for (const file of fileList) {
      try {
        await uploadMutation.mutateAsync(file);
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Ошибка при загрузке файла';
        setUploadError(message);
        break;
      }
    }
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      handleFiles(e.target.files);
      e.target.value = '';
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(false);
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      handleFiles(e.dataTransfer.files);
    }
  };

  const imageAttachments = attachments.filter((a) =>
    a.contentType.startsWith('image/')
  );
  const documentAttachments = attachments.filter(
    (a) => !a.contentType.startsWith('image/')
  );

  return (
    <div
      onDragOver={handleDragOver}
      onDragEnter={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      className={`space-y-2 pt-2 border-t border-subtle/50 transition-colors rounded-lg ${
        isDragOver ? 'bg-accent-surface/30 ring-1 ring-accent-solid/50 p-2' : ''
      }`}
    >
      <input
        type="file"
        ref={fileInputRef}
        onChange={handleFileInputChange}
        className="hidden"
        multiple
      />

      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-1.5 text-[10px] font-mono uppercase tracking-wider text-muted">
          <Paperclip className="w-3 h-3" />
          <span>Вложения {attachments.length > 0 && `(${attachments.length})`}</span>
        </div>

        <Button
          type="button"
          size="sm"
          variant="ghost"
          onClick={() => fileInputRef.current?.click()}
          disabled={uploadMutation.isPending}
          className="h-6 px-2 text-[11px] gap-1 text-muted hover:text-primary"
        >
          {uploadMutation.isPending ? (
            <Loader2 className="w-3 h-3 animate-spin" />
          ) : (
            <Plus className="w-3 h-3" />
          )}
          <span>Прикрепить</span>
        </Button>
      </div>

      {/* Uploading progress indicator */}
      {uploadMutation.isPending && (
        <div className="flex items-center gap-2 p-2 bg-canvas/40 border border-subtle/50 rounded-lg text-xs text-primary animate-pulse">
          <Loader2 className="w-3.5 h-3.5 animate-spin text-accent-solid" />
          <span className="font-sans text-[11px]">Загрузка файла в хранилище...</span>
        </div>
      )}

      {/* Error banner */}
      {uploadError && (
        <div className="p-2 bg-canvas border border-subtle rounded-lg text-[11px] text-primary flex items-center justify-between">
          <span className="truncate">{uploadError}</span>
          <button
            type="button"
            onClick={() => setUploadError(null)}
            className="text-muted hover:text-primary ml-2 font-mono"
          >
            ×
          </button>
        </div>
      )}

      {/* Loading state */}
      {isLoading && attachments.length === 0 && (
        <div className="space-y-1.5 py-1">
          <div className="h-10 bg-canvas/40 border border-subtle/40 rounded-lg animate-pulse" />
          <div className="h-10 bg-canvas/40 border border-subtle/40 rounded-lg animate-pulse" />
        </div>
      )}

      {/* Error fetching state */}
      {isError && (
        <div className="p-2 text-center text-xs text-muted">
          Не удалось загрузить список вложений
        </div>
      )}

      {/* Empty State */}
      {!isLoading && attachments.length === 0 && !uploadMutation.isPending && (
        <div
          onClick={() => fileInputRef.current?.click()}
          className="border border-dashed border-subtle/60 rounded-lg p-3 text-center cursor-pointer hover:bg-canvas/30 transition-colors"
        >
          <p className="text-[11px] text-muted font-sans">
            Перетащите файлы сюда или нажмите «Прикрепить»
          </p>
        </div>
      )}

      {/* Gallery for images */}
      {imageAttachments.length > 0 && (
        <div className="grid grid-cols-3 gap-2 pt-1">
          {imageAttachments.map((att: Attachment) => (
            <div
              key={att.id}
              className="group relative aspect-square rounded-lg border border-subtle overflow-hidden bg-canvas/40"
            >
              <img
                src={att.contentUrl}
                alt={att.fileName}
                className="w-full h-full object-cover cursor-pointer hover:scale-105 transition-transform"
                onClick={() => window.open(att.contentUrl, '_blank')}
                title={att.fileName}
                loading="lazy"
              />
              <div className="absolute inset-0 bg-primary/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-1.5 p-1">
                <a
                  href={`${att.contentUrl}?download=true`}
                  download={att.fileName}
                  className="p-1 rounded bg-surface/90 text-primary hover:bg-surface shadow-washi-sm transition-colors"
                  title="Скачать"
                  onClick={(e) => e.stopPropagation()}
                >
                  <Download className="w-3 h-3" />
                </a>
                <button
                  type="button"
                  onClick={(e) => {
                    e.stopPropagation();
                    deleteMutation.mutate(att.id);
                  }}
                  className="p-1 rounded bg-surface/90 text-primary hover:bg-surface shadow-washi-sm transition-colors"
                  title="Удалить"
                  disabled={deleteMutation.isPending}
                >
                  <Trash2 className="w-3 h-3" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* List for other files */}
      {documentAttachments.length > 0 && (
        <div className="space-y-1 pt-1">
          {documentAttachments.map((att: Attachment) => (
            <div
              key={att.id}
              className="flex items-center justify-between p-2 rounded-lg bg-canvas/40 border border-subtle/50 text-xs hover:bg-canvas/60 transition-colors"
            >
              <div className="flex items-center gap-2 min-w-0 mr-2">
                <FileText className="w-4 h-4 text-muted shrink-0" />
                <div className="min-w-0">
                  <p className="font-sans text-primary text-[11px] truncate leading-tight">
                    {att.fileName}
                  </p>
                  <span className="font-mono text-[9px] text-muted">
                    {formatFileSize(att.fileSize)}
                  </span>
                </div>
              </div>

              <div className="flex items-center gap-1 shrink-0">
                <a
                  href={`${att.contentUrl}?download=true`}
                  download={att.fileName}
                  className="p-1 rounded text-muted hover:text-primary hover:bg-canvas transition-colors"
                  title="Скачать"
                >
                  <Download className="w-3.5 h-3.5" />
                </a>
                <button
                  type="button"
                  onClick={() => deleteMutation.mutate(att.id)}
                  disabled={deleteMutation.isPending}
                  className="p-1 rounded text-muted hover:text-primary hover:bg-canvas transition-colors"
                  title="Удалить"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
