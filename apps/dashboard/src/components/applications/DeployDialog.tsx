import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogTrigger,
  DialogClose,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import {
  Rocket,
  Loader2,
  CheckCircle,
  AlertCircle,
  GitBranch,
  Globe,
  Terminal,
  Cpu,
  Cog,
  RotateCcw,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';
import { useCreateApplication } from '@/hooks/useCreateApplication';
import type { CreateApplicationInput } from '@/hooks/useCreateApplication';

/* ── Props ─────────────────────────────────────────────── */
export interface DeployDialogProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  children?: React.ReactNode;
}

/* ── Runtime options ──────────────────────────────────── */
const RUNTIME_OPTIONS = [
  { value: 'auto', label: 'Auto Detect (recommended)', description: 'Let CloudOS detect the runtime from your repository' },
  { value: 'node', label: 'Node.js', description: 'npm start, Next.js, or any Node server' },
  { value: 'python', label: 'Python', description: 'flask, django, or any Python app' },
  { value: 'go', label: 'Go', description: 'Compiled Go binary' },
  { value: 'static', label: 'Static Site', description: 'HTML, CSS, JS — no server needed' },
  { value: 'nextjs', label: 'Next.js', description: 'Next.js SSR application' },
  { value: 'laravel', label: 'Laravel', description: 'PHP Laravel application' },
  { value: 'docker', label: 'Docker', description: 'Custom Docker image' },
];

