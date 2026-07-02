import { useProviders } from '@/hooks/useCloudOS';
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
  ShieldCheck,
  CheckCircle2,
  XCircle,
  Globe,
  Tag,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import type { ProviderSpec, ProviderStatus } from '@cloudos/sdk';

export default function ProvidersPage() {
  usePageTitle('Providers');
  const { data, isLoading, error } = useProviders();
  const items = data?.items ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <ShieldCheck className="h-6 w-6 text-primary" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Providers</h1>
          <p className="mt-1 text-text-secondary">
            {data
              ? `${data.metadata.total} provider implementations registered`
              : 'CloudOS provider implementations'}
          </p>
        </div>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Failed to load providers</AlertTitle>
          <AlertDescription>
            {(error as Error)?.message || 'An unknown error occurred'}
          </AlertDescription>
        </Alert>
      )}

      <div className="grid gap-4 sm:grid-cols-2">
        {isLoading
          ? Array.from({ length: 3 }).map((_, i) => (
              <Card key={i}>
                <CardHeader>
                  <Skeleton className="h-5 w-40" />
                  <Skeleton className="mt-2 h-4 w-full" />
                </CardHeader>
                <CardContent className="space-y-3">
                  <Skeleton className="h-4 w-20" />
                  <Skeleton className="h-4 w-32" />
                </CardContent>
              </Card>
            ))
          : items.map((item) => {
              const spec = item.spec as unknown as ProviderSpec;
              const status = item.status as unknown as ProviderStatus;
              const capabilities = spec?.capabilities ?? [];

              return (
                <Card key={item.metadata.id}>
                  <CardHeader>
                    <div className="flex items-start justify-between">
                      <div>
                        <CardTitle className="text-lg">
                          {spec?.displayName || item.metadata.name}
                        </CardTitle>
                        <CardDescription className="mt-1 line-clamp-2">
                          {spec?.description}
                        </CardDescription>
                      </div>
                      <Badge
                        variant={status?.healthy ? 'success' : 'destructive'}
                        className="shrink-0"
                      >
                        {status?.healthy ? 'Healthy' : 'Unhealthy'}
                      </Badge>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    <div className="flex items-center gap-4 text-sm text-text-secondary">
                      <div className="flex items-center gap-1.5">
                        {status?.ready ? (
                          <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400" />
                        ) : (
                          <XCircle className="h-3.5 w-3.5 text-destructive" />
                        )}
                        <span>{status?.ready ? 'Ready' : 'Not ready'}</span>
                      </div>
                      <div className="flex items-center gap-1.5">
                        <Tag className="h-3.5 w-3.5" />
                        <span>{spec?.providerType}</span>
                      </div>
                      <span className="font-mono text-xs">
                        v{spec?.version}
                      </span>
                    </div>

                    {/* Capabilities */}
                    {capabilities.length > 0 && (
                      <div>
                        <span className="text-xs text-text-secondary">
                          Capabilities ({capabilities.length})
                        </span>
                        <div className="mt-1 flex flex-wrap gap-1.5">
                          {capabilities.map((cap) => (
                            <Badge key={cap.id} variant="secondary" className="text-xs">
                              {cap.id}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Platforms */}
                    {spec?.supportedPlatforms &&
                      spec.supportedPlatforms.length > 0 && (
                        <div className="flex flex-wrap gap-2 text-xs text-text-secondary">
                          {spec.supportedPlatforms.map((p) => (
                            <span
                              key={`${p.os}/${p.arch}`}
                              className="inline-flex items-center gap-1"
                            >
                              <Globe className="h-3 w-3" />
                              {p.os}/{p.arch}
                            </span>
                          ))}
                        </div>
                      )}
                  </CardContent>
                </Card>
              );
            })}
      </div>

      {!isLoading && items.length === 0 && !error && (
        <div className="flex min-h-[200px] items-center justify-center rounded-lg border border-dashed">
          <p className="text-sm text-text-secondary">
            No providers registered.
          </p>
        </div>
      )}
    </div>
  );
}
