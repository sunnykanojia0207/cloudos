import * as React from 'react';
import { motion } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import type { AppResource } from '@/hooks/useApplications';
import { cn, truncate } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { HealthIndicator } from '@/components/ui/health-indicator';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  ExternalLink,
  Terminal,
  GitCommitHorizontal,
  GitBranch,
  Clock,
  Globe,
  Box,
  MoreHorizontal,
  Rocket,
  List,
  Heart,
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';

/* ── Phase → HealthIndicator status map ─────────────────── */
function phaseToHealthStatus(phase: string): 'healthy' | 'deploying' | 'failed' | 'stopped' | 'running' {
  switch (phase) {
    case 'Running': return 'running';
    case 'Deploying': return 'deploying';
    case 'Failed': return 'failed';
    case 'Stopped': return 'stopped';
    default: return 'stopped';
  }
}

/* ── Environment badge style ────────────────────────────── */
const envStyles: Record<string, string> = {
  production: 'bg-success-subtle text-success border-success/20',
  staging: 'bg-warning-subtle text-warning border-warning/20',
  development: 'bg-accent-subtle text-accent border-accent/20',
  testing: 'bg-info-subtle text-info border-info/20',
};

/* ── Props ──────────────────────────────────────────────── */
export interface ApplicationCardProps {
  app: AppResource;
}