/* ── Component ─────────────────────────────────────────── */
export function DeployDialog({ open, onOpenChange, children }: DeployDialogProps) {
  const navigate = useNavigate();
  const createApp = useCreateApplication();

  // ── Form state ──
  const [repoUrl, setRepoUrl] = useState('');
  const [branch, setBranch] = useState('main');
  const [runtime, setRuntime] = useState('auto');
  const [appName, setAppName] = useState('');
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [buildCommand, setBuildCommand] = useState('');
  const [error, setError] = useState('');

  // ── Derived state ──
  const repoValid = repoUrl.trim().length > 0;
  const formValid = repoValid && !createApp.isPending;

  // ── Auto-generate app name from repo URL ──
  const generateAppName = useCallback((url: string): string => {
    try {
      // Extract repo name from git URL
      const match = url.match(/\/([^/]+?)(?:\.git)?$/);
      if (match) return match[1].toLowerCase().replace(/[^a-z0-9-]/g, '-');
    } catch {
      // fall through
    }
    return '';
  }, []);

  const handleRepoChange = (value: string) => {
    setRepoUrl(value);
    if (!appName || appName === generateAppName(repoUrl)) {
      setAppName(generateAppName(value));
    }
  };

  // ── Reset form ──
  const resetForm = useCallback(() => {
    setRepoUrl('');
    setBranch('main');
    setRuntime('auto');
    setAppName('');
    setShowAdvanced(false);
    setBuildCommand('');
    setError('');
  }, []);

  const handleOpenChange = (next: boolean) => {
    if (!next) {
      resetForm();
    }
    onOpenChange?.(next);
  };

  // ── Submit ──
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!repoValid) return;

    const effectiveRuntime = runtime === 'auto' ? 'node' : runtime;

    const payload: CreateApplicationInput = {
      id: appName || generateAppName(repoUrl) || `app-${Date.now()}`,
      name: appName || generateAppName(repoUrl) || `App ${Date.now()}`,
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

    try {
      const result = await createApp.mutateAsync(payload);
      handleOpenChange(false);
      // Navigate to the new application detail page
      navigate(`/applications/${result.metadata.id}`);
    } catch (err) {
      setError((err as Error)?.message || 'Failed to create application. Ensure the CloudOS kernel is running.');
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      {children && <DialogTrigger>{children}</DialogTrigger>}
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Rocket className="h-5 w-5 text-primary" />
            Deploy Application
          </DialogTitle>
          <DialogDescription>
            Deploy your application from a Git repository. CloudOS will clone, build, and run it automatically.
          </DialogDescription>
        </DialogHeader>

        <Separator />

        <form onSubmit={handleSubmit} className="space-y-5">
          {/* Repository URL */}
          <div className="space-y-2">
            <Label htmlFor="repo-url">Repository URL</Label>
            <div className="relative">
              <Globe className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted pointer-events-none" aria-hidden="true" />
              <Input
                id="repo-url"
                placeholder="https://github.com/user/my-app.git"
                value={repoUrl}
                onChange={(e) => handleRepoChange(e.target.value)}
                className="pl-8"
                aria-label="Git repository URL"
                autoFocus
              />
            </div>
          </div>

          {/* Branch */}
          <div className="space-y-2">
            <Label htmlFor="branch">Branch</Label>
            <div className="relative">
              <GitBranch className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted pointer-events-none" aria-hidden="true" />
              <Input
                id="branch"
                value={branch}
                onChange={(e) => setBranch(e.target.value)}
                className="pl-8"
                aria-label="Git branch"
              />
            </div>
          </div>

          {/* Runtime */}
          <div className="space-y-2">
            <Label htmlFor="runtime">Runtime</Label>
            <div className="relative">
              <Terminal className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted pointer-events-none z-10" aria-hidden="true" />
              <select
                id="runtime"
                value={runtime}
                onChange={(e) => setRuntime(e.target.value)}
                className="flex h-[34px] w-full rounded-md border border-border bg-surface pl-8 pr-8 text-small text-foreground placeholder:text-text-muted focus:outline-none focus:ring-2 focus:ring-accent focus:border-accent disabled:cursor-not-allowed disabled:opacity-50 appearance-none"
                aria-label="Runtime type"
              >
                {RUNTIME_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
              <ChevronDown className="absolute right-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-text-muted pointer-events-none" aria-hidden="true" />
            </div>
            <p className="text-caption text-text-muted">
              {RUNTIME_OPTIONS.find((o) => o.value === runtime)?.description ?? ''}
            </p>
          </div>

          {/* Advanced options */}
          <div>
            <button
              type="button"
              onClick={() => setShowAdvanced(!showAdvanced)}
              className="inline-flex items-center gap-1 text-small text-text-secondary hover:text-foreground transition-colors"
            >
              {showAdvanced ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
              Advanced Options
            </button>

            <AnimatePresence>
              {showAdvanced && (
                <motion.div
                  initial={{ height: 0, opacity: 0 }}
                  animate={{ height: 'auto', opacity: 1 }}
                  exit={{ height: 0, opacity: 0 }}
                  transition={{ duration: 0.15, ease: [0, 0, 0.2, 1] }}
                  className="overflow-hidden"
                >
                  <div className="pt-3 space-y-3">
                    <div className="space-y-2">
                      <Label htmlFor="app-name">Application Name</Label>
                      <Input
                        id="app-name"
                        placeholder="Auto-generated from repo"
                        value={appName}
                        onChange={(e) => setAppName(e.target.value)}
                        aria-label="Application name"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="build-cmd">Build Command</Label>
                      <Input
                        id="build-cmd"
                        placeholder="e.g. npm run build (optional)"
                        value={buildCommand}
                        onChange={(e) => setBuildCommand(e.target.value)}
                        aria-label="Build command"
                      />
                    </div>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>

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
            <DialogClose type="button" className="inline-flex items-center justify-center whitespace-nowrap rounded-md text-btn font-medium transition-all duration-150 h-7 px-2 text-small gap-1 bg-transparent border border-border text-foreground hover:border-border-hover hover:bg-surface active:scale-[0.97]">
              Cancel
            </DialogClose>
            <Button type="submit" variant="primary" size="sm" disabled={!formValid} className="gap-1.5">
              {createApp.isPending ? (
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
              ) : (
                <Rocket className="h-3.5 w-3.5" />
              )}
              {createApp.isPending ? 'Deploying...' : 'Deploy'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
