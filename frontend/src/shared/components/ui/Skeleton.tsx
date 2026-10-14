import React from 'react';
import { cn } from '@/shared/lib/cn';

export interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'rect' | 'circle' | 'text';
}

export const Skeleton: React.FC<SkeletonProps> = ({
  className,
  variant = 'rect',
  ...props
}) => {
  const variantStyles: Record<NonNullable<SkeletonProps['variant']>, string> = {
    rect: 'rounded-lg',
    circle: 'rounded-full',
    text: 'rounded h-4 w-3/4',
  };

  return (
    <div
      className={cn(
        'animate-pulse bg-accent-surface/30 border border-subtle/40',
        variantStyles[variant],
        className
      )}
      {...props}
    />
  );
};
