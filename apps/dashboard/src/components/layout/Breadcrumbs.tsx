import { useLocation, Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { ChevronRight } from 'lucide-react';
import { cn } from '@/lib/utils';

/* ── Label mapping ────────────────────────────────────────── */
const LABEL_MAP: Record<string, string> = {
  '': 'Home',
  applications: 'Applications',
  deployments: 'Deployments',
  projects: 'Projects',
  monitoring: 'Monitoring',
  workflows: 'Workflows',
  system: 'System',
  kernel: 'Kernel',
  capabilities: 'Capabilities',
  providers: 'Providers',
  resources: 'Resources',
  controllers: 'Controllers',
  plugins: 'Plugins',
  settings: 'Settings',
  overview: 'Overview',
  timeline: 'Timeline',
  logs: 'Logs',
  general: 'General',
  runtimes: 'Runtimes',
  buildpacks: 'Buildpacks',
  installed: 'Installed',
  catalog: 'Catalog',
  health: 'Health',
  dashboard: 'Dashboard',
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
      transition={{ duration: 0.2, ease: [0, 0, 0.2, 1] }}
      aria-label="Breadcrumb"
    >
      <ol className="flex items-center gap-1 text-small text-text-muted">
        {crumbs.map((crumb, idx) => {
          const isLast = idx === crumbs.length - 1;

          return (
            <li key={crumb.href} className="flex items-center gap-1">
              {idx > 0 && (
                <ChevronRight className="h-3 w-3 shrink-0 text-text-muted/50" aria-hidden="true" />
              )}

              {isLast ? (
                <span
                  className="rounded px-1.5 py-0.5 text-small font-medium text-foreground"
                  aria-current="page"
                >
                  {crumb.label}
                </span>
              ) : (
                <Link
                  to={crumb.href}
                  className="rounded px-1.5 py-0.5 text-small transition-colors duration-100 hover:text-foreground hover:bg-accent-subtle"
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