/* ── Component ──────────────────────────────────────────── */
export function ApplicationCard({ app }: ApplicationCardProps) {
  const navigate = useNavigate();

  const phase = app.status?.phase ?? 'Stopped';
  const health = app.status?.health ?? 'Unknown';
  const env = app.spec?.settings?.environment ?? app.status?.lastReport?.environment;
  const lastReport = app.status?.lastReport;
  const commitSha = lastReport?.commitSha;
  const deploymentNumber = lastReport?.deploymentNumber;
  const duration = lastReport?.duration;
  const url = app.status?.url;
  const repoUrl = app.spec?.source?.url;
  const branch = app.spec?.source?.branch ?? lastReport?.branch;
  const runtimeType = app.spec?.runtime?.type ?? lastReport?.detectedRuntime;
  const appId = app.metadata.id;

  const handleClick = () => navigate(`/applications/${appId}`);
  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -8 }}
      transition={{ duration: 0.2, ease: [0, 0, 0.2, 1] }}
    >
      <Card
        variant="interactive"
        className="group relative overflow-hidden cursor-pointer"
        onClick={handleClick}
        aria-label={`Application ${app.metadata.name}`}
      >
        {/* ── Top row: Icon + Name/ID + Status ─────────── */}
        <CardContent className="pb-0">
          <div className="flex items-start justify-between gap-3">
            <div className="flex items-center gap-3 min-w-0 flex-1">
              {/* App icon */}
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-border bg-surface">
                <Box className="h-4 w-4 text-text-secondary" />
              </div>

              {/* Name + ID */}
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <h3 className="text-body font-semibold text-foreground truncate">
                    {app.metadata.name}
                  </h3>
                  {env && (
                    <span className={cn(
                      'shrink-0 inline-flex items-center rounded-sm border px-1.5 py-0.5 text-caption font-medium uppercase tracking-wider',
                      envStyles[env] ?? 'bg-surface border-border text-text-muted',
                    )}>
                      {env}
                    </span>
                  )}
                </div>
                <p className="text-small text-text-muted truncate mt-0.5">
                  {appId}
                </p>
              </div>
            </div>

            {/* Phase status badge */}
            <Badge
              variant={phase === 'Running' ? 'subtle-success' : phase === 'Deploying' ? 'subtle-accent' : phase === 'Failed' ? 'subtle-danger' : 'subtle-neutral'}
              className={cn('shrink-0 gap-1.5', phase === 'Deploying' && 'animate-pulse')}
            >
              {phase === 'Deploying' && (
                <span className="relative flex h-1.5 w-1.5">
                  <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-current opacity-75" />
                  <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-current" />
                </span>
              )}
              {phase}
            </Badge>
          </div>
        </CardContent>

        {/* ── Meta row: Health + Runtime + Deploy info ─── */}
        <CardContent className="pb-0 pt-2.5">
          <div className="flex flex-wrap items-center gap-x-4 gap-y-1.5">
            {/* Health */}
            <HealthIndicator status={phaseToHealthStatus(phase)} showLabel size="sm" />

            {/* Runtime */}
            {runtimeType && (
              <span className="inline-flex items-center gap-1 text-small text-text-secondary">
                <Terminal className="h-3.5 w-3.5" />
                {runtimeType}
              </span>
            )}

            {/* Deployment number */}
            {deploymentNumber != null && (
              <span className="inline-flex items-center gap-1 text-small text-text-secondary">
                <Rocket className="h-3.5 w-3.5" />
                #{deploymentNumber}
              </span>
            )}

            {/* Duration */}
            {duration && (
              <span className="inline-flex items-center gap-1 text-small text-text-secondary">
                <Clock className="h-3.5 w-3.5" />
                {duration}
              </span>
            )}
          </div>
        </CardContent>

        {/* ── Details: Repo / Branch / Commit / URL ────── */}
        <CardContent className="pb-0 pt-2.5">
          <div className="flex flex-wrap items-center gap-x-4 gap-y-1.5">
            {/* Repository */}
            {repoUrl && (
              <span className="inline-flex items-center gap-1.5 text-small text-text-muted min-w-0">
                <Globe className="h-3.5 w-3.5 shrink-0" />
                <span className="truncate">{truncate(repoUrl.replace('https://', ''), 30)}</span>
              </span>
            )}

            {/* Branch */}
            {branch && (
              <span className="inline-flex items-center gap-1 text-small text-text-muted">
                <GitBranch className="h-3.5 w-3.5 shrink-0" />
                <span className="truncate">{branch}</span>
              </span>
            )}

            {/* Commit SHA */}
            {commitSha && (
              <span className="inline-flex items-center gap-1 text-small text-text-muted font-mono">
                <GitCommitHorizontal className="h-3.5 w-3.5 shrink-0" />
                {truncate(commitSha, 7)}
              </span>
            )}

            {/* URL */}
            {url && (
              <span className="inline-flex items-center gap-1 text-small text-accent min-w-0">
                <ExternalLink className="h-3.5 w-3.5 shrink-0" />
                <span className="truncate">{truncate(url, 30)}</span>
              </span>
            )}
          </div>
        </CardContent>

        {/* ── Footer: Actions ──────────────────────────── */}
        <CardContent className="pt-3">
          <div className="flex items-center gap-1 border-t border-border pt-2.5" onClick={(e) => e.stopPropagation()}>
            {/* Open (primary action) */}
            {url && (
              <Button variant="ghost" size="sm" className="h-7 gap-1 text-small" asChild>
                <a href={url} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="h-3.5 w-3.5" />
                  Open
                </a>
              </Button>
            )}

            {/* Logs */}
            <Button
              variant="ghost"
              size="sm"
              className="h-7 gap-1 text-small"
              onClick={() => navigate(`/applications/${appId}/logs`)}
            >
              <Terminal className="h-3.5 w-3.5" />
              Logs
            </Button>

            {/* Timeline */}
            <Button
              variant="ghost"
              size="sm"
              className="h-7 gap-1 text-small"
              onClick={() => navigate(`/applications/${appId}/timeline`)}
            >
              <List className="h-3.5 w-3.5" />
              Timeline
            </Button>

            {/* Status */}
            <Button
              variant="ghost"
              size="sm"
              className="h-7 gap-1 text-small"
              onClick={() => navigate(`/applications/${appId}`)}
            >
              <Heart className="h-3.5 w-3.5" />
              Status
            </Button>

            {/* Spacer */}
            <div className="flex-1" />

            {/* More actions dropdown */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="icon-ghost" size="icon-sm" aria-label="More actions">
                  <MoreHorizontal className="h-3.5 w-3.5" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onSelect={() => navigate(`/applications/${appId}`)}>
                  View Details
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={() => navigate(`/applications/${appId}/deployments`)}>
                  Deployments
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem onSelect={() => {}} className="text-danger">
                  Stop Application
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}
