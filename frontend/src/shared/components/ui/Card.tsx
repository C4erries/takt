import React from 'react';
import { cn } from '@/shared/lib/cn';

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'sheet' | 'subtle' | 'flat';
}

export const Card = React.forwardRef<HTMLDivElement, CardProps>(
  ({ className, variant = 'sheet', children, ...props }, ref) => {
    const variants: Record<NonNullable<CardProps['variant']>, string> = {
      // Standard Washi sheet with subtle border and delicate shadow
      sheet: 'bg-surface border border-subtle shadow-washi rounded-sheet',
      // Subtle container without shadow
      subtle: 'bg-surface/70 border border-subtle/80 rounded-xl',
      // Flat card on canvas
      flat: 'bg-transparent border border-subtle rounded-xl',
    };

    return (
      <div
        ref={ref}
        className={cn(variants[variant], 'transition-all duration-150', className)}
        {...props}
      >
        {children}
      </div>
    );
  }
);
Card.displayName = 'Card';

export const CardHeader = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn('flex flex-col space-y-1.5 p-5 pb-3', className)}
      {...props}
    />
  )
);
CardHeader.displayName = 'CardHeader';

export const CardTitle = React.forwardRef<
  HTMLHeadingElement,
  React.HTMLAttributes<HTMLHeadingElement>
>(({ className, ...props }, ref) => (
  <h3
    ref={ref}
    className={cn(
      'font-serif text-2xl font-normal leading-none tracking-tight text-primary',
      className
    )}
    {...props}
  />
));
CardTitle.displayName = 'CardTitle';

export const CardDescription = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLParagraphElement>
>(({ className, ...props }, ref) => (
  <p
    ref={ref}
    className={cn('text-xs text-muted font-sans', className)}
    {...props}
  />
));
CardDescription.displayName = 'CardDescription';

export const CardContent = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={cn('p-5 pt-0', className)} {...props} />
  )
);
CardContent.displayName = 'CardContent';

export const CardFooter = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn('flex items-center p-5 pt-0 border-t border-subtle/40 mt-4', className)}
      {...props}
    />
  )
);
CardFooter.displayName = 'CardFooter';
