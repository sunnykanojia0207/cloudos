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
  filter?: (value: string, search: string) => number;
}

function Command({ className, children, onSelect, filter: _filter }: CommandProps) {
  const [activeIndex, setActiveIndex] = React.useState(-1);
  const items = React.useRef<HTMLDivElement[]>([]);

  return (
    <CommandContext.Provider
      value={{ activeIndex, setActiveIndex, items, onSelectCallback: onSelect }}
    >
      <div
        className={cn(
          'flex h-full w-full flex-col overflow-hidden rounded-md bg-popover text-popover-foreground',
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
      <DialogContent className="overflow-hidden p-0 shadow-lg max-w-[540px]">
        <Command className="[&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:font-medium [&_[cmdk-group-heading]]:text-muted-foreground [&_[cmdk-group]:not([hidden])_~[cmdk-group]]:pt-0 [&_[cmdk-group]]:px-2 [&_[cmdk-input-wrapper]_svg]:h-5 [&_[cmdk-input-wrapper]_svg]:w-5 [&_[cmdk-input]]:h-12 [&_[cmdk-item]]:px-2 [&_[cmdk-item]]:py-3 [&_[cmdk-item]_svg]:h-5 [&_[cmdk-item]_svg]:w-5">
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
        className={cn('flex items-center border-b px-3', wrapperClassName)}
        cmdk-input-wrapper=""
      >
        <Search className="mr-2 h-4 w-4 shrink-0 opacity-50" />
        <input
          ref={ref}
          className={cn(
            'flex h-11 w-full rounded-md bg-transparent py-3 text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50',
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
    className={cn('max-h-[300px] overflow-y-auto overflow-x-hidden', className)}
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
    className={cn('py-6 text-center text-sm', className)}
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
        'overflow-hidden p-1 text-foreground [&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:py-1.5 [&_[cmdk-group-heading]]:text-xs [&_[cmdk-group-heading]]:font-medium [&_[cmdk-group-heading]]:text-muted-foreground',
        className,
      )}
      {...props}
    >
      {heading && (
        <div cmdk-group-heading="">
          {heading}
        </div>
      )}
      {children}
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
          'relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none aria-selected:bg-accent aria-selected:text-accent-foreground data-[disabled=true]:pointer-events-none data-[disabled=true]:opacity-50',
          isActive && 'bg-accent text-accent-foreground',
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
      'ml-auto text-xs tracking-widest text-muted-foreground',
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
