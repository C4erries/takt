import React from 'react';
import { cn } from '@/shared/lib/cn';

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'persimmon' | 'denim' | 'straw' | 'matcha' | 'subtle';
  size?: 'sm' | 'md';
  dot?: boolean;
}

export const Badge: React.FC<BadgeProps> = ({
  className,
  variant = 'subtle',
  size = 'md',
  dot = false,
  children,
  ...props
}) => {
  const variantStyles: Record<NonNullable<BadgeProps['variant']>, string> = {
    // Persimmon / warm peach (@home)
    persimmon: 'bg-tag-persimmon/70 text-primary border-tag-persimmon',
    // Indigo denim / muted grey-blue (project)
    denim: 'bg-tag-denim/70 text-primary border-tag-denim',
    // Rice straw (time estimates)
    straw: 'bg-tag-straw/80 text-primary border-tag-straw',
    // Matcha accent
    matcha: 'bg-accent-surface text-primary border-accent-surface',
    // Neutral subtle washi border
    subtle: 'bg-surface text-muted border-subtle',
  };

  const dotColors: Record<NonNullable<BadgeProps['variant']>, string> = {
    persimmon: 'bg-primary/60',
    denim: 'bg-primary/60',
    straw: 'bg-primary/60',
    matcha: 'bg-accent-solid',
    subtle: 'bg-text-muted',
  };

  const sizeStyles: Record<NonNullable<BadgeProps['size']>, string> = {
    sm: 'text-[11px] px-2 py-0.2 gap-1',
    md: 'text-xs px-2.5 py-0.5 gap-1.5',
  };

  return (
    <span
      className={cn(
        'inline-flex items-center font-medium rounded-full border transition-colors select-none leading-normal',
        variantStyles[variant],
        sizeStyles[size],
        className
      )}
      {...props}
    >
      {dot && <span className={cn('w-1.5 h-1.5 rounded-full shrink-0', dotColors[variant])} />}
      {children}
    </span>
  );
};
