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
        'inline-flex h-10 items-center justify-center rounded-md bg-muted p-1 text-muted-foreground',
        orientation === 'vertical' && 'flex-col h-auto w-fit',
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
}

const TabsTrigger = React.forwardRef<HTMLButtonElement, TabsTriggerProps>(
  ({ className, value: tabValue, ...props }, ref) => {
    const { value, onValueChange } = useTabsContext();
    const isActive = value === tabValue;

    return (
      <button
        ref={ref}
        type="button"
        role="tab"
        aria-selected={isActive}
        data-state={isActive ? 'active' : 'inactive'}
        data-orientation={useTabsContext().orientation}
        onClick={() => onValueChange(tabValue)}
        className={cn(
          'inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
          isActive
            ? 'bg-background text-foreground shadow-sm'
            : 'hover:text-foreground',
          className,
        )}
        {...props}
      />
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
          'mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
          className,
        )}
        {...props}
      />
    );
  },
);
TabsContent.displayName = 'TabsContent';

export { Tabs, TabsList, TabsTrigger, TabsContent };
