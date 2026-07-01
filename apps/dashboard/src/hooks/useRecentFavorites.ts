'use client';

import { useState, useCallback, useEffect, useRef } from 'react';

/* ── Types ──────────────────────────────────────────────── */
export interface RecentItem {
  id: string;
  type: string;
  label: string;
  subtitle: string;
  path: string;
  icon: string; // lucide icon name
  timestamp: number;
}

export interface FavoriteItem {
  id: string;
  type: string;
}

const RECENT_KEY = 'cloudos-recent';
const FAVORITES_KEY = 'cloudos-favorites';
const MAX_RECENT = 20;

/* ── Low-level storage helpers ──────────────────────────── */

function readRecent(): RecentItem[] {
  try {
    const raw = localStorage.getItem(RECENT_KEY);
    return raw ? (JSON.parse(raw) as RecentItem[]) : [];
  } catch {
    return [];
  }
}

function writeRecent(items: RecentItem[]) {
  try {
    localStorage.setItem(RECENT_KEY, JSON.stringify(items));
  } catch {
    // storage full or unavailable
  }
}

function readFavorites(): FavoriteItem[] {
  try {
    const raw = localStorage.getItem(FAVORITES_KEY);
    return raw ? (JSON.parse(raw) as FavoriteItem[]) : [];
  } catch {
    return [];
  }
}

function writeFavorites(items: FavoriteItem[]) {
  try {
    localStorage.setItem(FAVORITES_KEY, JSON.stringify(items));
  } catch {
    // storage full or unavailable
  }
}

/* ── Recent Items Hook ──────────────────────────────────── */
export function useRecentItems() {
  const [items, setItems] = useState<RecentItem[]>(readRecent);

  // Sync across tabs if localStorage changes
  useEffect(() => {
    const handler = (e: StorageEvent) => {
      if (e.key === RECENT_KEY) setItems(readRecent());
    };
    window.addEventListener('storage', handler);
    return () => window.removeEventListener('storage', handler);
  }, []);

  const pushItem = useCallback((item: Omit<RecentItem, 'timestamp'>) => {
    setItems((prev) => {
      const filtered = prev.filter((r) => !(r.id === item.id && r.type === item.type));
      const next = [{ ...item, timestamp: Date.now() }, ...filtered].slice(0, MAX_RECENT);
      writeRecent(next);
      return next;
    });
  }, []);

  const removeItem = useCallback((id: string, type: string) => {
    setItems((prev) => {
      const next = prev.filter((r) => !(r.id === id && r.type === type));
      writeRecent(next);
      return next;
    });
  }, []);

  const clearAll = useCallback(() => {
    setItems([]);
    writeRecent([]);
  }, []);

  return { items, pushItem, removeItem, clearAll };
}

/* ── Favorites Hook ─────────────────────────────────────── */
export function useFavorites() {
  const [items, setItems] = useState<FavoriteItem[]>(readFavorites);

  useEffect(() => {
    const handler = (e: StorageEvent) => {
      if (e.key === FAVORITES_KEY) setItems(readFavorites());
    };
    window.addEventListener('storage', handler);
    return () => window.removeEventListener('storage', handler);
  }, []);

  const isFavorite = useCallback(
    (id: string, type: string) => items.some((f) => f.id === id && f.type === type),
    [items],
  );

  const toggleFavorite = useCallback((id: string, type: string) => {
    setItems((prev) => {
      const exists = prev.some((f) => f.id === id && f.type === type);
      const next = exists
        ? prev.filter((f) => !(f.id === id && f.type === type))
        : [...prev, { id, type }];
      writeFavorites(next);
      return next;
    });
  }, []);

  return { items, isFavorite, toggleFavorite };
}
