import { NavLink } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import {
  LayoutDashboard,
  Folder,
  Database,
  Activity,
  Boxes,
  ShieldCheck,
  Cpu,
  Server,
  Puzzle,
  Settings2,
  ChevronLeft,
  ChevronRight,
  Sun,
  Moon,
  Monitor,
  AppWindow,
  type LucideIcon,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useTheme } from '@/components/theme/ThemeProvider';
import { useProjects } from '@/hooks/useCloudOS';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';

/* ── Types ────────────────────────────────────────────────── */
interface NavItem {
  label: string;
  to: string;
  icon: LucideIcon;
  badge?: 'projects' | 'none';
}

interface NavGroup {
  label: string;
  items: NavItem[];
}

interface SidebarProps {
  open: boolean;
  onToggle: () => void;
  className?: string;
}

/* ── Navigation groups ────────────────────────────────────── */
const NAV_GROUPS: NavGroup[] = [
  {
    label: 'General',
    items: [
      { label: 'Applications', to: '/', icon: AppWindow },
      { label: 'Dashboard', to: '/dashboard', icon: LayoutDashboard },
      { label: 'Projects', to: '/projects', icon: Folder, badge: 'projects' },
    ],
  },
  {
    label: 'Infrastructure',
    items: [
      { label: 'Resources', to: '/resources', icon: Database },
      { label: 'Controllers', to: '/controllers', icon: Activity },
      { label: 'Capabilities', to: '/capabilities', icon: Boxes },
      { label: 'Providers', to: '/providers', icon: ShieldCheck },
    ],
  },
  {
    label: 'System',
    items: [
      { label: 'Kernel', to: '/kernel', icon: Cpu },
      { label: 'System', to: '/system', icon: Server },
      { label: 'Plugins', to: '/plugins', icon: Puzzle },
    ],
  },
  {
    label: 'Settings',
    items: [
      { label: 'Settings', to: '/settings', icon: Settings2 },
    ],
  },
];

/* ── Sub-components ───────────────────────────────────────── */

function SidebarNavItem({
  item,
  collapsed,
}: {
  item: NavItem;
  collapsed: boolean;
}) {
  const { data: projects, isLoading: projectsLoading } = useProjects();

  const badgeCount =
    item.badge === 'projects'
      ? projects?.items?.length ?? 0
      : null;

  return (
    <div className={cn(collapsed && 'group relative flex justify-center')}>
      <NavLink
        to={item.to}
        end={item.to === '/'}
        className={({ isActive }: { isActive: boolean }) =>
          cn(
            'relative flex items-center rounded-md text-sm font-medium transition-all duration-150',
            collapsed
              ? 'justify-center px-0 py-2 mx-auto w-10'
              : 'gap-3 px-3 py-2',
            isActive
              ? 'bg-sidebar-accent text-foreground'
              : 'text-sidebar-foreground hover:bg-sidebar-muted hover:text-foreground/80',
          )
        }
      >
        <item.icon
          className={cn('h-4 w-4 shrink-0', collapsed && 'h-5 w-5')}
        />
        <AnimatePresence mode="wait">
          {!collapsed && (
            <motion.span
              initial={{ opacity: 0, width: 0 }}
              animate={{ opacity: 1, width: 'auto' }}
              exit={{ opacity: 0, width: 0 }}
              transition={{ duration: 0.15, ease: 'easeInOut' }}
              className="flex-1 truncate"
            >
              {item.label}
            </motion.span>
          )}
        </AnimatePresence>

        {/* Badge / skeleton */}
        {!collapsed && badgeCount !== null && (
          <span className="ml-auto">
            {projectsLoading ? (
              <Skeleton className="h-4 w-6 rounded-full" />
            ) : (
              <Badge
                variant="secondary"
                className="h-4 min-w-[1.25rem] rounded-full px-1.5 text-[10px] font-medium leading-none"
              >
                {badgeCount}
              </Badge>
            )}
          </span>
        )}
      </NavLink>

      {/* Inline tooltip for collapsed mode */}
      {collapsed && (
        <div className="pointer-events-none absolute left-full top-1/2 z-50 ml-3 -translate-y-1/2 whitespace-nowrap rounded-md border bg-popover px-2.5 py-1 text-xs text-popover-foreground shadow-md opacity-0 transition-opacity duration-150 group-hover:opacity-100">
          {item.label}
        </div>
      )}
    </div>
  );
}

