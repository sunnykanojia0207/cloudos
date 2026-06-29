import { useLocation, Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { ChevronRight } from 'lucide-react';
import { cn } from '@/lib/utils';

/* ── Label mapping ────────────────────────────────────────── */
const LABEL_MAP: Record<string, string> = {
  '': 'Home',
  dashboard: 'Dashboard',
  system: 'System',
  kernel: 'Kernel',
  capabilities: 'Capabilities',
  providers: 'Providers',
  resources: 'Resources',
  controllers: 'Controllers',
  plugins: 'Plugins',
  settings: 'Settings',
  projects: 'Projects',
};

function segmentLabel(seg: string): string {
  return LABEL_MAP[seg.toLowerCase()] ?? seg.charAt(0).toUpperCase() + seg.slice(1);
}

/* ── Component ────────────────────────────────────────────── */
export function Breadcrumbs() {
  const { pathname } = useLocation();
  const segments = pathname.split('/').filter(Boolean);

  const crumbs: { label: string; href: string }[] = [];

  // Root segment
  crumbs.push({ label: 'Home', href: '/' });

  // Path segments
  let cumulative = '';
  for (const seg of segments) {
    cumulative += `/${seg}`;
    crumbs.push({ label: segmentLabel(seg), href: cumulative });
  }

  return (
    <motion.nav
      initial={{ opacity: 0, x: -6 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ duration: 0.2, ease: 'easeOut' }}
      aria-label="Breadcrumb"
    >
      <ol className="flex items-center gap-1 text-xs text-muted-foreground/70">
        {crumbs.map((crumb, idx) => {
          const isLast = idx === crumbs.length - 1;

          return (
            <li key={crumb.href} className="flex items-center gap-1">
              {idx > 0 && (
                <ChevronRight className="h-3 w-3 shrink-0 text-muted-foreground/40" aria-hidden="true" />
              )}

              {isLast ? (
                <span
                  className={cn(
                    'rounded px-1.5 py-0.5 text-xs font-medium',
                    idx === 0
                      ? 'text-foreground/80'
                      : 'text-foreground/70',
                  )}
                  aria-current="page"
                >
                  {crumb.label}
                </span>
              ) : (
                <Link
                  to={crumb.href}
                  className={cn(
                    'rounded px-1.5 py-0.5 text-xs transition-colors duration-150',
                    'hover:bg-muted/60 hover:text-foreground/80',
                  )}
                >
                  {crumb.label}
                </Link>
              )}
            </li>
          );
        })}
      </ol>
    </motion.nav>
  );
}
