import * as React from 'react';

/**
 * Wrap matched characters in a `<mark>` element.
 * Returns a ReactNode fragment.
 */
export function highlightMatch(text: string, indices: number[]): React.ReactNode {
  if (!indices.length) return text;

  const set = new Set(indices);
  const parts: React.ReactNode[] = [];
  let i = 0;
  while (i < text.length) {
    if (set.has(i)) {
      let j = i;
      while (j < text.length && set.has(j)) j++;
      parts.push(
        <mark key={i} className="rounded-sm bg-accent/40 text-foreground font-medium">
          {text.slice(i, j)}
        </mark>,
      );
      i = j;
    } else {
      let j = i;
      while (j < text.length && !set.has(j)) j++;
      parts.push(<span key={i}>{text.slice(i, j)}</span>);
      i = j;
    }
  }
  return <>{parts}</>;
}
