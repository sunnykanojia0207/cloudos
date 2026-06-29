import { cn } from '@/lib/utils';

interface LoadingScreenProps {
  message?: string;
  className?: string;
}

export function LoadingScreen({
  message = 'Loading…',
  className,
}: LoadingScreenProps) {
  return (
    <div
      className={cn(
        'flex min-h-[400px] flex-col items-center justify-center gap-4',
        className,
      )}
    >
      <div className="relative">
        <div className="h-10 w-10 rounded-full border-2 border-muted" />
        <div className="absolute inset-0 h-10 w-10 animate-spin rounded-full border-2 border-t-primary" />
      </div>
      <p className="text-sm text-muted-foreground">{message}</p>
    </div>
  );
}
