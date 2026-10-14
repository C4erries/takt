import React from 'react';
import { Inbox } from 'lucide-react';
import { cn } from '@/shared/lib/cn';

export interface EmptyStateProps {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  action?: React.ReactNode;
  className?: string;
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  icon,
  title,
  description,
  action,
  className,
}) => {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center text-center p-8 py-12 rounded-xl border border-dashed border-subtle bg-surface/50 max-w-md mx-auto',
        className
      )}
    >
      <div className="w-12 h-12 rounded-full bg-accent-surface/30 flex items-center justify-center text-text-muted mb-4 border border-subtle/50">
        {icon || <Inbox className="w-6 h-6 stroke-[1.5]" />}
      </div>
      <h4 className="font-serif text-2xl font-normal text-primary mb-1 tracking-tight">
        {title}
      </h4>
      {description && (
        <p className="text-xs text-muted max-w-xs mb-6 font-sans leading-relaxed">
          {description}
        </p>
      )}
      {action && <div className="flex items-center gap-3">{action}</div>}
    </div>
  );
};
