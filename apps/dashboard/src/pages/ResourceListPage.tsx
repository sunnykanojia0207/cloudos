import { useParams, Link } from 'react-router-dom';
import { useResources } from '@/hooks/useCloudOS';
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
  ArrowLeft,
  Database,
  CheckCircle2,
  XCircle,
} from 'lucide-react';
import { Button } from '@/components/ui/button';

export default function ResourceListPage() {
  const { kind } = useParams<{ kind: string }>();
  const { data, isLoading, error } = useResources(kind ?? '');

  // ResourceObject items from the ResourceList
  const items = (data as any)?.items ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" asChild>
          <Link to="/resources">
            <ArrowLeft className="h-4 w-4" />
          </Link>
        </Button>
        <div className="flex items-center gap-2">
          <Database className="h-6 w-6 text-primary" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">{kind}</h1>
            <p className="mt-1 text-muted-foreground">
              {items.length > 0
                ? `${items.length} resource${items.length !== 1 ? 's' : ''} found`
                : 'Resources of this kind'}
            </p>
          </div>
        </div>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Failed to load resources</AlertTitle>
          <AlertDescription>
            {(error as Error)?.message || 'An unknown error occurred'}
          </AlertDescription>
        </Alert>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {isLoading
          ? Array.from({ length: 3 }).map((_, i) => (
              <Card key={i}>
                <CardHeader>
                  <Skeleton className="h-5 w-32" />
                </CardHeader>
                <CardContent>
                  <Skeleton className="h-4 w-24" />
                </CardContent>
              </Card>
            ))
          : items.map((item: any) => {
              const meta = item.metadata ?? {};
              const status = item.status ?? {};
              const isActive = status.phase === 'Active' || status.status === 'ready';
              return (
                <Link
                  key={meta.id}
                  to={`/resources/${kind}/${meta.id}`}
                >
                  <Card className="h-full transition-colors hover:border-primary/50">
                    <CardHeader>
                      <div className="flex items-start justify-between">
                        <CardTitle className="text-lg">
                          {meta.name || meta.id}
                        </CardTitle>
                        {isActive != null && (
                          <Badge variant={isActive ? 'success' : 'secondary'}>
                            {isActive ? 'Active' : 'Inactive'}
                          </Badge>
                        )}
                      </div>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-1 text-sm text-muted-foreground">
                        <div className="flex items-center gap-1.5">
                          {isActive ? (
                            <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400" />
                          ) : (
                            <XCircle className="h-3.5 w-3.5 text-destructive" />
                          )}
                          <span>ID: {meta.id}</span>
                        </div>
                        {meta.namespace && (
                          <div>
                            Namespace: {meta.namespace}
                          </div>
                        )}
                      </div>
                    </CardContent>
                  </Card>
                </Link>
              );
            })}
      </div>

      {!isLoading && items.length === 0 && !error && (
        <div className="flex min-h-[200px] items-center justify-center rounded-lg border border-dashed">
          <p className="text-sm text-muted-foreground">
            No resources of kind "{kind}" found.
          </p>
        </div>
      )}
    </div>
  );
}
