'use client';

import { useState } from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { TopNav } from './TopNav';
import { Sidebar } from './Sidebar';
import { StatusBar } from './StatusBar';
import { CommandPaletteProvider } from './CommandPalette';
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
  ease: 'easeOut' as const,
};

/* ── AppShell ─────────────────────────────────────────────── */
export function AppShell() {
  const [sidebarOpen, toggleSidebar] = useSidebarState();
  const location = useLocation();

  return (
    <CommandPaletteProvider>
      <div className="flex min-h-screen flex-col bg-background text-foreground">
        {/* ── Top Navigation ─────────────────────── */}
        <TopNav
          onToggleSidebar={toggleSidebar}
          sidebarOpen={sidebarOpen}
        />

        {/* ── Body ───────────────────────────────── */}
        <div className="flex flex-1 pt-12">
          {/* Sidebar */}
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
                'flex-1 overflow-auto transition-[margin] duration-200 ease-in-out',
                sidebarOpen ? 'ml-60' : 'ml-14',
                'pb-7', // space for StatusBar
              )}
            >
              <div className="mx-auto w-full max-w-7xl px-6 py-6">
                <Outlet />
              </div>
            </motion.main>
          </AnimatePresence>
        </div>

        {/* ── Status Bar ─────────────────────────── */}
        <StatusBar />
      </div>
    </CommandPaletteProvider>
  );
}
