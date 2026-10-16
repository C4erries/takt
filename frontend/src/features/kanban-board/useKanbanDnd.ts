import { useState, useCallback } from 'react';
import {
  useSensor,
  useSensors,
  PointerSensor,
  KeyboardSensor,
  type DragStartEvent,
  type DragEndEvent,
} from '@dnd-kit/core';
import { arrayMove, sortableKeyboardCoordinates } from '@dnd-kit/sortable';
import type { Task } from '@/shared/api';
import type { Column } from '@/entities/project';

export interface UseKanbanDndProps {
  tasks: Task[];
  columns: Column[];
  onMoveTask: (taskId: string, columnId: string, position: number) => void;
  onReorderColumn?: (columnId: string, position: number) => void;
}

export const useKanbanDnd = ({
  tasks,
  columns,
  onMoveTask,
  onReorderColumn,
}: UseKanbanDndProps) => {
  const [activeTask, setActiveTask] = useState<Task | null>(null);
  const [activeColumn, setActiveColumn] = useState<Column | null>(null);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 5,
      },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const handleDragStart = useCallback(
    (event: DragStartEvent) => {
      const activeData = event.active.data.current;
      if (activeData?.type === 'Column') {
        const col = columns.find((c) => c.id === event.active.id) ?? (activeData.column as Column);
        setActiveColumn(col);
        setActiveTask(null);
      } else {
        const task = tasks.find((t) => t.id === event.active.id) ?? (activeData?.task as Task);
        if (task) {
          setActiveTask(task);
          setActiveColumn(null);
        }
      }
    },
    [columns, tasks]
  );

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event;
      setActiveTask(null);
      setActiveColumn(null);

      if (!over) return;
      if (active.id === over.id) return;

      const activeId = String(active.id);
      const overId = String(over.id);
      const activeData = active.data.current;
      const overData = over.data.current;

      // 1. Column Drag & Drop
      if (activeData?.type === 'Column') {
        let targetColumnId: string | null = null;
        if (overData?.type === 'Column' || columns.some((c) => c.id === overId)) {
          targetColumnId = overId;
        } else if (overData?.type === 'Task') {
          targetColumnId = overData.task?.columnId ?? null;
        } else {
          const overTask = tasks.find((t) => t.id === overId);
          if (overTask?.columnId) {
            targetColumnId = overTask.columnId;
          }
        }

        if (!targetColumnId || targetColumnId === activeId) return;

        const sortedCols = [...columns].sort((a, b) => a.position - b.position);
        const activeIndex = sortedCols.findIndex((c) => c.id === activeId);
        const overIndex = sortedCols.findIndex((c) => c.id === targetColumnId);

        if (activeIndex === -1 || overIndex === -1 || activeIndex === overIndex) return;

        const newCols = arrayMove(sortedCols, activeIndex, overIndex);
        const targetIdx = newCols.findIndex((c) => c.id === activeId);

        let newPosition: number;
        if (targetIdx === 0) {
          newPosition = Math.max(0.1, newCols[1].position / 2.0);
        } else if (targetIdx === newCols.length - 1) {
          newPosition = newCols[targetIdx - 1].position + 1000.0;
        } else {
          const prevPos = newCols[targetIdx - 1].position;
          const nextPos = newCols[targetIdx + 1].position;
          newPosition = Math.abs(nextPos - prevPos) < 0.0001 ? prevPos + 0.5 : (prevPos + nextPos) / 2.0;
        }

        onReorderColumn?.(activeId, newPosition);
        return;
      }

      // 2. Task Drag & Drop
      if (activeData?.type === 'Task' || (!activeData?.type && tasks.some((t) => t.id === activeId))) {
        const currentTask = tasks.find((t) => t.id === activeId);
        if (!currentTask) return;

        // Check if dropped directly onto a column
        const isOverColumn = overData?.type === 'Column' || columns.some((c) => c.id === overId);
        if (isOverColumn) {
          const targetColumn = columns.find((c) => c.id === overId) ?? (overData?.column as Column | undefined);
          if (!targetColumn) return;

          if (currentTask.columnId === targetColumn.id) return;

          const colTasks = tasks
            .filter((t) => t.columnId === targetColumn.id && t.id !== activeId)
            .sort((a, b) => a.position - b.position);

          let newPosition = 1000.0;
          if (colTasks.length > 0) {
            newPosition = colTasks[colTasks.length - 1].position + 1000.0;
          }

          onMoveTask(activeId, targetColumn.id, newPosition);
          return;
        }

        // Dropped onto another task card
        const overTask = tasks.find((t) => t.id === overId) ?? (overData?.task as Task | undefined);
        if (!overTask || !overTask.columnId) return;

        const targetColId = overTask.columnId;

        // Same column reordering
        if (currentTask.columnId === targetColId) {
          const colTasks = tasks
            .filter((t) => t.columnId === targetColId)
            .sort((a, b) => a.position - b.position);

          const activeIndex = colTasks.findIndex((t) => t.id === activeId);
          const overIndex = colTasks.findIndex((t) => t.id === overId);

          if (activeIndex === -1 || overIndex === -1 || activeIndex === overIndex) return;

          const newTasks = arrayMove(colTasks, activeIndex, overIndex);
          const targetIdx = newTasks.findIndex((t) => t.id === activeId);

          let newPosition: number;
          if (targetIdx === 0) {
            newPosition = Math.max(0.1, newTasks[1].position / 2.0);
          } else if (targetIdx === newTasks.length - 1) {
            newPosition = newTasks[targetIdx - 1].position + 1000.0;
          } else {
            const prevPos = newTasks[targetIdx - 1].position;
            const nextPos = newTasks[targetIdx + 1].position;
            newPosition = Math.abs(nextPos - prevPos) < 0.0001 ? prevPos + 0.5 : (prevPos + nextPos) / 2.0;
          }

          onMoveTask(activeId, targetColId, newPosition);
          return;
        }

        // Cross-column moving onto a specific card
        const colTasks = tasks
          .filter((t) => t.columnId === targetColId && t.id !== activeId)
          .sort((a, b) => a.position - b.position);

        const overIndex = colTasks.findIndex((t) => t.id === overId);
        if (overIndex === -1) {
          const newPosition = overTask.position + 1000.0;
          onMoveTask(activeId, targetColId, newPosition);
          return;
        }

        let newPosition: number;
        if (overIndex === 0) {
          newPosition = Math.max(0.1, colTasks[0].position / 2.0);
        } else {
          const prevPos = colTasks[overIndex - 1].position;
          const nextPos = colTasks[overIndex].position;
          newPosition = Math.abs(nextPos - prevPos) < 0.0001 ? prevPos + 0.5 : (prevPos + nextPos) / 2.0;
        }

        onMoveTask(activeId, targetColId, newPosition);
      }
    },
    [tasks, columns, onMoveTask, onReorderColumn]
  );

  const handleDragCancel = useCallback(() => {
    setActiveTask(null);
    setActiveColumn(null);
  }, []);

  return {
    activeTask,
    activeColumn,
    sensors,
    handleDragStart,
    handleDragEnd,
    handleDragCancel,
  };
};
