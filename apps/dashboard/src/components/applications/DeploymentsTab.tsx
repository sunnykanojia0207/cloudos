import * as React from 'react';
import { useNavigate } from 'react-router-dom';
import type { DeploymentReport } from '@/hooks/useApplications';
import { cn, relativeTime, truncate } from '@/lib/utils';
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { HealthIndicator } from '@/components/ui/health-indicator';
import { EmptyState } from '@/components/ui/empty-state';
import {
  Clock,
  GitCommitHorizontal,
  Layers,
  Eye,
  GitCompare,
  Rocket,
} from 'lucide-react';

/* ── Helpers ──────────────────────────────────────────── */
function formatDateTime(dateStr?: string): string {
  if (!dateStr) return '\u2014';
  try {
    return new Date(dateStr).toLocaleString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return dateStr;
  }
}

/* ── Props ────────────────────────────────────────────── */
export interface DeploymentsTabProps {
  appId: string;
  deploymentHistory: DeploymentReport[];
}

/* ── Component ────────────────────────────────────────── */
export function DeploymentsTab({ appId, deploymentHistory }: DeploymentsTabProps) {
  const navigate = useNavigate();

  const sorted = React.useMemo(() => {
    if (!deploymentHistory || deploymentHistory.length === 0) return [];
    return [...deploymentHistory].sort(
      (a, b) => b.deploymentNumber - a.deploymentNumber,
    );
  }, [deploymentHistory]);

  if (sorted.length === 0) {
    return (
      <EmptyState
        icon={Rocket}
        title="No deployments yet"
        description="Deployments will appear here once the application is deployed."
        actions={[
          {
            label: 'Deploy Application',
            icon: Rocket,
            onClick: () => navigate(`/applications/${appId}/deploy`),
            variant: 'primary',
          },
        ]}
      />
    );
  }

  return (
    <div className="rounded-md border border-border overflow-hidden">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[130px]">Deployment</TableHead>
            <TableHead className="w-[110px]">Commit</TableHead>
            <TableHead className="w-[110px]">Status</TableHead>
            <TableHead className="w-[90px]">Duration</TableHead>
            <TableHead className="hidden md:table-cell">Started</TableHead>
            <TableHead className="w-[100px] text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {sorted.map((dep, index) => {
            const isLatest = index === 0;

            return (
              <TableRow key={dep.deploymentNumber}>
                {/* Deployment # */}
                <TableCell>
                  <button
                    type="button"
                    onClick={() => navigate(`/applications/${appId}/timeline?deployment=${dep.deploymentNumber}`)}
                    className={cn(
                      'inline-flex items-center gap-1.5 text-small font-medium',
                      'text-accent hover:text-accent-hover transition-colors',
                      'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded-sm',
                    )}
                    aria-label={`View timeline for deployment #${dep.deploymentNumber}`}
                  >
                    <Layers className="h-3.5 w-3.5" />
                    #{dep.deploymentNumber}
                    {isLatest && (
                      <Badge variant="subtle-success" className="ml-1 text-caption px-1.5 py-0 h-4 font-medium">
                        Latest
                      </Badge>
                    )}
                  </button>
                </TableCell>

                {/* Commit SHA */}
                <TableCell>
                  {dep.commitSha ? (
                    <span className="inline-flex items-center gap-1 text-caption font-mono text-text-secondary">
                      <GitCommitHorizontal className="h-3 w-3" />
                      {truncate(dep.commitSha, 7)}
                    </span>
                  ) : (
                    <span className="text-caption text-text-muted">\u2014</span>
                  )}
                </TableCell>

                {/* Status */}
                <TableCell>
                  <HealthIndicator
                    status={dep.buildSuccess ? 'healthy' : 'failed'}
                    showLabel
                    size="sm"
                  />
                </TableCell>

                {/* Duration */}
                <TableCell>
                  <span className="inline-flex items-center gap-1 text-caption text-text-muted tabular-nums">
                    <Clock className="h-3 w-3" />
                    {dep.duration || '\u2014'}
                  </span>
                </TableCell>

                {/* Started (desktop only) */}
                <TableCell className="hidden md:table-cell">
                  <span className="text-caption text-text-muted">
                    {formatDateTime(dep.startedAt)}
                  </span>
                </TableCell>

                {/* Actions */}
                <TableCell className="text-right">
                  <div className="flex items-center justify-end gap-1">
                    <Button
                      variant="icon-ghost"
                      size="icon-sm"
                      aria-label={`View timeline for deployment #${dep.deploymentNumber}`}
                      onClick={() => navigate(`/applications/${appId}/timeline?deployment=${dep.deploymentNumber}`)}
                    >
                      <Eye className="h-3.5 w-3.5" />
                    </Button>

                    <Button
                      variant="icon-ghost"
                      size="icon-sm"
                      disabled={dep.deploymentNumber <= 1}
                      aria-label={
                        dep.deploymentNumber <= 1
                          ? 'Cannot compare first deployment'
                          : `Compare deployment #${dep.deploymentNumber} with previous`
                      }
                      onClick={() => navigate(`/applications/${appId}/compare?from=${dep.deploymentNumber - 1}&to=${dep.deploymentNumber}`)}
                    >
                      <GitCompare className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </div>
  );
}
