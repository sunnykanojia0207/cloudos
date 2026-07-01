import { useState, useMemo, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { useApplications, type AppResource, type DeploymentReport } from '@/hooks/useApplications';
import { usePageTitle } from '@/hooks/usePageTitle';
import { Badge, StatusDot } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, type SelectOption } from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { EmptyState } from '@/components/ui/empty-state';
import { ErrorState } from '@/components/ui/error-state';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  Heart,
  Activity,
  Cpu,
  Timer,
  Gauge,
  Monitor,
  Search,
  RotateCcw,
  ExternalLink,
  Terminal,
  GitBranch,
  Server,
  Layers,
  ArrowUp,
  ArrowDown,
  TrendingUp,
  X,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  Play,
  Clock,
  Eye,
} from 'lucide-react';
import { cn, relativeTime, truncate } from '@/lib/utils';

/* ════════════════════════════════════════════════════════════
   Types
   ════════════════════════════════════════════════════════════ */

type HealthFilter = '' | 'healthy' | 'degraded' | 'failed';

/* ── Health helper ────────────────────────────────────────── */
function appHealth(app: AppResource): 'healthy' | 'degraded' | 'failed' | 'unknown' {
  const h = app.status?.health?.toLowerCase() ?? '';
  if (['healthy', 'running'].includes(h)) return 'healthy';
  if (['degraded', 'warning'].includes(h)) return 'degraded';
  if (['failed', 'error', 'unhealthy'].includes(h)) return 'failed';
  return 'unknown';
}

function healthColor(h: string): string {
  if (h === 'healthy') return 'bg-success';
  if (h === 'degraded') return 'bg-warning';
  if (h === 'failed') return 'bg-danger';
  return 'bg-text-muted';
}

function healthLabel(h: string): string {
  if (h === 'healthy') return 'Healthy';
  if (h === 'degraded') return 'Warning';
  if (h === 'failed') return 'Critical';
  return 'Unknown';
}

/* ════════════════════════════════════════════════════════════
   Inline Sparkline SVG Chart
   ════════════════════════════════════════════════════════════ */

interface SparklineProps {
  values: number[];
  width?: number;
  height?: number;
  color?: string;
  className?: string;
}

