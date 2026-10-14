import React, { forwardRef } from 'react';
import { Loader2 } from 'lucide-react';
import { cn } from '@/shared/lib/cn';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'solid' | 'outline' | 'ghost' | 'secondary';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      className,
      variant = 'solid',
      size = 'md',
      isLoading = false,
      disabled,
      children,
      leftIcon,
      rightIcon,
      ...props
    },
    ref
  ) => {
    const baseStyles =
      'inline-flex items-center justify-center font-medium transition-all duration-150 ease-spring-jk focus:outline-none focus-visible:ring-2 focus-visible:ring-accent-solid focus-visible:ring-offset-1 disabled:opacity-50 disabled:pointer-events-none select-none';

    const variantStyles: Record<NonNullable<ButtonProps['variant']>, string> = {
      // Solid matcha green
      solid:
        'bg-accent-solid text-surface hover:bg-accent-solid-hover active:scale-[0.98] shadow-washi-sm',
      // Clean 1px stationery outline
      outline:
        'border border-subtle bg-surface text-primary hover:bg-canvas hover:border-text-muted active:scale-[0.98]',
      // Subtle ghost for icon-buttons and calm toolbar actions
      ghost:
        'bg-transparent text-primary hover:bg-accent-surface/40 hover:text-primary active:scale-[0.98]',
      // Secondary matcha surface
      secondary:
        'bg-accent-surface text-primary hover:bg-accent-surface/80 active:scale-[0.98]',
    };

    const sizeStyles: Record<NonNullable<ButtonProps['size']>, string> = {
      sm: 'text-xs h-7 px-2.5 rounded-md gap-1.5',
      md: 'text-sm h-9 px-3.5 rounded-lg gap-2',
      lg: 'text-base h-11 px-5 rounded-lg gap-2.5',
    };

    return (
      <button
        ref={ref}
        disabled={disabled || isLoading}
        className={cn(baseStyles, variantStyles[variant], sizeStyles[size], className)}
        {...props}
      >
        {isLoading ? (
          <Loader2 className="w-4 h-4 animate-spin text-current" />
        ) : (
          leftIcon
        )}
        {children}
        {!isLoading && rightIcon}
      </button>
    );
  }
);

Button.displayName = 'Button';
