import React from 'react';
import { AlertCircle, RotateCcw } from 'lucide-react';
import { cn } from '@/shared/lib/cn';
import { Button } from './Button';

export interface ErrorBannerProps {
  title?: string;
  message: string;
  onRetry?: () => void;
  className?: string;
}

export const ErrorBanner: React.FC<ErrorBannerProps> = ({
  title = 'Что-то пошло не так',
  message,
  onRetry,
  className,
}) => {
  return (
    <div
      role="alert"
      className={cn(
        'flex items-start justify-between gap-3 p-4 rounded-xl border border-tag-persimmon/70 bg-tag-persimmon/20 text-primary transition-all duration-150',
        className
      )}
    >
      <div className="flex items-start gap-3">
        <div className="mt-0.5 rounded-full p-1 bg-tag-persimmon/50 text-text-primary">
          <AlertCircle className="w-4 h-4 stroke-[2]" />
        </div>
        <div className="flex flex-col">
          <span className="text-sm font-medium text-primary">{title}</span>
          <span className="text-xs text-muted mt-0.5 leading-relaxed">{message}</span>
        </div>
      </div>

      {onRetry && (
        <Button
          size="sm"
          variant="outline"
          onClick={onRetry}
          leftIcon={<RotateCcw className="w-3.5 h-3.5" />}
          className="shrink-0 bg-surface/90 hover:bg-surface text-xs"
        >
          Повторить
        </Button>
      )}
    </div>
  );
};
