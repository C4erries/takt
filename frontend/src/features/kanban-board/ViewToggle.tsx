import React from 'react';
import { LayoutGrid, List, Hash } from 'lucide-react';
import { cn } from '@/shared/lib/cn';
import { useShowTaskNumbers } from '@/entities/task';

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
  const [showTaskNumbers, setShowTaskNumbers] = useShowTaskNumbers();

  return (
    <div className={cn('inline-flex items-center gap-1.5', className)}>
      <div
        className="inline-flex items-center gap-1 p-0.5 rounded-lg bg-canvas/80 border border-subtle select-none"
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

      {/* Task Numbers Visibility Toggle */}
      <button
        type="button"
        onClick={() => setShowTaskNumbers(!showTaskNumbers)}
        className={cn(
          'inline-flex items-center gap-1 px-2 py-1 rounded-lg border text-xs font-medium transition-all duration-150 select-none',
          showTaskNumbers
            ? 'bg-surface text-accent-solid border-accent-solid/30 shadow-washi-sm'
            : 'bg-canvas/80 text-muted border-subtle hover:text-primary hover:bg-surface/50'
        )}
        title="Номера задач"
        aria-label="Номера задач"
        aria-pressed={showTaskNumbers}
      >
        <Hash className="w-3.5 h-3.5" />
        <span className="hidden sm:inline text-[11px]">#ID</span>
      </button>
    </div>
  );
};
