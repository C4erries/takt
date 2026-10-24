import React, { useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { Sparkles, Plus, ShieldCheck } from 'lucide-react';
import {
  Button,
  Badge,
  Skeleton,
  ErrorBanner,
  EmptyState,
} from '@/shared/components/ui';
import {
  useAutomationsQuery,
  useToggleAutomationMutation,
  useDeleteAutomationMutation,
  useCreateAutomationMutation,
  type Automation,
} from '@/entities/automation';
import { DEMO_COLUMNS, type Column } from '@/entities/project';
import { AutomationCard } from './AutomationCard';
import { RuleBuilderModal, RuleDetailModal } from './RuleBuilderModal';

export interface AutomationsHubProps {
  workspaceId?: string;
}

type FilterCategory = 'all' | 'column' | 'global';

export const AutomationsHub: React.FC<AutomationsHubProps> = ({ workspaceId }) => {
  const queryClient = useQueryClient();
  const { data: automations = [], isLoading, isError, error, refetch } = useAutomationsQuery(workspaceId);
  const toggleMutation = useToggleAutomationMutation(workspaceId);
  const deleteMutation = useDeleteAutomationMutation(workspaceId);
  const createMutation = useCreateAutomationMutation(workspaceId);

  const [filter, setFilter] = useState<FilterCategory>('all');
  const [isBuilderOpen, setIsBuilderOpen] = useState(false);
  const [detailRule, setDetailRule] = useState<Automation | null>(null);
  const [editingRule, setEditingRule] = useState<Automation | null>(null);

  // Collect columns from cache or fallback to default project columns
  const cachedCols = queryClient
    .getQueriesData<Column[]>({ queryKey: ['columns'] })
    .flatMap(([, cols]) => cols ?? []);
  const allColumns = cachedCols.length > 0 ? cachedCols : DEMO_COLUMNS;

  const filteredAutomations = automations.filter((auto) => {
    if (filter === 'column') return Boolean(auto.columnId);
    if (filter === 'global') return !auto.columnId;
    return true;
  });

  const activeCount = automations.filter((a) => a.isEnabled).length;

  const handleToggle = (auto: Automation) => {
    toggleMutation.mutate({
      automationId: auto.id,
      isEnabled: !auto.isEnabled,
    });
  };

  const handleDelete = (id: string) => {
    deleteMutation.mutate(id);
  };

  const handleDuplicate = (rule: Automation) => {
    createMutation.mutate({
      name: `${rule.name} (копия)`,
      triggerType: rule.triggerType,
      boardId: rule.boardId,
      columnId: rule.columnId,
      conditionsJson: rule.conditionsJson,
      actionsJson: rule.actionsJson,
      isEnabled: rule.isEnabled,
      ruleDefinition: rule.ruleDefinition,
    });
  };

  const handleEdit = (rule: Automation) => {
    setEditingRule(rule);
    setIsBuilderOpen(true);
  };

  return (
    <div className="flex-1 flex flex-col h-full overflow-hidden p-6 md:p-8 space-y-5">
      {/* Header */}
      <div className="flex items-center justify-between gap-4 flex-wrap pb-4 border-b border-subtle">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="font-serif text-2xl md:text-3xl font-normal text-primary">
              Хаб автоматизаций
            </h1>
            <Badge variant="matcha" size="sm" dot>
              {activeCount} активных
            </Badge>
          </div>
          <p className="text-xs text-muted mt-1">
            Low-Code правила «КОГДА → ЕСЛИ → ТО», SLA таймеры старения и расписания
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="solid"
            size="sm"
            onClick={() => {
              setEditingRule(null);
              setIsBuilderOpen(true);
            }}
            leftIcon={<Plus className="w-3.5 h-3.5" />}
          >
            Собрать правило
          </Button>
        </div>
      </div>

      {/* Protection Banner */}
      <div className="flex items-center gap-2.5 p-3 rounded-xl bg-accent-surface/40 border border-subtle text-xs text-primary">
        <ShieldCheck className="w-4 h-4 text-accent-solid shrink-0" />
        <span>
          Правила срабатывают мгновенно и защищены от зацикливания.
        </span>
      </div>

      {/* Filter Tabs */}
      <div className="flex items-center gap-2">
        {(['all', 'column', 'global'] as FilterCategory[]).map((cat) => {
          const labels: Record<FilterCategory, string> = {
            all: `Все (${automations.length})`,
            column: `Колоночные (${automations.filter((a) => a.columnId).length})`,
            global: `Глобальные (${automations.filter((a) => !a.columnId).length})`,
          };
          const isSelected = filter === cat;
          return (
            <button
              key={cat}
              type="button"
              onClick={() => setFilter(cat)}
              className={`px-3 py-1.5 rounded-lg text-xs transition-colors duration-150 ${
                isSelected
                  ? 'bg-accent-surface font-medium text-primary shadow-washi-sm'
                  : 'text-muted hover:text-primary hover:bg-surface/60'
              }`}
            >
              {labels[cat]}
            </button>
          );
        })}
      </div>

      {/* State 1: Loading */}
      {isLoading && (
        <div className="space-y-3 flex-1 overflow-y-auto">
          {[1, 2, 3].map((i) => (
            <div key={i} className="p-4 rounded-xl border border-subtle bg-surface/50 space-y-2">
              <Skeleton variant="text" className="w-48 h-5" />
              <Skeleton variant="text" className="w-3/4 h-4" />
            </div>
          ))}
        </div>
      )}

      {/* State 2: Error */}
      {!isLoading && isError && (
        <div className="py-6">
          <ErrorBanner
            title="Ошибка загрузки автоматизаций"
            message={error?.message || 'Не удалось связаться с сервером'}
            onRetry={() => refetch()}
          />
        </div>
      )}

      {/* State 3: Empty State */}
      {!isLoading && !isError && filteredAutomations.length === 0 && (
        <div className="flex-1 flex items-center justify-center py-12">
          <EmptyState
            title="Нет активных правил автоматизации"
            description="Создайте первое Low-Code правило для автоматизации переноса задач, смены приоритетов или SLA таймеров"
            action={
              <Button
                variant="solid"
                size="sm"
                onClick={() => {
                  setEditingRule(null);
                  setIsBuilderOpen(true);
                }}
              >
                Собрать правило
              </Button>
            }
            icon={<Sparkles className="w-8 h-8 text-accent-solid opacity-60" />}
          />
        </div>
      )}

      {/* State 4: Success (List of Rules) */}
      {!isLoading && !isError && filteredAutomations.length > 0 && (
        <div className="space-y-3 flex-1 overflow-y-auto pr-1">
          {filteredAutomations.map((auto, index) => {
            const col = allColumns.find((c) => c.id === auto.columnId);
            return (
              <AutomationCard
                key={auto.id}
                automation={auto}
                index={index}
                onToggle={handleToggle}
                onDelete={handleDelete}
                onDetail={(r) => setDetailRule(r)}
                onEdit={handleEdit}
                onDuplicate={handleDuplicate}
                columnName={col?.name}
                allColumns={allColumns}
              />
            );
          })}
        </div>
      )}

      {/* Rule Builder Modal */}
      <RuleBuilderModal
        isOpen={isBuilderOpen}
        onClose={() => {
          setIsBuilderOpen(false);
          setEditingRule(null);
        }}
        workspaceId={workspaceId}
        allColumns={allColumns}
        mode={editingRule ? 'edit' : 'create'}
        initialRule={editingRule}
      />

      {/* Rule Detail Modal */}
      <RuleDetailModal
        isOpen={Boolean(detailRule)}
        onClose={() => setDetailRule(null)}
        rule={detailRule}
        allColumns={allColumns}
        onEdit={handleEdit}
        onDuplicate={handleDuplicate}
      />
    </div>
  );
};

