import React, { useEffect } from 'react';
import {
  Search,
  Inbox,
  Zap,
  Plus,
  FolderClosed,
  CheckCircle2,
  Circle,
  Loader2,
  CornerDownLeft,
} from 'lucide-react';
import type { CommandPaletteModalProps, CommandItem } from './types';
import { useCommandPalette } from './useCommandPalette';

export const CommandPaletteModal: React.FC<CommandPaletteModalProps> = ({
  isOpen,
  onClose,
  workspaceId,
  onOpenQuickCapture,
}) => {
  const {
    query,
    setQuery,
    isLoading,
    items,
    selectedIndex,
    setSelectedIndex,
    handleKeyDown,
  } = useCommandPalette({
    isOpen,
    onClose,
    workspaceId,
    onOpenQuickCapture,
  });

  // Блокировка скролла страницы при открытой модалке
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  if (!isOpen) return null;

  const renderIcon = (item: CommandItem) => {
    switch (item.type) {
      case 'navigation':
        if (item.id === 'nav-inbox') return <Inbox className="w-4 h-4 text-muted" />;
        return <Zap className="w-4 h-4 text-muted" />;
      case 'action':
        return <Plus className="w-4 h-4 text-accent-solid" />;
      case 'project':
        return <FolderClosed className="w-4 h-4 text-muted" />;
      case 'task':
        return item.isCompleted ? (
          <CheckCircle2 className="w-4 h-4 text-accent-solid" />
        ) : (
          <Circle className="w-4 h-4 text-muted" />
        );
      default:
        return <Search className="w-4 h-4 text-muted" />;
    }
  };

  return (
    <div
      role="dialog"
      aria-modal="true"
      className="fixed inset-0 z-50 flex items-start justify-center pt-16 md:pt-24 px-4 bg-primary/25"
      onClick={onClose}
    >
      <div
        className="w-full max-w-xl bg-surface border border-subtle rounded-2xl shadow-washi-modal overflow-hidden flex flex-col max-h-[500px]"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        {/* Search Input Bar */}
        <div className="flex items-center gap-3 px-4 py-3.5 border-b border-subtle/60 bg-surface">
          <Search className="w-4 h-4 text-muted shrink-0" />
          <input
            autoFocus
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Поиск задач, проектов или команд..."
            className="flex-1 bg-transparent text-sm text-primary placeholder:text-muted/70 focus:outline-none"
          />
          {isLoading ? (
            <Loader2 className="w-4 h-4 text-accent-solid animate-spin shrink-0" />
          ) : (
            <kbd className="font-mono text-[10px] text-muted bg-canvas px-1.5 py-0.5 rounded border border-subtle">
              Esc
            </kbd>
          )}
        </div>

        {/* Results List */}
        <div className="flex-1 overflow-y-auto p-2 space-y-1">
          {items.length === 0 ? (
            <div className="py-12 text-center text-xs text-muted space-y-1">
              <p className="font-medium text-primary">Ничего не найдено</p>
              <p>По запросу «{query}» нет подходящих задач или проектов.</p>
            </div>
          ) : (
            items.map((item, index) => {
              const isSelected = index === selectedIndex;
              return (
                <div
                  key={item.id}
                  onClick={() => item.onSelect()}
                  onMouseEnter={() => setSelectedIndex(index)}
                  className={`flex items-center justify-between px-3 py-2 rounded-xl text-xs cursor-pointer transition-colors ${
                    isSelected
                      ? 'bg-accent-surface text-primary font-medium'
                      : 'text-primary hover:bg-canvas/40'
                  }`}
                >
                  <div className="flex items-center gap-3 min-w-0 mr-2">
                    <span className="shrink-0">{renderIcon(item)}</span>
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="truncate">{item.title}</span>
                        {item.type === 'project' && (
                          <span className="text-[10px] px-1.5 py-0.2 rounded-full border border-subtle font-mono text-muted">
                            Проект
                          </span>
                        )}
                      </div>
                      {item.subtitle && (
                        <p className="text-[11px] text-muted truncate mt-0.5 font-normal">
                          {item.subtitle}
                        </p>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center gap-2 shrink-0">
                    {item.badge && (
                      <kbd className="font-mono text-[10px] text-muted bg-canvas px-1.5 py-0.5 rounded border border-subtle">
                        {item.badge}
                      </kbd>
                    )}
                    {isSelected && (
                      <CornerDownLeft className="w-3.5 h-3.5 text-muted opacity-80" />
                    )}
                  </div>
                </div>
              );
            })
          )}
        </div>

        {/* Footer shortcuts */}
        <div className="flex items-center justify-between px-4 py-2 border-t border-subtle/50 bg-canvas/30 text-[10px] font-mono text-muted">
          <div className="flex items-center gap-3">
            <span>
              <kbd className="bg-canvas px-1 py-0.5 rounded border border-subtle">↑</kbd>{' '}
              <kbd className="bg-canvas px-1 py-0.5 rounded border border-subtle">↓</kbd> Навигация
            </span>
            <span>
              <kbd className="bg-canvas px-1 py-0.5 rounded border border-subtle">↵</kbd> Выбрать
            </span>
          </div>
          <span>
            <kbd className="bg-canvas px-1 py-0.5 rounded border border-subtle">Esc</kbd> Закрыть
          </span>
        </div>
      </div>
    </div>
  );
};
