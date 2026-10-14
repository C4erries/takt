import React, { useState } from 'react';
import { CheckCircle2, FolderOpen, Plus } from 'lucide-react';
import {
  Button,
  Skeleton,
  EmptyState,
  ErrorBanner,
} from '@/shared/components/ui';

export const StatesShowcase: React.FC = () => {
  const [stateView, setStateView] = useState<'success' | 'loading' | 'error' | 'empty'>('success');

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2 p-1.5 bg-canvas/70 rounded-xl border border-subtle w-fit flex-wrap">
        <Button
          size="sm"
          variant={stateView === 'success' ? 'solid' : 'ghost'}
          onClick={() => setStateView('success')}
        >
          1. Success State
        </Button>
        <Button
          size="sm"
          variant={stateView === 'loading' ? 'solid' : 'ghost'}
          onClick={() => setStateView('loading')}
        >
          2. Loading (Skeleton)
        </Button>
        <Button
          size="sm"
          variant={stateView === 'error' ? 'solid' : 'ghost'}
          onClick={() => setStateView('error')}
        >
          3. Error State
        </Button>
        <Button
          size="sm"
          variant={stateView === 'empty' ? 'solid' : 'ghost'}
          onClick={() => setStateView('empty')}
        >
          4. Empty State
        </Button>
      </div>

      <div className="p-6 rounded-2xl border border-subtle bg-surface">
        {stateView === 'success' && (
          <div className="space-y-4">
            <div className="flex items-center gap-2 text-accent-solid">
              <CheckCircle2 className="w-5 h-5 stroke-[2]" />
              <span className="text-sm font-medium">Контракт успешно синхронизирован</span>
            </div>
            <p className="text-xs text-muted">
              Все данные получены из API. Отображается готовый список задач и рабочий дашборд.
            </p>
          </div>
        )}

        {stateView === 'loading' && (
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              <Skeleton variant="circle" className="w-8 h-8" />
              <div className="space-y-1.5 flex-1">
                <Skeleton className="h-4 w-1/3" />
                <Skeleton className="h-3 w-1/4" />
              </div>
            </div>
            <div className="space-y-2 pt-2">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </div>
          </div>
        )}

        {stateView === 'error' && (
          <div className="space-y-4">
            <ErrorBanner
              title="Ошибка загрузки задач"
              message="Не удалось подключиться к серверу API (/api/v1/health). Проверьте сетевое соединение."
              onRetry={() => setStateView('success')}
            />
          </div>
        )}

        {stateView === 'empty' && (
          <EmptyState
            icon={<FolderOpen className="w-6 h-6 stroke-[1.5]" />}
            title="Входящие пусты"
            description="В вашей папке «Входящие» нет неразобранных задач. Нажмите кнопку ниже для быстрого захвата новой мысли."
            action={
              <Button variant="solid" leftIcon={<Plus className="w-4 h-4" />}>
                Новая задача
              </Button>
            }
          />
        )}
      </div>
    </div>
  );
};
