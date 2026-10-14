import React, { forwardRef } from 'react';
import { cn } from '@/shared/lib/cn';

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  leftAdornment?: React.ReactNode;
  rightAdornment?: React.ReactNode;
  error?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, type = 'text', leftAdornment, rightAdornment, error, disabled, ...props }, ref) => {
    return (
      <div className="w-full flex flex-col gap-1.5">
        <div
          className={cn(
            'group relative flex items-center w-full rounded-lg bg-surface border border-subtle transition-all duration-150 ease-spring-jk',
            'focus-within:border-accent-solid focus-within:ring-1 focus-within:ring-accent-solid/30',
            error && 'border-tag-persimmon focus-within:border-tag-persimmon focus-within:ring-tag-persimmon/40',
            disabled && 'opacity-60 bg-canvas cursor-not-allowed'
          )}
        >
          {leftAdornment && (
            <div className="flex items-center pl-3 text-muted select-none pointer-events-none">
              {leftAdornment}
            </div>
          )}

          <input
            ref={ref}
            type={type}
            disabled={disabled}
            className={cn(
              'w-full bg-transparent px-3 py-2 text-sm text-primary placeholder:text-muted/60 focus:outline-none disabled:cursor-not-allowed',
              leftAdornment && 'pl-2',
              rightAdornment && 'pr-2',
              className
            )}
            {...props}
          />

          {rightAdornment && (
            <div className="flex items-center pr-3 text-muted">{rightAdornment}</div>
          )}
        </div>

        {error && <span className="text-xs text-text-muted px-1">{error}</span>}
      </div>
    );
  }
);

Input.displayName = 'Input';
