import * as React from 'react';
import { motion } from 'framer-motion';
import type { TimelineResponse } from '@/hooks/useDeployments';
import { TimelineStep, type StepState, mapStatus } from '@/components/ui/timeline-step';
import { Badge } from '@/components/ui/badge';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { EmptyState } from '@/components/ui/empty-state';
import { Separator } from '@/components/ui/separator';
import {
  Activity,
  Clock,
  CheckCircle2,
  XCircle,
  RotateCcw,
  Timer,
} from 'lucide-react';

/* ── Stagger animation ─────────────────────────────────── */
const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { staggerChildren: 0.04 },
  },
};

/* ── Props ────────────────────────────────────────────── */
export interface TimelineStepContentProps {
  timeline: TimelineResponse | undefined;
  loading: boolean;
}

/* ── Component ────────────────────────────────────────── */
export function TimelineStepContent({ timeline, loading }: TimelineStepContentProps) {
  // ── Loading ──
  if (loading) {
    return (
      <div className="space-y-4">
        <Card>
          <CardHeader className="pb-3">
            <Skeleton className="h-5 w-48" />
          </CardHeader>
          <CardContent>
            <div className="space-y-5">
              {Array.from({ length: 6 }).map((_, i) => (
                <div key={i} className="flex gap-4">
                  <Skeleton className="h-4 w-4 rounded-full shrink-0 mt-0.5" />
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

  // ── Empty ──
  if (!timeline) {
    return (
      <EmptyState
        icon={Activity}
        title="No timeline data"
        description="Timeline data is not available for this deployment."
      />
    );
  }

  const steps = timeline.steps ?? [];
  const overallStatus = timeline.overallStatus ?? 'unknown';
  const overallSuccess = overallStatus.toLowerCase() === 'success';

  // Calculate total step duration from individual steps
  const succeededSteps = steps.filter((s) => mapStatus(s.status) === 'succeeded').length;
  const failedSteps = steps.filter((s) => mapStatus(s.status) === 'failed').length;

  return (
    <div className="space-y-5">
      {/* Header stats */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <CardTitle className="text-body font-semibold flex items-center gap-2">
                <Activity className="h-4 w-4 text-text-muted" />
                Execution Timeline
              </CardTitle>
            </div>
            <div className="flex items-center gap-2">
              <Badge variant={overallSuccess ? 'subtle-success' : 'subtle-danger'} className="gap-1.5">
                {overallSuccess
                  ? <CheckCircle2 className="h-3.5 w-3.5" />
                  : <XCircle className="h-3.5 w-3.5" />
                }
                {overallSuccess ? 'Succeeded' : 'Failed'}
              </Badge>
              {timeline.duration && (
                <Badge variant="subtle-neutral" className="gap-1">
                  <Timer className="h-3.5 w-3.5" />
                  {timeline.duration}
                </Badge>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4 text-small text-text-secondary">
            <span className="inline-flex items-center gap-1">
              <CheckCircle2 className="h-3.5 w-3.5 text-success" />
              {succeededSteps} of {steps.length} steps succeeded
            </span>
            {failedSteps > 0 && (
              <span className="inline-flex items-center gap-1">
                <XCircle className="h-3.5 w-3.5 text-danger" />
                {failedSteps} failed
              </span>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Timeline */}
      {steps.length === 0 ? (
        <EmptyState
          icon={Activity}
          title="No steps recorded"
          description="This deployment has no recorded steps."
        />
      ) : (
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          className="pl-1"
        >
          {steps.map((step, index) => (
            <TimelineStep
              key={step.id ?? index}
              state={mapStatus(step.status)}
              title={step.name}
              detail={step.result}
              error={step.error}
              isLast={index === steps.length - 1}
            />
          ))}
        </motion.div>
      )}
    </div>
  );
}
