import { useState, useCallback, useMemo, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Separator } from '@/components/ui/separator';
import { Badge } from '@/components/ui/badge';
import { useCreateApplication } from '@/hooks/useCreateApplication';
import { useDeploymentEvents } from '@/hooks/useDeploymentEvents';
import type { CreateApplicationInput } from '@/hooks/useCreateApplication';
import {
  Rocket,
  Loader2,
  CheckCircle2,
  XCircle,
  Circle,
  AlertCircle,
  GitBranch,
  Globe,
  Terminal,
  ExternalLink,
  RotateCcw,
  GitCompareArrows,
  RefreshCw,
  ArrowRight,
  BookOpen,
  Layers,
  Activity,
} from 'lucide-react';
import { cn } from '@/lib/utils';

/* ── Types ─────────────────────────────────────────────── */

type WizardPhase = 'input' | 'deploying' | 'complete' | 'failed';

/* ── Runtime detection ─────────────────────────────────── */

interface DetectedRuntime {
  value: string;
  label: string;
  buildpack: string;
  estimatedSeconds: number;
  confidence: 'high' | 'medium' | 'low';
}

const RUNTIME_DATABASE: DetectedRuntime[] = [
  { value: 'go', label: 'Go', buildpack: 'Go Buildpack', estimatedSeconds: 10, confidence: 'high' },
  { value: 'node', label: 'Node.js', buildpack: 'Node.js Buildpack', estimatedSeconds: 14, confidence: 'high' },
  { value: 'python', label: 'Python', buildpack: 'Python Buildpack', estimatedSeconds: 12, confidence: 'high' },
  { value: 'nextjs', label: 'Next.js', buildpack: 'Next.js Buildpack', estimatedSeconds: 18, confidence: 'high' },
  { value: 'static', label: 'Static Site', buildpack: 'Static Buildpack', estimatedSeconds: 5, confidence: 'high' },
  { value: 'laravel', label: 'Laravel', buildpack: 'PHP Buildpack', estimatedSeconds: 15, confidence: 'high' },
  { value: 'docker', label: 'Docker', buildpack: 'Docker Buildpack', estimatedSeconds: 20, confidence: 'high' },
];

// Keyword-based auto-detection from repo URL
function detectRuntimeFromUrl(url: string): DetectedRuntime | null {
  const lower = url.toLowerCase();

  // Framework-specific keywords (higher priority)
  if (/\bnextjs\b|\bnext\.js\b/.test(lower) || /\/next\./.test(lower))
    return RUNTIME_DATABASE.find((r) => r.value === 'nextjs')!;
  if (/\blaravel\b/.test(lower))
    return RUNTIME_DATABASE.find((r) => r.value === 'laravel')!;
  if (/\bdjango\b|flask|fastapi/.test(lower))
    return RUNTIME_DATABASE.find((r) => r.value === 'python')!;
  if (/\bdocker\b|Dockerfile/.test(lower))
    return RUNTIME_DATABASE.find((r) => r.value === 'docker')!;

  // File extension / name-based detection (lower confidence)
  if (/\.go$|\bgo-\b|golang/.test(lower))
    return { ...RUNTIME_DATABASE.find((r) => r.value === 'go')!, confidence: 'medium' };
  if (/\.py$|\bpython\b/.test(lower))
    return { ...RUNTIME_DATABASE.find((r) => r.value === 'python')!, confidence: 'medium' };
  if (/package\.json|node_modules|\.js$|node-|react|vue|angular|svelte/.test(lower))
    return { ...RUNTIME_DATABASE.find((r) => r.value === 'node')!, confidence: 'medium' };
  if (/\.html$|\.css$|jekyll|hugo|gatsby/.test(lower))
    return { ...RUNTIME_DATABASE.find((r) => r.value === 'static')!, confidence: 'medium' };

  return null;
}