/** Simple SVG line chart sparkline — no dependencies. */
function Sparkline({ values, width = 80, height = 28, color = '#5E6AD2', className }: SparklineProps) {
  if (values.length < 2) {
    // Single value or empty — just a flat line
    if (values.length === 1) {
      return (
        <svg width={width} height={height} className={className} aria-hidden="true">
          <line x1={0} y1={height / 2} x2={width} y2={height / 2} stroke={color} strokeWidth={1.5} />
        </svg>
      );
    }
    return null;
  }

  const padding = 2;
  const chartW = width - padding * 2;
  const chartH = height - padding * 2;
  const min = Math.min(...values);
  const max = Math.max(...values);
  const range = max - min || 1;

  const points = values.map((v, i) => {
    const x = padding + (i / (values.length - 1)) * chartW;
    const y = padding + chartH - ((v - min) / range) * chartH;
    return `${x},${y}`;
  });

  const pathD = `M ${points.join(' L ')}`;

  return (
    <svg width={width} height={height} className={className} aria-hidden="true">
      {/* Gradient fill under the line */}
      <defs>
        <linearGradient id={`spark-grad-${color.replace('#', '')}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={color} stopOpacity={0.2} />
          <stop offset="100%" stopColor={color} stopOpacity={0.02} />
        </linearGradient>
      </defs>
      <path
        d={`${pathD} L ${padding + chartW} ${height} L ${padding} ${height} Z`}
        fill={`url(#spark-grad-${color.replace('#', '')})`}
      />
      <path d={pathD} fill="none" stroke={color} strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round" />
      {/* Last point dot */}
      <circle cx={points[points.length - 1].split(',')[0]} cy={points[points.length - 1].split(',')[1]} r={2.5} fill={color} />
    </svg>
  );
}

/* ════════════════════════════════════════════════════════════
   Health Dot Strip (shows deployment outcomes as colored dots)
   ════════════════════════════════════════════════════════════ */

interface HealthDotStripProps {
  history: DeploymentReport[];
  max?: number;
}

function HealthDotStrip({ history, max = 12 }: HealthDotStripProps) {
  const recent = history.slice(-max).reverse();
  if (recent.length === 0) return <span className="text-caption text-text-muted">\u2014</span>;

  return (
    <span className="inline-flex items-center gap-0.5" role="img" aria-label={`Health history: ${recent.filter((d) => d.buildSuccess).length} success, ${recent.filter((d) => !d.buildSuccess).length} failure`}>
      {recent.map((d, i) => (
        <span
          key={i}
          className={cn(
            'inline-block h-2 w-2 rounded-full',
            d.buildSuccess ? 'bg-success' : 'bg-danger',
          )}
          aria-hidden="true"
        />
      ))}
      {history.length > max && (
        <span className="text-caption text-text-muted ml-0.5">+{history.length - max}</span>
      )}
    </span>
  );
}

/* ════════════════════════════════════════════════════════════
   Duration Trend (sparkline data from deployment durations)
   ════════════════════════════════════════════════════════════ */

function parseDurationToMs(dur: string): number {
  if (!dur) return 0;
  const match = dur.match(/^([\d.]+)(s|ms|m|h)?$/);
  if (!match) return 0;
  const val = parseFloat(match[1]);
  const unit = match[2] || 's';
  if (unit === 'ms') return val;
  if (unit === 's') return val * 1000;
  if (unit === 'm') return val * 60 * 1000;
  if (unit === 'h') return val * 3600 * 1000;
  return val;
}

function useDurationTrend(app: AppResource): number[] {
  return useMemo(() => {
    const history = app.status?.deploymentHistory ?? [];
    return history
      .filter((d) => d.duration)
      .map((d) => parseDurationToMs(d.duration))
      .filter((v) => v > 0);
  }, [app]);
}

/* ════════════════════════════════════════════════════════════
   Health Trend (sparkline of health over time — 1 for success, 0 for fail)
   ════════════════════════════════════════════════════════════ */

function useHealthTrend(app: AppResource): number[] {
  return useMemo(() => {
    const history = app.status?.deploymentHistory ?? [];
    return history.map((d) => (d.buildSuccess ? 1 : 0));
  }, [app]);
}

/* ════════════════════════════════════════════════════════════
   Filter options
   ════════════════════════════════════════════════════════════ */

const HEALTH_OPTIONS: SelectOption[] = [
  { value: '', label: 'All Health' },
  { value: 'healthy', label: 'Healthy' },
  { value: 'degraded', label: 'Warning' },
  { value: 'failed', label: 'Critical' },
];

const SORT_OPTIONS: SelectOption[] = [
  { value: 'name', label: 'Name (A-Z)' },
  { value: 'name-desc', label: 'Name (Z-A)' },
  { value: 'health', label: 'Health (worst first)' },
  { value: 'health-good', label: 'Health (best first)' },
];

/* ════════════════════════════════════════════════════════════
   Metric Card
   ════════════════════════════════════════════════════════════ */

interface MetricCardProps {
  icon: React.ReactNode;
  label: string;
  value: string;
  sublabel?: string;
  color?: 'default' | 'success' | 'warning' | 'danger';
  trend?: 'up' | 'down' | 'neutral';
}

function MetricCard({ icon, label, value, sublabel, color = 'default', trend }: MetricCardProps) {
  return (
    <div className={cn(
      'flex flex-col gap-1 rounded-md border p-3',
      color === 'success' ? 'border-success/20 bg-success-subtle/30' :
      color === 'warning' ? 'border-warning/20 bg-warning-subtle/30' :
      color === 'danger' ? 'border-danger/20 bg-danger-subtle/30' :
      'border-border bg-surface',
    )}>
      <div className="flex items-center justify-between">
        <span className="flex items-center gap-1.5 text-caption text-text-secondary">
          {icon}
          {label}
        </span>
        {trend === 'up' && <ArrowUp className="h-3 w-3 text-danger" aria-hidden="true" />}
        {trend === 'down' && <ArrowDown className="h-3 w-3 text-success" aria-hidden="true" />}
      </div>
      <span className={cn(
        'text-h2 font-semibold tabular-nums',
        color === 'success' ? 'text-success' :
        color === 'warning' ? 'text-warning' :
        color === 'danger' ? 'text-danger' :
        'text-foreground',
      )}>
        {value}
      </span>
      {sublabel && (
        <span className="text-caption text-text-muted">{sublabel}</span>
      )}
    </div>
  );
}

/* ════════════════════════════════════════════════════════════
   Application Table Row (Desktop & Tablet)
   ════════════════════════════════════════════════════════════ */

interface AppRowProps {
  app: AppResource;
  onSelect: (app: AppResource) => void;
}

function AppTableRow({ app, onSelect }: AppRowProps) {
  const health = appHealth(app);
  const lastReport = app.status?.lastReport;
  const history = app.status?.deploymentHistory ?? [];
  const durationTrend = useDurationTrend(app);
  const healthTrend = useHealthTrend(app);
  const duration = lastReport?.duration;
  const runtime = lastReport?.detectedRuntime || app.spec?.runtime?.type || '\u2014';
  const env = lastReport?.environment || '\u2014';

  // Compute avg duration
  const avgDuration = durationTrend.length > 0
    ? (durationTrend.reduce((a, b) => a + b, 0) / durationTrend.length / 1000).toFixed(1) + 's'
    : '\u2014';

  // Uptime (time since app creation)
  const uptime = app.metadata.createdAt
    ? relativeTime(app.metadata.createdAt)
    : '\u2014';

  return (
    <motion.tr
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      className="border-b border-border last:border-b-0 hover:bg-accent-subtle/30 transition-colors cursor-pointer"
      onClick={() => onSelect(app)}
      tabIndex={0}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(app); } }}
      role="button"
      aria-label={`View details for ${app.metadata.name}`}
    >
      {/* Application */}
      <td className="px-3 py-2.5 min-w-0">
        <div className="flex items-center gap-2">
          <span className={cn('inline-block h-2 w-2 rounded-full shrink-0', healthColor(health))} aria-hidden="true" />
          <div className="min-w-0">
            <span className="text-body text-foreground font-medium truncate block">{app.metadata.name}</span>
            <span className="text-caption text-text-muted truncate block">{truncate(app.metadata.id, 24)}</span>
          </div>
        </div>
      </td>

      {/* Health */}
      <td className="px-3 py-2.5 hidden sm:table-cell">
        <Badge variant={
          health === 'healthy' ? 'subtle-success' :
          health === 'degraded' ? 'subtle-warning' :
          health === 'failed' ? 'subtle-danger' :
          'subtle-neutral'
        } className="gap-1.5 text-caption">
          <StatusDot status={health === 'healthy' ? 'success' : health === 'degraded' ? 'warning' : health === 'failed' ? 'danger' : 'pending'} />
          {healthLabel(health)}
        </Badge>
      </td>

      {/* Avg Duration */}
      <td className="px-3 py-2.5 hidden md:table-cell text-small text-foreground tabular-nums">
        {avgDuration}
      </td>

      {/* Runtime */}
      <td className="px-3 py-2.5 hidden lg:table-cell text-small text-text-secondary">
        {runtime}
      </td>

      {/* Environment */}
      <td className="px-3 py-2.5 hidden lg:table-cell">
        {env !== '\u2014' ? (
          <Badge variant="subtle-neutral" className="text-caption uppercase tracking-wider">{env}</Badge>
        ) : (
          <span className="text-caption text-text-muted">\u2014</span>
        )}
      </td>

      {/* Health Trend (dot strip) */}
      <td className="px-3 py-2.5 hidden xl:table-cell">
        <HealthDotStrip history={history} />
      </td>

      {/* Duration Sparkline */}
      <td className="px-3 py-2.5 hidden xl:table-cell">
        {durationTrend.length >= 2 ? (
          <Sparkline
            values={durationTrend}
            color={health === 'failed' ? '#D45A5A' : '#5E6AD2'}
          />
        ) : (
          <span className="text-caption text-text-muted">\u2014</span>
        )}
      </td>

      {/* Last Deployment */}
      <td className="px-3 py-2.5 hidden sm:table-cell text-right">
        <div className="text-small text-foreground tabular-nums">
          {lastReport ? `#${lastReport.deploymentNumber}` : '\u2014'}
        </div>
        <div className="text-caption text-text-muted tabular-nums">
          {lastReport?.duration ? lastReport.duration : ''}
        </div>
      </td>
    </motion.tr>
  );
}

