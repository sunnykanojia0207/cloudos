import * as React from 'react';
import { createPortal } from 'react-dom';
import { Slot } from '@radix-ui/react-slot';
import { AnimatePresence, motion } from 'framer-motion';
import { cn } from '@/lib/utils';

/* ── Context ──────────────────────────────────────────── */
interface DropdownContextValue {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  triggerRef: React.RefObject<HTMLButtonElement | null>;
}

const DropdownContext = React.createContext<DropdownContextValue | null>(null);

function useDropdownContext(): DropdownContextValue {
  const ctx = React.useContext(DropdownContext);
  if (!ctx) throw new Error('Dropdown sub-components must be used within <DropdownMenu />');
  return ctx;
}

/* ── Root ──────────────────────────────────────────────── */
interface DropdownMenuProps {
  open?: boolean;
  defaultOpen?: boolean;
  onOpenChange?: (open: boolean) => void;
  children: React.ReactNode;
}

function DropdownMenu({
  open: controlledOpen,
  defaultOpen,
  onOpenChange: controlledOnOpenChange,
  children,
}: DropdownMenuProps) {
  const isControlled = controlledOpen !== undefined;
  const [uncontrolledOpen, setUncontrolledOpen] = React.useState(defaultOpen ?? false);
  const triggerRef = React.useRef<HTMLButtonElement | null>(null);

  const open = isControlled ? controlledOpen : uncontrolledOpen;

  const onOpenChange = React.useCallback(
    (next: boolean) => {
      if (!isControlled) setUncontrolledOpen(next);
      controlledOnOpenChange?.(next);
    },
    [isControlled, controlledOnOpenChange],
  );

  return (
    <DropdownContext.Provider value={{ open, onOpenChange, triggerRef }}>
      {children}
    </DropdownContext.Provider>
  );
}
DropdownMenu.displayName = 'DropdownMenu';

/* ── Trigger ──────────────────────────────────────────── */
interface DropdownMenuTriggerProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  asChild?: boolean;
}

const DropdownMenuTrigger = React.forwardRef<
  HTMLButtonElement,
  DropdownMenuTriggerProps
>(({ className, asChild = false, onClick, children, ...props }, ref) => {
  const { open, onOpenChange, triggerRef } = useDropdownContext();
  const Comp = asChild ? Slot : 'button';

  return (
    <Comp
      ref={(node) => {
        triggerRef.current = node as HTMLButtonElement;
        if (typeof ref === 'function') ref(node as HTMLButtonElement);
        else if (ref) (ref as React.MutableRefObject<HTMLButtonElement | null>).current = node as HTMLButtonElement;
      }}
      type={asChild ? undefined : 'button'}
      aria-haspopup="true"
      aria-expanded={open}
      onClick={(e) => {
        onClick?.(e);
        onOpenChange(!open);
      }}
      className={className}
      {...props}
    >
      {children}
    </Comp>
  );
});
DropdownMenuTrigger.displayName = 'DropdownMenuTrigger';

/* ── Content ──────────────────────────────────────────── */
interface DropdownMenuContentProps {
  align?: 'start' | 'end' | 'center';
  sideOffset?: number;
  className?: string;
  children?: React.ReactNode;
}

const DropdownMenuContent = React.forwardRef<HTMLDivElement, DropdownMenuContentProps>(
  ({ align = 'start', sideOffset = 8, className, children }, ref) => {
    const { open, onOpenChange, triggerRef } = useDropdownContext();
    const contentRef = React.useRef<HTMLDivElement | null>(null);

    // Close on Escape / click outside
    React.useEffect(() => {
      if (!open) return;
      const handler = (e: MouseEvent | TouchEvent) => {
        if (
          contentRef.current &&
          !contentRef.current.contains(e.target as Node) &&
          triggerRef.current &&
          !triggerRef.current.contains(e.target as Node)
        ) {
          onOpenChange(false);
        }
      };
      const keyHandler = (e: KeyboardEvent) => {
        if (e.key === 'Escape') {
          onOpenChange(false);
          triggerRef.current?.focus();
        }
      };
      document.addEventListener('mousedown', handler);
      document.addEventListener('keydown', keyHandler);
      return () => {
        document.removeEventListener('mousedown', handler);
        document.removeEventListener('keydown', keyHandler);
      };
    }, [open, onOpenChange, triggerRef]);

    // Calculate position based on trigger
    const [position, setPosition] = React.useState({ top: 0, left: 0 });
    React.useEffect(() => {
      if (!open || !triggerRef.current) return;
      const rect = triggerRef.current.getBoundingClientRect();
      let left = rect.left;
      if (align === 'end') left = rect.right - 200; // approximate width
      else if (align === 'center') left = rect.left + rect.width / 2 - 100;
      setPosition({ top: rect.bottom + sideOffset, left });
    }, [open, align, sideOffset, triggerRef]);

    return createPortal(
      <AnimatePresence>
        {open && (
          <motion.div
            ref={(node) => {
              contentRef.current = node;
              if (typeof ref === 'function') ref(node);
              else if (ref) (ref as React.MutableRefObject<HTMLDivElement | null>).current = node;
            }}
            initial={{ opacity: 0, y: -4 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -4 }}
            transition={{ duration: 0.15, ease: [0, 0, 0.2, 1] }}
            role="menu"
            style={{ top: position.top, left: position.left }}
            className={cn(
              'fixed z-[100] min-w-[200px] rounded-md border border-border bg-surface-elevated p-1 shadow-md',
              className,
            )}
          >
            {children}
          </motion.div>
        )}
      </AnimatePresence>,
      document.body,
    );
  },
);
DropdownMenuContent.displayName = 'DropdownMenuContent';

/* ── Item ──────────────────────────────────────────────── */
interface DropdownMenuItemProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  inset?: boolean;
}

const DropdownMenuItem = React.forwardRef<HTMLButtonElement, DropdownMenuItemProps>(
  ({ className, inset, children, onClick, ...props }, ref) => {
    const { onOpenChange } = useDropdownContext();

    return (
      <button
        ref={ref}
        type="button"
        role="menuitem"
        onClick={(e) => {
          onClick?.(e);
          onOpenChange(false);
        }}
        className={cn(
          'relative flex w-full cursor-default select-none items-center rounded-sm px-2 py-1.5 text-body outline-none',
          'text-foreground hover:bg-accent-subtle',
          'focus-visible:bg-accent-subtle focus-visible:outline-none',
          'data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
          inset && 'pl-8',
          className,
        )}
        {...props}
      >
        {children}
      </button>
    );
  },
);
DropdownMenuItem.displayName = 'DropdownMenuItem';

/* ── Separator ─────────────────────────────────────────── */
const DropdownMenuSeparator = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => (
  <div
    role="separator"
    className={cn('my-1 h-px bg-border', className)}
    {...props}
  />
);
DropdownMenuSeparator.displayName = 'DropdownMenuSeparator';

/* ── Label ─────────────────────────────────────────────── */
const DropdownMenuLabel = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & { inset?: boolean }
>(({ className, inset, ...props }, ref) => (
  <div
    ref={ref}
    className={cn(
      'px-2 py-1.5 text-caption font-medium text-text-muted',
      inset && 'pl-8',
      className,
    )}
    {...props}
  />
));
DropdownMenuLabel.displayName = 'DropdownMenuLabel';

export {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuLabel,
};
