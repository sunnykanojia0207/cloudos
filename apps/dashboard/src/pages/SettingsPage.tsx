import { Settings2 } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { usePageTitle } from '@/hooks/usePageTitle';

/**
 * Settings page — stub until the Configuration API is implemented.
 */
export default function SettingsPage() {
  usePageTitle('Settings');
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Settings2 className="h-6 w-6 text-primary" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
          <p className="mt-1 text-muted-foreground">
            CloudOS configuration
          </p>
        </div>
      </div>

      <div className="flex min-h-[300px] items-center justify-center rounded-lg border border-dashed">
        <div className="text-center">
          <Settings2 className="mx-auto h-12 w-12 text-muted-foreground/50" />
          <h3 className="mt-4 text-lg font-semibold">Configuration</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            The Configuration API is coming in Sprint 1.
          </p>
          <p className="mt-1 text-xs text-muted-foreground">
            Once implemented, this page will let you manage kernel and subsystem
            configuration.
          </p>
        </div>
      </div>
    </div>
  );
}
