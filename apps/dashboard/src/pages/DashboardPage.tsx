import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { usePageTitle } from '@/hooks/usePageTitle';
import {
  useHealth,
  useKernel,
  useVersion,
  useSystem,
  useControllers,
  useProjects,
} from '@/hooks/useCloudOS';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { cn } from '@/lib/utils';
import type { ReactNode } from 'react';
import {
  Activity,
  Box,
  Cpu,
  Heart,
  Layers,
  Terminal,
  Globe,
  HardDrive,
  Plus,
  XCircle,
} from 'lucide-react';

// ── Animation Variants ────────────────────────────────────────────────────

const containerVariants = {
  hidden: {},
  visible: {
    transition: {
      staggerChildren: 0.08,
    },
  },
};

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.4, ease: [0.25, 0.1, 0.25, 1] as const },
  },
};

// ── Helpers ────────────────────────────────────────────────────────────────

function formatUptime(uptime?: string): string {
  if (!uptime) return '\u2014';
  // uptime is typically like "2h45m10s" or "3d12h"
  return uptime;
}

function getStatusBg(status: string): string {
  switch (status) {
    case 'healthy':
    case 'running':
    case 'active':
      return 'bg-emerald-500';
    case 'degraded':
    case 'warning':
      return 'bg-amber-500';
    case 'unhealthy':
    case 'error':
    case 'stopped':
    default:
      return 'bg-red-500';
  }
}

// ── Sub-components ─────────────────────────────────────────────────────────

function StatusDot({ status, className }: { status: string; className?: string }) {
  return (
    <span
      className={cn(
        'inline-block h-2 w-2 rounded-full',
        getStatusBg(status),
        className,
      )}
      aria-hidden="true"
    />
  );
}

function StatCard({
  title,
  value,
  subtitle,
  icon: Icon,
  loading,
  error,
  action,
}: {
  title: string;
  value?: ReactNode;
  subtitle?: ReactNode;
  icon: React.ElementType;
  loading?: boolean;
  error?: boolean;
  action?: ReactNode;
}) {
  return (
    <motion.div variants={itemVariants}>
      <Card className="group relative overflow-hidden border-border/50 bg-card/50 backdrop-blur-sm transition-colors hover:border-border/80">
        <div className="pointer-events-none absolute inset-0 bg-gradient-to-br from-transparent via-transparent to-card-foreground/[0.02]" />
        <CardHeader className="flex flex-row items-start justify-between pb-2">
          <CardTitle className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
            {title}
          </CardTitle>
          <Icon className="h-4 w-4 text-muted-foreground/60" />
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="space-y-2">
              <Skeleton className="h-7 w-20" />
              {subtitle !== undefined && <Skeleton className="h-4 w-32" />}
            </div>
          ) : error ? (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <XCircle className="h-3.5 w-3.5 text-red-400" />
              <span>Unavailable</span>
            </div>
          ) : (
            <div className="flex items-end justify-between">
              <div>
                <div className="text-2xl font-semibold tracking-tight">
                  {value ?? '\u2014'}
                </div>
                {subtitle && (
                  <p className="mt-0.5 text-xs text-muted-foreground">
                    {subtitle}
                  </p>
                )}
              </div>
              {action && <div className="shrink-0">{action}</div>}
            </div>
          )}
        </CardContent>
      </Card>
    </motion.div>
  );
}

function HealthBadge({ status }: { status: string }) {
  const variant =
    status === 'healthy' || status === 'running' || status === 'active'
      ? 'success'
      : status === 'degraded' || status === 'warning'
        ? 'warning'
        : 'destructive';

  return <Badge variant={variant}>{status}</Badge>;
}

function QuickActionButton({
  icon: Icon,
  label,
  onClick,
}: {
  icon: React.ElementType;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="flex w-full items-center gap-3 rounded-lg border border-border/50 bg-card/30 px-4 py-3 text-sm font-medium text-foreground/80 transition-all hover:border-border/80 hover:bg-card/60 hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
    >
      <span className="flex h-7 w-7 items-center justify-center rounded-md bg-muted/50">
        <Icon className="h-3.5 w-3.5 text-muted-foreground" />
      </span>
      <span>{label}</span>
    </button>
  );
}

