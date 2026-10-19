import React, { forwardRef } from 'react';
import { Check } from 'lucide-react';
import { cn } from '@/shared/lib/cn';
import {
  motion,
  springTactile,
  checkboxPopScale,
  defaultFadeFallback,
  useReducedMotionPreference,
} from '@/shared/lib/motion';

export interface CheckboxProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type' | 'size'> {
  label?: React.ReactNode;
  description?: string;
}

export const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(
  ({ className, checked = false, disabled, label, description, onChange, id, ...props }, ref) => {
    const inputId = id || React.useId();
    const prefersReduced = useReducedMotionPreference();

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
          <motion.div
            animate={
              checked && !prefersReduced
                ? { scale: [...checkboxPopScale] }
                : { scale: 1 }
            }
            whileTap={prefersReduced || disabled ? undefined : { scale: 0.92 }}
            transition={springTactile}
            className={cn(
              'w-[18px] h-[18px] rounded-full border border-subtle bg-surface transition-colors duration-150 flex items-center justify-center',
              'group-hover:border-accent-solid/60',
              checked
                ? 'bg-accent-solid border-accent-solid'
                : 'bg-surface border-subtle',
              'peer-focus-visible:ring-2 peer-focus-visible:ring-accent-solid peer-focus-visible:ring-offset-1',
              className
            )}
          >
            <motion.span
              initial={false}
              animate={
                checked
                  ? { scale: prefersReduced ? 1 : [0, 1.1, 1], opacity: 1 }
                  : { scale: 0, opacity: 0 }
              }
              transition={
                prefersReduced
                  ? defaultFadeFallback
                  : { duration: 0.18, ease: 'easeOut' }
              }
              className="flex items-center justify-center pointer-events-none"
            >
              <Check className="w-3 h-3 text-surface stroke-[2.5]" />
            </motion.span>
          </motion.div>
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
