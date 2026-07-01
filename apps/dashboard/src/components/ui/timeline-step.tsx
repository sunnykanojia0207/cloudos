import * as React from 'react';
import { motion } from 'framer-motion';
import { Check, X, Circle, Minus } from 'lucide-react';
import { cn } from '@/lib/utils';

type StepState = 'pending' | 'running' | 'succeeded' | 'failed' | 'skipped' | 'cancelled';

interface TimelineStepProps {
  state: StepState;
  title: string;
  duration?: string;
  detail?: string;
  error?: string;
  isLast?: boolean;
  className?: string;
}

const STATE_CONFIG: Record<StepState, {
  icon: React.ElementType;
  color: string;
  bgColor: string;
  borderColor: string;
}> = {
  pending:   { icon: Circle,    color: 'text-text-muted',          bgColor: 'bg-surface',           borderColor: 'border-border' },
  running:   { icon: Circle,    color: 'text-accent',             bgColor: 'bg-accent',            borderColor: 'border-accent' },
  succeeded: { icon: Check,     color: 'text-success-foreground', bgColor: 'bg-success',           borderColor: 'border-success' },
  failed:    { icon: X,         color: 'text-danger-foreground',  bgColor: 'bg-danger',            borderColor: 'border-danger' },
  skipped:   { icon: Minus,     color: 'text-text-muted',         bgColor: 'bg-surface-elevated',  borderColor: 'border-border' },
  cancelled: { icon: Circle,    color: 'text-warning',            bgColor: 'bg-warning-subtle',    borderColor: 'border-warning' },
};

function TimelineStep({
  state,
  title,
  duration,
  detail,
  error: errorMsg,
  isLast,
  className,
}: TimelineStepProps) {
  const config = STATE_CONFIG[state];

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2, ease: [0, 0, 0.2, 1] }}
      className={cn('relative flex gap-4 pb-2', !isLast && 'pb-6', className)}
    >
      {/* Line + Node */}
      <div className="flex flex-col items-center">
        {/* Node */}
        <div
          className={cn(
            'relative z-10 flex h-4 w-4 items-center justify-center rounded-full border-2 shrink-0',
            config.bgColor,
            config.borderColor,
            state === 'running' && 'animate-pulse-accent',
          )}
          aria-label={state}
        >
          <config.icon className={cn('h-2.5 w-2.5', config.color)} />
        </div>

        {/* Connecting line */}
        {!isLast && (
          <div className="mt-1 w-px flex-1 bg-border" aria-hidden="true" />
        )}
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0 pb-2">
        <div className="flex items-center justify-between gap-2">
          <span className="text-body font-medium text-foreground">
            {title}
          </span>
          {duration && (
            <span className="shrink-0 text-caption text-text-muted tabular-nums">
              {duration}
            </span>
          )}
        </div>

        {detail && (
          <p className="mt-0.5 text-small text-text-secondary">
            {detail}
          </p>
        )}

        {errorMsg && (
          <p className="mt-1 text-small text-danger">
            → {errorMsg}
          </p>
        )}
      </div>
    </motion.div>
  );
}

/* ── Map API status string → StepState ─────────────────── */
function mapStatus(status: string): StepState {
  const lower = status?.toLowerCase() ?? '';
  if (['success', 'succeeded'].includes(lower)) return 'succeeded';
  if (['failure', 'failed', 'error'].includes(lower)) return 'failed';
  if (lower === 'running') return 'running';
  if (lower === 'skipped') return 'skipped';
  if (lower === 'cancelled') return 'cancelled';
  return 'pending';
}

export { TimelineStep, type StepState, mapStatus };
