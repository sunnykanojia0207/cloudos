import { useController, useControllerHealth } from '@/hooks/useCloudOS';
import { usePageTitle } from '@/hooks/usePageTitle';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import {
  AlertCircle,
  ArrowLeft,
  Activity,
  PlayCircle,
  StopCircle,
  AlertTriangle,
  RefreshCw,
  Hash,
  Clock,
  XCircle,
} from 'lucide-react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import type { ControllerDTO, ControllerHealthDTO } from '@cloudos/sdk';

const stateBadgeVariant = (state: string): 'default' | 'secondary' | 'destructive' | 'success' | 'outline' => {
  switch (state) {
    case 'running':
      return 'success';
    case 'stopped':
      return 'secondary';
    case 'failed':
      return 'destructive';
    default:
      return 'outline';
  }
};

const stateIcon = (state: string) => {
  switch (state) {
    case 'running':
      return <PlayCircle className="h-5 w-5 text-emerald-400" />;
    case 'stopped':
      return <StopCircle className="h-5 w-5 text-muted-foreground" />;
    case 'failed':
      return <AlertTriangle className="h-5 w-5 text-destructive" />;
    default:
      return <Activity className="h-5 w-5 text-muted-foreground" />;
  }
};

function formatTime(t?: string): string {
  if (!t) return 'Never';
  try {
    return new Date(t).toLocaleString();
  } catch {
    return t;
  }
}

export default function ControllerDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const {
    data: controller,
    isLoading,
    error,
  } = useController(id ?? '');

  usePageTitle(controller ? `${controller.name} • Controllers` : 'Controller Detail');

  const { data: health } = useControllerHealth(id ?? '');

  const ctrl = controller as ControllerDTO | undefined;
  const ctrlHealth = health as ControllerHealthDTO | undefined;

  return (
    <div className="space-y-6">
      {/* Back button */}
      <Button
        variant="ghost"
        className="gap-2 -ml-2"
        onClick={() => navigate('/controllers')}
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Controllers
      </Button>

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Failed to load controller</AlertTitle>
          <AlertDescription>
            {(error as Error)?.message || 'An unknown error occurred'}
          </AlertDescription>
        </Alert>
      )}

      {/* Header */}
      {isLoading ? (
        <div className="space-y-3">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-4 w-96" />
        </div>
      ) : ctrl ? (
        <div className="flex items-center gap-3">
          {stateIcon(ctrl.state)}
          <div>
            <h1 className="text-3xl font-bold tracking-tight">{ctrl.name}</h1>
            <p className="mt-1 flex items-center gap-2 text-muted-foreground">
              <span>Resource Kind:</span>
              <Badge variant="outline" className="font-mono text-xs">
                {ctrl.kind}
              </Badge>
              <Badge variant={stateBadgeVariant(ctrl.state)} className="ml-2">
                {ctrl.state}
              </Badge>
            </p>
          </div>
        </div>
      ) : null}

      {ctrl?.message && (
        <p className="text-sm text-muted-foreground">{ctrl.message}</p>
      )}

      {/* Health Details */}
      {ctrlHealth && (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">
                <div className="flex items-center gap-2">
                  <RefreshCw className="h-4 w-4 text-primary" />
                  Reconcile Count
                </div>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{ctrlHealth.reconcileCount}</div>
              <p className="text-xs text-muted-foreground">Total reconciliations</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">
                <div className="flex items-center gap-2">
                  <XCircle
                    className={cn(
                      'h-4 w-4',
                      ctrlHealth.errorCount > 0
                        ? 'text-destructive'
                        : 'text-muted-foreground',
                    )}
                  />
                  Error Count
                </div>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div
                className={cn(
                  'text-2xl font-bold',
                  ctrlHealth.errorCount > 0 && 'text-destructive',
                )}
              >
                {ctrlHealth.errorCount}
              </div>
              <p className="text-xs text-muted-foreground">Failed reconciliations</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">
                <div className="flex items-center gap-2">
                  <Clock className="h-4 w-4 text-primary" />
                  Last Reconciled
                </div>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-sm font-bold">
                {formatTime(ctrlHealth.lastReconciled)}
              </div>
              <p className="text-xs text-muted-foreground">
                {ctrlHealth.lastReconciled
                  ? `${Math.round(
                      (Date.now() - new Date(ctrlHealth.lastReconciled).getTime()) /
                        1000,
                    )}s ago`
                  : 'No reconciliation yet'}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">
                <div className="flex items-center gap-2">
                  <Hash className="h-4 w-4 text-primary" />
                  Resource Kind
                </div>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                <Badge variant="outline" className="font-mono text-lg">
                  {ctrlHealth.kind}
                </Badge>
              </div>
              <p className="text-xs text-muted-foreground">Watched resource kind</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Full Health JSON */}
      {ctrlHealth && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Controller Health</CardTitle>
            <CardDescription>Full health snapshot</CardDescription>
          </CardHeader>
          <CardContent>
            <pre className="overflow-x-auto rounded-lg bg-muted p-4 text-xs">
              {JSON.stringify(ctrlHealth, null, 2)}
            </pre>
          </CardContent>
        </Card>
      )}

      {!isLoading && !ctrl && !error && (
        <div className="flex min-h-[200px] items-center justify-center rounded-lg border border-dashed">
          <p className="text-sm text-muted-foreground">
            Controller "{id}" not found.
          </p>
        </div>
      )}
    </div>
  );
}