/* ════════════════════════════════════════════════════════════
   Mobile Card
   ════════════════════════════════════════════════════════════ */

function MobileAppCard({ app, onSelect }: AppRowProps) {
  const health = appHealth(app);
  const lastReport = app.status?.lastReport;
  const history = app.status?.deploymentHistory ?? [];
  const runtime = lastReport?.detectedRuntime || app.spec?.runtime?.type || '\u2014';

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      className="sm:hidden rounded-md border border-border bg-surface p-3 space-y-2 cursor-pointer hover:border-border-hover transition-colors"
      onClick={() => onSelect(app)}
      tabIndex={0}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(app); } }}
      role="button"
      aria-label={`View details for ${app.metadata.name}`}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className={cn('inline-block h-2 w-2 rounded-full shrink-0', healthColor(health))} aria-hidden="true" />
            <span className="text-body font-medium text-foreground truncate">{app.metadata.name}</span>
          </div>
          <span className="text-caption text-text-muted truncate block mt-0.5">{truncate(app.metadata.id, 24)}</span>
        </div>
        <Badge variant={
          health === 'healthy' ? 'subtle-success' :
          health === 'degraded' ? 'subtle-warning' :
          health === 'failed' ? 'subtle-danger' :
          'subtle-neutral'
        } className="gap-1 text-caption shrink-0">
          <StatusDot status={health === 'healthy' ? 'success' : health === 'degraded' ? 'warning' : health === 'failed' ? 'danger' : 'pending'} />
          {healthLabel(health)}
        </Badge>
      </div>

      <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-caption text-text-muted">
        <span className="tabular-nums">{lastReport ? `#${lastReport.deploymentNumber}` : '\u2014'}</span>
        {lastReport?.duration && <span className="tabular-nums">{lastReport.duration}</span>}
        <span>{runtime}</span>
        {history.length > 0 && <HealthDotStrip history={history} max={8} />}
      </div>
    </motion.div>
  );
}

