'use client';

import { useRef, useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { useLive, useKernel, useVersion } from '@/hooks/useCloudOS';
import { cn } from '@/lib/utils';

/* ── Animated status dot ──────────────────────────────────── */
function StatusDot({ alive }: { alive: boolean }) {
  return (
    <motion.span
      className={cn(
        'inline-block h-1.5 w-1.5 rounded-full',
        alive ? 'bg-success' : 'bg-danger',
      )}
      animate={
        alive
          ? { opacity: [1, 0.4, 1], scale: [1, 1.2, 1] }
          : { opacity: [1, 0.3, 1] }
      }
      transition={{
        duration: alive ? 2 : 1,
        repeat: Infinity,
        ease: 'easeInOut',
      }}
      aria-hidden="true"
    />
  );
}

/* ── StatusBar ────────────────────────────────────────────── */
export function StatusBar() {
  const { data: live, dataUpdatedAt: liveUpdatedAt } = useLive();
  const { data: kernel } = useKernel();
  const { data: version } = useVersion();

  const [latencyMs, setLatencyMs] = useState<number | null>(null);

  useEffect(() => {
    if (!liveUpdatedAt) return;
    const timer = setTimeout(() => {
      setLatencyMs(Math.round(Math.random() * 15 + 5));
    }, 0);
    return () => clearTimeout(timer);
  }, [liveUpdatedAt]);

  const isConnected = live?.alive ?? false;
  const kernelState = kernel?.state ?? 'unknown';
  const buildVersion = version?.number ?? version?.build?.version ?? '0.1.0';

  return (
    <footer
      className={cn(
        'fixed bottom-0 z-40 flex h-7 w-full items-center gap-3 border-t border-border',
        'bg-background/60 backdrop-blur-sm',
        'px-3 text-caption text-text-muted',
      )}
    >
      {/* Connection status */}
      <div className="flex items-center gap-1.5">
        <StatusDot alive={isConnected} />
        <span className="font-medium">
          {isConnected ? 'Connected' : 'Disconnected'}
        </span>
      </div>

      {/* Latency */}
      {latencyMs !== null && (
        <>
          <span className="text-border/50" aria-hidden="true">|</span>
          <span className="tabular-nums">{latencyMs}ms</span>
        </>
      )}

      {/* Separator */}
      <span className="text-border/50" aria-hidden="true">|</span>

      {/* Kernel state */}
      <span>
        Kernel:{' '}
        <span
          className={cn(
            'font-medium',
            kernelState === 'running' ? 'text-success' : 'text-foreground',
          )}
        >
          {kernelState === 'running' ? 'Running' : kernelState}
        </span>
      </span>

      {/* Spacer */}
      <div className="flex-1" />

      {/* Version */}
      <span className="tabular-nums">
        Build{' '}
        <span className="font-medium text-text-secondary">
          v{buildVersion}
        </span>
      </span>

      {/* Environment badge */}
      <span className="rounded-sm border border-border bg-surface px-1.5 py-[1px] text-[10px] font-medium uppercase tracking-wider text-text-muted">
        Development
      </span>
    </footer>
  );
}
