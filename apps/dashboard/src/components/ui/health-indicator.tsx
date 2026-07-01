import * as React from 'react';
import { cn } from '@/lib/utils';

type HealthStatus = 'healthy' | 'running' | 'degraded' | 'warning' | 'failed' | 'error' | 'stopped' | 'deploying' | 'pending';

interface HealthIndicatorProps {
  status: HealthStatus;
  showLabel?: boolean;
  size?: 'sm' | 'default' | 'lg';
  pulsing?: boolean;
  className?: string;
}

const STATUS_CONFIG: Record<HealthStatus, { color: string; label: string }> = {
  healthy:   { color: 'bg-success',                label: 'Healthy' },
  running:   { color: 'bg-success',                label: 'Running' },
  degraded:  { color: 'bg-warning',                label: 'Degraded' },
  warning:   { color: 'bg-warning',                label: 'Warning' },
  failed:    { color: 'bg-danger',                 label: 'Failed' },
  error:     { color: 'bg-danger',                 label: 'Error' },
  stopped:   { color: 'bg-text-muted',             label: 'Stopped' },
  deploying: { color: 'bg-accent',                 label: 'Deploying' },
  pending:   { color: 'bg-text-muted',             label: 'Pending' },
};

const SIZE_MAP = {
  sm: 'h-1.5 w-1.5',
  default: 'h-2 w-2',
  lg: 'h-2.5 w-2.5',
};

function HealthIndicator({
  status,
  showLabel = false,
  size = 'default',
  pulsing = false,
  className,
}: HealthIndicatorProps) {
  const config = STATUS_CONFIG[status] ?? STATUS_CONFIG.pending;

  return (
    <span className={cn('inline-flex items-center gap-1.5', className)}>
      <span
        className={cn(
          'inline-block shrink-0 rounded-full',
          SIZE_MAP[size],
          config.color,
          pulsing && 'animate-pulse-accent',
        )}
        aria-hidden="true"
      />
      {showLabel && (
        <span className="text-caption font-medium text-text-secondary">
          {config.label}
        </span>
      )}
    </span>
  );
}

export { HealthIndicator, type HealthStatus };
