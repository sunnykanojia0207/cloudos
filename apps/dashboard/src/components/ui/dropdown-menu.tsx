import * as React from 'react';
import { createPortal } from 'react-dom';
import { cn } from '@/lib/utils';

/* ── Context ──────────────────────────────────────────── */
interface DropdownMenuContextValue {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  activeIndex: number;
  setActiveIndex: React.Dispatch<React.SetStateAction<number>>;
  items: React.MutableRefObject<HTMLDivElement[]>;
  triggerRef: React.RefObject<HTMLButtonElement | null>;
}

const DropdownMenuContext = React.createContext<DropdownMenuContextValue | null>(null);

function useDropdownMenuContext(): DropdownMenuContextValue {
  const ctx = React.useContext(DropdownMenuContext);
  if (!ctx) throw new Error('DropdownMenu sub-components must be used within <DropdownMenu />');
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
  const [activeIndex, setActiveIndex] = React.useState(-1);
  const items = React.useRef<HTMLDivElement[]>([]);
  const triggerRef = React.useRef<HTMLButtonElement | null>(null);

  const open = isControlled ? controlledOpen : uncontrolledOpen;

  const onOpenChange = React.useCallback(
    (next: boolean) => {
      if (!isControlled) setUncontrolledOpen(next);
      if (!next) setActiveIndex(-1);
      controlledOnOpenChange?.(next);
    },
    [isControlled, controlledOnOpenChange],
  );

  return (
    <DropdownMenuContext.Provider
      value={{ open, onOpenChange, activeIndex, setActiveIndex, items, triggerRef }}
    >
      <div className="relative inline-flex">{children}</div>
    </DropdownMenuContext.Provider>
  );
}
DropdownMenu.displayName = 'DropdownMenu';

/* ── Trigger ──────────────────────────────────────────── */
const DropdownMenuTrigger = React.forwardRef<
  HTMLButtonElement,
  React.ButtonHTMLAttributes<HTMLButtonElement>
>(({ className, onClick, onKeyDown, children, ...props }, ref) => {
  const { open, onOpenChange, triggerRef } = useDropdownMenuContext();

  const setRefs = (node: HTMLButtonElement | null) => {
    triggerRef.current = node;
    if (typeof ref === 'function') ref(node);
    else if (ref) (ref as React.MutableRefObject<HTMLButtonElement | null>).current = node;
  };

  return (
    <button
      ref={setRefs}
      type="button"
      aria-haspopup="menu"
      aria-expanded={open}
      data-state={open ? 'open' : 'closed'}
      className={cn('inline-flex items-center justify-center', className)}
      onClick={(e) => {
        onClick?.(e);
        onOpenChange(!open);
      }}
      onKeyDown={(e) => {
        if (e.key === 'ArrowDown' || e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          if (!open) onOpenChange(true);
        }
        onKeyDown?.(e);
      }}
      {...props}
    >
      {children}
    </button>
  );
});
DropdownMenuTrigger.displayName = 'DropdownMenuTrigger';

/* ── Content (portal) ─────────────────────────────────── */
interface DropdownMenuContentProps extends React.HTMLAttributes<HTMLDivElement> {
  align?: 'start' | 'center' | 'end';
  sideOffset?: number;
  alignOffset?: number;
}

