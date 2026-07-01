import * as React from 'react';
import { cn } from '@/lib/utils';

interface TerminalProps {
  children?: React.ReactNode;
  className?: string;
  maxHeight?: string;
}

interface TerminalLineProps {
  timestamp?: string;
  level?: 'info' | 'success' | 'warn' | 'error';
  source?: string;
  step?: string;
  message: string;
  className?: string;
}

const LEVEL_ICONS: Record<string, string> = {
  info: '•',
  success: '✓',
  warn: '⚠',
  error: '✗',
};

const LEVEL_COLORS: Record<string, string> = {
  info: 'text-text-muted',
  success: 'text-success',
  warn: 'text-warning',
  error: 'text-danger',
};

/* ── Terminal Container ───────────────────────────────── */
function Terminal({ children, className, maxHeight }: TerminalProps) {
  return (
    <div
      className={cn(
        'rounded-sm border border-border bg-terminal p-3',
        'font-mono text-code leading-relaxed text-terminal-fg',
        'overflow-auto',
        className,
      )}
      style={maxHeight ? { maxHeight } : undefined}
      role="log"
      aria-label="Terminal output"
    >
      {children}
    </div>
  );
}

/* ── Terminal Line ────────────────────────────────────── */
function TerminalLine({
  timestamp,
  level = 'info',
  source,
  step,
  message,
  className,
}: TerminalLineProps) {
  return (
    <div className={cn('flex items-start gap-2', className)}>
      {/* Timestamp */}
      {timestamp && (
        <span className="shrink-0 text-code-sm text-text-muted tabular-nums">
          {timestamp}
        </span>
      )}

      {/* Level icon */}
      <span
        className={cn(
          'shrink-0 w-3 text-center',
          LEVEL_COLORS[level] ?? LEVEL_COLORS.info,
        )}
        aria-label={level}
      >
        {LEVEL_ICONS[level] ?? LEVEL_ICONS.info}
      </span>

      {/* Source */}
      {source && (
        <span className="shrink-0 text-code-sm text-accent font-medium">
          {source}
        </span>
      )}

      {/* Step */}
      {step && (
        <span className="shrink-0 text-code-sm text-text-muted">
          [{step}]
        </span>
      )}

      {/* Message */}
      <span className="flex-1 text-terminal-fg break-all">
        {message}
      </span>
    </div>
  );
}

/* ── Terminal Status ──────────────────────────────────── */
interface TerminalStatusProps {
  streaming?: boolean;
  paused?: boolean;
  error?: boolean;
  className?: string;
}

function TerminalStatus({
  streaming,
  paused,
  error,
  className,
}: TerminalStatusProps) {
  if (streaming) {
    return (
      <div className={cn('mt-2 flex items-center gap-1.5 text-code-sm text-success', className)}>
        <span className="inline-block h-1.5 w-1.5 rounded-full bg-success animate-pulse" aria-hidden="true" />
        <span>Live</span>
      </div>
    );
  }

  if (paused) {
    return (
      <div className={cn('mt-2 flex items-center gap-1.5 text-code-sm text-warning', className)}>
        <span>⏸</span>
        <span>Paused</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className={cn('mt-2 flex items-center gap-1.5 text-code-sm text-danger', className)}>
        <span>Connection lost. Reconnecting...</span>
      </div>
    );
  }

  return (
    <div className={cn('mt-2 flex items-center gap-1.5 text-code-sm text-text-muted', className)}>
      <span>Waiting for logs...</span>
      <span className="inline-block h-1 w-1 rounded-full bg-text-muted animate-pulse" aria-hidden="true" />
    </div>
  );
}

export { Terminal, TerminalLine, TerminalStatus };
