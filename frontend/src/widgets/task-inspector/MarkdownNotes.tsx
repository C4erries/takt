import React, { useState, useEffect, useRef } from 'react';
import { Loader2 } from 'lucide-react';

export interface MarkdownNotesProps {
  initialValue: string;
  onSave: (notes: string) => void;
  onUploadImage?: (file: File) => Promise<string | null>;
}

export const MarkdownNotes: React.FC<MarkdownNotesProps> = ({
  initialValue,
  onSave,
  onUploadImage,
}) => {
  const [isEditing, setIsEditing] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [value, setValue] = useState(initialValue);
  const textareaRef = useRef<HTMLTextAreaElement | null>(null);
  const debounceRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    setValue(initialValue);
    setIsEditing(false);
  }, [initialValue]);

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const next = e.target.value;
    setValue(next);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      onSave(next);
    }, 400);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Escape') {
      e.stopPropagation();
      setIsEditing(false);
      onSave(value);
    }
  };

  const handlePaste = async (e: React.ClipboardEvent<HTMLTextAreaElement>) => {
    if (!onUploadImage) return;

    const files = e.clipboardData?.files;
    if (!files || files.length === 0) return;

    const imageFile = Array.from(files).find((f) => f.type.startsWith('image/'));
    if (!imageFile) return;

    e.preventDefault();

    const textarea = textareaRef.current;
    const start = textarea?.selectionStart ?? value.length;
    const end = textarea?.selectionEnd ?? value.length;

    setIsUploading(true);
    try {
      const contentUrl = await onUploadImage(imageFile);
      if (contentUrl) {
        const altText = imageFile.name || 'image';
        const markdownImage = `![${altText}](${contentUrl})`;
        const newValue = value.slice(0, start) + markdownImage + value.slice(end);
        setValue(newValue);
        onSave(newValue);

        setTimeout(() => {
          if (textarea) {
            const newCursorPos = start + markdownImage.length;
            textarea.setSelectionRange(newCursorPos, newCursorPos);
            textarea.focus();
          }
        }, 0);
      }
    } catch {
      // Игнорируем ошибку загрузки, она логируется на верхнем уровне
    } finally {
      setIsUploading(false);
    }
  };

  const renderMarkdownPreview = (text: string) => {
    if (!text) {
      return <span className="text-muted italic">Добавить заметку в Markdown...</span>;
    }

    const regex = /!\[([^\]]*)\]\(([^)]+)\)/g;
    const elements: React.ReactNode[] = [];
    let lastIndex = 0;
    let match: RegExpExecArray | null;

    while ((match = regex.exec(text)) !== null) {
      const [fullMatch, alt, url] = match;
      const matchIndex = match.index;

      if (matchIndex > lastIndex) {
        elements.push(
          <span key={`text-${lastIndex}`}>{text.substring(lastIndex, matchIndex)}</span>
        );
      }

      elements.push(
        <span key={`img-${matchIndex}`} className="block my-2">
          <img
            src={url}
            alt={alt || 'вложение'}
            className="max-h-60 max-w-full rounded-lg border border-subtle object-contain bg-canvas/30 shadow-washi-sm hover:opacity-95 transition-opacity"
            loading="lazy"
            onClick={(e) => {
              e.stopPropagation();
              window.open(url, '_blank');
            }}
            title={alt ? `${alt} (нажмите для открытия)` : 'Нажмите для открытия оригинала'}
          />
          {alt && alt !== 'image' && (
            <span className="block text-[10px] text-muted font-mono mt-0.5 truncate">{alt}</span>
          )}
        </span>
      );

      lastIndex = matchIndex + fullMatch.length;
    }

    if (lastIndex < text.length) {
      elements.push(<span key={`text-${lastIndex}`}>{text.substring(lastIndex)}</span>);
    }

    return <>{elements}</>;
  };

  return (
    <div className="space-y-1.5 pt-2 border-t border-subtle/50">
      <div className="flex items-center justify-between text-[10px] font-mono uppercase tracking-wider text-muted">
        <span>Заметки (Markdown)</span>
        <span className="text-[9px] lowercase text-muted flex items-center gap-1.5">
          {isUploading && (
            <span className="text-accent-solid flex items-center gap-1 normal-case font-sans">
              <Loader2 className="w-3 h-3 animate-spin" />
              Загрузка изображения...
            </span>
          )}
          <span>{isEditing ? 'esc для выхода' : 'клик для правок'}</span>
        </span>
      </div>

      {isEditing ? (
        <textarea
          ref={textareaRef}
          autoFocus
          rows={5}
          value={value}
          onChange={handleChange}
          onPaste={handlePaste}
          onBlur={() => {
            setIsEditing(false);
            onSave(value);
          }}
          onKeyDown={handleKeyDown}
          placeholder="Подробное описание задачи в Markdown... (Ctrl+V для вставки скриншота)"
          className="w-full text-xs font-sans bg-canvas/50 border border-subtle rounded-lg p-2.5 text-primary focus:outline-none focus:border-accent-solid resize-y leading-relaxed"
        />
      ) : (
        <div
          onClick={() => setIsEditing(true)}
          className="min-h-[60px] text-xs text-primary leading-relaxed bg-canvas/30 hover:bg-canvas/50 p-2.5 rounded-lg border border-subtle/50 cursor-text transition-colors whitespace-pre-wrap break-words"
        >
          {renderMarkdownPreview(value)}
        </div>
      )}
    </div>
  );
};
