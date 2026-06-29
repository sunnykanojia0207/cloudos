import * as React from 'react';
import { motion } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import type { ProjectDTO } from '@cloudos/sdk';
import { cn, relativeTime, truncate, formatNumber } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from '@/components/ui/card';
import { Clock, Box, Globe, Layers } from 'lucide-react';

/* ── Helpers ──────────────────────────────────────────── */

const phaseConfig: Record<
  string,
  { variant: 'success' | 'warning' | 'secondary' | 'destructive'; label: string }
> = {
  Active: { variant: 'success', label: 'Active' },
  Creating: { variant: 'warning', label: 'Creating' },
  Archived: { variant: 'secondary', label: 'Archived' },
  Deleting: { variant: 'destructive', label: 'Deleting' },
};

const environmentLabel: Record<string, string> = {
  development: 'Dev',
  staging: 'Staging',
  production: 'Prod',
  testing: 'Test',
};

const environmentColor: Record<string, string> = {
  development:
    'bg-blue-500/10 text-blue-400 border-blue-500/20',
  staging:
    'bg-amber-500/10 text-amber-400 border-amber-500/20',
  production:
    'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
  testing:
    'bg-purple-500/10 text-purple-400 border-purple-500/20',
};

/* ── Props ────────────────────────────────────────────── */

export interface ProjectCardProps {
  project: ProjectDTO;
  view: 'grid' | 'list';
}

/* ── Component ────────────────────────────────────────── */

export function ProjectCard({ project, view }: ProjectCardProps) {
  const navigate = useNavigate();
  const phase = project.status?.phase ?? 'Creating';
  const config = phaseConfig[phase] ?? phaseConfig.Creating;
  const env = project.spec.environment;
  const lastActivity = project.status?.lastActivity;

  const handleClick = () => {
    navigate(`/projects/${project.metadata.id}`);
  };

  const envTag = (
    <span
      className={cn(
        'inline-flex items-center rounded-md border px-1.5 py-0.5 text-[11px] font-medium leading-none uppercase tracking-wider',
        environmentColor[env] ?? environmentColor.development,
      )}
    >
      {environmentLabel[env] ?? env}
    </span>
  );

  const statusBadge = (
    <Badge
      variant={config.variant}
      className={cn(
        'gap-1.5 select-none',
        phase === 'Creating' && 'animate-pulse',
      )}
    >
      {phase === 'Creating' && (
        <span className="relative flex h-1.5 w-1.5">
          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-current opacity-75" />
          <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-current" />
        </span>
      )}
      {config.label}
    </Badge>
  );

  const activityDisplay = lastActivity ? (
    <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
      <Clock className="h-3 w-3" />
      {relativeTime(lastActivity)}
    </span>
  ) : null;

  const resourceDisplay = project.status?.resourceCount != null ? (
    <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
      <Box className="h-3 w-3" />
      {formatNumber(project.status.resourceCount)} resources
    </span>
  ) : null;

  /* ── Grid Mode ──────────────────────────────────────── */

  if (view === 'grid') {
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
            'group cursor-pointer border-border/50 transition-colors',
            'hover:border-accent/50',
          )}
          onClick={handleClick}
          onKeyDown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault();
              handleClick();
            }
          }}
          tabIndex={0}
          role="link"
          aria-label={`Project ${project.spec.displayName}`}
        >
          <CardHeader className="pb-3">
            <div className="flex items-start justify-between gap-3">
              <div className="flex items-center gap-2.5 min-w-0">
                <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border bg-muted/50">
                  <Layers className="h-4 w-4 text-muted-foreground" />
                </div>
                <div className="min-w-0">
                  <CardTitle className="text-sm font-semibold leading-tight truncate">
                    {project.spec.displayName}
                  </CardTitle>
                  <CardDescription className="text-xs mt-0.5">
                    {project.metadata.id}
                  </CardDescription>
                </div>
              </div>
              {statusBadge}
            </div>
          </CardHeader>

          <CardContent className="pb-3">
            {project.spec.description && (
              <p className="text-xs text-muted-foreground leading-relaxed line-clamp-2 mb-2">
                {truncate(project.spec.description, 100)}
              </p>
            )}
            {envTag}
          </CardContent>

          <SeparatorLine />

          <CardFooter className="pt-3">
            <div className="flex w-full items-center justify-between gap-2">
              <div className="flex items-center gap-3">
                {resourceDisplay}
              </div>
              {activityDisplay}
            </div>
          </CardFooter>
        </Card>
      </motion.div>
    );
  }

  /* ── List Mode ──────────────────────────────────────── */

  return (
    <motion.div
      layout
      initial={{ opacity: 0, x: -8 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: -8 }}
      transition={{ duration: 0.2, ease: 'easeOut' }}
    >
      <div
        className={cn(
          'group flex items-center gap-4 rounded-lg border border-border/50 px-4 py-3',
          'cursor-pointer transition-colors hover:border-accent/50 hover:bg-accent/25',
        )}
        onClick={handleClick}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            handleClick();
          }
        }}
        tabIndex={0}
        role="link"
        aria-label={`Project ${project.spec.displayName}`}
      >
        {/* Icon */}
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md border bg-muted/50">
          <Layers className="h-3.5 w-3.5 text-muted-foreground" />
        </div>

        {/* Name + ID */}
        <div className="flex min-w-0 flex-1 items-center gap-3">
          <div className="min-w-0 flex-1">
            <span className="text-sm font-medium leading-tight truncate block">
              {project.spec.displayName}
            </span>
            <span className="text-xs text-muted-foreground truncate block">
              {project.metadata.id}
            </span>
          </div>
        </div>

        {/* Environment tag */}
        <div className="hidden sm:block shrink-0">{envTag}</div>

        {/* Status */}
        <div className="shrink-0">{statusBadge}</div>

        {/* Resource count */}
        {resourceDisplay && (
          <div className="hidden md:flex shrink-0 items-center gap-1">
            {resourceDisplay}
          </div>
        )}

        {/* Activity */}
        <div className="hidden lg:flex shrink-0 items-center gap-1">
          {activityDisplay}
        </div>
      </div>
    </motion.div>
  );
}

/* ── Separator (inline to avoid extra import complexity) ──── */

function SeparatorLine() {
  return (
    <div
      role="none"
      className="mx-6 h-px bg-border/50"
    />
  );
}
