import { useControllers } from '@/hooks/useCloudOS';
import { usePageTitle } from '@/hooks/usePageTitle';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import {
  AlertCircle,
  Activity,
  PlayCircle,
  StopCircle,
  AlertTriangle,
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import type { ControllerDTO } from '@cloudos/sdk';

const stateIcon = (state: string) => {
  switch (state) {
    case 'running':
      return <PlayCircle className="h-4 w-4 text-emerald-400" />;
    case 'stopped':
      return <StopCircle className="h-4 w-4 text-muted-foreground" />;
    case 'failed':
      return <AlertTriangle className="h-4 w-4 text-destructive" />;
    default:
      return <Activity className="h-4 w-4 text-muted-foreground" />;
  }
};

const stateVariant = (state: string): 'default' | 'secondary' | 'destructive' | 'success' | 'outline' => {
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

export default function ControllersPage() {
  usePageTitle('Controllers');
  const { data, isLoading, error } = useControllers();
  const navigate = useNavigate();
  const controllers = data?.controllers ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Activity className="h-6 w-6 text-primary" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Controllers</h1>
          <p className="mt-1 text-muted-foreground">
            {data
              ? `${data.total} controller${data.total !== 1 ? 's' : ''} registered`
              : 'CloudOS controller runtime'}
          </p>
        </div>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Failed to load controllers</AlertTitle>
          <AlertDescription>
            {(error as Error)?.message || 'An unknown error occurred'}
          </AlertDescription>
        </Alert>
      )}

      <div className="grid gap-4 sm:grid-cols-2">
        {isLoading
          ? Array.from({ length: 2 }).map((_, i) => (
              <Card key={i}>
                <CardHeader>
                  <Skeleton className="h-5 w-40" />
                </CardHeader>
                <CardContent className="space-y-3">
                  <Skeleton className="h-4 w-20" />
                  <Skeleton className="h-4 w-32" />
                </CardContent>
              </Card>
            ))
          : controllers.map((ctrl: ControllerDTO) => (
              <Card
                key={ctrl.name}
                className="cursor-pointer transition-colors hover:border-primary/50"
                onClick={() => navigate(`/controllers/${ctrl.name}`)}
              >
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <div>
                      <CardTitle className="text-lg flex items-center gap-2">
                        {stateIcon(ctrl.state)}
                        <span>{ctrl.name}</span>
                      </CardTitle>
                    </div>
                    <Badge variant={stateVariant(ctrl.state)} className="shrink-0">
                      {ctrl.state}
                    </Badge>
                  </div>
                </CardHeader>
                <CardContent className="space-y-2">
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <span className="font-mono text-xs">Kind:</span>
                    <Badge variant="outline" className="font-mono text-xs">
                      {ctrl.kind}
                    </Badge>
                  </div>

                  {ctrl.health && (
                    <div className="flex items-center gap-4 text-xs text-muted-foreground">
                      <span>
                        Reconciliations: {ctrl.health.reconcileCount}
                      </span>
                      {ctrl.health.errorCount > 0 && (
                        <span className="text-destructive">
                          Errors: {ctrl.health.errorCount}
                        </span>
                      )}
                    </div>
                  )}

                  {ctrl.message && (
                    <p className="text-xs text-muted-foreground">
                      {ctrl.message}
                    </p>
                  )}
                </CardContent>
              </Card>
            ))}
      </div>

      {!isLoading && controllers.length === 0 && !error && (
        <div className="flex min-h-[200px] items-center justify-center rounded-lg border border-dashed">
          <p className="text-sm text-muted-foreground">
            No controllers registered.
          </p>
        </div>
      )}
    </div>
  );
}
