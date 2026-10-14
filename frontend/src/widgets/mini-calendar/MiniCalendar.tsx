import React, { useState } from 'react';
import { ChevronLeft, ChevronRight, X } from 'lucide-react';
import { useTaskStore, useInboxTasks } from '@/features/task-navigation';
import { useTodayTasksQuery, useUpdateTaskMutation } from '@/entities/task';

export interface MiniCalendarProps {
  workspaceId?: string;
}

export const MiniCalendar: React.FC<MiniCalendarProps> = ({ workspaceId }) => {
  const [currentDate, setCurrentDate] = useState(() => new Date());

  const {
    selectedTaskId,
    selectedDateFilter,
    setSelectedDateFilter,
  } = useTaskStore();

  const { data: inboxTasks = [] } = useInboxTasks();
  const { data: todayTasks = [] } = useTodayTasksQuery(workspaceId);
  const updateTaskMutation = useUpdateTaskMutation(workspaceId);

  // Combine tasks to count due dates
  const allTasks = [...inboxTasks, ...todayTasks];

  const year = currentDate.getFullYear();
  const month = currentDate.getMonth();

  const monthName = currentDate.toLocaleDateString('ru-RU', {
    month: 'long',
    year: 'numeric',
  });

  const today = new Date();
  const todayIso = today.toISOString().split('T')[0];

  // Days calculation
  const firstDayOfWeek = (new Date(year, month, 1).getDay() + 6) % 7; // Monday = 0
  const daysInMonth = new Date(year, month + 1, 0).getDate();

  const days: (number | null)[] = [];
  for (let i = 0; i < firstDayOfWeek; i++) {
    days.push(null);
  }
  for (let d = 1; d <= daysInMonth; d++) {
    days.push(d);
  }

  const weekDays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];

  const handlePrevMonth = () => {
    setCurrentDate(new Date(year, month - 1, 1));
  };

  const handleNextMonth = () => {
    setCurrentDate(new Date(year, month + 1, 1));
  };

  const handleDayClick = (day: number) => {
    const formattedMonth = String(month + 1).padStart(2, '0');
    const formattedDay = String(day).padStart(2, '0');
    const dateStr = `${year}-${formattedMonth}-${formattedDay}`;

    if (selectedTaskId) {
      // Inspector open: set task deadline
      updateTaskMutation.mutate({
        taskId: selectedTaskId,
        data: { dueDate: dateStr },
      });
    } else {
      // Inspector closed: toggle date filter
      setSelectedDateFilter(selectedDateFilter === dateStr ? null : dateStr);
    }
  };

  return (
    <div className="w-80 shrink-0 bg-surface border border-subtle rounded-2xl p-4 shadow-washi select-none space-y-3">
      {/* Month header & navigation */}
      <div className="flex items-center justify-between border-b border-subtle/50 pb-2">
        <div className="flex items-center gap-1.5">
          <span className="font-serif text-sm font-normal text-primary capitalize">
            {monthName}
          </span>
          {selectedDateFilter && (
            <button
              type="button"
              onClick={() => setSelectedDateFilter(null)}
              className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded bg-accent-surface text-[10px] text-primary hover:bg-accent-surface/70"
              title="Сбросить фильтр по дате"
            >
              <span>{selectedDateFilter}</span>
              <X className="w-2.5 h-2.5" />
            </button>
          )}
        </div>

        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={handlePrevMonth}
            className="p-1 rounded text-muted hover:text-primary hover:bg-canvas transition-colors"
            title="Предыдущий месяц"
          >
            <ChevronLeft className="w-3.5 h-3.5" />
          </button>
          <button
            type="button"
            onClick={handleNextMonth}
            className="p-1 rounded text-muted hover:text-primary hover:bg-canvas transition-colors"
            title="Следующий месяц"
          >
            <ChevronRight className="w-3.5 h-3.5" />
          </button>
        </div>
      </div>

      {/* Weekday headers */}
      <div className="grid grid-cols-7 gap-1 text-center">
        {weekDays.map((wd) => (
          <span key={wd} className="text-[10px] font-mono text-muted">
            {wd}
          </span>
        ))}
      </div>

      {/* Days grid */}
      <div className="grid grid-cols-7 gap-1 text-center">
        {days.map((day, idx) => {
          if (day === null) {
            return <div key={`empty-${idx}`} className="h-7" />;
          }

          const formattedMonth = String(month + 1).padStart(2, '0');
          const formattedDay = String(day).padStart(2, '0');
          const dateStr = `${year}-${formattedMonth}-${formattedDay}`;

          const isToday = dateStr === todayIso;
          const isFilterActive = selectedDateFilter === dateStr;

          // Density marker: count tasks due on this date
          const dueTasks = allTasks.filter((t) => t.dueDate === dateStr);
          const taskCount = dueTasks.length;

          return (
            <button
              key={day}
              type="button"
              onClick={() => handleDayClick(day)}
              className={`h-7 flex flex-col items-center justify-center text-xs font-mono rounded relative transition-all ${
                isToday
                  ? 'bg-accent-solid text-surface font-semibold shadow-washi-sm'
                  : isFilterActive
                  ? 'bg-accent-surface text-primary border border-accent-solid font-medium'
                  : 'text-primary hover:bg-canvas/60'
              }`}
              title={
                selectedTaskId
                  ? `Назначить срок: ${dateStr}`
                  : taskCount > 0
                  ? `${taskCount} задач со сроком ${dateStr}`
                  : `Фильтр по ${dateStr}`
              }
            >
              <span>{day}</span>
              {taskCount > 0 && (
                <span className="flex gap-0.5 mt-0.5">
                  {Array.from({ length: Math.min(taskCount, 3) }).map((_, dotIdx) => (
                    <span
                      key={dotIdx}
                      className={`w-1 h-1 rounded-full ${
                        isToday ? 'bg-surface' : 'bg-accent-solid'
                      }`}
                    />
                  ))}
                </span>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
};
