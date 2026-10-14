import React, { forwardRef } from 'react';
import { Check } from 'lucide-react';
import { cn } from '@/shared/lib/cn';

export interface CheckboxProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type' | 'size'> {
  label?: React.ReactNode;
  description?: string;
}

export const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(
  ({ className, checked, disabled, label, description, onChange, id, ...props }, ref) => {
    const inputId = id || React.useId();

    return (
      <label
        htmlFor={inputId}
        className={cn(
          'inline-flex items-start gap-2.5 cursor-pointer select-none group',
          disabled && 'opacity-50 cursor-not-allowed'
        )}
      >
        <div className="relative flex items-center justify-center pt-0.5">
          <input
            id={inputId}
            ref={ref}
            type="checkbox"
            checked={checked}
            disabled={disabled}
            onChange={onChange}
            className="peer sr-only"
            {...props}
          />
          <div
            className={cn(
              'w-[18px] h-[18px] rounded-full border border-subtle bg-surface transition-all duration-150 ease-spring-jk flex items-center justify-center',
              'group-hover:border-accent-solid/60',
              'peer-checked:bg-accent-solid peer-checked:border-accent-solid',
              'peer-focus-visible:ring-2 peer-focus-visible:ring-accent-solid peer-focus-visible:ring-offset-1',
              className
            )}
          >
            <Check
              className={cn(
                'w-3 h-3 text-surface stroke-[2.5] transition-transform duration-150 scale-0',
                checked && 'scale-100'
              )}
            />
          </div>
        </div>

        {(label || description) && (
          <div className="flex flex-col text-sm leading-tight">
            {label && (
              <span
                className={cn(
                  'text-primary font-normal transition-colors',
                  checked && 'line-through text-muted'
                )}
              >
                {label}
              </span>
            )}
            {description && (
              <span className="text-xs text-muted mt-0.5">{description}</span>
            )}
          </div>
        )}
      </label>
    );
  }
);

Checkbox.displayName = 'Checkbox';
