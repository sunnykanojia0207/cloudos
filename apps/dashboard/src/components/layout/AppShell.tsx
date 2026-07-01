'use client';

import { useState } from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { TopNav } from './TopNav';
import { Sidebar } from './Sidebar';
import { StatusBar } from './StatusBar';
import { CommandPaletteProvider } from './CommandPalette';
import { OfflineBanner } from './OfflineBanner';
import { ToastProvider } from '@/components/ui/toast';
import { useKeyboardShortcuts } from '@/hooks/useKeyboardShortcuts';
import { cn } from '@/lib/utils';

/* ── Persisted sidebar state ──────────────────────────────── */
const STORAGE_KEY = 'sidebar-collapsed';

function useSidebarState(): [boolean, () => void] {
  const [open, setOpen] = useState(() => {
    if (typeof window === 'undefined') return true;
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored ? stored === 'true' : true;
  });

  const toggle = () => {
    setOpen((prev) => {
      const next = !prev;
      try {
        localStorage.setItem(STORAGE_KEY, String(next));
      } catch {
        // localStorage unavailable
      }
      return next;
    });
  };

  return [open, toggle];
}

/* ── Page transition variants ─────────────────────────────── */
const pageVariants = {
  initial: { opacity: 0, y: 8 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -4 },
};

const pageTransition = {
  duration: 0.2,
  ease: [0, 0, 0.2, 1] as [number, number, number, number],
};

/* ── Skip-to-content link ─────────────────────────────────── */
function SkipToContent() {
  return (
    <a
      href="#main-content"
      className="sr-only focus:not-sr-only focus:fixed focus:top-4 focus:left-4 focus:z-[100] focus:px-4 focus:py-2 focus:bg-surface-elevated focus:text-foreground focus:rounded-md focus:shadow-lg focus:outline-none focus:ring-2 focus:ring-ring"
    >
      Skip to content
    </a>
  );
}

/* ── AppShell ─────────────────────────────────────────────── */
export function AppShell() {
  const [sidebarOpen, toggleSidebar] = useSidebarState();
  const location = useLocation();

  // Register global keyboard shortcuts (g→page, / → search)
  useKeyboardShortcuts();

  return (
    <CommandPaletteProvider>
      <ToastProvider>
        <SkipToContent />
        <div className="flex min-h-screen flex-col bg-background text-foreground">
          {/* ── Top Navigation (fixed, h-12) ──────────── */}
          <TopNav
            onToggleSidebar={toggleSidebar}
            sidebarOpen={sidebarOpen}
          />

          {/* ── Body ──────────────────────────────────── */}
          <div className="flex flex-1 pt-12">
            {/* Sidebar (fixed) */}
            <Sidebar open={sidebarOpen} onToggle={toggleSidebar} />

            {/* Main content area */}
            <AnimatePresence mode="wait">
              <motion.main
                key={location.pathname}
                variants={pageVariants}
                initial="initial"
                animate="animate"
                exit="exit"
                transition={pageTransition}
                className={cn(
                  'flex-1 overflow-auto transition-[margin] duration-normal ease-standard',
                  sidebarOpen ? 'ml-60' : 'ml-14',
                  'pb-7', // space for StatusBar (h-7)
                )}
              >
                <OfflineBanner />
                <div
                  id="main-content"
                  className="mx-auto w-full max-w-page px-6 pt-8 pb-6"
                >
                  <Outlet />
                </div>
              </motion.main>
            </AnimatePresence>
          </div>

          {/* ── StatusBar (fixed, h-7) ────────────────── */}
          <StatusBar />
        </div>
      </ToastProvider>
    </CommandPaletteProvider>
  );
}
