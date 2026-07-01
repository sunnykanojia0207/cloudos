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
import { Clock, GitCommitHorizontal, Layers, Eye, GitCompare } from 'lucide-react';

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

  // Sort descending by deployment number (newest first)
  const sorted = React.useMemo(() => {
    if (!deploymentHistory || deploymentHistory.length === 0) return [];
    return [...deploymentHistory].sort(
      (a, b) => b.deploymentNumber - a.deploymentNumber,
    );
  }, [deploymentHistory]);

  if (sorted.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <Layers className="h-10 w-10 text-muted-foreground/30 mb-3" />
        <p className="text-sm text-muted-foreground">No deployments found.</p>
        <p className="text-xs text-muted-foreground/60 mt-1">
          Deployments will appear here once the application is deployed.
        </p>
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-border/50 overflow-hidden">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[120px]">Deployment</TableHead>
            <TableHead className="w-[110px]">Commit</TableHead>
            <TableHead className="w-[100px]">Status</TableHead>
            <TableHead className="w-[100px]">Duration</TableHead>
            <TableHead className="hidden md:table-cell">Started</TableHead>
            <TableHead className="hidden md:table-cell">Completed</TableHead>
            <TableHead className="w-[130px] text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {sorted.map((dep, index) => {
            const isLatest = index === 0;
            const isFirstDeployment = dep.deploymentNumber === 1;

            return (
              <TableRow key={dep.deploymentNumber}>
                {/* Deployment # */}
                <TableCell>
                  <button
                    type="button"
                    onClick={() =>
                      navigate(
                        `/applications/${appId}/timeline?deployment=${dep.deploymentNumber}`,
                      )
                    }
                    className={cn(
                      'inline-flex items-center gap-1.5 text-sm font-medium',
                      'text-primary hover:text-primary/80 transition-colors',
                      'focus-visible:outline-none focus-visible:underline',
                    )}
                    aria-label={`View timeline for deployment #${dep.deploymentNumber}`}
                  >
                    <Layers className="h-3.5 w-3.5" />
                    #{dep.deploymentNumber}
                    {isLatest && (
                      <Badge
                        variant="success"
                        className="ml-1 text-[10px] px-1.5 py-0 h-4"
                      >
                        Latest
                      </Badge>
                    )}
                  </button>
                </TableCell>

                {/* Commit SHA */}
                <TableCell>
                  {dep.commitSha ? (
                    <span className="inline-flex items-center gap-1 text-xs font-mono text-muted-foreground">
                      <GitCommitHorizontal className="h-3 w-3" />
                      {truncate(dep.commitSha, 7)}
                    </span>
                  ) : (
                    <span className="text-xs text-muted-foreground/50">\u2014</span>
                  )}
                </TableCell>

                {/* Build Success */}
                <TableCell>
                  <Badge
                    variant={dep.buildSuccess ? 'success' : 'destructive'}
                    className="gap-1 text-[11px] px-2 py-0.5"
                  >
                    {dep.buildSuccess ? (
                      <>
                        <span className="text-xs">✓</span>
                        Success
                      </>
                    ) : (
                      <>
                        <span className="text-xs">✗</span>
                        Failed
                      </>
                    )}
                  </Badge>
                </TableCell>

                {/* Duration */}
                <TableCell>
                  <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
                    <Clock className="h-3 w-3" />
                    {dep.duration || '\u2014'}
                  </span>
                </TableCell>

                {/* Started (desktop only) */}
                <TableCell className="hidden md:table-cell">
                  <span className="text-xs text-muted-foreground">
                    {formatDateTime(dep.startedAt)}
                  </span>
                </TableCell>

                {/* Completed (desktop only) */}
                <TableCell className="hidden md:table-cell">
                  <span className="text-xs text-muted-foreground">
                    {formatDateTime(dep.completedAt)}
                  </span>
                </TableCell>

                {/* Actions */}
                <TableCell className="text-right">
                  <div className="flex items-center justify-end gap-1">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 w-8 p-0"
                      aria-label={`View timeline for deployment #${dep.deploymentNumber}`}
                      onClick={() =>
                        navigate(
                          `/applications/${appId}/timeline?deployment=${dep.deploymentNumber}`,
                        )
                      }
                    >
                      <Eye className="h-3.5 w-3.5" />
                    </Button>

                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 w-8 p-0"
                      disabled={isFirstDeployment}
                      aria-label={
                        isFirstDeployment
                          ? 'Cannot compare first deployment'
                          : `Compare deployment #${dep.deploymentNumber} with previous`
                      }
                      onClick={() =>
                        navigate(
                          `/applications/${appId}/compare?from=${dep.deploymentNumber - 1}&to=${dep.deploymentNumber}`,
                        )
                      }
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