/* ════════════════════════════════════════════════════════════
   Detail Drawer (side panel)
   ════════════════════════════════════════════════════════════ */

interface DetailDrawerProps {
  app: AppResource;
  onClose: () => void;
}

function DetailDrawer({ app, onClose }: DetailDrawerProps) {
  const navigate = useNavigate();
  const health = appHealth(app);
  const lastReport = app.status?.lastReport;
  const history = app.status?.deploymentHistory ?? [];
  const durationTrend = useDurationTrend(app);
  const healthTrend = useHealthTrend(app);

  const recentDeployments = useMemo(
    () => [...history].reverse().slice(0, 5),
    [history],
  );

  // Compute health metrics
  const totalDeployments = history.length;
  const successfulDeployments = history.filter((d) => d.buildSuccess).length;
  const successRate = totalDeployments > 0
    ? `${Math.round((successfulDeployments / totalDeployments) * 100)}%`
    : '\u2014';

  return (
    <motion.aside
      initial={{ x: '100%', opacity: 0 }}
      animate={{ x: 0, opacity: 1 }}
      exit={{ x: '100%', opacity: 0 }}
      transition={{ duration: 0.2, ease: [0, 0, 0.2, 1] }}
      className="fixed right-0 top-12 z-40 h-[calc(100vh-3rem-1.75rem)] w-full sm:w-[420px] border-l border-border bg-surface shadow-lg overflow-y-auto"
      role="dialog"
      aria-modal="true"
      aria-label={`Details for ${app.metadata.name}`}
    >
      <div className="flex flex-col h-full">
        {/* ── Drawer Header ── */}
        <div className="flex items-start justify-between gap-3 px-4 pt-4 pb-3 border-b border-border">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <span className={cn('inline-block h-2.5 w-2.5 rounded-full shrink-0', healthColor(health))} aria-hidden="true" />
              <h2 className="text-h3 text-foreground truncate">{app.metadata.name}</h2>
            </div>
            <p className="text-small text-text-secondary mt-0.5 flex items-center gap-1.5">
              <Badge variant={
                health === 'healthy' ? 'subtle-success' :
                health === 'degraded' ? 'subtle-warning' :
                health === 'failed' ? 'subtle-danger' :
                'subtle-neutral'
              } className="gap-1 text-caption">
                <StatusDot status={health === 'healthy' ? 'success' : health === 'degraded' ? 'warning' : health === 'failed' ? 'danger' : 'pending'} />
                {healthLabel(health)}
              </Badge>
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="shrink-0 p-1 text-text-muted hover:text-foreground rounded-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            aria-label="Close detail panel"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <ScrollArea className="flex-1 px-4 py-4 space-y-5">
          {/* ── Health Overview ── */}
          <section aria-label="Health overview">
            <h3 className="text-body font-semibold text-foreground flex items-center gap-2 mb-3">
              <Heart className="h-4 w-4 text-text-muted" />
              Health Overview
            </h3>
            <div className="grid grid-cols-2 gap-2.5">
              <div className="rounded-md border border-border bg-surface p-2.5">
                <span className="text-caption text-text-secondary">Success Rate</span>
                <p className="text-body font-semibold text-foreground tabular-nums mt-0.5">{successRate}</p>
              </div>
              <div className="rounded-md border border-border bg-surface p-2.5">
                <span className="text-caption text-text-secondary">Deployments</span>
                <p className="text-body font-semibold text-foreground tabular-nums mt-0.5">{totalDeployments}</p>
              </div>
            </div>

            {/* Health trend sparkline */}
            {healthTrend.length >= 2 && (
              <div className="mt-3 rounded-md border border-border bg-surface p-2.5">
                <span className="text-caption text-text-secondary block mb-1">Health History (success rate)</span>
                <Sparkline values={healthTrend.map((v) => v * 100)} color="#2B9D5D" width={320} height={36} />
                <div className="flex justify-between text-caption text-text-muted mt-0.5">
                  <span>Earlier</span>
                  <span>Now</span>
                </div>
              </div>
            )}

            {/* Duration trend sparkline */}
            {durationTrend.length >= 2 && (
              <div className="mt-2.5 rounded-md border border-border bg-surface p-2.5">
                <span className="text-caption text-text-secondary block mb-1">Deployment Duration Trend</span>
                <Sparkline values={durationTrend} color="#5E6AD2" width={320} height={36} />
                <div className="flex justify-between text-caption text-text-muted mt-0.5">
                  <span>Earlier</span>
                  <span>Now</span>
                </div>
              </div>
            )}
          </section>

          <Separator />

          {/* ── Application Info ── */}
          <section aria-label="Application information">
            <h3 className="text-body font-semibold text-foreground flex items-center gap-2 mb-3">
              <Layers className="h-4 w-4 text-text-muted" />
              Application Info
            </h3>
            <div className="space-y-2.5">
              {/* ID */}
              <div className="flex items-center justify-between">
                <span className="text-small text-text-secondary">ID</span>
                <span className="text-small font-mono text-foreground truncate ml-2">{truncate(app.metadata.id, 28)}</span>
              </div>

              {/* Endpoint */}
              {app.status?.url && (
                <div className="flex items-center justify-between">
                  <span className="text-small text-text-secondary">Endpoint</span>
                  <a
                    href={app.status.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-small text-accent hover:text-accent-hover truncate ml-2 max-w-[200px]"
                  >
                    {app.status.url}
                  </a>
                </div>
              )}

              {/* Runtime */}
              <div className="flex items-center justify-between">
                <span className="text-small text-text-secondary">Runtime</span>
                <span className="text-small text-foreground">{lastReport?.detectedRuntime || app.spec?.runtime?.type || '\u2014'}</span>
              </div>

              {/* Phase */}
              <div className="flex items-center justify-between">
                <span className="text-small text-text-secondary">Phase</span>
                <span className="text-small text-foreground">{app.status?.phase || '\u2014'}</span>
              </div>

              {/* Created */}
              {app.metadata.createdAt && (
                <div className="flex items-center justify-between">
                  <span className="text-small text-text-secondary">Created</span>
                  <span className="text-small text-foreground tabular-nums">{relativeTime(app.metadata.createdAt)}</span>
                </div>
              )}
            </div>
          </section>

          <Separator />

          {/* ── Recent Deployments ── */}
          <section aria-label="Recent deployments">
            <h3 className="text-body font-semibold text-foreground flex items-center gap-2 mb-3">
              <GitBranch className="h-4 w-4 text-text-muted" />
              Recent Deployments
            </h3>
            {recentDeployments.length === 0 ? (
              <p className="text-small text-text-muted">No deployments yet.</p>
            ) : (
              <div className="space-y-1">
                {recentDeployments.map((d) => (
                  <button
                    key={d.deploymentNumber}
                    type="button"
                    onClick={() => {
                      onClose();
                      navigate(`/applications/${app.metadata.id}/deployments/${d.deploymentNumber}`);
                    }}
                    className="w-full flex items-center justify-between rounded-sm px-2.5 py-1.5 hover:bg-accent-subtle/50 transition-colors text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  >
                    <div className="flex items-center gap-2 min-w-0">
                      {d.buildSuccess
                        ? <CheckCircle2 className="h-3.5 w-3.5 text-success shrink-0" />
                        : <XCircle className="h-3.5 w-3.5 text-danger shrink-0" />
                      }
                      <span className="text-small text-foreground tabular-nums">#{d.deploymentNumber}</span>
                      {d.branch && <span className="text-caption text-text-muted truncate">{d.branch}</span>}
                    </div>
                    <span className="text-caption text-text-muted tabular-nums shrink-0 ml-2">
                      {d.duration} &middot; {relativeTime(d.completedAt || d.startedAt)}
                    </span>
                  </button>
                ))}
              </div>
            )}
          </section>

          <Separator />

          {/* ── Quick Actions ── */}
          <section aria-label="Quick actions">
            <h3 className="text-body font-semibold text-foreground flex items-center gap-2 mb-3">
              <Activity className="h-4 w-4 text-text-muted" />
              Quick Actions
            </h3>
            <div className="grid grid-cols-2 gap-2">
              <Button variant="secondary" size="sm" className="justify-start gap-2" onClick={() => { onClose(); navigate(`/applications/${app.metadata.id}`); }}>
                <Eye className="h-3.5 w-3.5" />
                Overview
              </Button>
              <Button variant="secondary" size="sm" className="justify-start gap-2" onClick={() => { onClose(); navigate(`/applications/${app.metadata.id}/timeline`); }}>
                <Terminal className="h-3.5 w-3.5" />
                Timeline
              </Button>
              <Button variant="secondary" size="sm" className="justify-start gap-2" onClick={() => { onClose(); navigate(`/applications/${app.metadata.id}/logs`); }}>
                <Terminal className="h-3.5 w-3.5" />
                Logs
              </Button>
              <Button variant="secondary" size="sm" className="justify-start gap-2" onClick={() => { onClose(); navigate(`/workflows/${lastReport?.workflowId}`); }} disabled={!lastReport?.workflowId}>
                <Activity className="h-3.5 w-3.5" />
                Workflow
              </Button>
            </div>
          </section>
        </ScrollArea>
      </div>
    </motion.aside>
  );
}

/* ════════════════════════════════════════════════════════════
   Loading / Empty / Error
   ════════════════════════════════════════════════════════════ */

function MonitoringSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <div className="space-y-1">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-4 w-60" />
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-[88px] rounded-md" />
        ))}
      </div>

      {/* Search + filter */}
      <div className="flex gap-2">
        <Skeleton className="h-[34px] w-64 rounded-md" />
        <Skeleton className="h-[34px] w-32 rounded-md" />
      </div>

      {/* Table skeleton */}
      <div className="rounded-md border border-border overflow-hidden">
        <div className="flex items-center gap-3 px-3 py-2 bg-surface border-b border-border">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-4" style={{ width: `${[30, 12, 10, 10, 12, 8][i]}%` }} />
          ))}
        </div>
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="flex items-center gap-3 px-3 py-2.5 border-b border-border">
            {Array.from({ length: 6 }).map((_, j) => (
              <Skeleton key={j} className="h-4" style={{ width: `${[30, 12, 10, 10, 12, 8][j]}%` }} />
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}

/* ════════════════════════════════════════════════════════════
   Main Page
   ════════════════════════════════════════════════════════════ */

export default function MonitoringPage() {
  usePageTitle('Monitoring');
  const navigate = useNavigate();
  const { data: applications, isLoading, error, refetch } = useApplications();

  const [search, setSearch] = useState('');
  const [healthFilter, setHealthFilter] = useState<HealthFilter>('');
  const [envFilter, setEnvFilter] = useState('');
  const [runtimeFilter, setRuntimeFilter] = useState('');
  const [sortOrder, setSortOrder] = useState('health');
  const [selectedApp, setSelectedApp] = useState<AppResource | null>(null);

  // ── Filter & sort options ──
  const filterOptions = useMemo(() => {
    const envs = new Set<string>();
    const runtimes = new Set<string>();
    for (const app of applications ?? []) {
      const env = app.status?.lastReport?.environment;
      if (env) envs.add(env);
      const rt = app.status?.lastReport?.detectedRuntime || app.spec?.runtime?.type;
      if (rt) runtimes.add(rt);
    }
    return {
      envOptions: [{ value: '', label: 'All Environments' }, ...Array.from(envs).map((v) => ({ value: v, label: v }))],
      runtimeOptions: [{ value: '', label: 'All Runtimes' }, ...Array.from(runtimes).map((v) => ({ value: v, label: v }))],
    };
  }, [applications]);

  // ── Compute summary stats ──
  const stats = useMemo(() => {
    const apps = applications ?? [];
    const total = apps.length;
    const healthy = apps.filter((a) => appHealth(a) === 'healthy').length;
    const degraded = apps.filter((a) => appHealth(a) === 'degraded').length;
    const failed = apps.filter((a) => appHealth(a) === 'failed').length;

    // Average deployment duration across all apps
    const allDurations: number[] = [];
    for (const app of apps) {
      for (const d of app.status?.deploymentHistory ?? []) {
        const ms = parseDurationToMs(d.duration);
        if (ms > 0) allDurations.push(ms);
      }
    }
    const avgDuration = allDurations.length > 0
      ? (allDurations.reduce((a, b) => a + b, 0) / allDurations.length / 1000).toFixed(1) + 's'
      : '\u2014';

    return { total, healthy, degraded, failed, avgDuration };
  }, [applications]);

  // ── Filtered + sorted ──
  const filtered = useMemo(() => {
    let result = [...(applications ?? [])];

    // Search
    if (search.trim()) {
      const q = search.toLowerCase();
      result = result.filter(
        (a) =>
          a.metadata.name.toLowerCase().includes(q) ||
          a.metadata.id.toLowerCase().includes(q) ||
          a.status?.lastReport?.repository?.toLowerCase().includes(q),
      );
    }

    // Health filter
    if (healthFilter) {
      result = result.filter((a) => appHealth(a) === healthFilter);
    }

    // Environment filter
    if (envFilter) {
      result = result.filter((a) => a.status?.lastReport?.environment === envFilter);
    }

    // Runtime filter
    if (runtimeFilter) {
      result = result.filter(
        (a) => (a.status?.lastReport?.detectedRuntime || a.spec?.runtime?.type) === runtimeFilter,
      );
    }

    // Sort
    result.sort((a, b) => {
      const hA = appHealth(a);
      const hB = appHealth(b);
      const order = ['failed', 'degraded', 'healthy', 'unknown'];

      if (sortOrder === 'health') return order.indexOf(hA) - order.indexOf(hB);
      if (sortOrder === 'health-good') return order.indexOf(hB) - order.indexOf(hA);
      if (sortOrder === 'name-desc') return b.metadata.name.localeCompare(a.metadata.name);
      return a.metadata.name.localeCompare(b.metadata.name);
    });

    return result;
  }, [applications, search, healthFilter, envFilter, runtimeFilter, sortOrder]);

  const hasActiveFilters = search || healthFilter || envFilter || runtimeFilter;

  const clearFilters = useCallback(() => {
    setSearch('');
    setHealthFilter('');
    setEnvFilter('');
    setRuntimeFilter('');
  }, []);

  // ── Keyboard: Escape closes drawer ──
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Escape' && selectedApp) {
        setSelectedApp(null);
      }
    },
    [selectedApp],
  );

  return (
    <div onKeyDown={handleKeyDown}>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className={cn('flex flex-col gap-6', selectedApp && 'sm:mr-[420px] transition-all duration-200')}
      >
        {/* ══════════ HEADER ══════════ */}
        <div className="flex flex-col gap-1">
          <h1 className="text-h1 text-foreground">Monitoring</h1>
          <p className="text-small text-text-secondary">
            {isLoading
              ? 'Loading...'
              : error
                ? 'Unable to load monitoring data'
                : `${stats.total} application${stats.total !== 1 ? 's' : ''} tracked`}
          </p>
        </div>

        {/* ══════════ HEALTH SUMMARY CARDS ══════════ */}
        {!isLoading && !error && stats.total > 0 && (
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
            <MetricCard
              icon={<Heart className="h-3.5 w-3.5" />}
              label="Healthy"
              value={String(stats.healthy)}
              color={stats.healthy > 0 ? 'success' : 'default'}
            />
            <MetricCard
              icon={<AlertTriangle className="h-3.5 w-3.5" />}
              label="Warning"
              value={String(stats.degraded)}
              color={stats.degraded > 0 ? 'warning' : 'default'}
            />
            <MetricCard
              icon={<XCircle className="h-3.5 w-3.5" />}
              label="Critical"
              value={String(stats.failed)}
              color={stats.failed > 0 ? 'danger' : 'default'}
            />
            <MetricCard
              icon={<Timer className="h-3.5 w-3.5" />}
              label="Avg Duration"
              value={stats.avgDuration}
              sublabel="Per deployment"
            />
            <MetricCard
              icon={<Activity className="h-3.5 w-3.5" />}
              label="Applications"
              value={String(stats.total)}
              sublabel="Total tracked"
            />
          </div>
        )}

        {/* ══════════ SEARCH + FILTERS ══════════ */}
        <div className="flex flex-col gap-3">
          <div className="flex items-center gap-2 flex-wrap">
            <div className="relative flex-1 min-w-[200px] max-w-sm">
              <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted pointer-events-none" />
              <Input
                placeholder="Search by app or repository..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-8 h-[34px]"
                aria-label="Search applications"
              />
            </div>

            <Select options={HEALTH_OPTIONS} value={healthFilter} onChange={(e) => setHealthFilter(e.target.value as HealthFilter)} className="w-36" aria-label="Health filter" />
            <Select options={filterOptions.envOptions} value={envFilter} onChange={(e) => setEnvFilter(e.target.value)} className="w-40" aria-label="Environment filter" />
            <Select options={filterOptions.runtimeOptions} value={runtimeFilter} onChange={(e) => setRuntimeFilter(e.target.value)} className="w-36" aria-label="Runtime filter" />
            <Select options={SORT_OPTIONS} value={sortOrder} onChange={(e) => setSortOrder(e.target.value)} className="w-40" aria-label="Sort order" />

            <Button variant="icon-ghost" size="icon-sm" onClick={() => refetch()} aria-label="Refresh monitoring data" title="Refresh">
              <RotateCcw className="h-3.5 w-3.5" />
            </Button>

            {hasActiveFilters && (
              <Button variant="ghost" size="sm" onClick={clearFilters} className="gap-1 text-small">
                Clear
              </Button>
            )}
          </div>
        </div>

        {/* ══════════ ERROR ══════════ */}
        {!isLoading && error && (
          <ErrorState
            title="Failed to load monitoring data"
            message={(error as Error)?.message || 'An unexpected error occurred.'}
            onRetry={() => refetch()}
          />
        )}

        {/* ══════════ EMPTY ══════════ */}
        {!isLoading && !error && applications && applications.length === 0 && (
          <EmptyState
            icon={Monitor}
            title="No applications deployed"
            description="Deploy an application to see health metrics and monitoring data."
            actions={[
              { label: 'Deploy Application', icon: Monitor, onClick: () => navigate('/applications/deploy'), variant: 'primary' },
            ]}
          />
        )}

        {/* ══════════ LOADING ══════════ */}
        {isLoading && <MonitoringSkeleton />}

        {/* ══════════ APPLICATION TABLE (desktop) ══════════ */}
        {!isLoading && !error && applications && applications.length > 0 && (
          <>
            {/* Filtered empty state */}
            {filtered.length === 0 && (
              <EmptyState
                icon={Search}
                title="No matching applications"
                description="Try adjusting your search or filters."
                actions={[{ label: 'Clear filters', onClick: clearFilters, variant: 'secondary' }]}
              />
            )}

            {/* Desktop+Tablet table */}
            <div className="hidden sm:block rounded-md border border-border overflow-hidden">
              <table className="w-full" role="grid" aria-label="Application health table">
                <thead>
                  <tr className="border-b border-border bg-surface text-caption font-medium text-text-secondary uppercase tracking-wider">
                    <th className="px-3 py-2 text-left w-[30%]">Application</th>
                    <th className="px-3 py-2 text-left w-[12%] hidden sm:table-cell">Health</th>
                    <th className="px-3 py-2 text-left w-[10%] hidden md:table-cell">Duration</th>
                    <th className="px-3 py-2 text-left w-[10%] hidden lg:table-cell">Runtime</th>
                    <th className="px-3 py-2 text-left w-[10%] hidden lg:table-cell">Env</th>
                    <th className="px-3 py-2 text-left w-[12%] hidden xl:table-cell">Trend</th>
                    <th className="px-3 py-2 text-left w-[8%] hidden xl:table-cell">Duration Sparkline</th>
                    <th className="px-3 py-2 text-right w-[8%] hidden sm:table-cell">Last Deploy</th>
                  </tr>
                </thead>
                <tbody>
                  {filtered.map((app) => (
                    <AppTableRow
                      key={app.metadata.id}
                      app={app}
                      onSelect={setSelectedApp}
                    />
                  ))}
                </tbody>
              </table>
            </div>

            {/* Mobile cards */}
            <div className="sm:hidden space-y-2">
              {filtered.map((app) => (
                <MobileAppCard
                  key={app.metadata.id}
                  app={app}
                  onSelect={setSelectedApp}
                />
              ))}
            </div>

            {/* Footer */}
            <div className="flex items-center justify-center text-caption text-text-muted pt-1">
              {hasActiveFilters && filtered.length !== (applications?.length ?? 0)
                ? `Showing ${filtered.length} of ${applications?.length ?? 0} applications`
                : `${applications?.length ?? 0} application${(applications?.length ?? 0) !== 1 ? 's' : ''}`}
            </div>
          </>
        )}
      </motion.div>

      {/* ══════════ DETAIL DRAWER ══════════ */}
      <AnimatePresence>
        {selectedApp && (
          <DetailDrawer
            key={selectedApp.metadata.id}
            app={selectedApp}
            onClose={() => setSelectedApp(null)}
          />
        )}
      </AnimatePresence>

      {/* Overlay when drawer is open (mobile) */}
      <AnimatePresence>
        {selectedApp && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-30 bg-black/40 sm:hidden"
            onClick={() => setSelectedApp(null)}
            aria-hidden="true"
          />
        )}
      </AnimatePresence>
    </div>
  );
}
