import { NavLink } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Box,
  GitBranch,
  Heart,
  GitMerge,
  Cpu,
  Settings,
  Puzzle,
  ChevronLeft,
  ChevronRight,
  Sun,
  Moon,
  Monitor,
  FolderKanban,
  Database,
  Activity as ActivityIcon,
  Container,
  Lightbulb,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useTheme } from '@/components/theme/ThemeProvider';
import { useVersion } from '@/hooks/useCloudOS';

/* ── Types ────────────────────────────────────────────────── */
interface NavItem {
  label: string;
  to: string;
  icon: React.ElementType;
}

interface NavGroup {
  items: NavItem[];
}

interface SidebarProps {
  open: boolean;
  onToggle: () => void;
  className?: string;
}

/* ── Navigation items (matching design system spec) ──────── */
const NAV_GROUPS: NavGroup[] = [
  {
    items: [
      { label: 'Applications', to: '/', icon: Box },
      { label: 'Deployments', to: '/deployments', icon: GitBranch },
    ],
  },
  {
    items: [
      { label: 'Monitoring', to: '/monitoring', icon: Heart },
      { label: 'Workflows', to: '/workflows', icon: GitMerge },
    ],
  },
  {
    items: [
      { label: 'Projects', to: '/projects', icon: FolderKanban },
    ],
  },
  {
    items: [
      { label: 'Resources', to: '/resources', icon: Database },
      { label: 'Controllers', to: '/controllers', icon: Container },
      { label: 'Providers', to: '/providers', icon: ActivityIcon },
      { label: 'Capabilities', to: '/capabilities', icon: Lightbulb },
    ],
  },
  {
    items: [
      { label: 'Kernel', to: '/kernel', icon: Cpu },
      { label: 'System', to: '/system', icon: Monitor },
      { label: 'Plugins', to: '/plugins', icon: Puzzle },
    ],
  },
  {
    items: [
      { label: 'Settings', to: '/settings', icon: Settings },
    ],
  },
];

/* ── Sidebar Nav Item ─────────────────────────────────────── */
function SidebarNavItem({
  item,
  collapsed,
}: {
  item: NavItem;
  collapsed: boolean;
}) {
  return (
    <div className={cn(collapsed && 'group relative flex justify-center')}>
      <NavLink
        to={item.to}
        end={item.to === '/'}
        className={({ isActive }) =>
          cn(
            'relative flex items-center rounded-sm transition-all duration-100',
            collapsed
              ? 'justify-center mx-auto w-9 h-9'
              : 'gap-2 px-4 py-2',
            'text-sidebar text-text-secondary hover:text-foreground',
            isActive && 'bg-accent-subtle text-accent',
          )
        }
        aria-label={collapsed ? item.label : undefined}
      >
        <item.icon className="h-4 w-4 shrink-0" />
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
      </NavLink>

      {/* Tooltip for collapsed mode */}
      {collapsed && (
        <div className="pointer-events-none absolute left-full top-1/2 z-50 ml-2 -translate-y-1/2 whitespace-nowrap rounded-sm border border-border bg-surface-elevated px-2.5 py-1 text-caption text-foreground shadow-md opacity-0 transition-opacity duration-150 group-hover:opacity-100">
          {item.label}
        </div>
      )}
    </div>
  );
}

/* ── Sidebar Root ─────────────────────────────────────────── */
export function Sidebar({ open, onToggle, className }: SidebarProps) {
  const { theme, setTheme } = useTheme();
  const { data: version } = useVersion();

  const cycleTheme = () => {
    const themes: Array<'dark' | 'light' | 'system'> = ['dark', 'light', 'system'];
    const idx = themes.indexOf(theme);
    setTheme(themes[(idx + 1) % themes.length]);
  };

  const ThemeIcon = theme === 'dark' ? Moon : theme === 'light' ? Sun : Monitor;
  const buildVersion = version?.number ?? version?.build?.version ?? '0.6.0-rc';

  return (
    <motion.aside
      initial={false}
      animate={{ width: open ? 240 : 56 }}
      transition={{ duration: 0.2, ease: 'easeInOut' }}
      className={cn(
        'fixed left-0 top-12 z-30 flex flex-col border-r border-border',
        'h-[calc(100vh-3rem-1.75rem)]', // viewport minus TopNav + StatusBar
        'bg-sidebar-bg',
        className,
      )}
      aria-label="Main navigation"
    >
      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto overflow-x-hidden py-4 scrollbar-none">
        {NAV_GROUPS.map((group, gIdx) => (
          <div key={gIdx}>
            {gIdx > 0 && (
              <div className="mx-4 my-3 h-px bg-border" aria-hidden="true" />
            )}
            <div className={cn('flex flex-col', open ? 'gap-0.5' : 'gap-1')}>
              {group.items.map((item) => (
                <SidebarNavItem
                  key={item.to}
                  item={item}
                  collapsed={!open}
                />
              ))}
            </div>
          </div>
        ))}
      </nav>

      {/* Bottom section — theme toggle + collapse + version */}
      <div
        className={cn(
          'flex shrink-0 items-center border-t border-border px-3 py-2.5',
          open ? 'justify-between gap-2' : 'flex-col gap-2',
        )}
      >
        {/* Theme toggle */}
        <button
          type="button"
          onClick={cycleTheme}
          className={cn(
            'flex items-center justify-center rounded-sm p-1.5',
            'text-text-muted hover:text-foreground hover:bg-accent-subtle',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
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
            'flex items-center justify-center rounded-sm p-1.5',
            'text-text-muted hover:text-foreground hover:bg-accent-subtle',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
          )}
          aria-label={open ? 'Collapse sidebar' : 'Expand sidebar'}
        >
          {open ? (
            <ChevronLeft className="h-3.5 w-3.5" />
          ) : (
            <ChevronRight className="h-3.5 w-3.5" />
          )}
        </button>

        {/* Version badge — only when expanded */}
        {open && (
          <span className="text-caption text-text-muted tabular-nums">
            v{buildVersion}
          </span>
        )}
      </div>
    </motion.aside>
  );
}
