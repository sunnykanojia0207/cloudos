import { motion } from 'framer-motion';
import { Menu, Sun, Moon, Monitor, Bell, Command as CommandIcon } from 'lucide-react';
import { useTheme } from '@/components/theme/ThemeProvider';
import { useCommandPalette } from '@/components/layout/CommandPalette';
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

/* ── Icon wrapper with hover animation ────────────────────── */
function IconButton({
  children,
  onClick,
  label,
  className,
}: {
  children: React.ReactNode;
  onClick?: () => void;
  label: string;
  className?: string;
}) {
  return (
    <motion.button
      type="button"
      whileHover={{ scale: 1.05 }}
      whileTap={{ scale: 0.95 }}
      onClick={onClick}
      aria-label={label}
      className={cn(
        'relative flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground/70 transition-colors hover:bg-muted/60 hover:text-foreground',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1',
        className,
      )}
    >
      {children}
    </motion.button>
  );
}

/* ── TopNav ───────────────────────────────────────────────── */
export function TopNav({ onToggleSidebar, sidebarOpen }: TopNavProps) {
  const { theme, setTheme } = useTheme();
  const { setOpen: setCommandOpen } = useCommandPalette();

  const cycleTheme = () => {
    const themes: Array<'dark' | 'light' | 'system'> = ['dark', 'light', 'system'];
    const idx = themes.indexOf(theme);
    setTheme(themes[(idx + 1) % themes.length]);
  };

  const ThemeIcon = theme === 'dark' ? Moon : theme === 'light' ? Sun : Monitor;

  return (
    <header
      className={cn(
        'fixed top-0 z-50 flex h-12 w-full items-center gap-1 border-b',
        'bg-background/80 backdrop-blur-xl supports-[backdrop-filter]:bg-background/60',
        'px-3',
      )}
    >
      {/* ── Left section ─────────────────────────── */}
      <div className="flex items-center gap-1.5">
        {/* Hamburger menu */}
        <IconButton
          label={sidebarOpen ? 'Close sidebar' : 'Open sidebar'}
          onClick={onToggleSidebar}
        >
          <Menu className="h-4 w-4" />
        </IconButton>

        {/* Breadcrumbs */}
        <div className="hidden sm:block">
          <Breadcrumbs />
        </div>
      </div>

      {/* ── Spacer ───────────────────────────────── */}
      <div className="flex-1" />

      {/* ── Right section ────────────────────────── */}
      <div className="flex items-center gap-0.5">
        {/* Command palette trigger with keyboard badge */}
        <button
          type="button"
          onClick={() => setCommandOpen(true)}
          className={cn(
            'flex items-center gap-1.5 rounded-md px-2 py-1.5 text-xs transition-colors',
            'text-muted-foreground/60 hover:bg-muted/60 hover:text-foreground/80',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1',
          )}
          aria-label="Open command palette (Ctrl+K)"
        >
          <CommandIcon className="h-3.5 w-3.5" />
          <kbd className="hidden rounded border border-border/50 bg-muted/50 px-1 py-[1px] text-[9px] font-medium tracking-wide text-muted-foreground/50 sm:inline-block">
            ⌘K
          </kbd>
        </button>

        {/* Separator */}
        <div className="mx-1.5 h-4 w-px bg-border/60" aria-hidden="true" />

        {/* Theme toggle */}
        <IconButton label={`Theme: ${theme}`} onClick={cycleTheme}>
          <ThemeIcon className="h-4 w-4" />
        </IconButton>

        {/* Notification bell */}
        <IconButton label="Notifications">
          <Bell className="h-4 w-4" />
        </IconButton>

        {/* Profile avatar dropdown */}
        <DropdownMenu>
          <DropdownMenuTrigger
            className={cn(
              'ml-1 flex items-center justify-center rounded-full transition-opacity hover:opacity-80',
              'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1',
            )}
            aria-label="User menu"
          >
            <Avatar className="h-7 w-7">
              <AvatarFallback className="bg-sidebar-accent text-[11px] font-semibold text-muted-foreground">
                CO
              </AvatarFallback>
            </Avatar>
          </DropdownMenuTrigger>

          <DropdownMenuContent align="end" sideOffset={8}>
            <DropdownMenuLabel className="font-normal">
              <div className="flex flex-col">
                <span className="text-sm font-medium text-foreground">
                  Cloud Operator
                </span>
                <span className="text-xs text-muted-foreground">
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
            <DropdownMenuItem onSelect={() => {}} className="text-destructive">
              Sign Out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
