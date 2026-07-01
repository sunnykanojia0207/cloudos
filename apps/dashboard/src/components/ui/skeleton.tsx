import { cn } from '@/lib/utils';

interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  /**
   * If true, renders a shimmer animation instead of pulse.
   * Use for large blocks that indicate content is loading.
   */
  shimmer?: boolean;
}

function Skeleton({
  className,
  shimmer,
  ...props
}: SkeletonProps) {
  return (
    <div
      className={cn(
        'rounded-sm bg-surface-elevated',
        shimmer
          ? 'animate-shimmer bg-gradient-to-r from-surface-elevated via-border to-surface-elevated bg-[length:200%_100%]'
          : 'animate-pulse',
        className,
      )}
      aria-hidden="true"
      {...props}
    />
  );
}

export { Skeleton };
