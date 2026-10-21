import React, { useState, useRef, useEffect } from 'react';
import { useSortable, SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Plus, MoreHorizontal, ArrowRight, GripVertical, Inbox, Check, Zap, Trash2, Pencil } from 'lucide-react';
import { Badge, Button } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import { motion, AnimatePresence, cardExitVariants } from '@/shared/lib/motion';
import type { Task } from '@/shared/api';
import {
  useProjectBoardsQuery,
  useMoveColumnMutation,
  useUpdateColumnMutation,
  useDeleteColumnMutation,
  type Column,
  type ColumnStatusCategory,
} from '@/entities/project';
import { KanbanCard } from './KanbanCard';
import { ColumnAutomationsModal } from './automations/ColumnAutomationsModal';

export interface KanbanColumnProps {
  column: Column;
  tasks: Task[];
  allColumns?: Column[];
  projectId?: string; workspaceId?: string;
  onToggleTask?: (taskId: string) => void; onSelectTask?: (taskId: string) => void;
  selectedTaskId?: string | null;
  onCreateTask?: (title: string, columnId: string) => Promise<void>;
  isOverlay?: boolean;
}

const CATEGORY_STYLES: Record<ColumnStatusCategory, { dot: string; label: string }> = {
  backlog: { dot: 'bg-muted', label: 'Бэклог' }, todo: { dot: 'bg-accent-solid', label: 'К выполнению' },
  in_progress: { dot: 'bg-tag-persimmon', label: 'В работе' }, done: { dot: 'bg-accent-solid', label: 'Готово' },
};

const COLUMN_COLORS = [
  { name: 'По умолчанию', hex: null }, { name: 'Матча', hex: '#566B44' }, { name: 'Хурма', hex: '#D97746' }, { name: 'Деним', hex: '#4A6B82' },
  { name: 'Солома', hex: '#B8943F' }, { name: 'Бамбук', hex: '#7C8E64' }, { name: 'Слива', hex: '#7E6B8F' }, { name: 'Суми', hex: '#44483E' },
] as const;

function getStoredColumnColor(columnId: string): string | null {
  try { return JSON.parse(localStorage.getItem('takt_column_colors') || '{}')[columnId] ?? null; } catch { return null; }
}

function setStoredColumnColor(columnId: string, color: string | null): void {
  try { const m = JSON.parse(localStorage.getItem('takt_column_colors') || '{}'); if (color) m[columnId] = color; else delete m[columnId]; localStorage.setItem('takt_column_colors', JSON.stringify(m)); } catch {}
}

