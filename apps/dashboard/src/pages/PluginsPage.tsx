import { Puzzle } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

/**
 * Plugins page — stub until the Plugin Discovery API (S1-004) is implemented.
 */
export default function PluginsPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Puzzle className="h-6 w-6 text-primary" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Plugins</h1>
          <p className="mt-1 text-muted-foreground">
            CloudOS plugin ecosystem
          </p>
        </div>
      </div>

      <div className="flex min-h-[300px] items-center justify-center rounded-lg border border-dashed">
        <div className="text-center">
          <Puzzle className="mx-auto h-12 w-12 text-muted-foreground/50" />
          <h3 className="mt-4 text-lg font-semibold">Plugin Discovery</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            The Plugin Discovery API is coming in Sprint 1.
          </p>
          <p className="mt-1 text-xs text-muted-foreground">
            Once implemented, this page will display installed and available
            plugins.
          </p>
        </div>
      </div>
    </div>
  );
}
