import React, { useState, useRef, useEffect } from 'react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { GripVertical, AlertCircle, Clock, Calendar, MoreHorizontal, GitFork } from 'lucide-react';
import { Checkbox, Badge } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import { motion, strikeThroughVariants } from '@/shared/lib/motion';
import type { Task, Priority } from '@/shared/api';
import {
  useProjectBoardsQuery,
  useProjectAllColumnsQuery,
  useMoveTaskMutation,
} from '@/entities/project';
import { useSubtasksQuery, formatTaskNumber, useShowTaskNumbers } from '@/entities/task';
import { MoveTaskModal } from './MoveTaskModal';

export interface KanbanCardProps {
  task: Task;
  onToggle?: (taskId: string) => void;
  onSelect?: (taskId: string) => void;
  isSelected?: boolean;
  isOverlay?: boolean;
  workspaceId?: string;
  projectId?: string;
  showTaskNumber?: boolean;
}

const PRIORITY_BADGES: Record<Priority, { label: string; variant: 'persimmon' | 'denim' | 'straw' | 'matcha' | 'subtle' } | null> = {
  pending: { label: 'В ожидании', variant: 'denim' },
  urgent: { label: 'В ожидании', variant: 'denim' },
  high: { label: 'Высокий', variant: 'persimmon' },
  medium: { label: 'Средний', variant: 'straw' },
  low: { label: 'Низкий', variant: 'subtle' },
  none: null,
};

const PRIORITY_STYLES: Record<Priority, string> = {
  pending: 'border-l-4 border-l-accent-denim',
  urgent: 'border-l-4 border-l-accent-denim',
  high: 'border-l-4 border-l-accent-persimmon',
  medium: 'border-l-4 border-l-accent-straw',
  low: 'border-l-4 border-l-accent-bamboo',
  none: 'border-l-transparent',
};