function SidebarGroup({
  group,
  collapsed,
}: {
  group: NavGroup;
  collapsed: boolean;
}) {
  return (
    <div className="px-2">
      {/* Group header — hidden when collapsed */}
      <AnimatePresence mode="wait">
        {!collapsed && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            transition={{ duration: 0.15 }}
            className="px-3 pb-1 pt-4"
          >
            <span className="text-[10px] font-semibold uppercase tracking-[0.12em] text-sidebar-foreground/50">
              {group.label}
            </span>
          </motion.div>
        )}
      </AnimatePresence>

      <div className={cn('flex flex-col', collapsed ? 'gap-1' : 'gap-0.5')}>
        {group.items.map((item) => (
          <SidebarNavItem key={item.to} item={item} collapsed={collapsed} />
        ))}
      </div>
    </div>
  );
}

/* ── Sidebar root ─────────────────────────────────────────── */
export function Sidebar({ open, onToggle, className }: SidebarProps) {
  const { theme, setTheme } = useTheme();

  const cycleTheme = () => {
    const themes: Array<'dark' | 'light' | 'system'> = [
      'dark',
      'light',
      'system',
    ];
    const idx = themes.indexOf(theme);
    setTheme(themes[(idx + 1) % themes.length]);
  };

  const ThemeIcon = theme === 'dark' ? Moon : theme === 'light' ? Sun : Monitor;

  return (
    <motion.aside
      initial={false}
      animate={{ width: open ? 240 : 56 }}
      transition={{ duration: 0.2, ease: 'easeInOut' }}
      className={cn(
        'fixed left-0 top-12 z-30 flex flex-col border-r bg-sidebar',
        'h-[calc(100vh-3rem-1.75rem)]', // viewport minus TopNav (h-12) and StatusBar (h-7)
        className,
      )}
      aria-label="Main navigation"
    >
      {/* Navigation links */}
      <nav className="flex-1 overflow-y-auto overflow-x-hidden py-4 scrollbar-none">
        {NAV_GROUPS.map((group) => (
          <SidebarGroup
            key={group.label}
            group={group}
            collapsed={!open}
          />
        ))}
      </nav>

      {/* Bottom section */}
      <div
        className={cn(
          'flex shrink-0 items-center border-t border-sidebar-border/60 px-3 py-2.5',
          open ? 'justify-between gap-2' : 'flex-col gap-2',
        )}
      >
        {/* Profile placeholder */}
        <div
          className={cn(
            'flex items-center gap-2.5 overflow-hidden',
            !open && 'flex-col',
          )}
        >
          <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-sidebar-accent text-[11px] font-semibold text-muted-foreground">
            CO
          </div>
          <AnimatePresence mode="wait">
            {open && (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="flex flex-col leading-tight"
              >
                <span className="text-xs font-medium text-foreground/80">
                  Cloud Operator
                </span>
                <span className="text-[10px] text-muted-foreground/60">
                  Admin
                </span>
              </motion.div>
            )}
          </AnimatePresence>
        </div>

        {/* Actions */}
        <div className={cn('flex', open ? 'gap-0.5' : 'flex-col gap-0.5')}>
          {/* Theme toggle */}
          <button
            type="button"
            onClick={cycleTheme}
            className={cn(
              'flex items-center justify-center rounded-md p-1.5 text-sidebar-foreground/60 transition-colors hover:bg-sidebar-accent hover:text-foreground',
              'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1',
            )}
            aria-label={`Theme: ${theme}`}
          >
            <ThemeIcon className="h-3.5 w-3.5" />
          </button>

          {/* Collapse toggle */}
          <button
            type="button"
            onClick={onToggle}
            className={cn(
              'flex items-center justify-center rounded-md p-1.5 text-sidebar-foreground/60 transition-colors hover:bg-sidebar-accent hover:text-foreground',
              'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1',
            )}
            aria-label={open ? 'Collapse sidebar' : 'Expand sidebar'}
          >
            {open ? (
              <ChevronLeft className="h-3.5 w-3.5" />
            ) : (
              <ChevronRight className="h-3.5 w-3.5" />
            )}
          </button>
        </div>
      </div>
    </motion.aside>
  );
}
