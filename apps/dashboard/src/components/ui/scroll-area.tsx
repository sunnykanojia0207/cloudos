import * as React from 'react';
import { cn } from '@/lib/utils';

interface ScrollAreaProps extends React.HTMLAttributes<HTMLDivElement> {
  orientation?: 'vertical' | 'horizontal' | 'both';
}

const ScrollArea = React.forwardRef<HTMLDivElement, ScrollAreaProps>(
  ({ className, orientation = 'vertical', children, ...props }, ref) => {
    const overflowClass =
      orientation === 'vertical'
        ? 'overflow-y-auto'
        : orientation === 'horizontal'
          ? 'overflow-x-auto'
          : 'overflow-auto';

    return (
      <div
        ref={ref}
        className={cn(
          overflowClass,
          'scrollbar-thin',
          className,
        )}
        {...props}
      >
        {children}
      </div>
    );
  },
);
ScrollArea.displayName = 'ScrollArea';

export { ScrollArea };
