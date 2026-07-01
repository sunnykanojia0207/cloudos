import * as React from 'react';
import { cn } from '@/lib/utils';

/* ── Context ──────────────────────────────────────────── */
interface AvatarContextValue {
  imageError: boolean;
  setImageError: (error: boolean) => void;
}

const AvatarContext = React.createContext<AvatarContextValue | null>(null);

function useAvatarContext(): AvatarContextValue {
  const ctx = React.useContext(AvatarContext);
  if (!ctx) throw new Error('Avatar sub-components must be used within <Avatar />');
  return ctx;
}

/* ── Root ──────────────────────────────────────────────── */
interface AvatarProps extends React.HTMLAttributes<HTMLDivElement> {
  size?: 'sm' | 'default' | 'lg';
}

const Avatar = React.forwardRef<HTMLDivElement, AvatarProps>(
  ({ className, size = 'default', ...props }, ref) => {
    const [imageError, setImageError] = React.useState(false);

    return (
      <AvatarContext.Provider value={{ imageError, setImageError }}>
        <div
          ref={ref}
          className={cn(
            'relative flex shrink-0 overflow-hidden rounded-full',
            size === 'sm' && 'h-6 w-6',
            size === 'default' && 'h-7 w-7',
            size === 'lg' && 'h-10 w-10',
            className,
          )}
          {...props}
        />
      </AvatarContext.Provider>
    );
  },
);
Avatar.displayName = 'Avatar';

/* ── Image ──────────────────────────────────────────────── */
interface AvatarImageProps extends React.ImgHTMLAttributes<HTMLImageElement> {}

const AvatarImage = React.forwardRef<HTMLImageElement, AvatarImageProps>(
  ({ className, onError, ...props }, ref) => {
    const { imageError, setImageError } = useAvatarContext();

    if (imageError) return null;

    return (
      <img
        ref={ref}
        className={cn(
          'absolute inset-0 aspect-square h-full w-full object-cover',
          className,
        )}
        onError={(e) => {
          setImageError(true);
          onError?.(e);
        }}
        {...props}
      />
    );
  },
);
AvatarImage.displayName = 'AvatarImage';

/* ── Fallback ──────────────────────────────────────────── */
interface AvatarFallbackProps extends React.HTMLAttributes<HTMLDivElement> {
  delayMs?: number;
}

const AvatarFallback = React.forwardRef<HTMLDivElement, AvatarFallbackProps>(
  ({ className, delayMs, children, ...props }, ref) => {
    const { imageError } = useAvatarContext();
    const [showFallback, setShowFallback] = React.useState(!delayMs);

    React.useEffect(() => {
      if (delayMs) {
        const timer = setTimeout(() => setShowFallback(true), delayMs);
        return () => clearTimeout(timer);
      }
    }, [delayMs]);

    if (!showFallback) return null;

    return (
      <div
        ref={ref}
        className={cn(
          'flex h-full w-full items-center justify-center rounded-full bg-accent-subtle text-caption font-medium text-text-secondary',
          className,
        )}
        {...props}
      >
        {children}
      </div>
    );
  },
);
AvatarFallback.displayName = 'AvatarFallback';

export { Avatar, AvatarImage, AvatarFallback };
