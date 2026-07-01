'use client';

import { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Wifi, WifiOff, RotateCcw } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';

/**
 * Monitors API connectivity and shows a global banner when offline.
 * Auto-retries every 10 seconds.
 */
export function useOnlineStatus() {
  const [isOnline, setIsOnline] = useState(true);
  const [lastChecked, setLastChecked] = useState<Date>(new Date());

  const check = useCallback(async () => {
    try {
      const BASE = import.meta.env.VITE_CLOUDOS_API_URL ?? '';
      const controller = new AbortController();
      const id = setTimeout(() => controller.abort(), 5000);
      const res = await fetch(`${BASE}/api/v1/ready`, {
        signal: controller.signal,
      });
      clearTimeout(id);
      setIsOnline(res.ok);
    } catch {
      setIsOnline(false);
    }
    setLastChecked(new Date());
  }, []);

  // Check every 10 seconds
  useEffect(() => {
    check();
    const interval = setInterval(check, 10_000);
    return () => clearInterval(interval);
  }, [check]);

  return { isOnline, lastChecked, retry: check };
}

export function OfflineBanner() {
  const { isOnline, retry } = useOnlineStatus();
  const [dismissed, setDismissed] = useState(false);

  // Re-show if status changes back to offline
  useEffect(() => {
    if (!isOnline) setDismissed(false);
  }, [isOnline]);

  return (
    <AnimatePresence>
      {!isOnline && !dismissed && (
        <motion.div
          initial={{ height: 0, opacity: 0 }}
          animate={{ height: 'auto', opacity: 1 }}
          exit={{ height: 0, opacity: 0 }}
          transition={{ duration: 0.2 }}
          className="overflow-hidden"
          role="alert"
          aria-live="assertive"
        >
          <div className="flex items-center gap-3 bg-danger px-4 py-1.5 text-small text-danger-foreground">
            <WifiOff className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
            <span className="flex-1 font-medium">
              Connection lost. Retrying automatically...
            </span>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => retry()}
              className="h-6 gap-1 text-caption text-danger-foreground hover:bg-danger/20"
            >
              <RotateCcw className="h-3 w-3" />
              Retry
            </Button>
            <button
              type="button"
              onClick={() => setDismissed(true)}
              className="text-danger-foreground/70 hover:text-danger-foreground p-0.5 rounded-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              aria-label="Dismiss offline warning"
            >
              <Wifi className="h-3 w-3" />
            </button>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