function getDueDateInfo(dueDateStr?: string | null) {
  if (!dueDateStr) return null;
  const due = new Date(dueDateStr);
  if (isNaN(due.getTime())) return null;

  const now = new Date();
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const endOfTomorrow = new Date(startOfToday.getTime() + 48 * 60 * 60 * 1000);
  const dueMidnight = new Date(due.getFullYear(), due.getMonth(), due.getDate());

  if (dueMidnight.getTime() < startOfToday.getTime()) {
    return { isOverdue: true, isSoon: false, text: due.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' }) };
  }
  if (due.getTime() <= endOfTomorrow.getTime()) {
    return { isOverdue: false, isSoon: true, text: dueMidnight.getTime() === startOfToday.getTime() ? 'Сегодня' : 'Завтра' };
  }
  return { isOverdue: false, isSoon: false, text: due.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' }) };
}

export const KanbanCard: React.FC<KanbanCardProps> = ({
  task,
  onToggle,
  onSelect,
  isSelected = false,
  isOverlay = false,
  workspaceId,
  projectId,
  showTaskNumber,
}) => {
  const effectiveProjectId = projectId || (task.projectId ?? undefined);
  const boardsQuery = useProjectBoardsQuery(workspaceId, effectiveProjectId);
  const boards = boardsQuery.data ?? [];
  const { data: boardGroups = [] } = useProjectAllColumnsQuery(workspaceId, boards);
  const moveTaskMutation = useMoveTaskMutation(workspaceId, effectiveProjectId);

  const currentBoard = boardGroups.find((bg) => bg.columns.some((c) => c.id === task.columnId));

  const subtasksQuery = useSubtasksQuery(
    (task.subtasksCount ?? 0) > 0 ? workspaceId : undefined,
    (task.subtasksCount ?? 0) > 0 ? task.id : undefined
  );
  const subtaskList = task.subtasks ?? subtasksQuery.data ?? [];
  const subtasksCount = task.subtasksCount ?? subtaskList.length;
  const subtasksCompletedCount = task.subtasksCompletedCount ?? subtaskList.filter((s) => s.isCompleted).length;
  const hasSubtasks = subtasksCount > 0;

  const [storedShowTaskNumbers] = useShowTaskNumbers();
  const effectiveShowTaskNumber = showTaskNumber ?? storedShowTaskNumbers;
  const [isMoveModalOpen, setIsMoveModalOpen] = useState(false);
  const [isMarkingCompleted, setIsMarkingCompleted] = useState(false);
  const completionTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (completionTimerRef.current) clearTimeout(completionTimerRef.current);
    };
  }, []);

  useEffect(() => {
    if (task.isCompleted) setIsMarkingCompleted(false);
  }, [task.isCompleted]);

  const handleToggle = () => {
    if (task.isCompleted) {
      setIsMarkingCompleted(false);
      if (completionTimerRef.current) clearTimeout(completionTimerRef.current);
      onToggle?.(task.id);
    } else {
      setIsMarkingCompleted(true);
      if (completionTimerRef.current) clearTimeout(completionTimerRef.current);
      completionTimerRef.current = setTimeout(() => onToggle?.(task.id), 350);
    }
  };

  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: task.id,
    data: { type: 'Task', task },
    disabled: isOverlay,
  });

  const style: React.CSSProperties = {
    transform: CSS.Translate.toString(transform),
    transition,
  };

  const priorityBadge = PRIORITY_BADGES[task.priority];
  const priorityStyle = PRIORITY_STYLES[task.priority] || PRIORITY_STYLES.none;
  const dueDateInfo = getDueDateInfo(task.dueDate);
  const taskShortNumber = formatTaskNumber(task.id);

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...(!isOverlay ? { ...attributes, ...listeners } : {})}
      onClick={() => onSelect?.(task.id)}
      className={cn(
        'group relative bg-surface border border-subtle rounded-lg p-3 select-none transition-shadow duration-150',
        priorityStyle,
        'hover:shadow-washi',
        isSelected && 'ring-1 ring-accent-solid/30 bg-accent-surface/30',
        isDragging && 'opacity-30 border-dashed border-subtle/80 shadow-none',
        isOverlay && 'shadow-washi-modal ring-1 ring-accent-solid/40 rotate-1 cursor-grabbing bg-surface',
        !isOverlay && 'cursor-grab active:cursor-grabbing'
      )}
    >
      <div className="flex items-start gap-2">
        {/* Drag handle */}
        <div
          className="mt-0.5 text-muted/40 group-hover:text-muted/70 p-0.5 -ml-1 rounded transition-colors shrink-0"
          title="Перетащить карточку"
        >
          <GripVertical className="w-3.5 h-3.5" />
        </div>

        {/* Task Details */}
        <div className="flex-1 min-w-0 space-y-1.5 pr-1">
          <div className="relative inline-block max-w-full">
            <p
              className={cn(
                'text-xs font-medium leading-snug break-words transition-colors duration-150',
                task.isCompleted
                  ? 'line-through text-muted opacity-70'
                  : isMarkingCompleted
                  ? 'text-muted opacity-80'
                  : 'text-primary'
              )}
            >
              {task.title}
            </p>
            <motion.span
              variants={strikeThroughVariants}
              initial="unchecked"
              animate={task.isCompleted || isMarkingCompleted ? 'checked' : 'unchecked'}
              className="absolute left-0 top-1/2 -translate-y-1/2 w-full h-[1.5px] bg-muted pointer-events-none rounded-full origin-left"
            />
          </div>

          {task.descriptionMarkdown && (
            <p className="text-[11px] text-muted line-clamp-2 leading-relaxed">
              {task.descriptionMarkdown}
            </p>
          )}

          {/* Badges / Metadata */}
          <div className="flex items-center gap-1.5 pt-0.5 flex-wrap">
            {/* WEEEK-style task ID badge (hidden by default) */}
            {effectiveShowTaskNumber && (
              <span className="font-mono text-[10px] text-muted font-medium bg-canvas/60 px-1.5 py-0.5 rounded border border-subtle/50">
                {taskShortNumber}
              </span>
            )}

            {priorityBadge && (
              <Badge variant={priorityBadge.variant} size="sm">
                {priorityBadge.label}
              </Badge>
            )}
            {dueDateInfo && (
              <div
                className={cn(
                  'flex items-center gap-1 font-mono text-[10px] px-1.5 py-0.5 rounded border',
                  dueDateInfo.isOverdue
                    ? 'bg-tag-persimmon/40 text-accent-persimmon border-accent-persimmon/40 font-medium'
                    : dueDateInfo.isSoon
                    ? 'bg-tag-straw/50 text-accent-straw border-accent-straw/40 font-medium'
                    : 'text-muted border-subtle/60 bg-canvas/40'
                )}
                title={dueDateInfo.isOverdue ? 'Срок истек!' : dueDateInfo.isSoon ? 'Приближается срок' : 'Дедлайн'}
              >
                {dueDateInfo.isOverdue ? (
                  <AlertCircle className="w-3 h-3 text-accent-persimmon shrink-0" />
                ) : dueDateInfo.isSoon ? (
                  <Clock className="w-3 h-3 text-accent-straw shrink-0" />
                ) : (
                  <Calendar className="w-3 h-3 text-muted shrink-0" />
                )}
                <span>{dueDateInfo.text}</span>
              </div>
            )}
            {task.tags && task.tags.length > 0 && (
              <div className="flex items-center gap-1 flex-wrap">
                {task.tags.map((tag) => (
                  <span
                    key={tag.id}
                    className="inline-flex items-center gap-1 text-[10px] font-medium px-2 py-0.5 bg-canvas/60 text-muted rounded-full border border-subtle"
                  >
                    {tag.colorHex && (
                      <span
                        className="w-1.5 h-1.5 rounded-full shrink-0"
                        style={{ backgroundColor: tag.colorHex }}
                      />
                    )}
                    <span>{tag.name}</span>
                  </span>
                ))}
              </div>
            )}
          </div>

          {/* Subtasks Progress & Compact List (WEEEK-style) */}
          {hasSubtasks && (
            <div className="pt-1.5 space-y-1.5 border-t border-subtle/40">
              {/* Progress bar and counter */}
              <div className="flex items-center gap-1.5 text-[10px] font-mono text-muted">
                <GitFork className="w-3 h-3 text-muted shrink-0" />
                <div className="flex-1 bg-canvas h-1 rounded-full overflow-hidden border border-subtle/40">
                  <div
                    className="bg-accent-solid h-full rounded-full transition-all duration-300"
                    style={{
                      width: `${subtasksCount > 0 ? (subtasksCompletedCount / subtasksCount) * 100 : 0}%`,
                    }}
                  />
                </div>
                <span className="shrink-0">{subtasksCompletedCount}/{subtasksCount}</span>
              </div>

              {/* Inline compact subtasks list with quick round checkboxes */}
              {subtaskList.length > 0 && (
                <div className="space-y-1 pt-0.5" onClick={(e) => e.stopPropagation()}>
                  {subtaskList.map((st) => (
                    <div
                      key={st.id}
                      className="flex items-center gap-2 group/st py-0.5"
                      onClick={(e) => e.stopPropagation()}
                    >
                      <div
                        className="shrink-0 scale-90"
                        onClick={(e) => e.stopPropagation()}
                      >
                        <Checkbox
                          checked={st.isCompleted}
                          onChange={() => onToggle?.(st.id)}
                          aria-label={`Завершить подзадачу ${st.title}`}
                        />
                      </div>
                      <span
                        className={cn(
                          'text-[11px] leading-tight truncate transition-colors',
                          st.isCompleted ? 'line-through text-muted/70' : 'text-primary'
                        )}
                      >
                        {st.title}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>

        {/* Actions & Checkbox */}
        <div
          className="pt-0.5 shrink-0 ml-auto flex items-center gap-1"
          onClick={(e) => e.stopPropagation()}
          onPointerDown={(e) => e.stopPropagation()}
        >
          {!isOverlay && (
            <>
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation();
                  setIsMoveModalOpen(true);
                }}
                className="opacity-0 group-hover:opacity-100 p-1 rounded hover:bg-canvas text-muted hover:text-primary transition-opacity"
                title="Переместить задачу..."
              >
                <MoreHorizontal className="w-3.5 h-3.5" />
              </button>

              <MoveTaskModal
                isOpen={isMoveModalOpen}
                onClose={() => setIsMoveModalOpen(false)}
                taskTitle={task.title}
                currentBoardId={currentBoard?.board.id}
                currentColumnId={task.columnId}
                boardGroups={boardGroups.map((bg) => ({
                  boardId: bg.board.id,
                  boardName: bg.boardName,
                  columns: bg.columns.map((c) => ({ id: c.id, name: c.name })),
                }))}
                onMove={(_targetBoardId, targetColumnId) => {
                  moveTaskMutation.mutate({
                    taskId: task.id,
                    data: {
                      columnId: targetColumnId,
                      position: task.position,
                      projectId: effectiveProjectId,
                    },
                  });
                }}
                isLoading={boardsQuery.isLoading}
              />
            </>
          )}

          <Checkbox
            checked={task.isCompleted || isMarkingCompleted}
            onChange={handleToggle}
            aria-label={`Завершить задачу ${task.title}`}
          />
        </div>
      </div>
    </div>
  );
};
