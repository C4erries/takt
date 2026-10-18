import React from 'react';
import { Zap, Clock, Sparkles, ArrowRight, Trash2 } from 'lucide-react';
import { Badge } from '@/shared/components/ui';
import type { Automation } from '@/entities/automation';
import { cn } from '@/shared/lib/cn';

export interface AutomationCardProps {
  automation: Automation;
  onToggle: (automation: Automation) => void;
  onDelete?: (automationId: string) => void;
  columnName?: string;
  isDeletable?: boolean;
}

export const AutomationCard: React.FC<AutomationCardProps> = ({
  automation,
  onToggle,
  onDelete,
  columnName,
  isDeletable = true,
}) => {
  const triggerType = automation.triggerType;
  const isTimeTrigger = triggerType === 'column_sla_timer' || triggerType === 'scheduled_sweep';

  const triggerLabel = (() => {
    switch (triggerType) {
      case 'task_entered_column':
        return columnName ? `Вход в [${columnName}]` : 'Вход в колонку';
      case 'task_completed_in_column':
        return columnName ? `Завершение в [${columnName}]` : 'Завершение в колонке';
      case 'column_sla_timer':
        return columnName ? `SLA таймер [${columnName}]` : 'SLA таймер';
      case 'scheduled_sweep':
        return columnName ? `Расписание [${columnName}]` : 'Ежедневное расписание';
      case 'task_field_changed':
        return 'Смена поля задачи';
      case 'task_completed':
        return 'Задача завершена';
      case 'task_updated':
        return 'Задача обновлена';
      default:
        return triggerType;
    }
  })();

  const conditions = automation.ruleDefinition?.conditions ?? [];
  const actions = automation.ruleDefinition?.actions ?? [];

  return (
    <div
      className={cn(
        'p-3.5 rounded-xl border transition-all duration-150 bg-surface/70',
        automation.isEnabled
          ? 'border-subtle shadow-washi-sm hover:border-subtle/90'
          : 'border-subtle/50 opacity-70'
      )}
    >
      <div className="flex items-start justify-between gap-3">
        {/* Left icon & details */}
        <div className="flex items-start gap-3 min-w-0 flex-1">
          <div className="w-8 h-8 rounded-lg bg-canvas border border-subtle flex items-center justify-center shrink-0 shadow-washi-sm mt-0.5">
            {isTimeTrigger ? (
              <Clock className="w-4 h-4 text-accent-solid" />
            ) : triggerType === 'task_completed' || triggerType === 'task_completed_in_column' ? (
              <Sparkles className="w-4 h-4 text-accent-solid" />
            ) : (
              <Zap className="w-4 h-4 text-accent-solid" />
            )}
          </div>

          <div className="space-y-1.5 min-w-0 flex-1">
            <div className="flex items-center gap-2 flex-wrap">
              <h4 className="font-serif text-sm font-normal text-primary">
                {automation.name}
              </h4>
              <Badge
                variant={automation.isEnabled ? 'matcha' : 'subtle'}
                size="sm"
                dot
              >
                {automation.isEnabled ? 'Вкл' : 'Выкл'}
              </Badge>
            </div>

            {/* Sentence Builder Visual Chain */}
            <div className="flex items-center gap-1.5 flex-wrap text-[11px] font-mono">
              <span className="text-muted text-[10px] uppercase">КОГДА</span>
              <span className="bg-canvas px-1.5 py-0.5 rounded border border-subtle text-primary">
                {triggerLabel}
              </span>

              {conditions.length > 0 && (
                <>
                  <span className="text-muted text-[10px] uppercase">ЕСЛИ</span>
                  {conditions.map((c, i) => (
                    <span
                      key={i}
                      className="bg-accent-surface/30 px-1.5 py-0.5 rounded border border-subtle text-primary"
                    >
                      {String(c.field)} {String(c.operator)} {c.value !== undefined ? String(c.value) : ''}
                    </span>
                  ))}
                </>
              )}

              {actions.length > 0 && (
                <>
                  <ArrowRight className="w-3 h-3 text-muted shrink-0" />
                  <span className="text-muted text-[10px] uppercase">ТО</span>
                  {actions.map((a, i) => (
                    <span
                      key={i}
                      className="bg-canvas px-1.5 py-0.5 rounded border border-subtle text-accent-solid font-medium"
                    >
                      {String(a.type)}
                    </span>
                  ))}
                </>
              )}
            </div>
          </div>
        </div>

        {/* Right actions: Toggle and Delete */}
        <div className="flex items-center gap-2 shrink-0">
          {isDeletable && onDelete && (
            <button
              type="button"
              onClick={() => onDelete(automation.id)}
              className="p-1 rounded text-muted hover:text-primary hover:bg-canvas transition-colors"
              title="Удалить правило"
            >
              <Trash2 className="w-3.5 h-3.5" />
            </button>
          )}

          <button
            type="button"
            role="switch"
            aria-checked={automation.isEnabled}
            onClick={() => onToggle(automation)}
            className={cn(
              'w-10 h-5 flex items-center rounded-full p-0.5 transition-colors duration-200 cursor-pointer',
              automation.isEnabled ? 'bg-accent-solid' : 'bg-canvas border border-subtle'
            )}
            title={automation.isEnabled ? 'Выключить' : 'Включить'}
          >
            <span
              className={cn(
                'w-4 h-4 rounded-full bg-surface shadow-washi-sm transition-transform duration-200',
                automation.isEnabled ? 'translate-x-5' : 'translate-x-0'
              )}
            />
          </button>
        </div>
      </div>
    </div>
  );
};
