import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatUptime(uptime: string): string {
  if (!uptime) return '\u2014';
  return uptime;
}

export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export function truncate(str: string, maxLength: number): string {
  if (str.length <= maxLength) return str;
  return str.slice(0, maxLength) + '\u2026';
}

export function relativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const seconds = Math.floor(diff / 1000);
  if (seconds < 5) return 'just now';
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}d ago`;
  return date.toLocaleDateString();
}

export function formatNumber(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return n.toLocaleString();
}

/* ── Fuzzy matching ────────────────────────────────────────── */
export interface FuzzyMatchResult {
  score: number;
  indices: number[]; // character indices that matched
}

/**
 * Score how well `query` matches `text`.
 *
 * Returns null if no match, or an object with:
 * - `score`: higher is better (prefix > contiguous > fuzzy)
 * - `indices`: character positions in `text` that matched
 */
export function fuzzyMatch(text: string, query: string): FuzzyMatchResult | null {
  if (!query) return null;
  const tl = text.toLowerCase();
  const ql = query.toLowerCase();
  const tLen = text.length;
  const qLen = query.length;

  if (qLen > tLen) return null;

  // Exact match → max score
  if (tl === ql) return { score: 1000, indices: Array.from({ length: tLen }, (_, i) => i) };

  // Prefix match → high score
  if (tl.startsWith(ql)) return { score: 900, indices: Array.from({ length: qLen }, (_, i) => i) };

  // Contains substring → medium-high score
  const idx = tl.indexOf(ql);
  if (idx !== -1) return { score: 700 - idx, indices: Array.from({ length: qLen }, (_, i) => idx + i) };

  // Fuzzy (character-skip) match
  let qi = 0;
  const indices: number[] = [];
  for (let ti = 0; ti < tLen && qi < qLen; ti++) {
    if (tl[ti] === ql[qi]) {
      indices.push(ti);
      qi++;
    }
  }
  if (qi === qLen) {
    // Score based on how spread out the matches are (closer = better)
    const spread = indices[indices.length - 1] - indices[0] + 1;
    const ideal = qLen;
    const penalty = spread - ideal;
    const score = Math.max(300 - penalty, 1);
    return { score, indices };
  }

  return null;
}

/**
 * highlightMatch is in search-utils.tsx (needs JSX).
 */
