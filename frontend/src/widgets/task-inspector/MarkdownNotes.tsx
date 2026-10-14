import React, { useState, useEffect, useRef } from 'react';

export interface MarkdownNotesProps {
  initialValue: string;
  onSave: (notes: string) => void;
}

export const MarkdownNotes: React.FC<MarkdownNotesProps> = ({
  initialValue,
  onSave,
}) => {
  const [isEditing, setIsEditing] = useState(false);
  const [value, setValue] = useState(initialValue);
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

  return (
    <div className="space-y-1.5 pt-2 border-t border-subtle/50">
      <div className="flex items-center justify-between text-[10px] font-mono uppercase tracking-wider text-muted">
        <span>Заметки (Markdown)</span>
        <span className="text-[9px] lowercase text-muted">
          {isEditing ? 'esc для выхода' : 'клик для правок'}
        </span>
      </div>

      {isEditing ? (
        <textarea
          autoFocus
          rows={5}
          value={value}
          onChange={handleChange}
          onBlur={() => {
            setIsEditing(false);
            onSave(value);
          }}
          onKeyDown={handleKeyDown}
          placeholder="Подробное описание задачи в Markdown..."
          className="w-full text-xs font-sans bg-canvas/50 border border-subtle rounded-lg p-2.5 text-primary focus:outline-none focus:border-accent-solid resize-y leading-relaxed"
        />
      ) : (
        <div
          onClick={() => setIsEditing(true)}
          className="min-h-[60px] text-xs text-primary leading-relaxed bg-canvas/30 hover:bg-canvas/50 p-2.5 rounded-lg border border-subtle/50 cursor-text transition-colors whitespace-pre-wrap break-words"
        >
          {value ? (
            <span>{value}</span>
          ) : (
            <span className="text-muted italic">Добавить заметку в Markdown...</span>
          )}
        </div>
      )}
    </div>
  );
};
