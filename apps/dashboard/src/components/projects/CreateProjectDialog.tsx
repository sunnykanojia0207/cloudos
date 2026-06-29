import * as React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Loader2, CheckCircle, AlertCircle } from 'lucide-react';
import { useCreateProject } from '@/hooks/useCloudOS';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogFooter,
  DialogTitle,
  DialogDescription,
  DialogClose,
} from '@/components/ui/dialog';
import { Separator } from '@/components/ui/separator';

/* ── Constants ────────────────────────────────────────── */

const ENVIRONMENTS = [
  { value: 'development', label: 'Development' },
  { value: 'staging', label: 'Staging' },
  { value: 'production', label: 'Production' },
  { value: 'testing', label: 'Testing' },
] as const;

const ID_REGEX = /^[a-z0-9-]+$/;

/* ── Helpers ──────────────────────────────────────────── */

function deriveId(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9-]/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');
}

/* ── Props ────────────────────────────────────────────── */

export interface CreateProjectDialogProps {
  /** Optional controlled open state */
  open?: boolean;
  /** Callback when open state changes */
  onOpenChange?: (open: boolean) => void;
  children?: React.ReactNode;
}

/* ── Component ────────────────────────────────────────── */

export function CreateProjectDialog({
  open: controlledOpen,
  onOpenChange: controlledOnOpenChange,
  children,
}: CreateProjectDialogProps) {
  const [internalOpen, setInternalOpen] = React.useState(false);

  const isControlled = controlledOpen !== undefined;
  const open = isControlled ? controlledOpen : internalOpen;
  const onOpenChange = isControlled
    ? (controlledOnOpenChange ?? (() => {}))
    : setInternalOpen;

  // ── Form state ──
  const [displayName, setDisplayName] = React.useState('');
  const [id, setId] = React.useState('');
  const [description, setDescription] = React.useState('');
  const [environment, setEnvironment] = React.useState('development');
  const [idManuallyEdited, setIdManuallyEdited] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);

  const createProject = useCreateProject();

  // ── Derived validation ──
  const nameValid = displayName.trim().length > 0;
  const idValid = ID_REGEX.test(id) && id.length > 0;
  const formValid = nameValid && idValid && !createProject.isPending;

  // ── Auto-derive ID from name (unless manually edited) ──
  const handleNameChange = (value: string) => {
    setDisplayName(value);
    if (!idManuallyEdited) {
      setId(deriveId(value));
    }
    setError(null);
  };

  const handleIdChange = (value: string) => {
    setIdManuallyEdited(true);
    setId(value);
    setError(null);
  };

  // ── Reset form when dialog opens ──
  const handleOpenChange = React.useCallback(
    (next: boolean) => {
      onOpenChange(next);
      if (next) {
        setDisplayName('');
        setId('');
        setDescription('');
        setEnvironment('development');
        setIdManuallyEdited(false);
        setError(null);
      }
    },
    [onOpenChange],
  );

  // ── Submit ──
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formValid) return;

    setError(null);

    try {
      await createProject.mutateAsync({
        id,
        displayName: displayName.trim(),
        description: description.trim() || undefined,
        environment: environment as string,
      });
      handleOpenChange(false);
    } catch (err: unknown) {
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to create project. Please try again.';
      setError(message);
    }
  };

  // ── Render ──
  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      {children && <DialogTrigger>{children}</DialogTrigger>}

      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Create Project</DialogTitle>
          <DialogDescription>
            Set up a new project to organize your cloud resources.
          </DialogDescription>
        </DialogHeader>

        <Separator />

        <form onSubmit={handleSubmit} className="space-y-5">
          {/* Project Name */}
          <div className="space-y-2">
            <Label htmlFor="project-name">
              Project Name <span className="text-destructive">*</span>
            </Label>
            <Input
              id="project-name"
              placeholder="My Project"
              value={displayName}
              onChange={(e) => handleNameChange(e.target.value)}
              disabled={createProject.isPending}
              required
              autoFocus
              className="h-9"
            />
          </div>

          {/* Project ID */}
          <div className="space-y-2">
            <Label htmlFor="project-id">
              Project ID <span className="text-destructive">*</span>
            </Label>
            <Input
              id="project-id"
              placeholder="my-project"
              value={id}
              onChange={(e) => handleIdChange(e.target.value)}
              disabled={createProject.isPending}
              required
              className={cn(
                'h-9',
                id && !idValid && 'border-destructive/50 focus-visible:ring-destructive',
              )}
              aria-invalid={id.length > 0 && !idValid}
              aria-describedby="project-id-helper"
            />
            <p
              id="project-id-helper"
              className={cn(
                'text-xs',
                id && !idValid ? 'text-destructive' : 'text-muted-foreground',
              )}
            >
              Must be unique, lowercase letters and hyphens
              {id && !idValid && ' — invalid characters detected'}
            </p>
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="project-description">Description</Label>
            <textarea
              id="project-description"
              placeholder="Optional description for your project"
              value={description}
              onChange={(e) => {
                setDescription(e.target.value);
                setError(null);
              }}
              disabled={createProject.isPending}
              rows={3}
              className={cn(
                'flex w-full rounded-md border border-input bg-background px-3 py-2',
                'text-sm ring-offset-background',
                'placeholder:text-muted-foreground',
                'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
                'disabled:cursor-not-allowed disabled:opacity-50',
                'resize-none min-h-[72px]',
              )}
            />
          </div>

          {/* Environment */}
          <fieldset className="space-y-3">
            <legend className="text-sm font-medium leading-none">Environment</legend>
            <div className="grid grid-cols-2 gap-2">
              {ENVIRONMENTS.map((env) => (
                <EnvironmentOption
                  key={env.value}
                  value={env.value}
                  label={env.label}
                  selected={environment === env.value}
                  onSelect={() => {
                    setEnvironment(env.value);
                    setError(null);
                  }}
                  disabled={createProject.isPending}
                />
              ))}
            </div>
          </fieldset>

          {/* Error alert */}
          <AnimatePresence mode="wait">
            {error && (
              <motion.div
                initial={{ opacity: 0, y: -4 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -4 }}
                transition={{ duration: 0.15 }}
                role="alert"
                className="flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2.5 text-sm text-destructive"
              >
                <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
                <span>{error}</span>
              </motion.div>
            )}
          </AnimatePresence>

          <DialogFooter className="gap-2 sm:gap-0">
            <DialogClose>
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={createProject.isPending}
              >
                Cancel
              </Button>
            </DialogClose>
            <Button
              type="submit"
              size="sm"
              disabled={!formValid}
              className="min-w-[100px]"
            >
              {createProject.isPending ? (
                <span className="inline-flex items-center gap-2">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Creating…
                </span>
              ) : (
                <span className="inline-flex items-center gap-2">
                  <CheckCircle className="h-4 w-4" />
                  Create
                </span>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

/* ── Sub-components ──────────────────────────────────── */

interface EnvironmentOptionProps {
  value: string;
  label: string;
  selected: boolean;
  onSelect: () => void;
  disabled?: boolean;
}

function EnvironmentOption({
  value,
  label,
  selected,
  onSelect,
  disabled,
}: EnvironmentOptionProps) {
  const id = `env-${value}`;

  return (
    <div
      role="radio"
      aria-checked={selected}
      tabIndex={disabled ? -1 : 0}
      onClick={disabled ? undefined : onSelect}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          if (!disabled) onSelect();
        }
      }}
      className={cn(
        'flex cursor-pointer items-center gap-2 rounded-md border px-3 py-2.5 text-sm transition-colors',
        selected
          ? 'border-primary/50 bg-primary/10 text-primary'
          : 'border-input bg-background hover:bg-accent/50',
        disabled && 'pointer-events-none opacity-50',
      )}
    >
      <input
        type="radio"
        id={id}
        name="environment"
        value={value}
        checked={selected}
        onChange={onSelect}
        disabled={disabled}
        className="sr-only"
      />
      <span
        className={cn(
          'flex h-3.5 w-3.5 shrink-0 items-center justify-center rounded-full border transition-colors',
          selected ? 'border-primary bg-primary' : 'border-muted-foreground/40',
        )}
      >
        {selected && (
          <motion.span
            layoutId="env-dot"
            className="h-1.5 w-1.5 rounded-full bg-primary-foreground"
            transition={{ duration: 0.15 }}
          />
        )}
      </span>
      <Label htmlFor={id} className="cursor-pointer text-xs font-normal leading-none">
        {label}
      </Label>
    </div>
  );
}
