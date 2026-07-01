import * as React from 'react';
import { motion } from 'framer-motion';
import { useNavigate, NavLink } from 'react-router-dom';
import type { AppResource } from '@/hooks/useApplications';
import { cn, truncate } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  ExternalLink,
  Terminal,
  GitCommitHorizontal,
  Clock,
  Globe,
  Layers,
} from 'lucide-react';

/* ── Helpers ──────────────────────────────────────────── */

const phaseConfig: Record<
  string,
  { variant: 'success' | 'warning' | 'secondary' | 'destructive'; label: string }
> = {
  Running: { variant: 'success', label: 'Running' },
  Deploying: { variant: 'warning', label: 'Deploying' },
  Stopped: { variant: 'secondary', label: 'Stopped' },
  Failed: { variant: 'destructive', label: 'Failed' },
};

const healthConfig: Record<
  string,
  { variant: 'success' | 'warning' | 'destructive'; icon: string }
> = {
  Healthy: { variant: 'success', icon: '●' },
  Degraded: { variant: 'warning', icon: '●' },
  Error: { variant: 'destructive', icon: '●' },
};

const environmentConfig: Record<string, string> = {
  development: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
  staging: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
  production: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
  testing: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
};

/* ── Props ────────────────────────────────────────────── */

export interface ApplicationCardProps {
  app: AppResource;
}

/* ── Component ────────────────────────────────────────── */

export function ApplicationCard({ app }: ApplicationCardProps) {
  const navigate = useNavigate();

  const phase = app.status?.phase ?? 'Stopped';
  const phaseConf = phaseConfig[phase] ?? phaseConfig.Stopped;

  const health = app.status?.health ?? 'Unknown';
  const healthConf = healthConfig[health] ?? { variant: 'secondary' as const, icon: '●' };

  const env = app.status?.lastReport?.environment;
  const lastReport = app.status?.lastReport;
  const commitSha = lastReport?.commitSha;
  const deploymentNumber = lastReport?.deploymentNumber;
  const duration = lastReport?.duration;
  const url = app.status?.url;
  const runtimeType = app.spec?.runtime?.type;

  const handleClick = () => {
    navigate(`/applications/${app.metadata.id}`);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleClick();
    }
  };

  return (
    <motion.div
      layout
      initial={{ opacity: 0, scale: 0.96 }}
      animate={{ opacity: 1, scale: 1 }}
      exit={{ opacity: 0, scale: 0.96 }}
      transition={{ duration: 0.2, ease: 'easeOut' }}
    >
      <Card
        className={cn(
          'group cursor-pointer border-border/50 transition-all duration-200',
          'hover:border-accent/30 hover:-translate-y-0.5',
        )}
        onClick={handleClick}
        onKeyDown={handleKeyDown}
        tabIndex={0}
        role="link"
        aria-label={`Application ${app.metadata.name}`}
      >
        {/* Header */}
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between gap-3">
            <div className="flex items-center gap-2.5 min-w-0">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border bg-muted/50">
                <Layers className="h-4 w-4 text-muted-foreground" />
              </div>
              <div className="min-w-0">
                <CardTitle className="text-sm font-semibold leading-tight truncate">
                  {app.metadata.name}
                </CardTitle>
                <CardDescription className="text-xs mt-0.5">
                  {app.metadata.id}
                </CardDescription>
              </div>
            </div>
            <Badge
              variant={phaseConf.variant}
              className={cn(
                'gap-1.5 select-none shrink-0',
                phase === 'Deploying' && 'animate-pulse',
              )}
            >
              {phase === 'Deploying' && (
                <span className="relative flex h-1.5 w-1.5">
                  <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-current opacity-75" />
                  <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-current" />
                </span>
              )}
              {phaseConf.label}
            </Badge>
          </div>
        </CardHeader>

        {/* Content */}
        <CardContent className="pb-3 space-y-2.5">
          {/* Health + Runtime + Environment row */}
          <div className="flex flex-wrap items-center gap-2">
            {/* Health badge */}
            <Badge
              variant={healthConf.variant}
              className="gap-1 text-[11px] px-2 py-0.5 select-none"
            >
              <span className="text-[10px]" aria-hidden="true">
                {healthConf.icon}
              </span>
              {health}
            </Badge>

            {/* Runtime type */}
            {runtimeType && (
              <span className="inline-flex items-center gap-1 rounded-md border border-border/50 px-1.5 py-0.5 text-[11px] font-medium text-muted-foreground">
                <Terminal className="h-3 w-3" />
                {runtimeType}
              </span>
            )}

            {/* Environment tag */}
            {env && (
              <span
                className={cn(
                  'inline-flex items-center rounded-md border px-1.5 py-0.5 text-[11px] font-medium leading-none uppercase tracking-wider',
                  environmentConfig[env] ?? environmentConfig.development,
                )}
              >
                {env}
              </span>
            )}
          </div>

          {/* Details row */}
          <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
            {commitSha && (
              <span className="inline-flex items-center gap-1">
                <GitCommitHorizontal className="h-3 w-3" />
                {truncate(commitSha, 7)}
              </span>
            )}

            {deploymentNumber != null && (
              <span className="inline-flex items-center gap-1">
                <Layers className="h-3 w-3" />
                #{deploymentNumber}
              </span>
            )}

            {duration && (
              <span className="inline-flex items-center gap-1">
                <Clock className="h-3 w-3" />
                {duration}
              </span>
            )}

            {url && (
              <span className="inline-flex items-center gap-1 min-w-0">
                <Globe className="h-3 w-3 shrink-0" />
                <span className="truncate">{truncate(url, 30)}</span>
              </span>
            )}
          </div>
        </CardContent>

        {/* Separator */}
        <div role="none" className="mx-6 h-px bg-border/50" />

        {/* Footer — Quick action buttons */}
        <CardFooter className="pt-3">
          <div className="flex w-full items-center gap-1.5" onClick={(e) => e.stopPropagation()}>
            {url && (
              <Button variant="ghost" size="sm" className="h-8 gap-1 text-xs" asChild>
                <NavLink to={url} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="h-3.5 w-3.5" />
                  Open
                </NavLink>
              </Button>
            )}

            <Button
              variant="ghost"
              size="sm"
              className="h-8 gap-1 text-xs"
              asChild
            >
              <NavLink to={`/applications/${app.metadata.id}/logs`}>
                <Terminal className="h-3.5 w-3.5" />
                Logs
              </NavLink>
            </Button>

            <Button
              variant="ghost"
              size="sm"
              className="h-8 gap-1 text-xs"
              asChild
            >
              <NavLink to={`/applications/${app.metadata.id}/timeline`}>
                <Clock className="h-3.5 w-3.5" />
                Timeline
              </NavLink>
            </Button>

            <Button
              variant="ghost"
              size="sm"
              className="h-8 gap-1 text-xs"
              asChild
            >
              <NavLink to={`/applications/${app.metadata.id}`}>
                <Globe className="h-3.5 w-3.5" />
                Status
              </NavLink>
            </Button>
          </div>
        </CardFooter>
      </Card>
    </motion.div>
  );
}
