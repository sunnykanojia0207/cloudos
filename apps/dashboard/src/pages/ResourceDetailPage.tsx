import { useParams, Link } from 'react-router-dom';
import { useResource } from '@/hooks/useCloudOS';
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
  ArrowLeft,
  Database,
  Clock,
  CalendarDays,
  Hash,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import type { ResourceObject, ResourceSpec, ResourceStatus, ResourceMeta } from '@cloudos/sdk';

export default function ResourceDetailPage() {
  const { kind, id } = useParams<{ kind: string; id: string }>();
  usePageTitle(id ? `${id} • ${kind} • Resources` : 'Resource Detail');
  const { data, isLoading, error } = useResource(kind ?? '', id ?? '');

  const resource = data as ResourceObject<ResourceSpec, ResourceStatus> | undefined;
  const meta = resource?.metadata ?? ({} as ResourceMeta);
  const spec = resource?.spec ?? ({} as ResourceSpec);
  const status = resource?.status ?? ({} as ResourceStatus);

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" asChild>
          <Link to={`/resources/${kind}`}>
            <ArrowLeft className="h-4 w-4" />
          </Link>
        </Button>
        <div className="flex items-center gap-2">
          <Database className="h-6 w-6 text-primary" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">
              {meta.name || id}
            </h1>
            <p className="mt-1 text-muted-foreground">
              {kind} / {id}
            </p>
          </div>
        </div>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Failed to load resource</AlertTitle>
          <AlertDescription>
            {(error as Error)?.message || 'An unknown error occurred'}
          </AlertDescription>
        </Alert>
      )}

      {isLoading ? (
        <div className="space-y-4">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-32 w-full" />
        </div>
      ) : data ? (
        <>
          {/* Metadata summary */}
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm text-muted-foreground">
                  Kind
                </CardTitle>
              </CardHeader>
              <CardContent>
                <Badge variant="outline">{kind}</Badge>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm text-muted-foreground">
                  Resource Version
                </CardTitle>
              </CardHeader>
              <CardContent className="flex items-center gap-2">
                <Hash className="h-4 w-4 text-muted-foreground" />
                <span className="font-mono text-sm">
                  {meta.resourceVersion ?? '—'}
                </span>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm text-muted-foreground">
                  Created
                </CardTitle>
              </CardHeader>
              <CardContent className="flex items-center gap-2">
                <CalendarDays className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm">
                  {meta.createdAt
                    ? new Date(meta.createdAt).toLocaleString()
                    : '—'}
                </span>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm text-muted-foreground">
                  Updated
                </CardTitle>
              </CardHeader>
              <CardContent className="flex items-center gap-2">
                <Clock className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm">
                  {meta.updatedAt
                    ? new Date(meta.updatedAt).toLocaleString()
                    : '—'}
                </span>
              </CardContent>
            </Card>
          </div>

          {/* Status */}
          {Object.keys(status).length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle>Status</CardTitle>
              </CardHeader>
              <CardContent>
                <pre className="overflow-x-auto rounded-lg bg-muted p-4 font-mono text-xs">
                  {JSON.stringify(status, null, 2)}
                </pre>
              </CardContent>
            </Card>
          )}

          {/* Spec */}
          {Object.keys(spec).length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle>Spec</CardTitle>
              </CardHeader>
              <CardContent>
                <pre className="overflow-x-auto rounded-lg bg-muted p-4 font-mono text-xs">
                  {JSON.stringify(spec, null, 2)}
                </pre>
              </CardContent>
            </Card>
          )}

          {/* Labels & Annotations */}
          {meta.labels && Object.keys(meta.labels).length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle>Labels</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-wrap gap-2">
                  {Object.entries(meta.labels).map(([k, v]) => (
                    <Badge key={k} variant="outline">
                      {k}: {v as string}
                    </Badge>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </>
      ) : null}
    </div>
  );
}