function validateRepoUrl(url: string): { valid: boolean; reason?: string } {
  if (!url.trim()) return { valid: false };
  // Must start with http(s):// or git@
  if (!/^https?:\/\//.test(url) && !/^git@/.test(url))
    return { valid: false, reason: 'URL must start with https:// or git@' };
  // Must have a host
  try {
    const u = new URL(url.startsWith('git@') ? `https://${url.slice(4)}` : url);
    if (!u.hostname.includes('.')) return { valid: false, reason: 'Invalid hostname' };
  } catch {
    return { valid: false, reason: 'Invalid URL format' };
  }
  return { valid: true };
}

function generateAppId(url: string): string {
  try {
    const match = url.match(/\/([^/]+?)(?:\.git)?$/);
    if (match) return match[1].toLowerCase().replace(/[^a-z0-9-]/g, '-').replace(/^-+|-+$/g, '');
  } catch { /* ignore */ }
  return '';
}

function generateAppName(url: string): string {
  const id = generateAppId(url);
  return id
    ? id.split('-').map((s) => s.charAt(0).toUpperCase() + s.slice(1)).join(' ')
    : '';
}

/* ── Deployment plan ───────────────────────────────────── */

const DEPLOYMENT_PLAN = [
  { id: 'validate',   name: 'Validate Application',   action: 'validate',        narrative: 'Validate application configuration and dependencies' },
  { id: 'clone',      name: 'Clone Source Repository', action: 'source.clone',    narrative: 'Clone repository and checkout branch' },
  { id: 'install',    name: 'Install Dependencies',    action: 'build.install',   narrative: 'Install project dependencies' },
  { id: 'build',      name: 'Build Artifact',          action: 'build.execute',   narrative: 'Build the application binary or bundle' },
  { id: 'deploy',     name: 'Deploy Application',      action: 'provider.deploy', narrative: 'Deploy to runtime and allocate resources' },
  { id: 'healthcheck', name: 'Health Check',           action: 'health.check',    narrative: 'Verify application is responding correctly' },
  { id: 'complete',   name: 'Complete Deployment',     action: 'complete',        narrative: 'Finalize deployment and update status' },
];

const STEP_NARRATIVES: Record<string, string> = {
  validate:     'Validating application configuration...',
  clone:        'Cloning repository...',
  install:      'Installing dependencies...',
  build:        'Building application...',
  deploy:       'Deploying application...',
  healthcheck:  'Running health check...',
  complete:     'Finalizing deployment...',
};

/* ── Props ─────────────────────────────────────────────── */

export interface DeploymentWizardDialogProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  children?: React.ReactNode;
}

/* ── Component ─────────────────────────────────────────── */

