import React from 'react';

export const MiniCalendar: React.FC = () => {
  const today = new Date();
  const currentMonthName = today.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' });
  const currentDay = today.getDate();

  // Days in month calculation
  const year = today.getFullYear();
  const month = today.getMonth();
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

  return (
    <div className="w-72 shrink-0 bg-surface border border-subtle rounded-2xl p-4 shadow-washi select-none space-y-3">
      {/* Month header */}
      <div className="flex items-center justify-between border-b border-subtle/50 pb-2">
        <span className="font-serif text-sm font-normal text-primary capitalize">
          {currentMonthName}
        </span>
        <span className="text-[10px] font-mono uppercase tracking-wider text-muted">
          Hobonichi
        </span>
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
            return <div key={`empty-${idx}`} className="h-6" />;
          }

          const isToday = day === currentDay;

          return (
            <div
              key={day}
              className={`h-6 flex items-center justify-center text-xs font-mono rounded transition-colors ${
                isToday
                  ? 'bg-accent-solid text-surface font-semibold shadow-washi-sm'
                  : 'text-primary hover:bg-canvas/60'
              }`}
            >
              {day}
            </div>
          );
        })}
      </div>
    </div>
  );
};
