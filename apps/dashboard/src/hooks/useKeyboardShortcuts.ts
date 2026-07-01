import { useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';

type ShortcutMap = Record<string, { handler: () => void; description: string }>;

/**
 * Global keyboard shortcuts hook.
 *
 * Supports chained shortcuts like "g a" (press g, then a within 800ms).
 * Also supports single-key shortcuts like "/", "Escape".
 */
export function useKeyboardShortcuts(
  extraShortcuts?: ShortcutMap,
) {
  const navigate = useNavigate();

  // ── Chained shortcut state ──
  useEffect(() => {
    let chainBuffer = '';
    let chainTimer: ReturnType<typeof setTimeout> | null = null;

    const CHAIN_TIMEOUT = 800; // ms between key presses
    const CHAIN_MAP: Record<string, () => void> = {
      ga: () => navigate('/'),
      gd: () => navigate('/deployments'),
      gw: () => navigate('/workflows'),
      gm: () => navigate('/monitoring'),
      gs: () => navigate('/system'),
      gp: () => navigate('/projects'),
      gt: () => navigate('/settings'),
    };

    const handler = (e: globalThis.KeyboardEvent) => {
      const key = e.key.toLowerCase();

      // Ignore if user is typing in an input/textarea/select
      const target = e.target as HTMLElement;
      const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.tagName === 'SELECT' || target.isContentEditable;
      if (isInput) return;

      // Handle Escape → close all
      if (key === 'escape') {
        // This will be handled by any open dialogs via their own handlers
        return;
      }

      // Handle "/" → focus search
      if (key === '/') {
        e.preventDefault();
        // Dispatch a custom event that TopNav listens to
        window.dispatchEvent(new CustomEvent('cloudos:focus-search'));
        return;
      }

      // Chain handling: "g" followed by another key
      if (chainBuffer === 'g' && key.length === 1 && key >= 'a' && key <= 'z') {
        e.preventDefault();
        const combo = `g${key}`;
        if (CHAIN_MAP[combo]) {
          CHAIN_MAP[combo]();
        }
        chainBuffer = '';
        if (chainTimer) clearTimeout(chainTimer);
        chainTimer = null;
        return;
      }

      if (key === 'g') {
        e.preventDefault();
        chainBuffer = 'g';
        if (chainTimer) clearTimeout(chainTimer);
        chainTimer = setTimeout(() => {
          chainBuffer = '';
        }, CHAIN_TIMEOUT);
        return;
      }

      // Not a chain start — reset
      if (chainBuffer) {
        chainBuffer = '';
        if (chainTimer) clearTimeout(chainTimer);
        chainTimer = null;
      }

      // Extra shortcuts
      if (extraShortcuts && extraShortcuts[key]) {
        e.preventDefault();
        extraShortcuts[key].handler();
      }
    };

    document.addEventListener('keydown', handler);
    return () => {
      document.removeEventListener('keydown', handler);
      if (chainTimer) clearTimeout(chainTimer);
    };
  }, [navigate, extraShortcuts]);
}
