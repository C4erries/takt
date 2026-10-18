import React, { useState, useEffect, useRef } from 'react';
import { Loader2 } from 'lucide-react';

export interface MarkdownNotesProps {
  initialValue: string;
  onSave: (notes: string) => void;
  onUploadImage?: (file: File) => Promise<string | null>;
}

function renderInline(text: string): React.ReactNode[] {
  const TOKEN_REGEX =
    /(!\[([^\]]*)\]\(([^)]+)\))|(\[([^\]]+)\]\(([^)]+)\))|(`([^`]+)`)|(\*\*([^*]+)\*\*|__([^_]+)__)|(~~([^~]+)~~)|(\*([^*]+)\*|_([^_]+)_)/g;
  const nodes: React.ReactNode[] = [];
  let lastIdx = 0;
  let match: RegExpExecArray | null;

  while ((match = TOKEN_REGEX.exec(text)) !== null) {
    if (match.index > lastIdx) {
      nodes.push(text.substring(lastIdx, match.index));
    }

    if (match[1]) {
      // Image: ![alt](url)
      const alt = match[2];
      const url = match[3];
      nodes.push(
        <span key={`img-${match.index}`} className="block my-2">
          <img
            src={url}
            alt={alt || 'вложение'}
            className="max-h-60 max-w-full rounded-lg border border-subtle object-contain bg-canvas/30 shadow-washi-sm hover:opacity-95 transition-opacity cursor-pointer"
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
    } else if (match[4]) {
      // Link: [text](url)
      const linkText = match[5];
      const linkUrl = match[6];
      nodes.push(
        <a
          key={`link-${match.index}`}
          href={linkUrl}
          target="_blank"
          rel="noopener noreferrer"
          onClick={(e) => e.stopPropagation()}
          className="text-accent-solid underline hover:opacity-80 transition-opacity"
        >
          {linkText}
        </a>
      );
    } else if (match[7]) {
      // Inline code: `code`
      nodes.push(
        <code
          key={`code-${match.index}`}
          className="font-mono text-[11px] px-1 py-0.5 rounded bg-canvas/60 text-primary border border-subtle/50"
        >
          {match[8]}
        </code>
      );
    } else if (match[9]) {
      // Bold: **text** or __text__
      nodes.push(
        <strong key={`b-${match.index}`} className="font-semibold text-primary">
          {match[10] || match[11]}
        </strong>
      );
    } else if (match[12]) {
      // Strikethrough: ~~text~~
      nodes.push(
        <del key={`del-${match.index}`} className="line-through text-muted">
          {match[13]}
        </del>
      );
    } else if (match[14]) {
      // Italic: *text* or _text_
      nodes.push(
        <em key={`em-${match.index}`} className="italic">
          {match[15] || match[16]}
        </em>
      );
    }

    lastIdx = match.index + match[0].length;
  }

  if (lastIdx < text.length) {
    nodes.push(text.substring(lastIdx));
  }

  return nodes;
}

function renderMarkdownPreview(text: string): React.ReactNode {
  if (!text) {
    return <span className="text-muted italic">Добавить заметку в Markdown...</span>;
  }

  const lines = text.split('\n');
  return (
    <div className="space-y-1">
      {lines.map((line, idx) => {
        if (!line.trim()) {
          return <div key={idx} className="h-2" />;
        }
        if (line.startsWith('### ')) {
          return (
            <h3 key={idx} className="font-semibold text-xs text-primary pt-1">
              {renderInline(line.slice(4))}
            </h3>
          );
        }
        if (line.startsWith('## ')) {
          return (
            <h2 key={idx} className="font-semibold text-sm text-primary pt-1.5">
              {renderInline(line.slice(3))}
            </h2>
          );
        }
        if (line.startsWith('# ')) {
          return (
            <h1 key={idx} className="font-bold text-base text-primary pt-2">
              {renderInline(line.slice(2))}
            </h1>
          );
        }
        if (line.startsWith('> ')) {
          return (
            <blockquote
              key={idx}
              className="border-l-2 border-accent-solid/60 pl-2.5 my-1 text-muted italic"
            >
              {renderInline(line.slice(2))}
            </blockquote>
          );
        }
        if (line.startsWith('- ') || line.startsWith('* ')) {
          return (
            <div key={idx} className="flex items-start gap-1.5 my-0.5 pl-1">
              <span className="text-accent-solid shrink-0 select-none">•</span>
              <span className="flex-1">{renderInline(line.slice(2))}</span>
            </div>
          );
        }
        return (
          <p key={idx} className="leading-relaxed">
            {renderInline(line)}
          </p>
        );
      })}
    </div>
  );
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

  // Critical fix: do not close editing mode or overwrite value if user is actively editing
  useEffect(() => {
    if (!isEditing) {
      setValue(initialValue);
    }
  }, [initialValue, isEditing]);

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
      // Ignored
    } finally {
      setIsUploading(false);
    }
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
          className="min-h-[60px] text-xs text-primary leading-relaxed bg-canvas/30 hover:bg-canvas/50 p-2.5 rounded-lg border border-subtle/50 cursor-text transition-colors break-words"
        >
          {renderMarkdownPreview(value)}
        </div>
      )}
    </div>
  );
};
