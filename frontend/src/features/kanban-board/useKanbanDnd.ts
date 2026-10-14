import { useState, useCallback } from 'react';
import {
  useSensor,
  useSensors,
  PointerSensor,
  KeyboardSensor,
  type DragStartEvent,
  type DragEndEvent,
} from '@dnd-kit/core';
import { sortableKeyboardCoordinates } from '@dnd-kit/sortable';
import type { Task } from '@/shared/api';
import type { Column } from '@/entities/project';

export interface UseKanbanDndProps {
  tasks: Task[];
  columns: Column[];
  onMoveTask: (taskId: string, columnId: string, position: number) => void;
}

export const useKanbanDnd = ({
  tasks,
  columns,
  onMoveTask,
}: UseKanbanDndProps) => {
  const [activeTask, setActiveTask] = useState<Task | null>(null);

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
      const task = tasks.find((t) => t.id === event.active.id);
      if (task) {
        setActiveTask(task);
      }
    },
    [tasks]
  );

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event;
      setActiveTask(null);

      if (!over) return;
      if (active.id === over.id) return;

      const activeId = String(active.id);
      const overId = String(over.id);

      const currentTask = tasks.find((t) => t.id === activeId);
      if (!currentTask) return;

      // Check if dropped directly onto a column
      const targetColumn = columns.find((c) => c.id === overId);
      if (targetColumn) {
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
      const overTask = tasks.find((t) => t.id === overId);
      if (!overTask || !overTask.columnId) return;

      const targetColId = overTask.columnId;
      const colTasks = tasks
        .filter((t) => t.columnId === targetColId && t.id !== activeId)
        .sort((a, b) => a.position - b.position);

      const overIndex = colTasks.findIndex((t) => t.id === overId);
      if (overIndex === -1) {
        // Fallback: append after overTask
        const newPosition = overTask.position + 1000.0;
        onMoveTask(activeId, targetColId, newPosition);
        return;
      }

      let newPosition: number;
      if (overIndex === 0) {
        // Placed before the first task in column
        newPosition = Math.max(0.1, colTasks[0].position / 2.0);
      } else {
        // Placed between overIndex - 1 and overIndex
        const prevPos = colTasks[overIndex - 1].position;
        const nextPos = colTasks[overIndex].position;
        newPosition = (prevPos + nextPos) / 2.0;
      }

      onMoveTask(activeId, targetColId, newPosition);
    },
    [tasks, columns, onMoveTask]
  );

  const handleDragCancel = useCallback(() => {
    setActiveTask(null);
  }, []);

  return {
    activeTask,
    sensors,
    handleDragStart,
    handleDragEnd,
    handleDragCancel,
  };
};