// ── Loading Skeletons ──────────────────────────────────────────────────────

function StatCardsSkeleton() {
  return (
    <>
      {Array.from({ length: 4 }).map((_, i) => (
        <motion.div key={i} variants={itemVariants}>
          <Card className="border-border/50 bg-card/50">
            <CardHeader className="flex flex-row items-start justify-between pb-2">
              <Skeleton className="h-3 w-24" />
              <Skeleton className="h-4 w-4" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-7 w-20" />
              <Skeleton className="mt-1.5 h-3 w-28" />
            </CardContent>
          </Card>
        </motion.div>
      ))}
    </>
  );
}

// ── Main Page ──────────────────────────────────────────────────────────────

export default function DashboardPage() {
  usePageTitle('Dashboard');
  const navigate = useNavigate();

  const { data: health, isLoading: healthLoading, error: healthError } = useHealth();
  const { data: kernel, isLoading: kernelLoading, error: kernelError } = useKernel();
  const { data: version, isLoading: versionLoading } = useVersion();
  const { data: system, isLoading: systemLoading } = useSystem();
  const { data: controllers, isLoading: controllersLoading, error: controllersError } = useControllers();
  const { data: projects, isLoading: projectsLoading, error: projectsError } = useProjects();

  const healthComponents = health?.components ? Object.entries(health.components) : [];
  const controllerCount = controllers?.total ?? controllers?.controllers?.length ?? 0;
  const projectCount = projects?.metadata?.total ?? projects?.items?.length ?? 0;

  const allRunning =
    controllers?.controllers?.every((c) => c.state === 'running') ?? false;
  const allControllersHealthy =
    controllers?.controllers?.every(
      (c) => c.health?.state === 'healthy' || c.health?.state === 'running',
    ) ?? false;

  return (
    <motion.div
      className="mx-auto max-w-6xl space-y-8 pb-12 pt-2"
      variants={containerVariants}
      initial="hidden"
      animate="visible"
    >
      {/* ── Welcome Section ────────────────────────────────────────────── */}
      <motion.div variants={itemVariants} className="flex flex-col gap-1.5">
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">
            Welcome to CloudOS
          </h1>
          {versionLoading ? (
            <Skeleton className="h-5 w-20 rounded-full" />
          ) : version?.number ? (
            <Badge
              variant="secondary"
              className="gap-1 rounded-full px-2.5 py-0.5 text-[11px] font-medium"
            >
              <Layers className="h-3 w-3" />
              v{version.number}
            </Badge>
          ) : null}
        </div>
        <p className="text-sm text-muted-foreground sm:text-base">
          Your infrastructure operating system
        </p>
      </motion.div>

      {/* ── Stats Grid (2×2) ───────────────────────────────────────────── */}
      <section>
        <div className="mb-3 flex items-center gap-2">
          <h2 className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
            Overview
          </h2>
          <Separator className="flex-1" />
        </div>

        {kernelLoading && healthLoading && controllersLoading && projectsLoading ? (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <StatCardsSkeleton />
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {/* 1. Kernel State */}
            <StatCard
              title="Kernel State"
              icon={Cpu}
              loading={kernelLoading}
              error={!!kernelError}
              value={
                kernel?.state ? (
                  <span className="flex items-center gap-2">
                    <StatusDot status={kernel.state} />
                    <span className="capitalize">{kernel.state}</span>
                  </span>
                ) : undefined
              }
              subtitle={
                kernel?.uptime
                  ? `Uptime ${formatUptime(kernel.uptime)}`
                  : undefined
              }
            />

            {/* 2. System Health */}
            <StatCard
              title="System Health"
              icon={Heart}
              loading={healthLoading}
              error={!!healthError}
              value={
                health?.overall?.status ? (
                  <span className="flex items-center gap-2">
                    <StatusDot status={health.overall.status} />
                    <span className="capitalize">{health.overall.status}</span>
                  </span>
                ) : undefined
              }
              subtitle={
                healthComponents.length > 0
                  ? `${healthComponents.length} component${healthComponents.length === 1 ? '' : 's'}`
                  : undefined
              }
            />

            {/* 3. Controllers */}
            <StatCard
              title="Controllers"
              icon={Terminal}
              loading={controllersLoading}
              error={!!controllersError}
              value={
                !controllersLoading
                  ? String(controllerCount)
                  : undefined
              }
              subtitle={
                !controllersLoading && controllerCount > 0
                  ? allControllersHealthy
                    ? 'All healthy'
                    : allRunning
                      ? 'All running'
                      : `${controllers?.controllers?.filter((c) => c.state === 'running').length ?? 0} running`
                  : undefined
              }
            />

            {/* 4. Projects */}
            <StatCard
              title="Projects"
              icon={Box}
              loading={projectsLoading}
              error={!!projectsError}
              value={
                !projectsLoading
                  ? String(projectCount)
                  : undefined
              }
              subtitle={
                !projectsLoading && projectCount > 0
                  ? `${projects?.items?.filter((p) => p.status?.phase === 'Active').length ?? 0} active`
                  : projectCount === 0
                    ? 'No projects yet'
                    : undefined
              }
              action={
                !projectsLoading ? (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7 rounded-md text-muted-foreground/60 hover:text-foreground"
                    onClick={() => navigate('/projects')}
                    aria-label="Create new project"
                  >
                    <Plus className="h-3.5 w-3.5" />
                  </Button>
                ) : undefined
              }
            />
          </div>
        )}
      </section>

      {/* ── System Health + Quick Actions (side by side) ───────────────── */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* System Health */}
        <motion.div variants={itemVariants} className="lg:col-span-2">
          <Card className="border-border/50 bg-card/50 backdrop-blur-sm">
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Activity className="h-4 w-4 text-muted-foreground" />
                  <CardTitle className="text-sm font-medium uppercase tracking-wider text-muted-foreground">
                    System Health
                  </CardTitle>
                </div>
                {healthLoading ? (
                  <Skeleton className="h-6 w-24 rounded-full" />
                ) : health?.overall?.status ? (
                  <div className="flex items-center gap-2">
                    <StatusDot status={health.overall.status} className="h-2.5 w-2.5" />
                    <Badge
                      variant={
                        health.overall.status === 'healthy'
                          ? 'success'
                          : health.overall.status === 'degraded'
                            ? 'warning'
                            : 'destructive'
                      }
                      className="text-[11px] font-semibold uppercase tracking-wider"
                    >
                      {health.overall.status}
                    </Badge>
                  </div>
                ) : healthError ? (
                  <Badge variant="destructive" className="text-[11px]">
                    Unreachable
                  </Badge>
                ) : null}
              </div>
            </CardHeader>
            <CardContent>
              {healthLoading ? (
                <div className="space-y-3">
                  {Array.from({ length: 4 }).map((_, i) => (
                    <div key={i} className="flex items-center justify-between rounded-md border border-border/30 p-3">
                      <Skeleton className="h-4 w-24" />
                      <Skeleton className="h-5 w-16 rounded-full" />
                    </div>
                  ))}
                </div>
              ) : healthError ? (
                <div className="flex flex-col items-center gap-2 py-8 text-center text-sm text-muted-foreground">
                  <XCircle className="h-8 w-8 text-red-400/60" />
                  <p>Failed to load health data</p>
                  <p className="text-xs text-muted-foreground/60">
                    Ensure the CloudOS kernel is running
                  </p>
                </div>
              ) : healthComponents.length > 0 ? (
                <div className="space-y-1">
                  {healthComponents.map(([name, report]) => (
                    <div
                      key={name}
                      className="flex items-center justify-between rounded-md border border-border/30 px-3 py-2.5 transition-colors hover:border-border/60"
                    >
                      <div className="flex items-center gap-2.5">
                        <StatusDot status={report.status} className="h-2 w-2" />
                        <span className="text-sm font-medium capitalize">
                          {name.replace(/-/g, ' ')}
                        </span>
                      </div>
                      <HealthBadge status={report.status} />
                    </div>
                  ))}
                </div>
              ) : (
                <div className="py-8 text-center text-sm text-muted-foreground">
                  No health data available
                </div>
              )}
            </CardContent>
          </Card>
        </motion.div>

        {/* Quick Actions */}
        <motion.div variants={itemVariants}>
          <Card className="border-border/50 bg-card/50 backdrop-blur-sm">
            <CardHeader className="pb-3">
              <div className="flex items-center gap-2">
                <Activity className="h-4 w-4 text-muted-foreground" />
                <CardTitle className="text-sm font-medium uppercase tracking-wider text-muted-foreground">
                  Quick Actions
                </CardTitle>
              </div>
            </CardHeader>
            <CardContent>
              <div className="flex flex-col gap-2">
                <QuickActionButton
                  icon={Box}
                  label="Create Project"
                  onClick={() => navigate('/projects')}
                />
                <QuickActionButton
                  icon={Globe}
                  label="View Resources"
                  onClick={() => navigate('/resources')}
                />
                <QuickActionButton
                  icon={Terminal}
                  label="View Controllers"
                  onClick={() => navigate('/controllers')}
                />
                <QuickActionButton
                  icon={Activity}
                  label="System Overview"
                  onClick={() => navigate('/system')}
                />
              </div>
            </CardContent>
          </Card>
        </motion.div>
      </div>

      {/* ── System Info ────────────────────────────────────────────────── */}
      <motion.div variants={itemVariants}>
        <Card className="border-border/50 bg-card/50 backdrop-blur-sm">
          <CardHeader className="pb-3">
            <div className="flex items-center gap-2">
              <Cpu className="h-4 w-4 text-muted-foreground" />
              <CardTitle className="text-sm font-medium uppercase tracking-wider text-muted-foreground">
                System Info
              </CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            {systemLoading ? (
              <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
                {Array.from({ length: 4 }).map((_, i) => (
                  <div key={i} className="space-y-1.5">
                    <Skeleton className="h-3 w-16" />
                    <Skeleton className="h-5 w-20" />
                  </div>
                ))}
              </div>
            ) : system ? (
              <div className="grid grid-cols-2 gap-y-4 gap-x-6 sm:grid-cols-4">
                <div>
                  <p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground/70">
                    Go Version
                  </p>
                  <p className="mt-0.5 font-mono text-sm tabular-nums text-foreground/80">
                    {system.goVersion || '\u2014'}
                  </p>
                </div>
                <div>
                  <p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground/70">
                    OS / Arch
                  </p>
                  <p className="mt-0.5 flex items-center gap-1.5 text-sm text-foreground/80">
                    <Globe className="h-3.5 w-3.5 text-muted-foreground/60" />
                    {system.os}/{system.arch}
                  </p>
                </div>
                <div>
                  <p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground/70">
                    CPU Cores
                  </p>
                  <p className="mt-0.5 flex items-center gap-1.5 font-mono text-sm tabular-nums text-foreground/80">
                    <HardDrive className="h-3.5 w-3.5 text-muted-foreground/60" />
                    {system.numCpu ?? '\u2014'}
                  </p>
                </div>
                <div>
                  <p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground/70">
                    Goroutines
                  </p>
                  <p className="mt-0.5 flex items-center gap-1.5 font-mono text-sm tabular-nums text-foreground/80">
                    <Cpu className="h-3.5 w-3.5 text-muted-foreground/60" />
                    {system.numGoroutine ?? '\u2014'}
                  </p>
                </div>
              </div>
            ) : (
              <div className="py-4 text-center text-sm text-muted-foreground">
                No system info available
              </div>
            )}
          </CardContent>
        </Card>
      </motion.div>
    </motion.div>
  );
}
