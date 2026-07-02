import { useParams, Link } from 'react-router-dom';
import { useCapability } from '@/hooks/useCloudOS';
import { usePageTitle } from '@/hooks/usePageTitle';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import {
  AlertCircle,
  ArrowLeft,
  Boxes,
  CheckCircle2,
  XCircle,
} from 'lucide-react';
import { Button } from '@/components/ui/button';

export default function CapabilityDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { data, isLoading, error } = useCapability(id ?? '');

  usePageTitle(data ? `${data.spec.displayName} • Capabilities` : 'Capability Detail');

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" asChild>
          <Link to="/capabilities">
            <ArrowLeft className="h-4 w-4" />
          </Link>
        </Button>
        <div className="flex items-center gap-2">
          <Boxes className="h-6 w-6 text-primary" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">
              {data?.spec?.displayName ?? id}
            </h1>
            <p className="mt-1 text-text-secondary">
              Capability: {data?.metadata?.id ?? id}
            </p>
          </div>
        </div>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Failed to load capability</AlertTitle>
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
          <div className="grid gap-4 sm:grid-cols-3">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm text-text-secondary">
                  Status
                </CardTitle>
              </CardHeader>
              <CardContent>
                <Badge
                  variant={
                    data.status?.status === 'stable' ? 'success' : 'warning'
                  }
                >
                  {data.status?.status ?? 'unknown'}
                </Badge>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm text-text-secondary">
                  Available
                </CardTitle>
              </CardHeader>
              <CardContent>
                {data.status?.available ? (
                  <div className="flex items-center gap-1.5 text-emerald-400">
                    <CheckCircle2 className="h-4 w-4" />
                    <span className="font-semibold">Yes</span>
                  </div>
                ) : (
                  <div className="flex items-center gap-1.5 text-destructive">
                    <XCircle className="h-4 w-4" />
                    <span className="font-semibold">No</span>
                  </div>
                )}
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm text-text-secondary">
                  Providers
                </CardTitle>
              </CardHeader>
              <CardContent>
                <span className="text-2xl font-bold">
                  {data.status?.providerCount ?? 0}
                </span>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Description</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-text-secondary">
                {data.spec?.description || 'No description available.'}
              </p>
            </CardContent>
          </Card>

          {/* Operations */}
          {data.spec?.operations && data.spec.operations.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle>Operations</CardTitle>
                <CardDescription>
                  Every operation this capability exposes
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="divide-y">
                  {data.spec.operations.map((op: any) => (
                    <div
                      key={op.name}
                      className="flex items-center justify-between py-3 first:pt-0 last:pb-0"
                    >
                      <div>
                        <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-sm">
                          {op.name}
                        </code>
                        <p className="mt-0.5 text-xs text-text-secondary">
                          {op.description || ''}
                        </p>
                      </div>
                      <div className="flex items-center gap-2 text-xs text-text-secondary">
                        {op.httpMethod && (
                          <Badge variant="outline" className="font-mono">
                            {op.httpMethod}
                          </Badge>
                        )}
                        {op.path && (
                          <code className="font-mono">{op.path}</code>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Tags */}
          {data.spec?.tags && data.spec.tags.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle>Tags</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-wrap gap-2">
                  {data.spec.tags.map((tag: string) => (
                    <Badge key={tag} variant="secondary">
                      {tag}
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
