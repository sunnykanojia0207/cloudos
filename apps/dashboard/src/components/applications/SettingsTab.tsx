import * as React from 'react';
import type { AppResource } from '@/hooks/useApplications';
import { cn } from '@/lib/utils';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  BookOpen,
  Box,
  Tag,
  GitBranch,
  Globe,
  Terminal,
  Settings,
  Layers,
  Variable,
  Hash,
  AlertTriangle,
  Trash2,
  ExternalLink,
  Cpu,
} from 'lucide-react';

/* ── Detail Row ───────────────────────────────────────── */
interface DetailRowProps {
  icon: React.ReactNode;
  label: string;
  value: React.ReactNode;
}

function DetailRow({ icon, label, value }: DetailRowProps) {
  return (
    <div className="flex items-start gap-3 py-1.5">
      <span className="mt-0.5 shrink-0 text-text-muted">{icon}</span>
      <div className="min-w-0 flex-1">
        <span className="block text-caption text-text-secondary">{label}</span>
        <span className="mt-0.5 block text-small font-medium text-foreground break-all">
          {value ?? '\u2014'}
        </span>
      </div>
    </div>
  );
}

/* ── Props ────────────────────────────────────────────── */
export interface SettingsTabProps {
  app: AppResource;
}

/* ── Component ────────────────────────────────────────── */
export function SettingsTab({ app }: SettingsTabProps) {
  const metadata = app.metadata;
  const spec = app.spec;
  const envVars = spec?.settings;
  const envVarEntries = envVars ? Object.entries(envVars) : [];

  return (
    <div className="space-y-5">
      {/* ── Metadata ── */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-body font-semibold flex items-center gap-2">
            <Tag className="h-4 w-4 text-text-muted" />
            Application Metadata
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-x-6">
            <DetailRow icon={<Hash className="h-4 w-4" />} label="ID" value={metadata.id} />
            <DetailRow icon={<Tag className="h-4 w-4" />} label="Name" value={metadata.name} />
            <DetailRow icon={<BookOpen className="h-4 w-4" />} label="API Version" value={app.apiVersion} />
            <DetailRow icon={<Box className="h-4 w-4" />} label="Kind" value={app.kind} />
            {metadata.createdAt && (
              <DetailRow
                icon={<Hash className="h-4 w-4" />}
                label="Created At"
                value={new Date(metadata.createdAt).toLocaleString()}
              />
            )}
          </div>

          {/* Labels */}
          {metadata.labels && Object.keys(metadata.labels).length > 0 && (
            <>
              <Separator className="my-3" />
              <span className="text-caption text-text-secondary font-medium">Labels</span>
              <div className="flex flex-wrap gap-1.5 mt-1.5">
                {Object.entries(metadata.labels).map(([key, val]) => (
                  <span
                    key={key}
                    className={cn(
                      'inline-flex items-center rounded-md border border-border',
                      'bg-surface-elevated px-2 py-0.5 text-caption font-mono text-text-secondary',
                    )}
                  >
                    {key}: {val}
                  </span>
                ))}
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* ── Source Configuration ── */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-body font-semibold flex items-center gap-2">
            <GitBranch className="h-4 w-4 text-text-muted" />
            Source Configuration
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-x-6">
            <DetailRow icon={<GitBranch className="h-4 w-4" />} label="Source Type" value={spec?.source?.type} />
            {spec?.source?.url && (
              <DetailRow
                icon={<Globe className="h-4 w-4" />}
                label="Source URL"
                value={
                  <a href={spec.source.url} target="_blank" rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 text-accent hover:text-accent-hover">
                    {spec.source.url}
                    <ExternalLink className="h-3 w-3" />
                  </a>
                }
              />
            )}
            {spec?.source?.branch && <DetailRow icon={<GitBranch className="h-4 w-4" />} label="Branch" value={spec.source.branch} />}
            {spec?.source?.path && <DetailRow icon={<Terminal className="h-4 w-4" />} label="Path" value={spec.source.path} />}
          </div>
        </CardContent>
      </Card>

      {/* ── Runtime Configuration ── */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-body font-semibold flex items-center gap-2">
            <Terminal className="h-4 w-4 text-text-muted" />
            Runtime Configuration
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-x-6">
            <DetailRow icon={<Terminal className="h-4 w-4" />} label="Runtime Type" value={spec?.runtime?.type} />
            {spec?.runtime?.command && (
              <DetailRow
                icon={<Terminal className="h-4 w-4" />}
                label="Command"
                value={<code className="rounded bg-surface-elevated px-1.5 py-0.5 text-code font-mono">{spec.runtime.command}</code>}
              />
            )}
            {spec?.runtime?.port != null && (
              <DetailRow icon={<Hash className="h-4 w-4" />} label="Port" value={spec.runtime.port} />
            )}
          </div>
        </CardContent>
      </Card>

      {/* ── Deployment Settings ── */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-body font-semibold flex items-center gap-2">
            <Layers className="h-4 w-4 text-text-muted" />
            Deployment Settings
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-x-6">
            {spec?.deployment?.port != null && (
              <DetailRow icon={<Hash className="h-4 w-4" />} label="Deployment Port" value={spec.deployment.port} />
            )}
          </div>
        </CardContent>
      </Card>

      {/* ── Environment Variables ── */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-body font-semibold flex items-center gap-2">
            <Variable className="h-4 w-4 text-text-muted" />
            Environment Variables
          </CardTitle>
        </CardHeader>
        <CardContent>
          {envVarEntries.length > 0 ? (
            <div className="space-y-1">
              {envVarEntries.map(([key, value]) => (
                <div
                  key={key}
                  className={cn(
                    'flex items-center gap-3 rounded-md border border-border',
                    'bg-surface px-3 py-2 text-small font-mono',
                  )}
                >
                  <span className="text-accent font-medium shrink-0">{key}</span>
                  <span className="text-text-muted">=</span>
                  <span className="text-foreground break-all">{value}</span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-small text-text-muted py-2">No environment variables configured.</p>
          )}
        </CardContent>
      </Card>

      {/* ── Danger Zone ── */}
      <Card className="border-danger/30">
        <CardHeader className="pb-2">
          <CardTitle className="text-body font-semibold flex items-center gap-2 text-danger">
            <AlertTriangle className="h-4 w-4" />
            Danger Zone
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-small text-text-secondary mb-3">
            Irreversible actions that will affect this application. Proceed with caution.
          </p>
          <div className="flex items-center justify-between rounded-md border border-danger/30 bg-danger-subtle px-4 py-3">
            <div className="min-w-0">
              <span className="block text-small font-medium text-foreground">Delete this application</span>
              <span className="block text-caption text-text-secondary">Permanently remove the application and all its deployments.</span>
            </div>
            <Button variant="danger" size="sm" className="shrink-0 gap-1.5 ml-3">
              <Trash2 className="h-3.5 w-3.5" />
              Delete
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
