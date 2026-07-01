import { motion } from 'framer-motion';
import {
  Menu,
  Bell,
  Search,
  Command as CommandIcon,
} from 'lucide-react';
import { useTheme } from '@/components/theme/ThemeProvider';
import { useCommandPalette } from '@/components/layout/CommandPalette';
import { useHealth } from '@/hooks/useCloudOS';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuLabel,
} from '@/components/ui/dropdown-menu';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { cn } from '@/lib/utils';
import { Breadcrumbs } from './Breadcrumbs';

/* ── Props ────────────────────────────────────────────────── */
interface TopNavProps {
  onToggleSidebar: () => void;
  sidebarOpen: boolean;
}

/* ── TopNav ───────────────────────────────────────────────── */
export function TopNav({ onToggleSidebar }: TopNavProps) {
  const { theme } = useTheme();
  const { setOpen: setCommandOpen } = useCommandPalette();
  const { data: health } = useHealth();

  const isHealthy = health?.overall?.status === 'healthy' || health?.overall?.status === 'running';
  const isDegraded = health?.overall?.status === 'degraded' || health?.overall?.status === 'warning';
  const healthColor = isHealthy ? 'bg-success' : isDegraded ? 'bg-warning' : 'bg-danger';

  return (
    <header
      className={cn(
        'fixed top-0 z-50 flex h-12 w-full items-center gap-2 border-b border-border',
        'bg-topnav-bg/80 backdrop-blur-xl supports-[backdrop-filter]:bg-topnav-bg/60',
        'px-4',
      )}
    >
      {/* ── Left section ─────────────────────────── */}
      <div className="flex items-center gap-2">
        {/* Hamburger */}
        <button
          type="button"
          onClick={onToggleSidebar}
          aria-label="Toggle sidebar"
          className={cn(
            'flex h-8 w-8 items-center justify-center rounded-md',
            'text-text-muted hover:text-foreground hover:bg-accent-subtle',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
          )}
        >
          <Menu className="h-4 w-4" />
        </button>

        {/* Breadcrumbs */}
        <div className="hidden sm:block">
          <Breadcrumbs />
        </div>
      </div>

      {/* ── Spacer ──────────────────────────────── */}
      <div className="flex-1" />

      {/* ── Right section ───────────────────────── */}
      <div className="flex items-center gap-1">
        {/* Cmd+K search trigger */}
        <button
          type="button"
          onClick={() => setCommandOpen(true)}
          className={cn(
            'flex items-center gap-2 rounded-md border border-border bg-surface px-2.5 py-1.5',
            'text-small text-text-muted hover:text-foreground hover:border-border-hover',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
            'min-w-[160px] sm:min-w-[200px]',
          )}
          aria-label="Open command palette (Ctrl+K)"
        >
          <Search className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
          <span className="flex-1 text-left">Search...</span>
          <kbd className="hidden rounded border border-border/50 bg-muted/50 px-1 py-[1px] text-[10px] font-medium text-text-muted sm:inline-block">
            ⌘K
          </kbd>
        </button>

        {/* Kernel health indicator dot */}
        <span
          className={cn(
            'inline-block h-2 w-2 rounded-full',
            healthColor,
          )}
          aria-label={
            isHealthy ? 'System healthy' : isDegraded ? 'System degraded' : 'System error'
          }
          title={
            isHealthy ? 'Healthy' : isDegraded ? 'Degraded' : 'Error'
          }
        />

        {/* Notification bell */}
        <button
          type="button"
          aria-label="Notifications"
          className={cn(
            'flex h-8 w-8 items-center justify-center rounded-md',
            'text-text-muted hover:text-foreground hover:bg-accent-subtle',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
          )}
        >
          <Bell className="h-4 w-4" />
        </button>

        {/* Profile avatar */}
        <DropdownMenu>
          <DropdownMenuTrigger
            className={cn(
              'flex items-center justify-center rounded-full transition-opacity hover:opacity-80',
              'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
            )}
            aria-label="User menu"
          >
            <Avatar>
              <AvatarFallback className="bg-accent-subtle text-caption font-medium text-accent">
                CO
              </AvatarFallback>
            </Avatar>
          </DropdownMenuTrigger>

          <DropdownMenuContent align="end" sideOffset={8}>
            <DropdownMenuLabel className="font-normal">
              <div className="flex flex-col">
                <span className="text-body font-medium text-foreground">
                  Cloud Operator
                </span>
                <span className="text-small text-text-secondary">
                  admin@cloudos.io
                </span>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onSelect={() => {}}>Profile</DropdownMenuItem>
            <DropdownMenuItem onSelect={() => {}}>
              Keyboard Shortcuts
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onSelect={() => {}} className="text-danger">
              Sign Out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
