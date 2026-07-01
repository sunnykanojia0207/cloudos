import * as React from 'react';
import type { AppResource } from '@/hooks/useApplications';
import { cn } from '@/lib/utils';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { HealthIndicator } from '@/components/ui/health-indicator';
import { EmptyState } from '@/components/ui/empty-state';
import {
  Monitor,
  Cpu,
  MemoryStick,
  Heart,
  Activity,
  Gauge,
  Timer,
  TrendingUp,
  ArrowUp,
} from 'lucide-react';

/* ── Metric Card ──────────────────────────────────────── */
interface MetricCardProps {
  icon: React.ReactNode;
  label: string;
  value: string;
  sublabel?: string;
  trend?: 'up' | 'down' | 'neutral';
  color?: 'default' | 'success' | 'warning' | 'danger';
}

const TREND_ICONS = {
  up: { icon: ArrowUp, color: 'text-danger' },
  down: { icon: TrendingUp, color: 'text-success' },
  neutral: { icon: Activity, color: 'text-text-muted' },
};

function MetricCard({ icon, label, value, sublabel, trend, color = 'default' }: MetricCardProps) {
  const trendCfg = trend ? TREND_ICONS[trend] : null;

  return (
    <div className={cn(
      'flex flex-col gap-1.5 rounded-md border p-3',
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
        {trendCfg && (
          <trendCfg.icon className={cn('h-3 w-3', trendCfg.color)} aria-hidden="true" />
        )}
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

/* ── Props ────────────────────────────────────────────── */
export interface MonitoringTabProps {
  app: AppResource;
}

/* ── Component ────────────────────────────────────────── */
export function MonitoringTab({ app }: MonitoringTabProps) {
  const phase = app.status?.phase ?? 'Stopped';
  const health = app.status?.health ?? 'Unknown';
  const url = app.status?.url;
  const hasData = app.status?.deploymentCount != null && app.status.deploymentCount > 0;

  // Simulated / derived metrics (in a real app, these would come from a metrics API)
  const isRunning = phase === 'Running';
  const uptime = isRunning ? '12h 34m' : '\u2014';
  const requests = isRunning ? '2.4K' : '\u2014';
  const latency = isRunning ? '45ms' : '\u2014';
  const cpu = isRunning ? '23%' : '\u2014';
  const memory = isRunning ? '128 MB' : '\u2014';

  if (!hasData) {
    return (
      <EmptyState
        icon={Monitor}
        title="No monitoring data"
        description="Deploy the application to see health metrics and monitoring data."
      />
    );
  }

  return (
    <div className="space-y-5">
      {/* ── Quick Health Cards ── */}
      <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
        <MetricCard
          icon={<Heart className="h-3.5 w-3.5" />}
          label="Health"
          value={health}
          color={health === 'Healthy' ? 'success' : health === 'Degraded' ? 'warning' : health === 'Error' ? 'danger' : 'default'}
          trend={health === 'Healthy' ? 'up' : health === 'Degraded' ? 'down' : 'neutral'}
        />
        <MetricCard
          icon={<Activity className="h-3.5 w-3.5" />}
          label="Requests"
          value={requests}
          sublabel="Last 24h"
          trend="up"
        />
        <MetricCard
          icon={<Timer className="h-3.5 w-3.5" />}
          label="Latency"
          value={latency}
          sublabel="P50"
          trend={latency === '45ms' ? 'down' : 'neutral'}
        />
        <MetricCard
          icon={<Cpu className="h-3.5 w-3.5" />}
          label="CPU"
          value={cpu}
          sublabel="Current"
          trend={cpu === '23%' ? 'down' : 'up'}
        />
        <MetricCard
          icon={<MemoryStick className="h-3.5 w-3.5" />}
          label="Memory"
          value={memory}
          sublabel="Current"
          trend="neutral"
        />
        <MetricCard
          icon={<Gauge className="h-3.5 w-3.5" />}
          label="Uptime"
          value={uptime}
          sublabel="Since last deploy"
          color={isRunning ? 'success' : 'default'}
        />
      </div>

      {/* ── Health Status Detail ── */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-body font-semibold flex items-center gap-2">
            <Heart className="h-4 w-4 text-text-muted" />
            Health Status
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {/* Status */}
            <div className="flex items-center justify-between rounded-md border border-border bg-surface px-3 py-2.5">
              <span className="text-small text-text-secondary">Status</span>
              <HealthIndicator
                status={
                  phase === 'Running' ? 'running' :
                  phase === 'Deploying' ? 'deploying' :
                  phase === 'Failed' ? 'failed' :
                  'stopped'
                }
                showLabel
                size="sm"
              />
            </div>

            {/* Endpoint */}
            {url && (
              <div className="flex items-center justify-between rounded-md border border-border bg-surface px-3 py-2.5">
                <span className="text-small text-text-secondary">Endpoint</span>
                <a
                  href={url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-small text-accent hover:text-accent-hover truncate max-w-[200px]"
                >
                  {url}
                </a>
              </div>
            )}

            {/* Phase */}
            <div className="flex items-center justify-between rounded-md border border-border bg-surface px-3 py-2.5">
              <span className="text-small text-text-secondary">Phase</span>
              <span className="text-small text-foreground font-medium">{phase}</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
