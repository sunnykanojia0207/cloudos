import * as React from 'react';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';

/* ── Types ─────────────────────────────────────────────── */
export interface GraphNodeProps {
  label: string;
  status: 'success' | 'failed' | 'running' | 'pending' | 'skipped';
  isLast?: boolean;
}

export type GraphNodeStatus = GraphNodeProps['status'];

export interface NodeDefinition {
  label: string;
  key: string;
}

/* ── Node status config ────────────────────────────────── */
const NODE_STATUS: Record<string, { color: string; bg: string; label: string }> = {
  success: { color: 'border-success bg-success/10', bg: 'bg-success', label: 'Success' },
  failed:  { color: 'border-danger bg-danger/10',   bg: 'bg-danger',  label: 'Failed' },
  running: { color: 'border-accent bg-accent/10',   bg: 'bg-accent',  label: 'Running' },
  pending: { color: 'border-border bg-surface',      bg: 'bg-text-muted', label: 'Pending' },
  skipped: { color: 'border-border bg-surface-elevated', bg: 'bg-text-muted', label: 'Skipped' },
};

/* ── Default node definitions ──────────────────────────── */
export const DEFAULT_NODES: NodeDefinition[] = [
  { label: 'Validate',     key: 'validate' },
  { label: 'Clone',        key: 'clone' },
  { label: 'Detect',       key: 'detect' },
  { label: 'Install',      key: 'install' },
  { label: 'Build',        key: 'build' },
  { label: 'Deploy',       key: 'deploy' },
  { label: 'Health Check', key: 'health' },
  { label: 'Complete',     key: 'complete' },
];

/* ── GraphNode Component ───────────────────────────────── */
export function GraphNode({ label, status, isLast }: GraphNodeProps) {
  const cfg = NODE_STATUS[status] ?? NODE_STATUS.pending;

  return (
    <div className="flex items-start gap-3">
      {/* Node + line */}
      <div className="flex flex-col items-center">
        <div className={cn(
          'flex h-5 w-5 items-center justify-center rounded-full border-2 shrink-0',
          cfg.color,
          status === 'running' && 'animate-pulse-accent',
        )}>
          <div className={cn('h-2 w-2 rounded-full', cfg.bg)} />
        </div>
        {!isLast && <div className="mt-0.5 w-px flex-1 bg-border min-h-[24px]" aria-hidden="true" />}
      </div>

      {/* Label + status */}
      <div className="flex items-center gap-2 min-w-0 py-0.5">
        <span className="text-small text-foreground font-medium">{label}</span>
        <Badge variant={
          status === 'success' ? 'subtle-success' :
          status === 'failed' ? 'subtle-danger' :
          status === 'running' ? 'subtle-accent' :
          'subtle-neutral'
        } className="text-caption font-medium">
          {NODE_STATUS[status]?.label ?? status}
        </Badge>
      </div>
    </div>
  );
}