const DropdownMenuContent = React.forwardRef<HTMLDivElement, DropdownMenuContentProps>(
  ({ className, align = 'center', sideOffset = 8, alignOffset = 0, children, ...props }, ref) => {
    const { open, onOpenChange, activeIndex, setActiveIndex, items, triggerRef } =
      useDropdownMenuContext();
    const contentRef = React.useRef<HTMLDivElement | null>(null);
    const [position, setPosition] = React.useState<{ top: number; left: number } | null>(null);

    // Calculate position
    React.useEffect(() => {
      if (!open || !triggerRef.current) return;
      const triggerRect = triggerRef.current.getBoundingClientRect();
      let left = triggerRect.left;
      if (align === 'center') left = triggerRect.left + triggerRect.width / 2;
      else if (align === 'end') left = triggerRect.right;
      left += alignOffset;

      setPosition({
        top: triggerRect.bottom + sideOffset,
        left,
      });
    }, [open, align, sideOffset, alignOffset, triggerRef]);

    // Keyboard navigation
    const handleKeyDown = (e: React.KeyboardEvent) => {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setActiveIndex((prev) => {
          const next = prev + 1;
          const max = items.current.length - 1;
          return next > max ? 0 : next;
        });
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault();
        setActiveIndex((prev) => {
          const next = prev - 1;
          const max = items.current.length - 1;
          return next < 0 ? max : next;
        });
      }
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        const el = items.current[activeIndex];
        el?.click();
        onOpenChange(false);
      }
      if (e.key === 'Escape') {
        e.preventDefault();
        onOpenChange(false);
        triggerRef.current?.focus();
      }
    };

    // Close on outside click
    React.useEffect(() => {
      if (!open) return;
      const handler = (e: MouseEvent) => {
        if (
          contentRef.current &&
          !contentRef.current.contains(e.target as Node) &&
          triggerRef.current &&
          !triggerRef.current.contains(e.target as Node)
        ) {
          onOpenChange(false);
        }
      };
      document.addEventListener('mousedown', handler);
      return () => document.removeEventListener('mousedown', handler);
    }, [open, onOpenChange, triggerRef]);

    // Focus first item on open
    React.useEffect(() => {
      if (open) {
        setActiveIndex(0);
      }
    }, [open, setActiveIndex]);

    const content = open ? (
      <div
        ref={(node) => {
          contentRef.current = node;
          if (typeof ref === 'function') ref(node);
          else if (ref) (ref as React.MutableRefObject<HTMLDivElement | null>).current = node;
        }}
        role="menu"
        aria-orientation="vertical"
        data-state={open ? 'open' : 'closed'}
        onKeyDown={handleKeyDown}
        style={
          position
            ? { position: 'fixed', top: position.top, left: position.left }
            : undefined
        }
        className={cn(
          'z-50 min-w-[8rem] overflow-hidden rounded-md border bg-popover p-1 text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95',
          className,
        )}
        {...props}
      >
        {children}
      </div>
    ) : null;

    if (!content) return null;

    return createPortal(content, document.body);
  },
);
DropdownMenuContent.displayName = 'DropdownMenuContent';

/* ── Item ──────────────────────────────────────────────── */
interface DropdownMenuItemProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'onSelect'> {
  inset?: boolean;
  disabled?: boolean;
  onSelect?: (e: Event) => void;
}

const DropdownMenuItem = React.forwardRef<HTMLDivElement, DropdownMenuItemProps>(
  ({ className, inset, disabled, onSelect, onClick, children, ...props }, ref) => {
    const { activeIndex, items, onOpenChange } = useDropdownMenuContext();
    const itemRef = React.useRef<HTMLDivElement | null>(null);
    const indexRef = React.useRef(-1);

    React.useEffect(() => {
      const idx = items.current.length;
      indexRef.current = idx;
      items.current.push(itemRef.current!);
      return () => {
        items.current.splice(idx, 1);
      };
    }, [items]);

    const isActive = activeIndex === indexRef.current;

    return (
      <div
        ref={(node) => {
          itemRef.current = node;
          if (typeof ref === 'function') ref(node);
          else if (ref) (ref as React.MutableRefObject<HTMLDivElement | null>).current = node;
        }}
        role="menuitem"
        aria-disabled={disabled}
        data-highlighted={isActive ? '' : undefined}
        tabIndex={isActive ? 0 : -1}
        onClick={(e) => {
          if (disabled) return;
          onClick?.(e);
          onSelect?.(e as unknown as Event);
          onOpenChange(false);
        }}
        className={cn(
          'relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none transition-colors focus:bg-accent focus:text-accent-foreground data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground',
          isActive && 'bg-accent text-accent-foreground',
          inset && 'pl-8',
          disabled && 'pointer-events-none opacity-50',
          className,
        )}
        {...props}
      >
        {children}
      </div>
    );
  },
);
DropdownMenuItem.displayName = 'DropdownMenuItem';

/* ── Separator ──────────────────────────────────────────── */
const DropdownMenuSeparator = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    role="separator"
    aria-orientation="horizontal"
    className={cn('-mx-1 my-1 h-px bg-muted', className)}
    {...props}
  />
));
DropdownMenuSeparator.displayName = 'DropdownMenuSeparator';

/* ── Label ──────────────────────────────────────────────── */
const DropdownMenuLabel = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & { inset?: boolean }
>(({ className, inset, ...props }, ref) => (
  <div
    ref={ref}
    className={cn(
      'px-2 py-1.5 text-sm font-semibold',
      inset && 'pl-8',
      className,
    )}
    {...props}
  />
));
DropdownMenuLabel.displayName = 'DropdownMenuLabel';

/* ── Group ──────────────────────────────────────────────── */
const DropdownMenuGroup = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div ref={ref} className={cn('', className)} role="group" {...props} />
));
DropdownMenuGroup.displayName = 'DropdownMenuGroup';

export {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuLabel,
  DropdownMenuGroup,
};