export const KanbanColumn: React.FC<KanbanColumnProps> = ({
  column, tasks, allColumns, projectId, workspaceId, onToggleTask, onSelectTask, selectedTaskId, onCreateTask, isOverlay = false,
}) => {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging, isOver } =
    useSortable({ id: column.id, data: { type: 'Column', column }, disabled: isOverlay });

  const style: React.CSSProperties = { transform: CSS.Translate.toString(transform), transition };

  const [isAdding, setIsAdding] = useState(false);
  const [newTitle, setNewTitle] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isAutomationsOpen, setIsAutomationsOpen] = useState(false);
  const [columnColor, setColumnColor] = useState<string | null>(() => getStoredColumnColor(column.id));
  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const [editedTitle, setEditedTitle] = useState(column.name);
  const menuRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const titleInputRef = useRef<HTMLInputElement>(null);

  const boardsQuery = useProjectBoardsQuery(workspaceId, projectId);
  const moveColumnMutation = useMoveColumnMutation(workspaceId, column.boardId);
  const updateColumnMutation = useUpdateColumnMutation(workspaceId, column.boardId);
  const deleteColumnMutation = useDeleteColumnMutation(workspaceId, column.boardId);
  const otherBoards = (boardsQuery.data ?? []).filter((b) => b.id !== column.boardId);

  useEffect(() => {
    if (!isMenuOpen) return;
    const onOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) setIsMenuOpen(false);
    };
    document.addEventListener('mousedown', onOutside);
    return () => document.removeEventListener('mousedown', onOutside);
  }, [isMenuOpen]);

  useEffect(() => {
    if (isAdding) inputRef.current?.focus();
  }, [isAdding]);

  useEffect(() => {
    if (isEditingTitle) {
      titleInputRef.current?.focus();
      titleInputRef.current?.select();
    }
  }, [isEditingTitle]);

  const startRename = () => {
    setEditedTitle(column.name);
    setIsEditingTitle(true);
  };

  const handleSaveTitle = () => {
    if (!isEditingTitle) return;
    setIsEditingTitle(false);
    const trimmed = editedTitle.trim();
    if (trimmed && trimmed !== column.name) {
      updateColumnMutation.mutate({ columnId: column.id, data: { name: trimmed } });
    } else {
      setEditedTitle(column.name);
    }
  };

  const handleCancelTitle = () => {
    setIsEditingTitle(false);
    setEditedTitle(column.name);
  };

  const categoryConfig = CATEGORY_STYLES[column.statusCategory] || CATEGORY_STYLES.todo;

  const handleQuickAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = newTitle.trim();
    if (!trimmed || isSubmitting) return;

    setIsSubmitting(true);
    try {
      if (onCreateTask) await onCreateTask(trimmed, column.id);
      setNewTitle('');
      inputRef.current?.focus();
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleSelectColor = (hex: string | null) => {
    setColumnColor(hex);
    setStoredColumnColor(column.id, hex);
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn(
        'w-72 shrink-0 flex flex-col bg-canvas/60 border border-subtle rounded-xl p-3 transition-colors duration-150',
        isOver && !isDragging && 'border-accent-solid/50 bg-accent-surface/20',
        isDragging && 'opacity-30 border-dashed border-subtle/80 shadow-none',
        isOverlay && 'shadow-washi-modal ring-1 ring-accent-solid/40 rotate-1 cursor-grabbing bg-canvas/90'
      )}
    >
      {/* Color Accent Bar */}
      <div className="h-1 -mx-3 -mt-3 mb-2.5 rounded-t-xl transition-colors duration-150" style={{ backgroundColor: columnColor || 'transparent' }} />

      {/* Column Header */}
      <div className="flex items-center justify-between px-1 pb-2.5 select-none">
        <div
          {...(!isOverlay && !isEditingTitle ? { ...attributes, ...listeners } : {})}
          className="flex items-center gap-2 min-w-0 cursor-grab active:cursor-grabbing flex-1 mr-1"
          title={isEditingTitle ? undefined : `Перетащить колонку ${column.name}`}
        >
          {!isOverlay && (
            <div className="text-muted/40 hover:text-muted p-0.5 -ml-1 rounded transition-colors shrink-0">
              <GripVertical className="w-3.5 h-3.5" />
            </div>
          )}
          <span className={cn('w-2 h-2 rounded-full shrink-0', categoryConfig.dot)} />
          {isEditingTitle ? (
            <input
              ref={titleInputRef}
              type="text"
              value={editedTitle}
              onChange={(e) => setEditedTitle(e.target.value)}
              onBlur={handleSaveTitle}
              onKeyDown={(e) => {
                if (e.key === 'Enter') { e.preventDefault(); handleSaveTitle(); }
                else if (e.key === 'Escape') { e.preventDefault(); handleCancelTitle(); }
              }}
              onPointerDown={(e) => e.stopPropagation()}
              onClick={(e) => e.stopPropagation()}
              className="bg-surface border border-subtle rounded px-2 py-0.5 text-xs font-semibold text-primary uppercase focus:outline-none focus:border-accent-solid min-w-0 flex-1"
              autoFocus
            />
          ) : (
            <h3
              onDoubleClick={(e) => { e.stopPropagation(); startRename(); }}
              className="text-xs font-semibold tracking-wide uppercase text-primary truncate cursor-text"
              title="Двойной клик для переименования"
            >
              {column.name}
            </h3>
          )}
          {column.isInbox && (
            <span title="Входящая колонка проекта" className="inline-flex items-center shrink-0">
              <Inbox className="w-3.5 h-3.5 text-accent-solid" />
            </span>
          )}
          <Badge variant="subtle" size="sm" className="font-mono text-[10px] px-1.5 py-0">
            {tasks.length}
          </Badge>
        </div>

        <div className="flex items-center gap-0.5" onPointerDown={(e) => e.stopPropagation()}>
          <Button
            variant="ghost"
            size="sm"
            disabled={tasks.length > 0 || deleteColumnMutation.isPending}
            onClick={(e) => {
              e.stopPropagation();
              if (tasks.length === 0 && window.confirm(`Удалить колонку «${column.name}»?`)) {
                deleteColumnMutation.mutate(column.id);
              }
            }}
            className={cn('h-6 w-6 p-0 transition-colors', tasks.length === 0 ? 'text-muted hover:text-destructive hover:bg-destructive/10' : 'text-muted/40 cursor-not-allowed')}
            title={tasks.length > 0 ? 'Нельзя удалить колонку с задачами. Сначала переместите или удалите задачи.' : `Удалить колонку «${column.name}»`}
            aria-label={`Удалить колонку ${column.name}`}
          >
            <Trash2 className="w-3.5 h-3.5" />
          </Button>

          <div className="relative" ref={menuRef}>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsMenuOpen((prev) => !prev)}
              className={cn('h-6 w-6 p-0 text-muted hover:text-primary', isMenuOpen && 'bg-canvas/60 text-primary')}
              title="Меню колонки"
              aria-label={`Меню колонки ${column.name}`}
            >
              <MoreHorizontal className="w-3.5 h-3.5" />
            </Button>

            {isMenuOpen && (
              <div className="absolute right-0 top-full mt-1 w-56 z-30 bg-surface border border-subtle rounded-xl shadow-washi p-1.5 space-y-1 text-xs animate-in fade-in zoom-in-95 duration-100">
                {/* 0. Rename Column */}
                <button
                  type="button"
                  onClick={() => { setIsMenuOpen(false); startRename(); }}
                  className="w-full flex items-center gap-2 px-2.5 py-1.5 rounded-lg text-left text-primary hover:bg-canvas/60 transition-colors group"
                >
                  <Pencil className="w-3.5 h-3.5 text-muted group-hover:text-primary shrink-0" />
                  <span>Переименовать</span>
                </button>

                {/* 1. Toggle Inbox Status */}
                <button
                  type="button"
                  onClick={() => {
                    setIsMenuOpen(false);
                    updateColumnMutation.mutate({ columnId: column.id, data: { isInbox: !column.isInbox } });
                  }}
                  className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-left text-primary hover:bg-canvas/60 transition-colors group"
                >
                  <span className="flex items-center gap-2 truncate">
                    <Inbox className={cn('w-3.5 h-3.5 shrink-0', column.isInbox ? 'text-accent-solid' : 'text-muted')} />
                    <span className="truncate">{column.isInbox ? 'Снять статус Inbox' : 'Сделать входящей (Inbox)'}</span>
                  </span>
                  {column.isInbox && <Check className="w-3 h-3 text-accent-solid shrink-0 ml-1.5" />}
                </button>

                {/* 2. Column Automations */}
                <button
                  type="button"
                  onClick={() => { setIsMenuOpen(false); setIsAutomationsOpen(true); }}
                  className="w-full flex items-center gap-2 px-2.5 py-1.5 rounded-lg text-left text-primary hover:bg-canvas/60 transition-colors group"
                >
                  <Zap className="w-3.5 h-3.5 text-accent-solid shrink-0" />
                  <span>Автоматизации колонки (⚡)</span>
                </button>

                {/* 3. Move to Board */}
                <div className="px-2 py-1 text-[10px] font-semibold text-muted tracking-wider uppercase border-t border-subtle/40 pt-1.5 mt-1">
                  Переместить на доску
                </div>
                {otherBoards.length === 0 ? (
                  <div className="px-2 py-1 text-[11px] text-muted italic">Нет других досок</div>
                ) : (
                  <div className="space-y-0.5">
                    {otherBoards.map((targetBoard) => (
                      <button
                        key={targetBoard.id}
                        type="button"
                        onClick={() => {
                          setIsMenuOpen(false);
                          moveColumnMutation.mutate({ columnId: column.id, targetBoardId: targetBoard.id });
                        }}
                        className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-left text-primary hover:bg-canvas/60 transition-colors group"
                      >
                        <span className="truncate">{targetBoard.name}</span>
                        <ArrowRight className="w-3 h-3 text-muted group-hover:text-primary shrink-0 ml-1.5" />
                      </button>
                    ))}
                  </div>
                )}

                {/* 4. Column Color Picker */}
                <div className="px-2 py-1 text-[10px] font-semibold text-muted tracking-wider uppercase border-t border-subtle/40 pt-1.5 mt-1">
                  Цвет колонки
                </div>
                <div className="grid grid-cols-4 gap-1.5 px-2 py-1">
                  {COLUMN_COLORS.map((c) => (
                    <button
                      key={c.name}
                      type="button"
                      onClick={() => handleSelectColor(c.hex)}
                      title={c.name}
                      className={cn('w-6 h-6 rounded-full flex items-center justify-center transition-all border shadow-washi-sm', c.hex === columnColor ? 'ring-2 ring-accent-solid scale-105 border-surface' : 'border-subtle/80 hover:scale-105', !c.hex && 'bg-canvas/60 text-muted')}
                      style={c.hex ? { backgroundColor: c.hex } : undefined}
                    >
                      {!c.hex && <span className="text-[10px] font-mono leading-none">✕</span>}
                    </button>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Inline Task Add row */}
      {!isOverlay && (
        <div className="mb-2">
          {isAdding ? (
            <form onSubmit={handleQuickAdd} className="p-2 rounded-xl bg-surface border border-subtle shadow-washi space-y-2 animate-in fade-in duration-100">
              <input
                ref={inputRef}
                type="text"
                autoFocus
                value={newTitle}
                onChange={(e) => setNewTitle(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Escape') { setIsAdding(false); setNewTitle(''); } }}
                placeholder="Название задачи... (Enter)"
                disabled={isSubmitting}
                className="w-full px-2.5 py-1.5 text-xs bg-surface border border-subtle rounded-md text-primary placeholder:text-muted/60 focus:outline-none focus:border-accent-solid"
              />
              <div className="flex items-center justify-end gap-1.5">
                <Button type="button" variant="ghost" size="sm" onClick={() => { setIsAdding(false); setNewTitle(''); }} className="text-[11px] h-6 px-2 text-muted hover:text-primary">Отмена</Button>
                <Button type="submit" variant="solid" size="sm" isLoading={isSubmitting} disabled={!newTitle.trim()} className="text-[11px] h-6 px-2.5">Добавить</Button>
              </div>
            </form>
          ) : (
            <button
              type="button"
              onClick={() => setIsAdding(true)}
              className="w-full text-xs text-muted hover:text-primary hover:bg-surface/80 rounded-lg py-1.5 px-2.5 flex items-center gap-1.5 transition-colors border border-transparent hover:border-subtle/50 select-none"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>Добавить задачу</span>
            </button>
          )}
        </div>
      )}

      {/* Droppable Task List */}
      <SortableContext items={tasks.map((t) => t.id)} strategy={verticalListSortingStrategy}>
        <div className="flex-1 min-h-[140px] space-y-2 overflow-y-auto pr-0.5">
          <AnimatePresence initial={false}>
            {tasks.map((task) => (
              <motion.div key={task.id} layout={!isOverlay && !isDragging} variants={cardExitVariants} initial="initial" animate="animate" exit="exit">
                <KanbanCard
                  task={task}
                  workspaceId={workspaceId}
                  projectId={projectId}
                  isSelected={selectedTaskId === task.id}
                  isOverlay={isOverlay}
                  onSelect={onSelectTask}
                  onToggle={onToggleTask}
                />
              </motion.div>
            ))}
          </AnimatePresence>

          {tasks.length === 0 && !isAdding && (
            <div
              className={cn(
                'h-24 rounded-lg border border-dashed border-subtle/70 flex items-center justify-center text-center p-3 select-none',
                isOver ? 'border-accent-solid/60 bg-accent-surface/30' : 'bg-surface/30'
              )}
            >
              <p className="text-[11px] text-muted">Перетащите задачу сюда</p>
            </div>
          )}
        </div>
      </SortableContext>

      {/* Column Automations Modal */}
      {!isOverlay && (
        <ColumnAutomationsModal
          isOpen={isAutomationsOpen}
          onClose={() => setIsAutomationsOpen(false)}
          column={column}
          allColumns={allColumns || [column]}
          workspaceId={workspaceId}
        />
      )}
    </div>
  );
};
