import * as React from 'react';
import { cn } from '@/lib/utils';

/* ── Context ──────────────────────────────────────────── */
interface TooltipContextValue {
  open: boolean;
  setOpen: (open: boolean) => void;
  delayDuration: number;
  triggerRef: React.RefObject<HTMLButtonElement | null>;
}

const TooltipContext = React.createContext<TooltipContextValue | null>(null);

function useTooltipContext(): TooltipContextValue {
  const ctx = React.useContext(TooltipContext);
  if (!ctx) throw new Error('Tooltip sub-components must be used within <Tooltip />');
  return ctx;
}

/* ── Provider (global default) ────────────────────────── */
interface TooltipProviderProps {
  children: React.ReactNode;
  delayDuration?: number;
}

function TooltipProvider({
  children,
  delayDuration = 300,
}: TooltipProviderProps) {
  return (
    <React.Fragment>
      {children}
    </React.Fragment>
  );
}
TooltipProvider.displayName = 'TooltipProvider';

/* ── Root ──────────────────────────────────────────────── */
interface TooltipProps {
  children: React.ReactNode;
  delayDuration?: number;
  open?: boolean;
  defaultOpen?: boolean;
  onOpenChange?: (open: boolean) => void;
}

function Tooltip({
  children,
  delayDuration = 300,
  open: controlledOpen,
  defaultOpen,
  onOpenChange: controlledOnOpenChange,
}: TooltipProps) {
  const isControlled = controlledOpen !== undefined;
  const [uncontrolledOpen, setUncontrolledOpen] = React.useState(
    defaultOpen ?? false,
  );
  const triggerRef = React.useRef<HTMLButtonElement | null>(null);

  const open = isControlled ? controlledOpen : uncontrolledOpen;

  const setOpen = React.useCallback(
    (next: boolean) => {
      if (!isControlled) setUncontrolledOpen(next);
      controlledOnOpenChange?.(next);
    },
    [isControlled, controlledOnOpenChange],
  );

  return (
    <TooltipContext.Provider value={{ open, setOpen, delayDuration, triggerRef }}>
      <div className="relative inline-flex">
        {children}
      </div>
    </TooltipContext.Provider>
  );
}
Tooltip.displayName = 'Tooltip';

/* ── Trigger ──────────────────────────────────────────── */
const TooltipTrigger = React.forwardRef<
  HTMLButtonElement,
  React.ButtonHTMLAttributes<HTMLButtonElement>
>(({ className, onMouseEnter, onMouseLeave, onFocus, onBlur, children, ...props }, ref) => {
  const { setOpen, delayDuration, triggerRef } = useTooltipContext();
  const showTimeout = React.useRef<number | null>(null);
  const hideTimeout = React.useRef<number | null>(null);

  const handleMouseEnter = (e: React.MouseEvent<HTMLButtonElement>) => {
    if (hideTimeout.current !== null) window.clearTimeout(hideTimeout.current);
    showTimeout.current = window.setTimeout(() => setOpen(true), delayDuration);
    onMouseEnter?.(e);
  };

  const handleMouseLeave = (e: React.MouseEvent<HTMLButtonElement>) => {
    if (showTimeout.current !== null) window.clearTimeout(showTimeout.current);
    hideTimeout.current = window.setTimeout(() => setOpen(false), 0);
    onMouseLeave?.(e);
  };

  const handleFocus = (e: React.FocusEvent<HTMLButtonElement>) => {
    if (hideTimeout.current !== null) window.clearTimeout(hideTimeout.current);
    setOpen(true);
    onFocus?.(e);
  };

  const handleBlur = (e: React.FocusEvent<HTMLButtonElement>) => {
    setOpen(false);
    onBlur?.(e);
  };

  React.useEffect(() => {
    return () => {
      if (showTimeout.current !== null) window.clearTimeout(showTimeout.current);
      if (hideTimeout.current !== null) window.clearTimeout(hideTimeout.current);
    };
  }, []);

  const setRefs = (node: HTMLButtonElement | null) => {
    triggerRef.current = node;
    if (typeof ref === 'function') ref(node);
    else if (ref) (ref as React.MutableRefObject<HTMLButtonElement | null>).current = node;
  };

  return (
    <button
      ref={setRefs}
      type="button"
      className={cn('inline-flex items-center', className)}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      onFocus={handleFocus}
      onBlur={handleBlur}
      aria-describedby="tooltip-content"
      {...props}
    >
      {children}
    </button>
  );
});
TooltipTrigger.displayName = 'TooltipTrigger';

/* ── Content ──────────────────────────────────────────── */
const TooltipContent = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & {
    side?: 'top' | 'bottom' | 'left' | 'right';
    align?: 'start' | 'center' | 'end';
  }
>(({ className, side = 'top', align = 'center', children, ...props }, ref) => {
  const { open } = useTooltipContext();

  const sideClasses: Record<string, string> = {
    top: 'bottom-full left-1/2 -translate-x-1/2 mb-2',
    bottom: 'top-full left-1/2 -translate-x-1/2 mt-2',
    left: 'right-full top-1/2 -translate-y-1/2 mr-2',
    right: 'left-full top-1/2 -translate-y-1/2 ml-2',
  };

  const alignClasses: Record<string, string> = {
    start: 'left-0 -translate-x-0',
    center: 'left-1/2 -translate-x-1/2',
    end: 'right-0 left-auto -translate-x-0',
  };

  if (!open) return null;

  return (
    <div
      ref={ref}
      role="tooltip"
      id="tooltip-content"
      data-side={side}
      data-align={align}
      className={cn(
        'absolute z-50 pointer-events-none',
        sideClasses[side],
        align === 'center' ? '' : alignClasses[align],
        'overflow-hidden rounded-md border bg-popover px-3 py-1.5 text-sm text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95',
        className,
      )}
      {...props}
    >
      {children}
    </div>
  );
});
TooltipContent.displayName = 'TooltipContent';

export { TooltipProvider, Tooltip, TooltipTrigger, TooltipContent };
