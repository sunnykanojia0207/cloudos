import * as React from 'react';
import { Search } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Dialog, DialogContent, DialogTitle, DialogDescription } from './dialog';

/* ── Context ──────────────────────────────────────────── */
interface CommandContextValue {
  activeIndex: number;
  setActiveIndex: React.Dispatch<React.SetStateAction<number>>;
  items: React.MutableRefObject<HTMLDivElement[]>;
  onSelectCallback?: (value: string) => void;
}

const CommandContext = React.createContext<CommandContextValue | null>(null);

function useCommandContext(): CommandContextValue {
  const ctx = React.useContext(CommandContext);
  if (!ctx) throw new Error('Command sub-components must be used within <Command />');
  return ctx;
}

/* ── Root ──────────────────────────────────────────────── */
interface CommandProps {
  className?: string;
  children?: React.ReactNode;
  onSelect?: (value: string) => void;
}

function Command({ className, children, onSelect }: CommandProps) {
  const [activeIndex, setActiveIndex] = React.useState(-1);
  const items = React.useRef<HTMLDivElement[]>([]);

  return (
    <CommandContext.Provider
      value={{ activeIndex, setActiveIndex, items, onSelectCallback: onSelect }}
    >
      <div
        className={cn(
          'flex h-full w-full flex-col overflow-hidden rounded-lg bg-surface-elevated text-foreground',
          className,
        )}
      >
        {children}
      </div>
    </CommandContext.Provider>
  );
}
Command.displayName = 'Command';

/* ── Dialog (command palette wrapper) ─────────────────── */
interface CommandDialogProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  children?: React.ReactNode;
}

function CommandDialog({ open, onOpenChange, children }: CommandDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTitle className="sr-only">Command Palette</DialogTitle>
      <DialogDescription className="sr-only">
        Search and run commands
      </DialogDescription>
      <DialogContent size="wide" className="overflow-hidden p-0 shadow-lg max-w-[640px]">
        <Command className="[&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:py-1.5 [&_[cmdk-group-heading]]:text-caption [&_[cmdk-group-heading]]:font-medium [&_[cmdk-group-heading]]:text-text-muted [&_[cmdk-group]:not([hidden])_~[cmdk-group]]:pt-2 [&_[cmdk-group]]:px-2 [&_[cmdk-input-wrapper]_svg]:h-5 [&_[cmdk-input-wrapper]_svg]:w-5 [&_[cmdk-input]]:h-12 [&_[cmdk-item]]:px-2 [&_[cmdk-item]]:py-2.5 [&_[cmdk-item]_svg]:h-4 [&_[cmdk-item]_svg]:w-4">
          {children}
        </Command>
      </DialogContent>
    </Dialog>
  );
}
CommandDialog.displayName = 'CommandDialog';

/* ── Input ─────────────────────────────────────────────── */
interface CommandInputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {
  wrapperClassName?: string;
}

const CommandInput = React.forwardRef<HTMLInputElement, CommandInputProps>(
  ({ className, wrapperClassName, ...props }, ref) => {
    const { activeIndex, setActiveIndex, items } = useCommandContext();

    return (
      <div
        className={cn('flex items-center border-b border-border px-4', wrapperClassName)}
        cmdk-input-wrapper=""
      >
        <Search className="mr-3 h-4 w-4 shrink-0 text-text-muted" />
        <input
          ref={ref}
          className={cn(
            'flex h-12 w-full rounded-md bg-transparent py-3 text-body outline-none placeholder:text-text-muted disabled:cursor-not-allowed disabled:opacity-50',
            'font-medium',
            className,
          )}
          onInput={() => {
            setActiveIndex(-1);
          }}
          onKeyDown={(e) => {
            if (e.key === 'ArrowDown') {
              e.preventDefault();
              setActiveIndex((prev: number) => {
                const next = prev + 1;
                const max = items.current.length - 1;
                return next > max ? 0 : next;
              });
            }
            if (e.key === 'ArrowUp') {
              e.preventDefault();
              setActiveIndex((prev: number) => {
                const next = prev - 1;
                const max = items.current.length - 1;
                return next < 0 ? max : next;
              });
            }
            if (e.key === 'Enter') {
              e.preventDefault();
              const el = items.current[activeIndex];
              el?.click();
            }
            props.onKeyDown?.(e);
          }}
          {...props}
        />
      </div>
    );
  },
);
CommandInput.displayName = 'CommandInput';

/* ── List ──────────────────────────────────────────────── */
const CommandList = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn('max-h-[min(480px,60vh)] overflow-y-auto overflow-x-hidden p-2', className)}
    {...props}
  />
));
CommandList.displayName = 'CommandList';

/* ── Empty ─────────────────────────────────────────────── */
const CommandEmpty = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn('py-8 text-center text-small text-text-muted', className)}
    {...props}
  />
));
CommandEmpty.displayName = 'CommandEmpty';

/* ── Group ─────────────────────────────────────────────── */
interface CommandGroupProps extends React.HTMLAttributes<HTMLDivElement> {
  heading?: string;
}

const CommandGroup = React.forwardRef<HTMLDivElement, CommandGroupProps>(
  ({ className, heading, children, ...props }, ref) => (
    <div
      ref={ref}
      className={cn(
        'overflow-hidden text-foreground',
        className,
      )}
      {...props}
    >
      {heading && (
        <div cmdk-group-heading="" className="px-2 py-1.5 text-caption font-medium text-text-muted">
          {heading}
        </div>
      )}
      <div className="flex flex-col gap-0.5">
        {children}
      </div>
    </div>
  ),
);
CommandGroup.displayName = 'CommandGroup';

/* ── Item ──────────────────────────────────────────────── */
interface CommandItemProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'onSelect'> {
  value?: string;
  onSelect?: (value: string) => void;
  disabled?: boolean;
}

const CommandItem = React.forwardRef<HTMLDivElement, CommandItemProps>(
  ({ className, value, onSelect, disabled, onClick, children, ...props }, ref) => {
    const { activeIndex, items, onSelectCallback } = useCommandContext();
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
        role="option"
        aria-selected={isActive}
        data-disabled={disabled ? true : undefined}
        data-active={isActive ? true : undefined}
        onClick={(e) => {
          if (disabled) return;
          onClick?.(e);
          if (value) {
            onSelect?.(value);
            onSelectCallback?.(value);
          }
        }}
        className={cn(
          'relative flex cursor-default select-none items-center rounded-sm px-2 py-2.5 text-body outline-none',
          'aria-selected:bg-accent-subtle data-[active=true]:bg-accent-subtle',
          'data-[disabled=true]:pointer-events-none data-[disabled=true]:opacity-50',
          className,
        )}
        {...props}
      >
        {children}
      </div>
    );
  },
);
CommandItem.displayName = 'CommandItem';

/* ── Shortcut ──────────────────────────────────────────── */
const CommandShortcut = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLSpanElement>) => (
  <span
    className={cn(
      'ml-auto text-caption tracking-widest text-text-muted',
      className,
    )}
    {...props}
  />
);
CommandShortcut.displayName = 'CommandShortcut';

export {
  Command,
  CommandDialog,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
  CommandShortcut,
};
