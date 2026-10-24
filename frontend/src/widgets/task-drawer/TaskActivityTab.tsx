import React from 'react';
import {
  ArrowRightLeft,
  CheckCircle2,
  AlertCircle,
  Calendar,
  Tag,
  Type,
  ListPlus,
  Archive,
  RotateCcw,
  Clock,
  Bot,
  User,
} from 'lucide-react';
import { Skeleton, ErrorBanner, EmptyState } from '@/shared/components/ui';
import type { TaskActivity } from '@/entities/task';
import { useTaskActivitiesQuery } from '@/entities/task';

export interface TaskActivityTabProps {
  workspaceId?: string;
  taskId?: string;
  isActive?: boolean;
}

const PRIORITY_NAMES: Record<string, string> = {
  urgent: 'В ожидании',
  pending: 'В ожидании',
  high: 'Высокий',
  medium: 'Средний',
  low: 'Низкий',
  none: 'Обычный',
};

function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  if (isNaN(date.getTime())) return dateStr;

  const now = new Date();
  const diffSec = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (diffSec < 60) return 'только что';
  if (diffSec < 3600) {
    const mins = Math.floor(diffSec / 60);
    return `${mins} мин назад`;
  }
  if (diffSec < 86400) {
    const hours = Math.floor(diffSec / 3600);
    return `${hours} ч назад`;
  }

  const timeStr = date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
  const startOfYesterday = startOfToday - 86400 * 1000;
  const itemTime = date.getTime();

  if (itemTime >= startOfYesterday && itemTime < startOfToday) {
    return `вчера в ${timeStr}`;
  }

  return `${date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })} в ${timeStr}`;
}

function getActivityIcon(type: string) {
  switch (type) {
    case 'column_moved':
      return <ArrowRightLeft className="w-3.5 h-3.5 text-accent-solid" />;
    case 'status_changed':
      return <CheckCircle2 className="w-3.5 h-3.5 text-accent-solid" />;
    case 'priority_changed':
      return <AlertCircle className="w-3.5 h-3.5 text-accent-solid" />;
    case 'due_date_set':
    case 'due_date_cleared':
      return <Calendar className="w-3.5 h-3.5 text-accent-solid" />;
    case 'tag_added':
    case 'tag_removed':
      return <Tag className="w-3.5 h-3.5 text-accent-solid" />;
    case 'title_changed':
      return <Type className="w-3.5 h-3.5 text-accent-solid" />;
    case 'subtask_added':
      return <ListPlus className="w-3.5 h-3.5 text-accent-solid" />;
    case 'task_archived':
      return <Archive className="w-3.5 h-3.5 text-accent-solid" />;
    case 'task_unarchived':
      return <RotateCcw className="w-3.5 h-3.5 text-accent-solid" />;
    default:
      return <Clock className="w-3.5 h-3.5 text-muted" />;
  }
}

