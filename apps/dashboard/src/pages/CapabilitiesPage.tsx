import { Link } from 'react-router-dom';
import { useCapabilities } from '@/hooks/useCloudOS';
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
  Boxes,
  CheckCircle2,
  XCircle,
  ArrowRight,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import type { CapabilitySpec, CapabilityStatus } from '@cloudos/sdk';

function CapabilityCard({
  item,
}: {
  item: { metadata: { id: string; name: string }; spec: CapabilitySpec; status?: CapabilityStatus };
}) {
  const { metadata, spec, status } = item;
  const available = status?.available ?? false;
  const ops = spec.operations ?? [];
  const tags = spec.tags ?? [];

  return (
    <Link to={`/capabilities/${metadata.id}`}>
      <Card className="h-full transition-colors hover:border-primary/50">
        <CardHeader>
          <div className="flex items-start justify-between">
            <div>
              <CardTitle className="text-lg">{spec.displayName}</CardTitle>
              <CardDescription className="mt-1 line-clamp-2">
                {spec.description}
              </CardDescription>
            </div>
            <Badge
              variant={available ? 'success' : 'secondary'}
              className="shrink-0"
            >
              {available ? 'Available' : 'No Provider'}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-3">
          {/* Status row */}
          <div className="flex items-center gap-4 text-sm">
            <div className="flex items-center gap-1.5">
              {status?.status === 'stable' ? (
                <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400" />
              ) : (
                <XCircle className="h-3.5 w-3.5 text-amber-400" />
              )}
              <span className="capitalize text-muted-foreground">
                {status?.status ?? 'unknown'}
              </span>
            </div>
            <span className="text-muted-foreground">
              Providers: {status?.providerCount ?? 0}
            </span>
          </div>

          {/* Tags */}
          {tags.length > 0 && (
            <div className="flex flex-wrap gap-1.5">
              {tags.slice(0, 4).map((tag) => (
                <Badge key={tag} variant="outline" className="text-xs">
                  {tag}
                </Badge>
              ))}
              {tags.length > 4 && (
                <Badge variant="outline" className="text-xs">
                  +{tags.length - 4}
                </Badge>
              )}
            </div>
          )}

          {/* Operations preview */}
          <div>
            <span className="text-xs text-muted-foreground">
              Operations ({ops.length})
            </span>
            <div className="mt-1 flex flex-wrap gap-1">
              {ops.slice(0, 4).map((op) => (
                <code
                  key={op.name}
                  className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs"
                >
                  {op.name}
                </code>
              ))}
              {ops.length > 4 && (
                <span className="text-xs text-muted-foreground">
                  +{ops.length - 4} more
                </span>
              )}
            </div>
          </div>

          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <span>View details</span>
            <ArrowRight className="h-3 w-3" />
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}

export default function CapabilitiesPage() {
  const { data, isLoading, error } = useCapabilities();
  const items = data?.items ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Boxes className="h-6 w-6 text-primary" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Capabilities</h1>
          <p className="mt-1 text-muted-foreground">
            {data
              ? `${data.metadata.total} capability interfaces registered in the kernel`
              : 'CloudOS capability interfaces'}
          </p>
        </div>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Failed to load capabilities</AlertTitle>
          <AlertDescription>
            {(error as Error)?.message || 'An unknown error occurred'}
          </AlertDescription>
        </Alert>
      )}

      <div
        className={cn(
          'grid gap-4 sm:grid-cols-2',
          items.length > 2 && 'lg:grid-cols-3',
        )}
      >
        {isLoading
          ? Array.from({ length: 5 }).map((_, i) => (
              <Card key={i}>
                <CardHeader>
                  <Skeleton className="h-5 w-32" />
                  <Skeleton className="mt-2 h-4 w-full" />
                </CardHeader>
                <CardContent className="space-y-3">
                  <Skeleton className="h-4 w-24" />
                  <div className="flex gap-1.5">
                    {Array.from({ length: 3 }).map((_, j) => (
                      <Skeleton key={j} className="h-5 w-16 rounded-full" />
                    ))}
                  </div>
                  <Skeleton className="h-4 w-20" />
                </CardContent>
              </Card>
            ))
          : items.map((item) => (
              <CapabilityCard
                key={item.metadata.id}
                item={item as unknown as { metadata: { id: string; name: string }; spec: CapabilitySpec; status?: CapabilityStatus }}
              />
            ))}
      </div>

      {!isLoading && items.length === 0 && !error && (
        <div className="flex min-h-[200px] items-center justify-center rounded-lg border border-dashed">
          <p className="text-sm text-muted-foreground">
            No capabilities registered.
          </p>
        </div>
      )}
    </div>
  );
}
