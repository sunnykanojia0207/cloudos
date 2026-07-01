import * as React from 'react';
import type { AppResource } from '@/hooks/useApplications';
import { cn } from '@/lib/utils';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
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
} from 'lucide-react';

/* ── Detail Row ───────────────────────────────────────── */

interface DetailRowProps {
  icon: React.ReactNode;
  label: string;
  value: React.ReactNode;
}

function DetailRow({ icon, label, value }: DetailRowProps) {
  return (
    <div className="flex items-start gap-3 py-2">
      <span className="mt-0.5 shrink-0 text-muted-foreground">{icon}</span>
      <div className="min-w-0 flex-1">
        <span className="block text-xs text-muted-foreground">{label}</span>
        <span className="mt-0.5 block text-sm font-medium text-foreground/90 break-all">
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
    <div className="space-y-6">
      {/* ── Metadata ── */}
      <Card className="border-border/50">
        <CardHeader className="pb-3">
          <CardTitle className="text-base font-semibold flex items-center gap-2">
            <Tag className="h-4 w-4 text-muted-foreground" />
            Application Metadata
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-x-6">
            <DetailRow
              icon={<Hash className="h-4 w-4" />}
              label="ID"
              value={metadata.id}
            />
            <DetailRow
              icon={<Tag className="h-4 w-4" />}
              label="Name"
              value={metadata.name}
            />
            <DetailRow
              icon={<BookOpen className="h-4 w-4" />}
              label="API Version"
              value={app.apiVersion}
            />
            <DetailRow
              icon={<Box className="h-4 w-4" />}
              label="Kind"
              value={app.kind}
            />
            {metadata.createdAt && (
              <DetailRow
                icon={<Tag className="h-4 w-4" />}
                label="Created At"
                value={new Date(metadata.createdAt).toLocaleString()}
              />
            )}
          </div>

          {/* Labels */}
          {metadata.labels && Object.keys(metadata.labels).length > 0 && (
            <>
              <Separator className="my-3" />
              <div className="space-y-1">
                <span className="text-xs text-muted-foreground font-medium">
                  Labels
                </span>
                <div className="flex flex-wrap gap-1.5 mt-1">
                  {Object.entries(metadata.labels).map(([key, val]) => (
                    <span
                      key={key}
                      className={cn(
                        'inline-flex items-center rounded-md border border-border/40',
                        'bg-muted/30 px-2 py-0.5 text-[11px] font-mono text-muted-foreground',
                      )}
                    >
                      {key}: {val}
                    </span>
                  ))}
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* ── Source Configuration ── */}
      <Card className="border-border/50">
        <CardHeader className="pb-3">
          <CardTitle className="text-base font-semibold flex items-center gap-2">
            <GitBranch className="h-4 w-4 text-muted-foreground" />
            Source Configuration
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-x-6">
            <DetailRow
              icon={<GitBranch className="h-4 w-4" />}
              label="Source Type"
              value={spec?.source?.type}
            />
            {spec?.source?.url && (
              <DetailRow
                icon={<Globe className="h-4 w-4" />}
                label="Source URL"
                value={spec.source.url}
              />
            )}
            {spec?.source?.branch && (
              <DetailRow
                icon={<GitBranch className="h-4 w-4" />}
                label="Branch"
                value={spec.source.branch}
              />
            )}
            {spec?.source?.path && (
              <DetailRow
                icon={<Terminal className="h-4 w-4" />}
                label="Path"
                value={spec.source.path}
              />
            )}
          </div>
        </CardContent>
      </Card>

      {/* ── Runtime Configuration ── */}
      <Card className="border-border/50">
        <CardHeader className="pb-3">
          <CardTitle className="text-base font-semibold flex items-center gap-2">
            <Terminal className="h-4 w-4 text-muted-foreground" />
            Runtime Configuration
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-x-6">
            <DetailRow
              icon={<Terminal className="h-4 w-4" />}
              label="Runtime Type"
              value={spec?.runtime?.type}
            />
            {spec?.runtime?.command && (
              <DetailRow
                icon={<Terminal className="h-4 w-4" />}
                label="Command"
                value={
                  <code className="rounded bg-muted/50 px-1.5 py-0.5 text-xs font-mono">
                    {spec.runtime.command}
                  </code>
                }
              />
            )}
            {spec?.runtime?.port != null && (
              <DetailRow
                icon={<Hash className="h-4 w-4" />}
                label="Port"
                value={spec.runtime.port}
              />
            )}
          </div>
        </CardContent>
      </Card>

      {/* ── Deployment Settings ── */}
      <Card className="border-border/50">
        <CardHeader className="pb-3">
          <CardTitle className="text-base font-semibold flex items-center gap-2">
            <Layers className="h-4 w-4 text-muted-foreground" />
            Deployment Settings
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-x-6">
            {spec?.deployment?.port != null && (
              <DetailRow
                icon={<Hash className="h-4 w-4" />}
                label="Deployment Port"
                value={spec.deployment.port}
              />
            )}
          </div>
        </CardContent>
      </Card>

      {/* ── Environment Variables ── */}
      <Card className="border-border/50">
        <CardHeader className="pb-3">
          <CardTitle className="text-base font-semibold flex items-center gap-2">
            <Variable className="h-4 w-4 text-muted-foreground" />
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
                    'flex items-center gap-3 rounded-md border border-border/40',
                    'bg-muted/20 px-3 py-2 text-sm font-mono',
                  )}
                >
                  <span className="text-primary font-medium shrink-0">{key}</span>
                  <span className="text-muted-foreground">=</span>
                  <span className="text-foreground/80 break-all">{value}</span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground py-2">
              No environment variables configured.
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
