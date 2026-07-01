import { useApplications } from '@/hooks/useApplications';
import { ApplicationCard } from '@/components/applications/ApplicationCard';
import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Layers, RotateCcw, AlertCircle } from 'lucide-react';
import { motion, type Variants } from 'framer-motion';

// ── Animation Variants ────────────────────────────────────────────────────

const containerVariants: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { staggerChildren: 0.06 },
  },
};

const fadeUpVariants: Variants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.4, ease: [0.25, 0.1, 0.25, 1] },
  },
};

const cardVariants: Variants = {
  hidden: { opacity: 0, y: 16 },
  visible: (i: number) => ({
    opacity: 1,
    y: 0,
    transition: { duration: 0.35, delay: i * 0.05, ease: [0.25, 0.1, 0.25, 1] },
  }),
};

// ── Loading Skeleton Grid ─────────────────────────────────────────────────

function ApplicationSkeletonGrid() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      {Array.from({ length: 4 }).map((_, i) => (
        <motion.div
          key={i}
          variants={cardVariants}
          initial="hidden"
          animate="visible"
          custom={i}
        >
          <Card className="border-border/50">
            <CardHeader className="pb-3">
              <div className="flex items-start justify-between gap-3">
                <div className="flex items-center gap-2.5 min-w-0">
                  <Skeleton className="h-9 w-9 rounded-lg shrink-0" />
                  <div className="min-w-0 space-y-1.5">
                    <Skeleton className="h-4 w-32" />
                    <Skeleton className="h-3 w-20" />
                  </div>
                </div>
                <Skeleton className="h-5 w-16 rounded-full shrink-0" />
              </div>
            </CardHeader>
            <CardContent className="space-y-2.5">
              <div className="flex flex-wrap items-center gap-2">
                <Skeleton className="h-5 w-14 rounded-full" />
                <Skeleton className="h-5 w-16 rounded-md" />
                <Skeleton className="h-5 w-20 rounded-md" />
              </div>
              <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
                <Skeleton className="h-3 w-16" />
                <Skeleton className="h-3 w-12" />
                <Skeleton className="h-3 w-14" />
              </div>
            </CardContent>
            <div className="mx-6 h-px bg-border/50" />
            <div className="p-4 pt-3">
              <div className="flex items-center gap-1.5">
                <Skeleton className="h-8 w-16 rounded-md" />
                <Skeleton className="h-8 w-16 rounded-md" />
                <Skeleton className="h-8 w-16 rounded-md" />
                <Skeleton className="h-8 w-16 rounded-md" />
              </div>
            </div>
          </Card>
        </motion.div>
      ))}
    </div>
  );
}

// ── Empty State ───────────────────────────────────────────────────────────

function EmptyState() {
  return (
    <motion.div
      variants={fadeUpVariants}
      initial="hidden"
      animate="visible"
      className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border/40 bg-card/10 px-6 py-20 text-center"
    >
      <div className="mb-5 flex h-14 w-14 items-center justify-center rounded-2xl border border-border/40 bg-muted/30">
        <Layers className="h-7 w-7 text-muted-foreground/50" />
      </div>
      <h3 className="mb-1.5 text-lg font-semibold tracking-tight text-foreground/90">
        No applications yet
      </h3>
      <p className="mb-6 max-w-sm text-sm leading-relaxed text-muted-foreground">
        Applications deployed through CloudOS will appear here. Create a project
        and deploy an application to get started.
      </p>
      <Button variant="outline" size="sm" className="gap-1.5" asChild>
        <a href="/projects">
          <Layers className="h-3.5 w-3.5" />
          Go to Projects
        </a>
      </Button>
    </motion.div>
  );
}

// ── Error State ───────────────────────────────────────────────────────────

function ErrorState({
  message,
  onRetry,
}: {
  message: string;
  onRetry: () => void;
}) {
  return (
    <motion.div
      variants={fadeUpVariants}
      initial="hidden"
      animate="visible"
      className="flex flex-col items-center justify-center rounded-xl border border-dashed border-red-500/20 bg-red-500/5 px-6 py-16 text-center"
    >
      <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-xl border border-red-500/20 bg-red-500/10">
        <AlertCircle className="h-6 w-6 text-red-400" />
      </div>
      <h3 className="mb-1.5 text-base font-semibold text-foreground/90">
        Failed to load applications
      </h3>
      <p className="mb-4 max-w-md text-sm text-muted-foreground">
        {message}
      </p>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          className="gap-1.5"
          onClick={onRetry}
        >
          <RotateCcw className="h-3.5 w-3.5" />
          Retry
        </Button>
        <Button variant="ghost" size="sm" className="gap-1.5" asChild>
          <a href="/">
            <Layers className="h-3.5 w-3.5" />
            Go to Dashboard
          </a>
        </Button>
      </div>
    </motion.div>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────

export default function ApplicationsPage() {
  const {
    data: applications,
    isLoading,
    error,
    refetch,
  } = useApplications();

  const appCount = applications?.length ?? 0;

  return (
    <motion.div
      className="flex flex-col gap-6 p-6"
      variants={containerVariants}
      initial="hidden"
      animate="visible"
    >
      {/* ── Page Header ─────────────────────────────────────────────── */}
      <motion.div variants={fadeUpVariants} className="flex flex-col gap-1">
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">
            Applications
          </h1>
          {!isLoading && !error && (
            <span className="inline-flex items-center rounded-md border border-border/50 bg-muted/30 px-2 py-0.5 text-xs font-medium text-muted-foreground tabular-nums">
              {appCount} {appCount === 1 ? 'app' : 'apps'}
            </span>
          )}
        </div>
        <p className="text-sm text-muted-foreground">
          CloudOS
        </p>
      </motion.div>

      {/* ── Loading State ───────────────────────────────────────────── */}
      {isLoading && <ApplicationSkeletonGrid />}

      {/* ── Error State ─────────────────────────────────────────────── */}
      {!isLoading && error && (
        <ErrorState
          message={
            (error as Error)?.message ||
            'An unexpected error occurred while loading applications.'
          }
          onRetry={() => refetch()}
        />
      )}

      {/* ── Empty State ─────────────────────────────────────────────── */}
      {!isLoading && !error && appCount === 0 && <EmptyState />}

      {/* ── Application Grid ────────────────────────────────────────── */}
      {!isLoading && !error && appCount > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {applications!.map((app, i) => (
            <motion.div
              key={app.metadata.id}
              variants={cardVariants}
              initial="hidden"
              animate="visible"
              custom={i}
            >
              <ApplicationCard app={app} />
            </motion.div>
          ))}
        </div>
      )}
    </motion.div>
  );
}
