import { Link } from 'react-router-dom';
import { useResourceKinds } from '@/hooks/useCloudOS';
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
import { AlertCircle, Database, ArrowRight } from 'lucide-react';

export default function ResourcesPage() {
  usePageTitle('Resources');
  const { data, isLoading, error } = useResourceKinds();
  const kinds = data?.kinds ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Database className="h-6 w-6 text-primary" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Resources</h1>
          <p className="mt-1 text-text-secondary">
            {data
              ? `${data.total} resource kinds registered in the engine`
              : 'CloudOS Resource Engine'}
          </p>
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
                  <Skeleton className="mt-2 h-4 w-20" />
                </CardHeader>
                <CardContent>
                  <Skeleton className="h-4 w-24" />
                </CardContent>
              </Card>
            ))
          : kinds.map((kind) => (
              <Link key={kind.name} to={`/resources/${kind.name}`}>
                <Card className="h-full transition-colors hover:border-primary/50">
                  <CardHeader>
                    <div className="flex items-start justify-between">
                      <CardTitle className="text-lg">{kind.name}</CardTitle>
                      <Badge variant={kind.namespaced ? 'secondary' : 'outline'}>
                        {kind.namespaced ? 'Namespaced' : 'Global'}
                      </Badge>
                    </div>
                    {kind.versions && kind.versions.length > 0 && (
                      <CardDescription>
                        API: {kind.versions.join(', ')}
                      </CardDescription>
                    )}
                  </CardHeader>
                  <CardContent>
                    <div className="flex items-center gap-1 text-sm text-text-secondary">
                      <span>View resources</span>
                      <ArrowRight className="h-3 w-3" />
                    </div>
                  </CardContent>
                </Card>
              </Link>
            ))}
      </div>

      {!isLoading && kinds.length === 0 && !error && (
        <div className="flex min-h-[200px] items-center justify-center rounded-lg border border-dashed">
          <p className="text-sm text-text-secondary">
            No resource kinds registered.
          </p>
        </div>
      )}
    </div>
  );
}
