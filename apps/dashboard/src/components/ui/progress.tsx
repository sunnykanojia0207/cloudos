import * as React from 'react';
import { cn } from '@/lib/utils';

interface ProgressProps extends React.HTMLAttributes<HTMLDivElement> {
  value?: number;
  max?: number;
  variant?: 'default' | 'success' | 'warning' | 'danger';
}

const Progress = React.forwardRef<HTMLDivElement, ProgressProps>(
  ({ className, value = 0, max = 100, variant = 'default', ...props }, ref) => {
    const pct = Math.min(Math.max((value / max) * 100, 0), 100);
    const barColor =
      variant === 'success'
        ? 'bg-success'
        : variant === 'warning'
          ? 'bg-warning'
          : variant === 'danger'
            ? 'bg-danger'
            : 'bg-accent';

    return (
      <div
        ref={ref}
        role="progressbar"
        aria-valuenow={value}
        aria-valuemin={0}
        aria-valuemax={max}
        className={cn(
          'h-1.5 w-full overflow-hidden rounded-full bg-surface-elevated',
          className,
        )}
        {...props}
      >
        <div
          className={cn('h-full rounded-full transition-all duration-normal ease-standard', barColor)}
          style={{ width: `${pct}%` }}
        />
      </div>
    );
  },
);
Progress.displayName = 'Progress';

export { Progress };
