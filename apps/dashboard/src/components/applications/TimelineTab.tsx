import * as React from 'react';
import { motion } from 'framer-motion';
import type { TimelineResponse } from '@/hooks/useDeployments';
import { TimelineStep, type StepState, mapStatus } from '@/components/ui/timeline-step';
import { Badge } from '@/components/ui/badge';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { EmptyState } from '@/components/ui/empty-state';
import {
  Activity,
  Clock,
  CheckCircle2,
  XCircle,
} from 'lucide-react';

/* ── Stagger Animation ─────────────────────────────────── */
const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { staggerChildren: 0.05 },
  },
};

/* ── Props ────────────────────────────────────────────── */
export interface TimelineTabProps {
  timeline: TimelineResponse | undefined;
  loading: boolean;
}

/* ── Component ────────────────────────────────────────── */
export function TimelineTab({ timeline, loading }: TimelineTabProps) {
  // ── Loading state ──
  if (loading) {
    return (
      <div className="space-y-4">
        <Card>
          <CardHeader className="pb-3">
            <Skeleton className="h-5 w-48" />
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
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

  // ── Empty state ──
  if (!timeline) {
    return (
      <EmptyState
        icon={Activity}
        title="No timeline data"
        description="Deploy the application to see the deployment timeline."
      />
    );
  }

  const steps = timeline.steps ?? [];
  const overallStatus = timeline.overallStatus ?? 'unknown';
  const overallSuccess = overallStatus.toLowerCase() === 'success';

  return (
    <div className="space-y-5">
      {/* Header card */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <CardTitle className="text-body font-semibold">
                Deployment{' '}
                <span className="text-accent">#{timeline.deploymentNumber}</span>
              </CardTitle>
              {timeline.workflowId && (
                <span className="text-caption text-text-muted hidden sm:inline">
                  Workflow: {timeline.workflowId}
                </span>
              )}
            </div>
            <div className="flex items-center gap-2">
              <Badge
                variant={overallSuccess ? 'subtle-success' : 'subtle-danger'}
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
                <Badge variant="subtle-neutral" className="gap-1">
                  <Clock className="h-3.5 w-3.5" />
                  {timeline.duration}
                </Badge>
              )}
            </div>
          </div>
        </CardHeader>
      </Card>

      {/* Timeline steps */}
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
