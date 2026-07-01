import * as React from 'react';
import { motion } from 'framer-motion';
import type { TimelineResponse } from '@/hooks/useDeployments';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import {
  CheckCircle2,
  XCircle,
  Clock,
  AlertTriangle,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';

/* ── Helpers ──────────────────────────────────────────── */

const statusConfig: Record<
  string,
  { icon: React.ReactNode; color: string; bgColor: string }
> = {
  success: {
    icon: <CheckCircle2 className="h-4 w-4" />,
    color: 'text-emerald-400',
    bgColor: 'bg-emerald-500/10',
  },
  failure: {
    icon: <XCircle className="h-4 w-4" />,
    color: 'text-red-400',
    bgColor: 'bg-red-500/10',
  },
  error: {
    icon: <XCircle className="h-4 w-4" />,
    color: 'text-red-400',
    bgColor: 'bg-red-500/10',
  },
  running: {
    icon: (
      <span className="relative flex h-4 w-4 items-center justify-center">
        <span className="absolute inline-flex h-3 w-3 animate-ping rounded-full bg-amber-400/60" />
        <span className="relative inline-flex h-2 w-2 rounded-full bg-amber-400" />
      </span>
    ),
    color: 'text-amber-400',
    bgColor: 'bg-amber-500/10',
  },
  pending: {
    icon: <Clock className="h-4 w-4" />,
    color: 'text-muted-foreground',
    bgColor: 'bg-muted/40',
  },
  skipped: {
    icon: <Clock className="h-4 w-4" />,
    color: 'text-muted-foreground/50',
    bgColor: 'bg-muted/20',
  },
};

function getStepStatusConfig(status: string) {
  const lower = status?.toLowerCase() ?? '';
  return (
    statusConfig[lower] ?? {
      icon: <Clock className="h-4 w-4" />,
      color: 'text-muted-foreground',
      bgColor: 'bg-muted/40',
    }
  );
}

/* ── Props ────────────────────────────────────────────── */

export interface TimelineTabProps {
  timeline: TimelineResponse | undefined;
  loading: boolean;
}

/* ── Stagger Animation Variants ───────────────────────── */

const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.06,
    },
  },
};

const stepVariants = {
  hidden: { opacity: 0, x: -12 },
  visible: {
    opacity: 1,
    x: 0,
    transition: { duration: 0.3, ease: 'easeOut' as const },
  },
};

/* ── Expandable Error ─────────────────────────────────── */

function ExpandableError({ message }: { message: string }) {
  const [expanded, setExpanded] = React.useState(false);

  if (!message) return null;

  return (
    <div className="mt-2">
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className={cn(
          'inline-flex items-center gap-1 text-xs font-medium',
          'text-red-400/80 hover:text-red-400 transition-colors',
          'focus-visible:outline-none focus-visible:underline',
        )}
      >
        {expanded ? (
          <ChevronDown className="h-3 w-3" />
        ) : (
          <ChevronRight className="h-3 w-3" />
        )}
        {expanded ? 'Hide error' : 'Show error'}
      </button>
      {expanded && (
        <pre className="mt-1 rounded-md bg-red-950/50 border border-red-500/20 p-3 text-xs text-red-300 font-mono whitespace-pre-wrap overflow-x-auto">
          {message}
        </pre>
      )}
    </div>
  );
}

/* ── Component ────────────────────────────────────────── */

export function TimelineTab({ timeline, loading }: TimelineTabProps) {
  // ── Loading state ──
  if (loading) {
    return (
      <div className="space-y-4">
        <Card className="border-border/50">
          <CardHeader className="pb-3">
            <Skeleton className="h-5 w-48" />
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="flex gap-4">
                  <Skeleton className="h-8 w-8 rounded-full shrink-0" />
                  <div className="flex-1 space-y-2">
                    <Skeleton className="h-4 w-32" />
                    <Skeleton className="h-3 w-20" />
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // ── Empty state ──
  if (!timeline) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <Clock className="h-10 w-10 text-muted-foreground/30 mb-3" />
        <p className="text-sm text-muted-foreground">No timeline data available.</p>
        <p className="text-xs text-muted-foreground/60 mt-1">
          Deploy the application to see the deployment timeline.
        </p>
      </div>
    );
  }

  const steps = timeline.steps ?? [];
  const overallStatus = timeline.overallStatus ?? 'unknown';
  const overallSuccess = overallStatus.toLowerCase() === 'success';

  return (
    <div className="space-y-6">
      {/* Header */}
      <Card className="border-border/50">
        <CardHeader className="pb-3">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <CardTitle className="text-base font-semibold">
                Deployment{' '}
                <span className="text-primary">#{timeline.deploymentNumber}</span>
              </CardTitle>
              {timeline.workflowId && (
                <span className="text-xs text-muted-foreground">
                  Workflow: {timeline.workflowId}
                </span>
              )}
            </div>
            <div className="flex items-center gap-2">
              <Badge
                variant={overallSuccess ? 'success' : 'destructive'}
                className="gap-1.5"
              >
                {overallSuccess ? (
                  <CheckCircle2 className="h-3.5 w-3.5" />
                ) : (
                  <XCircle className="h-3.5 w-3.5" />
                )}
                {overallSuccess ? 'Success' : 'Failed'}
              </Badge>
              {timeline.duration && (
                <Badge variant="secondary" className="gap-1">
                  <Clock className="h-3.5 w-3.5" />
                  {timeline.duration}
                </Badge>
              )}
            </div>
          </div>
        </CardHeader>
      </Card>

      {/* Timeline */}
      {steps.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <AlertTriangle className="h-8 w-8 text-muted-foreground/30 mb-2" />
          <p className="text-sm text-muted-foreground">No steps recorded in this deployment.</p>
        </div>
      ) : (
        <motion.div
          className="relative"
          variants={containerVariants}
          initial="hidden"
          animate="visible"
        >
          {steps.map((step, index) => {
            const stepConfig = getStepStatusConfig(step.status);
            const isLast = index === steps.length - 1;

            return (
              <motion.div
                key={step.id ?? index}
                variants={stepVariants}
                className="relative flex gap-4 pb-6 last:pb-0"
              >
                {/* Vertical connector line + dot */}
                <div className="flex flex-col items-center">
                  {/* Dot */}
                  <div
                    className={cn(
                      'relative z-10 flex h-8 w-8 shrink-0 items-center justify-center rounded-full border',
                      stepConfig.bgColor,
                      stepConfig.color,
                      'border-border/50',
                    )}
                  >
                    {stepConfig.icon}
                  </div>

                  {/* Connecting line */}
                  {!isLast && (
                    <div className="mt-1 w-px flex-1 bg-border/60" />
                  )}
                </div>

                {/* Content */}
                <div className="min-w-0 flex-1 pb-4">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-medium text-foreground/90">
                      {step.name}
                    </span>
                    <Badge
                      variant={
                        step.status === 'success'
                          ? 'success'
                          : step.status === 'running'
                            ? 'warning'
                            : step.status === 'failure' || step.status === 'error'
                              ? 'destructive'
                              : 'secondary'
                      }
                      className="text-[10px] px-1.5 py-0 h-4"
                    >
                      {step.status}
                    </Badge>
                  </div>

                  {step.result && (
                    <p className="mt-1 text-xs text-muted-foreground">
                      {step.result}
                    </p>
                  )}

                  {step.error && <ExpandableError message={step.error} />}
                </div>
              </motion.div>
            );
          })}
        </motion.div>
      )}
    </div>
  );
}
