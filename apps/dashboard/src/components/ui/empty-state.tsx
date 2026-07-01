import * as React from 'react';
import { cn } from '@/lib/utils';
import { Button } from './button';

export interface EmptyStateAction {
  label: string;
  onClick: () => void;
  icon?: React.ElementType;
  variant?: 'primary' | 'secondary' | 'ghost';
}

interface EmptyStateProps {
  icon?: React.ElementType;
  title: string;
  description?: string;
  actions?: EmptyStateAction[];
  className?: string;
}

function EmptyState({
  icon: Icon,
  title,
  description,
  actions,
  className,
}: EmptyStateProps) {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center rounded-md border border-dashed border-border px-6 py-16 text-center',
        className,
      )}
    >
      {Icon && (
        <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-md bg-surface-elevated border border-border">
          <Icon className="h-6 w-6 text-text-muted" aria-hidden="true" />
        </div>
      )}
      <h3 className="text-h3 text-foreground mb-1">{title}</h3>
      {description && (
        <p className="text-small text-text-secondary max-w-md mb-6 leading-relaxed">
          {description}
        </p>
      )}
      {actions && actions.length > 0 && (
        <div className="flex items-center gap-2" role="group" aria-label="Available actions">
          {actions.map((action, idx) => (
            <Button
              key={idx}
              variant={action.variant ?? 'primary'}
              size="sm"
              onClick={action.onClick}
            >
              {action.icon && <action.icon className="h-3.5 w-3.5" />}
              {action.label}
            </Button>
          ))}
        </div>
      )}
    </div>
  );
}

export { EmptyState };
