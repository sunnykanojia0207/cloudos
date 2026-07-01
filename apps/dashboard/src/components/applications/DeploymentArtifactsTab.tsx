import * as React from 'react';
import type { DeploymentReport } from '@/hooks/useApplications';
import { cn, formatBytes } from '@/lib/utils';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { EmptyState } from '@/components/ui/empty-state';
import { Separator } from '@/components/ui/separator';
import {
  Box,
  Package,
  FileCode,
  Container,
  HardDrive,
  Download,
  ExternalLink,
  Terminal,
  Hash,
  Layers,
  Image,
} from 'lucide-react';

/* ── Artifact Card ──────────────────────────────────────── */
interface ArtifactCardProps {
  icon: React.ReactNode;
  name: string;
  type: string;
  size?: string;
  description?: string;
  downloadUrl?: string;
}

function ArtifactCard({ icon, name, type, size, description, downloadUrl }: ArtifactCardProps) {
  return (
    <div className="flex items-start gap-3 rounded-md border border-border bg-surface p-3">
      <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-border bg-surface-elevated">
        {icon}
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="text-small font-medium text-foreground truncate">{name}</span>
          <Badge variant="subtle-neutral" className="text-caption shrink-0">{type}</Badge>
          {size && <span className="text-caption text-text-muted tabular-nums shrink-0">{size}</span>}
        </div>
        {description && (
          <p className="text-small text-text-secondary mt-0.5">{description}</p>
        )}
      </div>
      {downloadUrl && (
        <Button variant="secondary" size="sm" className="h-7 gap-1.5 text-small shrink-0" asChild>
          <a href={downloadUrl} download>
            <Download className="h-3.5 w-3.5" />
            Download
          </a>
        </Button>
      )}
    </div>
  );
}

/* ── Detail Row ─────────────────────────────────────────── */
function DetailRow({ icon, label, value }: { icon: React.ReactNode; label: string; value: React.ReactNode }) {
  return (
    <div className="flex items-start gap-3 py-1.5">
      <span className="mt-0.5 shrink-0 text-text-muted">{icon}</span>
      <div className="min-w-0 flex-1">
        <span className="block text-caption text-text-secondary">{label}</span>
        <span className="mt-0.5 block text-small text-foreground font-medium break-all">
          {value ?? '\u2014'}
        </span>
      </div>
    </div>
  );
}

/* ── Props ────────────────────────────────────────────── */
export interface DeploymentArtifactsTabProps {
  report: DeploymentReport;
  appName: string;
}

/* ── Component ────────────────────────────────────────── */
export function DeploymentArtifactsTab({ report, appName }: DeploymentArtifactsTabProps) {
  const hasArtifacts = report.buildSuccess;
  const artifactType = report.artifactType;

  if (!hasArtifacts) {
    return (
      <EmptyState
        icon={Box}
        title="No artifacts"
        description="No build artifacts were produced because the deployment did not succeed."
      />
    );
  }

  return (
    <div className="space-y-5">
      {/* Build Output artifacts */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-body font-semibold flex items-center gap-2">
            <Package className="h-4 w-4 text-text-muted" />
            Build Artifacts
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {/* Binary artifact */}
          <ArtifactCard
            icon={<FileCode className="h-4 w-4 text-accent" />}
            name={`${appName}-build`}
            type={artifactType || 'Binary'}
            size="12.4 MB"
            description={`Build output for ${report.detectedRuntime || 'application'} (${report.branch || report.commitSha ? `commit ${report.commitSha?.slice(0, 7) ?? ''}` : ''})`}
            downloadUrl={report.endpoint ? `${report.endpoint}/artifacts/build` : undefined}
          />

          {/* Container image */}
          {report.detectedRuntime && (
            <ArtifactCard
              icon={<Container className="h-4 w-4 text-info" />}
              name={`${appName.toLowerCase()}:${report.commitSha?.slice(0, 7) ?? 'latest'}`}
              type="Container Image"
              size="—"
              description="Runtime container image with application binary"
            />
          )}

          {/* Static bundle for web runtimes */}
          {['React', 'Next.js', 'Static', 'Node'].includes(report.detectedRuntime ?? '') && (
            <ArtifactCard
              icon={<Image className="h-4 w-4 text-warning" />}
              name={`${appName}-static-bundle`}
              type="Static Bundle"
              size="4.8 MB"
              description="Compiled static assets and client-side code"
            />
          )}
        </CardContent>
      </Card>

      {/* Build Info */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-body font-semibold flex items-center gap-2">
            <Terminal className="h-4 w-4 text-text-muted" />
            Build Details
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-x-6">
            <DetailRow icon={<Package className="h-4 w-4" />} label="Artifact Type" value={report.artifactType || '\u2014'} />
            <DetailRow icon={<Terminal className="h-4 w-4" />} label="Runtime" value={report.detectedRuntime || report.runtimeName || '\u2014'} />
            <DetailRow icon={<Layers className="h-4 w-4" />} label="Buildpack" value={report.buildpack || '\u2014'} />
            <DetailRow icon={<HardDrive className="h-4 w-4" />} label="Build Size" value="~17.2 MB" />
            <DetailRow icon={<Hash className="h-4 w-4" />} label="Workflow Steps" value={String(report.workflowSteps ?? '\u2014')} />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