function renderActivityDescription(activity: TaskActivity) {
  const p = activity.payload || {};

  switch (activity.activityType) {
    case 'column_moved':
      if (p.from_column && p.to_column) {
        return (
          <span>
            Перенесена из колонки <span className="font-medium text-primary">«{String(p.from_column)}»</span> в{' '}
            <span className="font-medium text-primary">«{String(p.to_column)}»</span>
          </span>
        );
      }
      return <span>Перемещена в колонку <span className="font-medium text-primary">«{String(p.to_column || '')}»</span></span>;

    case 'status_changed':
      return p.is_completed ? (
        <span className="text-primary font-medium">Отмечена как выполненная</span>
      ) : (
        <span>Снята отметка о выполнении</span>
      );

    case 'priority_changed': {
      const fromP = PRIORITY_NAMES[String(p.from_priority || '')] || String(p.from_priority || '');
      const toP = PRIORITY_NAMES[String(p.to_priority || '')] || String(p.to_priority || '');
      return (
        <span>
          Приоритет изменён с <span className="font-medium text-primary">{fromP || '—'}</span> на{' '}
          <span className="font-medium text-primary">{toP}</span>
        </span>
      );
    }

    case 'due_date_set': {
      const date = p.due_date ? new Date(String(p.due_date)).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' }) : '';
      return <span>Назначен срок выполнения: <span className="font-medium text-primary">{date}</span></span>;
    }

    case 'due_date_cleared':
      return <span>Срок выполнения очищен</span>;

    case 'tag_added':
      return (
        <span className="inline-flex items-center gap-1">
          Добавлен тег{' '}
          <span className="font-medium text-primary inline-flex items-center gap-1 bg-canvas px-1.5 py-0.5 rounded border border-subtle">
            {Boolean(p.color_hex) && (
              <span className="w-1.5 h-1.5 rounded-full" style={{ backgroundColor: String(p.color_hex) }} />
            )}
            «{String(p.tag_name || '')}»
          </span>
        </span>
      );

    case 'tag_removed':
      return <span>Удалён тег <span className="font-medium text-primary">«{String(p.tag_name || '')}»</span></span>;

    case 'title_changed':
      return (
        <span>
          Название изменено на <span className="font-medium text-primary">«{String(p.to_title || '')}»</span>
        </span>
      );

    case 'subtask_added':
      return (
        <span>
          Создана подзадача <span className="font-medium text-primary">«{String(p.title || '')}»</span>
        </span>
      );

    case 'task_archived':
      return <span>Задача сгружена в архив</span>;

    case 'task_unarchived':
      return (
        <span>
          Задача восстановлена из архива
          {p.column_name ? <span> в колонку <span className="font-medium text-primary">«{String(p.column_name)}»</span></span> : ''}
        </span>
      );

    default:
      return <span>{activity.activityType}</span>;
  }
}

export const TaskActivityTab: React.FC<TaskActivityTabProps> = ({
  workspaceId,
  taskId,
  isActive = true,
}) => {
  const { data: activities = [], isLoading, isError, error, refetch } = useTaskActivitiesQuery(
    workspaceId,
    taskId,
    isActive
  );

  if (isLoading) {
    return (
      <div className="space-y-4 py-2">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="flex items-start gap-3">
            <Skeleton variant="circle" className="w-7 h-7 shrink-0 mt-0.5" />
            <div className="space-y-1.5 flex-1">
              <Skeleton variant="text" className="w-3/4 h-3.5" />
              <Skeleton variant="text" className="w-1/3 h-3" />
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (isError) {
    return (
      <div className="py-4">
        <ErrorBanner
          title="Не удалось загрузить историю"
          message={error?.message || 'Ошибка связи с сервером'}
          onRetry={() => refetch()}
        />
      </div>
    );
  }

  if (activities.length === 0) {
    return (
      <div className="py-8">
        <EmptyState
          title="История изменений пуста"
          description="Перемещение между колонками, смена дедлайна, статуса и тегов будут отражаться здесь."
          icon={<Clock className="w-7 h-7 text-muted opacity-60" />}
        />
      </div>
    );
  }

  return (
    <div className="relative pl-4 space-y-4 py-1">
      {/* Vertical Timeline Guide Line */}
      <div className="absolute left-6 top-3 bottom-3 w-px bg-subtle" />

      {activities.map((act) => {
        const isSystem = !act.userId;
        const author = act.userName || (isSystem ? 'Система' : 'Пользователь');

        return (
          <div key={act.id} className="relative flex items-start gap-3 group">
            {/* Event Icon Node */}
            <div className="relative z-10 w-6 h-6 rounded-full bg-surface border border-subtle flex items-center justify-center shrink-0 shadow-washi-sm mt-0.5">
              {getActivityIcon(act.activityType)}
            </div>

            {/* Event Details */}
            <div className="flex-1 min-w-0 text-xs space-y-0.5">
              <div className="text-primary leading-snug">
                {renderActivityDescription(act)}
              </div>

              <div className="flex items-center gap-1.5 text-[11px] font-mono text-muted">
                <span className="inline-flex items-center gap-1">
                  {isSystem ? <Bot className="w-3 h-3 text-muted/80" /> : <User className="w-3 h-3 text-muted/80" />}
                  <span>{author}</span>
                </span>
                <span>•</span>
                <span>{formatRelativeTime(act.createdAt)}</span>
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
};
