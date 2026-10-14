import React from 'react';
import { CheckCircle2, Zap, ArrowRightLeft, Sparkles } from 'lucide-react';
import { Modal, Badge, Skeleton, ErrorBanner } from '@/shared/components/ui';
import {
  useAutomationsQuery,
  useToggleAutomationMutation,
  type Automation,
} from '@/entities/automation';
import { cn } from '@/shared/lib/cn';

export interface AutomationsModalProps {
  isOpen: boolean;
  onClose: () => void;
  workspaceId?: string;
}

interface PresetMeta {
  icon: React.ReactNode;
  displayName: string;
  description: string;
  ruleTag: string;
}

const PRESET_METAS: Record<string, PresetMeta> = {
  'Done & Archive': {
    icon: <CheckCircle2 className="w-5 h-5 text-accent-solid" />,
    displayName: 'Done & Archive',
    description:
      'Автоматический перенос выполненной задачи в колонку «Готово» с фиксацией времени завершения.',
    ruleTag: 'task_completed → колонка Готово',
  },
  'Auto-Focus': {
    icon: <Zap className="w-5 h-5 text-accent-solid" />,
    displayName: 'Auto-Focus',
    description:
      'Задачи с дедлайном «Сегодня» автоматически всплывают в фокус дня для концентрации внимания.',
    ruleTag: 'due_date <= today → фокус дня',
  },
  'Inbox Zero': {
    icon: <ArrowRightLeft className="w-5 h-5 text-accent-solid" />,
    displayName: 'Inbox Zero',
    description:
      'При назначении проекта задача из «Входящих» автоматически перемещается в первую колонку проекта («К выполнению»).',
    ruleTag: 'назначен проект → первая колонка',
  },
};

export const AutomationsModal: React.FC<AutomationsModalProps> = ({
  isOpen,
  onClose,
  workspaceId,
}) => {
  const { data: automations = [], isLoading, isError, error, refetch } = useAutomationsQuery(workspaceId);
  const toggleMutation = useToggleAutomationMutation(workspaceId);

  const handleToggle = (automation: Automation) => {
    toggleMutation.mutate({
      automationId: automation.id,
      isEnabled: !automation.isEnabled,
    });
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Автоматизации GTD"
      description="Фоновые правила продуктивности для бесшовного порядка в задачах и проектах"
      className="max-w-xl"
    >
      <div className="space-y-4">
        {/* Banner about EventBus */}
        <div className="flex items-start gap-3 p-3 rounded-xl bg-accent-surface/40 border border-subtle/60 text-xs text-primary leading-relaxed">
          <Sparkles className="w-4 h-4 text-accent-solid shrink-0 mt-0.5" />
          <p>
            Автоматизации срабатывают асинхронно через шину событий <span className="font-mono text-[11px] text-muted">EventBus</span> с защитой от зацикливания (Max Depth = 1).
          </p>
        </div>

        {/* State 1: Loading */}
        {isLoading && (
          <div className="space-y-3 py-2">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="p-4 rounded-xl border border-subtle/60 bg-canvas/40 space-y-2"
              >
                <div className="flex items-center justify-between">
                  <Skeleton variant="text" className="w-36 h-5" />
                  <Skeleton variant="rect" className="w-12 h-6 rounded-full" />
                </div>
                <Skeleton variant="text" className="w-4/5 h-4" />
              </div>
            ))}
          </div>
        )}

        {/* State 2: Error */}
        {!isLoading && isError && (
          <div className="py-4">
            <ErrorBanner
              title="Не удалось загрузить автоматизации"
              message={error?.message || 'Ошибка сети'}
              onRetry={() => refetch()}
            />
          </div>
        )}

        {/* State 3: Automation Cards */}
        {!isLoading && !isError && (
          <div className="space-y-3">
            {automations.map((auto) => {
              const meta = PRESET_METAS[auto.name] || {
                icon: <Sparkles className="w-5 h-5 text-accent-solid" />,
                displayName: auto.name,
                description: 'Пользовательское правило автоматизации GTD.',
                ruleTag: auto.triggerType,
              };

              return (
                <div
                  key={auto.id}
                  className={cn(
                    'p-4 rounded-xl border transition-all duration-150',
                    'bg-canvas/50 hover:bg-canvas/70',
                    auto.isEnabled
                      ? 'border-subtle shadow-washi-sm'
                      : 'border-subtle/60 opacity-80'
                  )}
                >
                  <div className="flex items-start justify-between gap-4">
                    {/* Left: Icon & Meta */}
                    <div className="flex items-start gap-3 min-w-0">
                      <div className="w-9 h-9 rounded-lg bg-surface border border-subtle/80 flex items-center justify-center shrink-0 shadow-washi-sm">
                        {meta.icon}
                      </div>

                      <div className="space-y-1 min-w-0">
                        <div className="flex items-center gap-2 flex-wrap">
                          <h4 className="font-serif text-base font-normal text-primary">
                            {meta.displayName}
                          </h4>
                          <Badge
                            variant={auto.isEnabled ? 'matcha' : 'subtle'}
                            size="sm"
                            dot
                          >
                            {auto.isEnabled ? 'Активно' : 'Выключено'}
                          </Badge>
                        </div>

                        <p className="text-xs text-muted leading-relaxed">
                          {meta.description}
                        </p>

                        <div className="pt-1">
                          <span className="font-mono text-[10px] text-muted bg-surface/80 px-2 py-0.5 rounded border border-subtle/50 inline-block">
                            {meta.ruleTag}
                          </span>
                        </div>
                      </div>
                    </div>

                    {/* Right: Toggle Switch */}
                    <button
                      type="button"
                      role="switch"
                      aria-checked={auto.isEnabled}
                      onClick={() => handleToggle(auto)}
                      className={cn(
                        'w-11 h-6 shrink-0 flex items-center rounded-full p-0.5 transition-colors duration-200 cursor-pointer focus:outline-none focus-visible:ring-2 focus-visible:ring-accent-solid',
                        auto.isEnabled ? 'bg-accent-solid' : 'bg-subtle'
                      )}
                      title={auto.isEnabled ? 'Выключить правило' : 'Включить правило'}
                    >
                      <span
                        className={cn(
                          'w-5 h-5 rounded-full bg-surface shadow-washi-sm transition-transform duration-200',
                          auto.isEnabled ? 'translate-x-5' : 'translate-x-0'
                        )}
                      />
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </Modal>
  );
};
