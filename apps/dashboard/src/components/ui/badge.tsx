import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const badgeVariants = cva(
  'inline-flex items-center gap-1 rounded-sm border px-2 py-0.5 text-badge font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2',
  {
    variants: {
      variant: {
        // Filled variants — strong emphasis
        'filled-success':
          'border-transparent bg-success text-success-foreground',
        'filled-warning':
          'border-transparent bg-warning text-warning-foreground',
        'filled-danger':
          'border-transparent bg-danger text-danger-foreground',
        'filled-info':
          'border-transparent bg-info text-info-foreground',
        'filled-neutral':
          'border-transparent bg-text-muted text-text-inverse',
        'filled-accent':
          'border-transparent bg-accent text-accent-foreground',

        // Subtle variants — moderate emphasis
        'subtle-success':
          'border-transparent bg-success-subtle text-success',
        'subtle-warning':
          'border-transparent bg-warning-subtle text-warning',
        'subtle-danger':
          'border-transparent bg-danger-subtle text-danger',
        'subtle-info':
          'border-transparent bg-info-subtle text-info',
        'subtle-neutral':
          'border border-border text-text-muted',
        'subtle-accent':
          'border-transparent bg-accent-subtle text-accent',

        // Legacy aliases for backward compatibility
        default:
          'border-transparent bg-accent text-accent-foreground',
        secondary:
          'border-transparent bg-accent-subtle text-accent',
        destructive:
          'border-transparent bg-danger text-danger-foreground',
        outline:
          'border-border text-text-secondary',
        success:
          'border-transparent bg-success-subtle text-success',
        warning:
          'border-transparent bg-warning-subtle text-warning',
      },
    },
    defaultVariants: {
      variant: 'subtle-neutral',
    },
  },
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}

export { Badge, badgeVariants };

/* ── StatusDot — inline status indicator ──────────────── */
interface StatusDotProps {
  status: 'success' | 'warning' | 'danger' | 'stopped' | 'deploying' | 'pending';
  className?: string;
  pulsing?: boolean;
}

function StatusDot({ status, className, pulsing }: StatusDotProps) {
  const colorMap: Record<string, string> = {
    success: 'bg-success',
    warning: 'bg-warning',
    danger: 'bg-danger',
    stopped: 'bg-text-muted',
    deploying: 'bg-accent',
    pending: 'bg-text-muted',
  };

  return (
    <span
      className={cn(
        'inline-block h-2 w-2 rounded-full shrink-0',
        colorMap[status] ?? 'bg-text-muted',
        pulsing && 'animate-pulse-accent',
        className,
      )}
      aria-hidden="true"
    />
  );
}

export { StatusDot };
