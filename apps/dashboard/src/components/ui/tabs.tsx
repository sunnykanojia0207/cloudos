import * as React from 'react';
import { cn } from '@/lib/utils';

/* ── Context ──────────────────────────────────────────── */
interface TabsContextValue {
  value: string;
  onValueChange: (value: string) => void;
  orientation: 'horizontal' | 'vertical';
}

const TabsContext = React.createContext<TabsContextValue | null>(null);

function useTabsContext(): TabsContextValue {
  const ctx = React.useContext(TabsContext);
  if (!ctx) throw new Error('Tabs sub-components must be used within <Tabs />');
  return ctx;
}

/* ── Root ──────────────────────────────────────────────── */
interface TabsProps {
  value?: string;
  defaultValue?: string;
  onValueChange?: (value: string) => void;
  orientation?: 'horizontal' | 'vertical';
  className?: string;
  children?: React.ReactNode;
}

function Tabs({
  value: controlledValue,
  defaultValue,
  onValueChange: controlledOnValueChange,
  orientation = 'horizontal',
  className,
  children,
  ...props
}: TabsProps & React.HTMLAttributes<HTMLDivElement>) {
  const isControlled = controlledValue !== undefined;
  const [uncontrolledValue, setUncontrolledValue] = React.useState(
    defaultValue ?? '',
  );

  const value = isControlled ? controlledValue : uncontrolledValue;

  const onValueChange = React.useCallback(
    (next: string) => {
      if (!isControlled) setUncontrolledValue(next);
      controlledOnValueChange?.(next);
    },
    [isControlled, controlledOnValueChange],
  );

  return (
    <TabsContext.Provider value={{ value, onValueChange, orientation }}>
      <div
        className={cn(
          orientation === 'vertical' ? 'flex gap-4' : 'flex flex-col',
          className,
        )}
        data-orientation={orientation}
        {...props}
      >
        {children}
      </div>
    </TabsContext.Provider>
  );
}
Tabs.displayName = 'Tabs';

/* ── List ──────────────────────────────────────────────── */
const TabsList = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => {
  const { orientation } = useTabsContext();
  return (
    <div
      ref={ref}
      role="tablist"
      data-orientation={orientation}
      className={cn(
        'flex border-b border-border',
        orientation === 'vertical' && 'flex-col border-b-0 border-r',
        'gap-1',
        className,
      )}
      {...props}
    />
  );
});
TabsList.displayName = 'TabsList';

/* ── Trigger ──────────────────────────────────────────── */
interface TabsTriggerProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  value: string;
  count?: number;
}

const TabsTrigger = React.forwardRef<HTMLButtonElement, TabsTriggerProps>(
  ({ className, value: tabValue, count, children, ...props }, ref) => {
    const { value, onValueChange } = useTabsContext();
    const isActive = value === tabValue;

    return (
      <button
        ref={ref}
        type="button"
        role="tab"
        aria-selected={isActive}
        data-state={isActive ? 'active' : 'inactive'}
        onClick={() => onValueChange(tabValue)}
        className={cn(
          'inline-flex items-center justify-center whitespace-nowrap px-4 py-2',
          'text-tab text-text-secondary transition-colors duration-150',
          'border-b-2 border-transparent -mb-px',
          'hover:text-foreground',
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset',
          'disabled:pointer-events-none disabled:opacity-50',
          isActive && 'border-accent text-foreground',
          className,
        )}
        {...props}
      >
        {children}
        {count !== undefined && (
          <span className="ml-1.5 rounded-sm bg-accent-subtle px-1.5 py-0.5 text-caption text-text-secondary tabular-nums">
            {count}
          </span>
        )}
      </button>
    );
  },
);
TabsTrigger.displayName = 'TabsTrigger';

/* ── Content ──────────────────────────────────────────── */
interface TabsContentProps extends React.HTMLAttributes<HTMLDivElement> {
  value: string;
}

const TabsContent = React.forwardRef<HTMLDivElement, TabsContentProps>(
  ({ className, value: tabValue, ...props }, ref) => {
    const { value } = useTabsContext();
    const isActive = value === tabValue;

    if (!isActive) return null;

    return (
      <div
        ref={ref}
        role="tabpanel"
        data-state={isActive ? 'active' : 'inactive'}
        className={cn(
          'pt-4 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset',
          className,
        )}
        {...props}
      />
    );
  },
);
TabsContent.displayName = 'TabsContent';

export { Tabs, TabsList, TabsTrigger, TabsContent };
