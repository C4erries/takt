import React from 'react';
import { LayoutGrid, List } from 'lucide-react';
import { cn } from '@/shared/lib/cn';

export type ViewMode = 'kanban' | 'list';

export interface ViewToggleProps {
  view: ViewMode;
  onViewChange: (view: ViewMode) => void;
  className?: string;
}

export const ViewToggle: React.FC<ViewToggleProps> = ({
  view,
  onViewChange,
  className,
}) => {
  return (
    <div
      className={cn(
        'inline-flex items-center gap-1 p-0.5 rounded-lg bg-canvas/80 border border-subtle select-none',
        className
      )}
      role="group"
      aria-label="Переключение вида"
    >
      <button
        type="button"
        onClick={() => onViewChange('kanban')}
        className={cn(
          'flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium transition-all duration-150',
          view === 'kanban'
            ? 'bg-surface text-primary shadow-washi-sm border border-subtle/40'
            : 'text-muted hover:text-primary hover:bg-surface/50'
        )}
        title="Канбан-доска"
      >
        <LayoutGrid className="w-3.5 h-3.5" />
        <span>Доска</span>
      </button>

      <button
        type="button"
        onClick={() => onViewChange('list')}
        className={cn(
          'flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium transition-all duration-150',
          view === 'list'
            ? 'bg-surface text-primary shadow-washi-sm border border-subtle/40'
            : 'text-muted hover:text-primary hover:bg-surface/50'
        )}
        title="Плоский список задач"
      >
        <List className="w-3.5 h-3.5" />
        <span>Список</span>
      </button>
    </div>
  );
};
