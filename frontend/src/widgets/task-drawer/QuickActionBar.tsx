import React, { useRef } from 'react';
import { Calendar, Paperclip, Tag as TagIcon, KanbanSquare, X } from 'lucide-react';

export interface QuickActionBarProps {
  dueDate?: string | null;
  onDueDateChange: (date: string | null) => void;
  onAttachClick: () => void;
  onToggleTagInput: () => void;
  columns?: { id: string; name: string }[];
  selectedColumnId?: string | null;
  onColumnChange?: (columnId: string) => void;
}

export const QuickActionBar: React.FC<QuickActionBarProps> = ({
  dueDate,
  onDueDateChange,
  onAttachClick,
  onToggleTagInput,
  columns = [],
  selectedColumnId,
  onColumnChange,
}) => {
  const dateInputRef = useRef<HTMLInputElement | null>(null);

  const handleDateButtonClick = () => {
    if (dateInputRef.current) {
      if (typeof dateInputRef.current.showPicker === 'function') {
        dateInputRef.current.showPicker();
      } else {
        dateInputRef.current.focus();
      }
    }
  };

  return (
    <div className="flex items-center gap-2 flex-wrap p-2.5 bg-canvas/40 border border-subtle/60 rounded-xl text-xs select-none">
      {/* Due Date capsule */}
      <div className="inline-flex items-center gap-1">
        <button
          type="button"
          onClick={handleDateButtonClick}
          className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-surface border border-subtle/80 text-muted hover:text-primary transition-colors shadow-washi-sm cursor-pointer"
          title="Назначить срок"
        >
          <Calendar className="w-3.5 h-3.5 text-accent-solid" />
          <span className="font-mono text-[11px]">{dueDate ? dueDate : '+ Срок'}</span>
        </button>
        {dueDate && (
          <button
            type="button"
            onClick={() => onDueDateChange(null)}
            className="p-1 text-muted hover:text-primary rounded-md hover:bg-canvas transition-colors"
            title="Очистить срок"
          >
            <X className="w-3 h-3" />
          </button>
        )}
        <input
          ref={dateInputRef}
          type="date"
          value={dueDate ?? ''}
          onChange={(e) => onDueDateChange(e.target.value || null)}
          className="sr-only"
        />
      </div>

      {/* Attach File capsule */}
      <button
        type="button"
        onClick={onAttachClick}
        className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-surface border border-subtle/80 text-muted hover:text-primary transition-colors shadow-washi-sm"
        title="Прикрепить файл"
      >
        <Paperclip className="w-3.5 h-3.5 text-accent-solid" />
        <span>+ Файл</span>
      </button>

      {/* Tag capsule */}
      <button
        type="button"
        onClick={onToggleTagInput}
        className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-surface border border-subtle/80 text-muted hover:text-primary transition-colors shadow-washi-sm"
        title="Добавить тег"
      >
        <TagIcon className="w-3.5 h-3.5 text-accent-solid" />
        <span>+ Тег</span>
      </button>

      {/* Column/Status Selector (if project has columns) */}
      {columns.length > 0 && onColumnChange && (
        <div className="flex items-center gap-1.5 px-2 py-0.5 rounded-lg bg-surface border border-subtle/80 text-muted shadow-washi-sm">
          <KanbanSquare className="w-3.5 h-3.5 text-accent-solid" />
          <select
            value={selectedColumnId ?? ''}
            onChange={(e) => onColumnChange(e.target.value)}
            className="bg-transparent text-[11px] font-sans text-primary focus:outline-none cursor-pointer"
          >
            {columns.map((col) => (
              <option key={col.id} value={col.id}>
                {col.name}
              </option>
            ))}
          </select>
        </div>
      )}
    </div>
  );
};
