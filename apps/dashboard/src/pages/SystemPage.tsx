import { useSystem } from '@/hooks/useCloudOS';
import { usePageTitle } from '@/hooks/usePageTitle';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import {
  AlertCircle,
  Server,
  Monitor,
  Cpu,
  Users,
  GitBranch,
  Wrench,
} from 'lucide-react';

const fields: Array<{
  key: string;
  label: string;
  icon: typeof Server;
  render: (val: string | number) => string;
}> = [
  {
    key: 'os',
    label: 'Operating System',
    icon: Monitor,
    render: (v) => String(v),
  },
  {
    key: 'arch',
    label: 'Architecture',
    icon: Cpu,
    render: (v) => String(v),
  },
  {
    key: 'goVersion',
    label: 'Go Version',
    icon: GitBranch,
    render: (v) => String(v),
  },
  {
    key: 'numCpu',
    label: 'CPU Cores',
    icon: Cpu,
    render: (v) => String(v),
  },
  {
    key: 'numGoroutine',
    label: 'Goroutines',
    icon: Users,
    render: (v) => Number(v).toLocaleString(),
  },
  {
    key: 'compiler',
    label: 'Compiler',
    icon: Wrench,
    render: (v) => String(v),
  },
  {
    key: 'hostname',
    label: 'Hostname',
    icon: Server,
    render: (v) => String(v),
  },
];

export default function SystemPage() {
  usePageTitle('System');
  const { data, isLoading, error } = useSystem();

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Server className="h-6 w-6 text-primary" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">System</h1>
          <p className="mt-1 text-text-secondary">
            Go runtime and operating system information
          </p>
        </div>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Failed to load system info</AlertTitle>
          <AlertDescription>
            {(error as Error)?.message || 'An unknown error occurred'}
          </AlertDescription>
        </Alert>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {isLoading
          ? Array.from({ length: 7 }).map((_, i) => (
              <Card key={i}>
                <CardHeader>
                  <Skeleton className="h-4 w-24" />
                </CardHeader>
                <CardContent>
                  <Skeleton className="h-8 w-32" />
                </CardContent>
              </Card>
            ))
          : fields.map(({ key, label, icon: Icon, render }) => {
              const value = data?.[key as keyof typeof data];
              return (
                <Card key={key}>
                  <CardHeader className="flex flex-row items-center justify-between pb-2">
                    <CardTitle className="text-sm font-medium text-text-secondary">
                      {label}
                    </CardTitle>
                    <Icon className="h-4 w-4 text-text-secondary" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-xl font-bold">
                      {value != null && value !== ''
                        ? render(value)
                        : '—'}
                    </div>
                  </CardContent>
                </Card>
              );
            })}
      </div>
    </div>
  );
}
