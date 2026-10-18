import React, { useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { CheckCircle2, Zap, ArrowRightLeft, Sparkles, Plus } from 'lucide-react';
import { Modal, Badge, Skeleton, ErrorBanner, Button } from '@/shared/components/ui';
import {
  useAutomationsQuery,
  useToggleAutomationMutation,
  type Automation,
} from '@/entities/automation';
import { DEMO_COLUMNS, type Column } from '@/entities/project';
import { cn } from '@/shared/lib/cn';
import { RuleBuilderModal } from './RuleBuilderModal';

export interface AutomationsModalProps {
  isOpen: boolean;
  onClose: () => void;
  workspaceId?: string;
  onNavigateToHub?: () => void;
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
    description: 'Автоматический перенос выполненной задачи в колонку «Готово».',
    ruleTag: 'task_completed → колонка Готово',
  },
  'Auto-Focus': {
    icon: <Zap className="w-5 h-5 text-accent-solid" />,
    displayName: 'Auto-Focus',
    description: 'Задачи с дедлайном «Сегодня» автоматически всплывают в фокус дня.',
    ruleTag: 'due_date <= today → фокус дня',
  },
  'Inbox Zero': {
    icon: <ArrowRightLeft className="w-5 h-5 text-accent-solid" />,
    displayName: 'Inbox Zero',
    description: 'При назначении проекта задача перемещается в первую колонку проекта.',
    ruleTag: 'назначен проект → первая колонка',
  },
};

export const AutomationsModal: React.FC<AutomationsModalProps> = ({
  isOpen,
  onClose,
  workspaceId,
  onNavigateToHub,
}) => {
  const queryClient = useQueryClient();
  const { data: automations = [], isLoading, isError, error, refetch } = useAutomationsQuery(workspaceId);
  const toggleMutation = useToggleAutomationMutation(workspaceId);
  const [isBuilderOpen, setIsBuilderOpen] = useState(false);

  const cachedCols = queryClient
    .getQueriesData<Column[]>({ queryKey: ['columns'] })
    .flatMap(([, cols]) => cols ?? []);
  const allColumns = cachedCols.length > 0 ? cachedCols : DEMO_COLUMNS;

  const handleToggle = (automation: Automation) => {
    toggleMutation.mutate({
      automationId: automation.id,
      isEnabled: !automation.isEnabled,
    });
  };

  return (
    <>
      <Modal
        isOpen={isOpen}
        onClose={onClose}
        title="Автоматизации GTD"
        description="Фоновые правила продуктивности и Low-Code конструктор"
        className="max-w-xl"
      >
        <div className="space-y-4">
          <div className="flex items-center justify-between p-3 rounded-xl bg-accent-surface/40 border border-subtle text-xs text-primary">
            <div className="flex items-center gap-2">
              <Sparkles className="w-4 h-4 text-accent-solid shrink-0" />
              <span>Правила срабатывают мгновенно и защищены от зацикливания.</span>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setIsBuilderOpen(true)}
              leftIcon={<Plus className="w-3.5 h-3.5" />}
              className="text-[11px] h-7 bg-surface/80 shadow-washi-sm"
            >
              Собрать своё
            </Button>
          </div>

          {isLoading && (
            <div className="space-y-3 py-2">
              {[1, 2, 3].map((i) => (
                <div key={i} className="p-4 rounded-xl border border-subtle bg-canvas/40 space-y-2">
                  <div className="flex items-center justify-between">
                    <Skeleton variant="text" className="w-36 h-5" />
                    <Skeleton variant="rect" className="w-12 h-6 rounded-full" />
                  </div>
                  <Skeleton variant="text" className="w-4/5 h-4" />
                </div>
              ))}
            </div>
          )}

          {!isLoading && isError && (
            <div className="py-4">
              <ErrorBanner
                title="Не удалось загрузить автоматизации"
                message={error?.message || 'Ошибка сети'}
                onRetry={() => refetch()}
              />
            </div>
          )}

          {!isLoading && !isError && (
            <div className="space-y-3 max-h-[50vh] overflow-y-auto pr-1">
              {automations.map((auto) => {
                const meta = PRESET_METAS[auto.name] || {
                  icon: <Sparkles className="w-5 h-5 text-accent-solid" />,
                  displayName: auto.name,
                  description: 'Пользовательское Low-Code правило автоматизации GTD.',
                  ruleTag: auto.triggerType,
                };

                return (
                  <div
                    key={auto.id}
                    className={cn(
                      'p-3.5 rounded-xl border transition-all duration-150 bg-canvas/50 hover:bg-canvas/70',
                      auto.isEnabled ? 'border-subtle shadow-washi-sm' : 'border-subtle/60 opacity-80'
                    )}
                  >
                    <div className="flex items-start justify-between gap-4">
                      <div className="flex items-start gap-3 min-w-0">
                        <div className="w-8 h-8 rounded-lg bg-surface border border-subtle flex items-center justify-center shrink-0 shadow-washi-sm">
                          {meta.icon}
                        </div>
                        <div className="space-y-1 min-w-0">
                          <div className="flex items-center gap-2 flex-wrap">
                            <h4 className="font-serif text-sm font-normal text-primary">
                              {meta.displayName}
                            </h4>
                            <Badge variant={auto.isEnabled ? 'matcha' : 'subtle'} size="sm" dot>
                              {auto.isEnabled ? 'Активно' : 'Выкл'}
                            </Badge>
                          </div>
                          <p className="text-[11px] text-muted leading-relaxed">
                            {meta.description}
                          </p>
                          <div className="pt-0.5">
                            <span className="font-mono text-[10px] text-muted bg-surface/80 px-2 py-0.5 rounded border border-subtle/50 inline-block">
                              {meta.ruleTag}
                            </span>
                          </div>
                        </div>
                      </div>

                      <button
                        type="button"
                        role="switch"
                        aria-checked={auto.isEnabled}
                        onClick={() => handleToggle(auto)}
                        className={cn(
                          'w-10 h-5 shrink-0 flex items-center rounded-full p-0.5 transition-colors duration-200 cursor-pointer',
                          auto.isEnabled ? 'bg-accent-solid' : 'bg-surface border border-subtle'
                        )}
                        title={auto.isEnabled ? 'Выключить' : 'Включить'}
                      >
                        <span
                          className={cn(
                            'w-4 h-4 rounded-full bg-surface shadow-washi-sm transition-transform duration-200',
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

          {onNavigateToHub && (
            <div className="flex justify-between items-center pt-2 border-t border-subtle">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  onClose();
                  onNavigateToHub();
                }}
                className="text-xs text-muted hover:text-primary"
              >
                Открыть хаб автоматизаций →
              </Button>
              <Button variant="solid" size="sm" onClick={onClose}>
                Закрыть
              </Button>
            </div>
          )}
        </div>
      </Modal>

      <RuleBuilderModal
        isOpen={isBuilderOpen}
        onClose={() => setIsBuilderOpen(false)}
        workspaceId={workspaceId}
        allColumns={allColumns}
      />
    </>
  );
};