export function DeploymentWizardDialog({ open, onOpenChange, children }: DeploymentWizardDialogProps) {
  const navigate = useNavigate();
  const createApp = useCreateApplication();

  // ── Phase state ──
  const [phase, setPhase] = useState<WizardPhase>('input');

  // ── Form state ──
  const [repoUrl, setRepoUrl] = useState('');
  const [branch, setBranch] = useState('main');
  const [selectedRuntime, setSelectedRuntime] = useState<string | null>(null);
  const [buildCommand, setBuildCommand] = useState('');
  const [appId, setAppId] = useState('');
  const [appName, setAppName] = useState('');
  const [error, setError] = useState('');

  // ── Deploy result state ──
  const [createdAppId, setCreatedAppId] = useState<string | null>(null);
  const [createdAppName, setCreatedAppName] = useState<string | null>(null);
  const [createdUrl, setCreatedUrl] = useState<string | null>(null);
  const [workflowId, setWorkflowId] = useState<string | null>(null);
  const [deployDuration, setDeployDuration] = useState<string | null>(null);
  const [deployRuntime, setDeployRuntime] = useState<string | null>(null);

  // ── Derived / computed ──
  const validation = useMemo(() => validateRepoUrl(repoUrl), [repoUrl]);
  const detectedRuntime = useMemo(() => detectRuntimeFromUrl(repoUrl), [repoUrl]);
  const effectiveRuntime = selectedRuntime || detectedRuntime?.value || 'node';
  const runtimeInfo = RUNTIME_DATABASE.find((r) => r.value === effectiveRuntime);

  const repoName = useMemo(() => generateAppName(repoUrl), [repoUrl]);
  const repoId = useMemo(() => generateAppId(repoUrl), [repoUrl]);

  useEffect(() => {
    if (repoName && !appName) setAppName(repoName);
    if (repoId && !appId) setAppId(repoId);
  }, [repoName, repoId, appName, appId]);

  // ── Live deployment tracking ──
  const liveState = useDeploymentEvents(phase === 'deploying' ? workflowId : null);

  // Sync SSE state with our phase
  useEffect(() => {
    if (liveState.finished && phase === 'deploying') {
      if (liveState.phase === 'completed' || liveState.phase === 'succeeded') {
        setPhase('complete');
        setDeployDuration(liveState.duration);
        if (liveState.url) setCreatedUrl(liveState.url);
      } else if (liveState.phase === 'failed') {
        setPhase('failed');
        setError(liveState.error || 'Deployment failed');
      }
    }
  }, [liveState.finished, liveState.phase, liveState.duration, liveState.url, liveState.error, phase]);

  // ── Step statuses for live deployment ──
  const deploySteps = useMemo(() => {
    const completedNodes = liveState.completedNodes ?? [];
    const failedNodes = liveState.failedNodes ?? [];
    const currentNode = liveState.currentNode;

    return DEPLOYMENT_PLAN.map((step) => {
      let status = 'pending';
      if (failedNodes.includes(step.id)) status = 'failed';
      else if (completedNodes.includes(step.id)) status = 'completed';
      else if (currentNode === step.id) status = 'running';
      return { ...step, status };
    });
  }, [liveState]);

  // ── Reset form ──
  const resetForm = useCallback(() => {
    setPhase('input');
    setRepoUrl('');
    setBranch('main');
    setSelectedRuntime(null);
    setBuildCommand('');
    setAppId('');
    setAppName('');
    setError('');
    setCreatedAppId(null);
    setCreatedAppName(null);
    setCreatedUrl(null);
    setWorkflowId(null);
    setDeployDuration(null);
    setDeployRuntime(null);
  }, []);

  const handleOpenChange = (next: boolean) => {
    if (!next) {
      resetForm();
    }
    onOpenChange?.(next);
  };

  // ── Deploy ──
  const handleDeploy = async () => {
    if (!validation.valid) return;
    setError('');

    const id = appId || repoId || `app-${Date.now()}`;
    const name = appName || repoName || `App ${Date.now()}`;

    const payload: CreateApplicationInput = {
      id,
      name,
      source: {
        url: repoUrl.trim(),
        branch: branch.trim() || 'main',
      },
      runtime: {
        type: effectiveRuntime,
      },
    };

    if (buildCommand.trim()) {
      payload.build = { command: buildCommand.trim() };
    }

    setPhase('deploying');
    setCreatedAppId(id);
    setCreatedAppName(name);

    try {
      const result = await createApp.mutateAsync(payload);
      // Store workflow ID for SSE tracking
      const wfId = (result as any)?.status?.currentDeploymentId;
      if (wfId) setWorkflowId(wfId);
      setDeployRuntime(effectiveRuntime);

      // Set up polling to detect when app transitions to Running/Failed
      const poll = setInterval(async () => {
        try {
          const { cloudos } = await import('@/lib/sdk');
          const updated = await cloudos.getResource('Application', id) as any;
          const phase = updated?.status?.phase;
          if (phase === 'Running') {
            clearInterval(poll);
            setPhase('complete');
            setCreatedUrl(updated.status?.url || null);
            setDeployDuration(updated.status?.lastReport?.duration || null);
          } else if (phase === 'Failed') {
            clearInterval(poll);
            setPhase('failed');
            setError(updated.status?.lastReport?.errors?.[0] || 'Deployment failed');
          }
        } catch {
          // Polling error — stop if component unmounts
        }
      }, 2000);
    } catch (err) {
      setPhase('failed');
      setError((err as Error)?.message || 'Failed to deploy application. Ensure the CloudOS kernel is running.');
    }
  };

  // ── Actions ──
  const handleOpenApp = () => {
    if (createdUrl) window.open(createdUrl, '_blank');
  };

  const handleViewDetails = () => {
    if (createdAppId) {
      handleOpenChange(false);
      navigate(`/applications/${createdAppId}`);
    }
  };

  const handleRetry = () => {
    setPhase('input');
    setError('');
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      {children && <DialogTrigger>{children}</DialogTrigger>}
      <DialogContent className={cn(
        'sm:max-w-lg transition-all duration-300',
        phase === 'complete' && 'sm:max-w-xl',
        phase === 'deploying' && 'sm:max-w-xl',
      )}>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {phase === 'input' && <Rocket className="h-5 w-5 text-primary" />}
            {phase === 'deploying' && <Loader2 className="h-5 w-5 text-accent animate-spin" />}
            {phase === 'complete' && <CheckCircle2 className="h-5 w-5 text-success" />}
            {phase === 'failed' && <XCircle className="h-5 w-5 text-danger" />}
            {phase === 'input' && 'Deploy Application'}
            {phase === 'deploying' && 'Deploying...'}
            {phase === 'complete' && 'Deployment Complete'}
            {phase === 'failed' && 'Deployment Failed'}
          </DialogTitle>
          <DialogDescription>
            {phase === 'input' && 'Paste a Git URL to deploy. CloudOS will detect the stack, build it, and give you a running URL.'}
            {phase === 'deploying' && 'CloudOS is building and deploying your application.'}
            {phase === 'complete' && 'Your application is running and ready to use.'}
            {phase === 'failed' && 'The deployment did not complete successfully.'}
          </DialogDescription>
        </DialogHeader>

        <Separator />

        {/* ════════════════════ INPUT PHASE ════════════════════ */}
        {phase === 'input' && (
          <div className="space-y-5">
            {/* Repository URL */}
            <div className="space-y-2">
              <label htmlFor="wiz-repo-url" className="text-small font-medium text-foreground">
                Repository URL
              </label>
              <div className="relative">
                <Globe className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted pointer-events-none" aria-hidden="true" />
                <Input
                  id="wiz-repo-url"
                  placeholder="https://github.com/user/my-app.git"
                  value={repoUrl}
                  onChange={(e) => {
                    setRepoUrl(e.target.value);
                    setAppName('');
                    setAppId('');
                  }}
                  className={cn(
                    'pl-8 pr-10',
                    repoUrl && validation.valid && 'border-success/50',
                    repoUrl && !validation.valid && 'border-danger/50',
                  )}
                  aria-label="Git repository URL"
                  autoFocus
                />
                {/* Validation indicator */}
                {repoUrl && (
                  <span className="absolute right-2.5 top-1/2 -translate-y-1/2">
                    {validation.valid
                      ? <CheckCircle2 className="h-4 w-4 text-success" />
                      : <AlertCircle className="h-4 w-4 text-danger" />
                    }
                  </span>
                )}
              </div>
              {repoUrl && !validation.valid && validation.reason && (
                <p className="text-caption text-danger">{validation.reason}</p>
              )}
              {repoUrl && validation.valid && (
                <p className="text-caption text-success flex items-center gap-1">
                  <CheckCircle2 className="h-3 w-3" />
                  Repository URL is valid
                </p>
              )}
            </div>

            {/* Auto-detected stack */}
            <AnimatePresence>
              {validation.valid && (
                <motion.div
                  initial={{ height: 0, opacity: 0 }}
                  animate={{ height: 'auto', opacity: 1 }}
                  exit={{ height: 0, opacity: 0 }}
                  transition={{ duration: 0.2, ease: [0, 0, 0.2, 1] }}
                  className="overflow-hidden space-y-4"
                >
                  {/* Detected Stack */}
                  <div className="rounded-md border border-border bg-surface p-3">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Terminal className="h-4 w-4 text-text-muted" />
                        <span className="text-small text-text-secondary">Detected Stack</span>
                      </div>
                      {detectedRuntime ? (
                        <Badge variant="subtle-success" className="gap-1">
                          <CheckCircle2 className="h-3 w-3" />
                          {detectedRuntime.label}
                        </Badge>
                      ) : (
                        <Badge variant="subtle-accent">Auto-detect</Badge>
                      )}
                    </div>

                    {detectedRuntime && (
                      <div className="mt-2 grid grid-cols-2 gap-2 text-small">
                        <div className="text-text-secondary">Buildpack</div>
                        <div className="text-foreground font-medium text-right">{detectedRuntime.buildpack}</div>
                        <div className="text-text-secondary">Est. Deploy Time</div>
                        <div className="text-foreground font-medium text-right tabular-nums">
                          ~{runtimeInfo?.estimatedSeconds ?? detectedRuntime.estimatedSeconds}s
                        </div>
                        <div className="text-text-secondary">Confidence</div>
                        <div className="text-right">
                          <Badge variant={
                            detectedRuntime.confidence === 'high' ? 'subtle-success' :
                            detectedRuntime.confidence === 'medium' ? 'subtle-accent' :
                            'subtle-neutral'
                          } className="text-caption">
                            {detectedRuntime.confidence === 'high' ? 'High' :
                             detectedRuntime.confidence === 'medium' ? 'Medium' : 'Low'}
                          </Badge>
                        </div>
                      </div>
                    )}
                  </div>

                  {/* Deployment Plan Preview */}
                  <div className="rounded-md border border-border bg-surface p-3">
                    <div className="flex items-center gap-2 mb-2">
                      <Layers className="h-4 w-4 text-text-muted" />
                      <span className="text-small text-text-secondary">Deployment Plan</span>
                      <span className="text-caption text-text-muted ml-auto">{DEPLOYMENT_PLAN.length} steps</span>
                    </div>
                    <div className="space-y-1">
                      {DEPLOYMENT_PLAN.map((step) => (
                        <div key={step.id} className="flex items-center gap-2 text-small text-text-secondary">
                          <CheckCircle2 className="h-3 w-3 text-text-muted shrink-0" />
                          <span>{step.name}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>

            {/* Configuration panel (collapsible) */}
            <details className="group">
              <summary className="flex cursor-pointer items-center gap-1 text-small text-text-secondary hover:text-foreground transition-colors list-none">
                <span className="group-open:rotate-90 transition-transform">▶</span>
                Configure
              </summary>
              <div className="pt-3 space-y-3">
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-1.5">
                    <label htmlFor="wiz-branch" className="text-caption text-text-secondary font-medium">Branch</label>
                    <div className="relative">
                      <GitBranch className="absolute left-2 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-text-muted pointer-events-none" />
                      <Input
                        id="wiz-branch"
                        value={branch}
                        onChange={(e) => setBranch(e.target.value)}
                        className="pl-7 h-[32px] text-small"
                        aria-label="Git branch"
                      />
                    </div>
                  </div>
                  <div className="space-y-1.5">
                    <label htmlFor="wiz-appname" className="text-caption text-text-secondary font-medium">App Name</label>
                    <Input
                      id="wiz-appname"
                      placeholder={repoName || 'Auto-generated'}
                      value={appName}
                      onChange={(e) => setAppName(e.target.value)}
                      className="h-[32px] text-small"
                      aria-label="Application name"
                    />
                  </div>
                </div>
                <div className="space-y-1.5">
                  <label htmlFor="wiz-buildcmd" className="text-caption text-text-secondary font-medium">Build Command <span className="text-text-muted">(optional)</span></label>
                  <Input
                    id="wiz-buildcmd"
                    placeholder="e.g. npm run build"
                    value={buildCommand}
                    onChange={(e) => setBuildCommand(e.target.value)}
                    className="h-[32px] text-small"
                    aria-label="Build command"
                  />
                </div>
              </div>
            </details>

            {/* Error */}
            <AnimatePresence>
              {error && (
                <motion.div
                  initial={{ opacity: 0, y: -4 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -4 }}
                  className="flex items-start gap-2 rounded-md border border-danger/30 bg-danger/5 px-3 py-2 text-small text-danger"
                  role="alert"
                >
                  <AlertCircle className="h-4 w-4 mt-0.5 shrink-0" />
                  <span>{error}</span>
                </motion.div>
              )}
            </AnimatePresence>

            {/* Footer */}
            <div className="flex items-center justify-end gap-2 pt-1">
              <button
                type="button"
                onClick={() => handleOpenChange(false)}
                className="inline-flex items-center justify-center whitespace-nowrap rounded-md h-7 px-2 text-small gap-1 bg-transparent border border-border text-foreground hover:border-border-hover hover:bg-surface active:scale-[0.97] transition-all duration-150"
              >
                Cancel
              </button>
              <Button
                variant="primary"
                size="sm"
                className="gap-1.5"
                disabled={!validation.valid || createApp.isPending}
                onClick={handleDeploy}
              >
                {createApp.isPending ? (
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                ) : (
                  <Rocket className="h-3.5 w-3.5" />
                )}
                Deploy
              </Button>
            </div>
          </div>
        )}

        {/* ════════════════════ DEPLOYING PHASE ════════════════════ */}
        {phase === 'deploying' && (
          <div className="space-y-5">
            {/* Progress */}
            <div className="space-y-2">
              <div className="relative h-1.5 w-full overflow-hidden rounded-full bg-surface-elevated">
                <motion.div
                  className="absolute inset-y-0 left-0 bg-accent rounded-full"
                  initial={{ width: 0 }}
                  animate={{ width: `${Math.min(liveState.progress > 0 ? liveState.progress * 100 : 
                    deploySteps.filter(s => s.status === 'completed').length / deploySteps.length * 100, 100)}%` }}
                  transition={{ duration: 0.4, ease: [0, 0, 0.2, 1] }}
                />
              </div>
            </div>

            {/* Step-by-step narrative */}
            <div className="space-y-1">
              {deploySteps.map((step, i) => (
                <motion.div
                  key={step.id}
                  initial={{ opacity: 0, x: -8 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: i * 0.03 }}
                  className={cn(
                    'flex items-start gap-3 rounded-md px-3 py-2.5 transition-colors',
                    step.status === 'running' && 'bg-accent/5',
                    step.status === 'failed' && 'bg-danger/5',
                    step.status === 'completed' && 'bg-success/5',
                  )}
                >
                  {/* Step icon */}
                  <div className="mt-0.5">
                    {step.status === 'completed' && (
                      <motion.div
                        initial={{ scale: 0 }}
                        animate={{ scale: 1 }}
                        transition={{ type: 'spring', stiffness: 500, damping: 30 }}
                      >
                        <CheckCircle2 className="h-4 w-4 text-success" />
                      </motion.div>
                    )}
                    {step.status === 'running' && (
                      <Loader2 className="h-4 w-4 text-accent animate-spin" />
                    )}
                    {step.status === 'failed' && (
                      <XCircle className="h-4 w-4 text-danger" />
                    )}
                    {step.status === 'pending' && (
                      <Circle className="h-4 w-4 text-text-muted" />
                    )}
                  </div>

                  {/* Step content */}
                  <div className="flex-1 min-w-0">
                    <div className={cn(
                      'text-small font-medium',
                      step.status === 'completed' && 'text-foreground',
                      step.status === 'running' && 'text-accent',
                      step.status === 'failed' && 'text-danger',
                      step.status === 'pending' && 'text-text-muted',
                    )}>
                      {step.name}
                    </div>
                    {step.status === 'running' && (
                      <motion.p
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        className="text-caption text-text-secondary mt-0.5"
                      >
                        {STEP_NARRATIVES[step.id] || step.narrative}
                      </motion.p>
                    )}
                    {step.status === 'completed' && (
                      <p className="text-caption text-success mt-0.5">Complete</p>
                    )}
                    {step.status === 'failed' && (
                      <p className="text-caption text-danger mt-0.5">
                        {liveState.error || 'Failed'}
                      </p>
                    )}
                  </div>

                  {/* Duration indicator for completed steps */}
                  {step.status === 'completed' && (
                    <CheckCircle2 className="h-3.5 w-3.5 text-success shrink-0" />
                  )}
                  {step.status === 'running' && (
                    <span className="text-caption text-accent animate-pulse shrink-0">In progress...</span>
                  )}
                </motion.div>
              ))}
            </div>

            {/* Duration so far */}
            {liveState.duration && (
              <p className="text-caption text-text-muted text-center">
                Elapsed: {liveState.duration}
              </p>
            )}
          </div>
        )}

        {/* ════════════════════ COMPLETE PHASE ════════════════════ */}
        {phase === 'complete' && (
          <div className="space-y-5">
            {/* Celebration card */}
            <motion.div
              initial={{ scale: 0.95, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              transition={{ type: 'spring', stiffness: 300, damping: 25 }}
              className="rounded-xl border-2 border-success/30 bg-success/5 p-5 text-center"
            >
              <motion.div
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                transition={{ type: 'spring', stiffness: 500, damping: 20, delay: 0.1 }}
                className="mx-auto mb-3 flex h-14 w-14 items-center justify-center rounded-full bg-success/20"
              >
                <CheckCircle2 className="h-7 w-7 text-success" />
              </motion.div>
              <h3 className="text-h3 text-foreground mb-1">{createdAppName || 'Deployment Complete'}</h3>
              <p className="text-small text-text-secondary mb-4">Your application is running</p>

              {/* Summary grid */}
              <div className="mx-auto max-w-sm rounded-lg border border-border bg-surface p-4 text-left">
                <div className="space-y-2 text-small">
                  <div className="flex justify-between">
                    <span className="text-text-secondary">Application</span>
                    <span className="text-foreground font-medium">{createdAppName || createdAppId}</span>
                  </div>
                  {deployRuntime && (
                    <div className="flex justify-between">
                      <span className="text-text-secondary">Runtime</span>
                      <span className="text-foreground font-medium">
                        {RUNTIME_DATABASE.find((r) => r.value === deployRuntime)?.label || deployRuntime}
                      </span>
                    </div>
                  )}
                  <div className="flex justify-between">
                    <span className="text-text-secondary">Buildpack</span>
                    <span className="text-foreground font-medium">{runtimeInfo?.buildpack || 'Auto-detected'}</span>
                  </div>
                  {deployDuration && (
                    <div className="flex justify-between">
                      <span className="text-text-secondary">Duration</span>
                      <span className="text-foreground font-medium tabular-nums">{deployDuration}</span>
                    </div>
                  )}
                  {createdUrl && (
                    <div className="flex justify-between items-center">
                      <span className="text-text-secondary">URL</span>
                      <a
                        href={createdUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-accent hover:text-accent-hover inline-flex items-center gap-1 truncate max-w-[200px]"
                      >
                        <span className="truncate">{createdUrl.replace(/^https?:\/\//, '')}</span>
                        <ExternalLink className="h-3 w-3 shrink-0" />
                      </a>
                    </div>
                  )}
                </div>
              </div>
            </motion.div>

            {/* Actions */}
            <div className="flex flex-wrap items-center justify-center gap-2">
              <Button variant="primary" size="sm" className="gap-1.5" onClick={handleViewDetails}>
                <ArrowRight className="h-3.5 w-3.5" />
                View Details
              </Button>
              {createdUrl && (
                <Button variant="secondary" size="sm" className="gap-1.5" onClick={handleOpenApp}>
                  <ExternalLink className="h-3.5 w-3.5" />
                  Open Application
                </Button>
              )}
              <Button variant="secondary" size="sm" className="gap-1.5" onClick={handleViewDetails}>
                <Terminal className="h-3.5 w-3.5" />
                View Logs
              </Button>
            </div>
          </div>
        )}

        {/* ════════════════════ FAILED PHASE ════════════════════ */}
        {phase === 'failed' && (
          <div className="space-y-5">
            {/* Error card */}
            <div className="rounded-xl border-2 border-danger/30 bg-danger/5 p-5">
              <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-danger/20 mx-auto">
                <XCircle className="h-6 w-6 text-danger" />
              </div>
              <h3 className="text-h3 text-foreground mb-1 text-center">Deployment Failed</h3>

              {error && (
                <div className="mt-3 rounded-md border border-danger/20 bg-danger/10 px-3 py-2.5">
                  <p className="text-small text-danger">{error}</p>
                </div>
              )}

              {/* Failed steps timeline */}
              <div className="mt-4 space-y-1">
                {deploySteps.map((step) => (
                  <div
                    key={step.id}
                    className={cn(
                      'flex items-center gap-3 rounded-md px-3 py-2',
                      step.status === 'failed' && 'bg-danger/10',
                    )}
                  >
                    {step.status === 'failed' ? <XCircle className="h-4 w-4 text-danger shrink-0" /> :
                     step.status === 'completed' ? <CheckCircle2 className="h-4 w-4 text-success shrink-0" /> :
                     <Circle className="h-4 w-4 text-text-muted shrink-0" />}
                    <span className={cn(
                      'text-small flex-1',
                      step.status === 'failed' ? 'text-danger font-medium' :
                      step.status === 'completed' ? 'text-foreground' : 'text-text-muted',
                    )}>
                      {step.name}
                    </span>
                  </div>
                ))}
              </div>
            </div>

            {/* Actions */}
            <div className="flex flex-wrap items-center justify-center gap-2">
              <Button variant="primary" size="sm" className="gap-1.5" onClick={handleRetry}>
                <RotateCcw className="h-3.5 w-3.5" />
                Retry
              </Button>
              <Button variant="secondary" size="sm" className="gap-1.5" onClick={handleViewDetails}>
                <GitCompareArrows className="h-3.5 w-3.5" />
                Compare Previous
              </Button>
              <Button variant="secondary" size="sm" className="gap-1.5" onClick={handleViewDetails}>
                <Terminal className="h-3.5 w-3.5" />
                View Logs
              </Button>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
