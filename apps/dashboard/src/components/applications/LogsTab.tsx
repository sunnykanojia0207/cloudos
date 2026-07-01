import * as React from 'react';
import { useApplicationLogs, type LogEvent } from '@/hooks/useApplicationLogs';
import { cn } from '@/lib/utils';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Pause,
  Play,
  Download,
  Trash2,
  Search,
  Eye,
  EyeOff,
  Terminal,
} from 'lucide-react';

/* ── Helpers ──────────────────────────────────────────── */

const levelColors: Record<string, string> = {
  ERROR: 'text-red-400',
  WARN: 'text-amber-400',
  INFO: 'text-foreground/90',
  DEBUG: 'text-muted-foreground/60',
  TRACE: 'text-muted-foreground/40',
};

function formatTimestamp(ts: string): string {
  if (!ts) return '';
  try {
    return new Date(ts).toLocaleTimeString(undefined, {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      fractionalSecondDigits: 3,
    });
  } catch {
    return ts;
  }
}

function matchesSearch(log: LogEvent, searchTerm: string): boolean {
  if (!searchTerm) return true;
  const term = searchTerm.toLowerCase();
  return (
    log.message.toLowerCase().includes(term) ||
    log.level.toLowerCase().includes(term) ||
    log.source.toLowerCase().includes(term) ||
    (log.step ?? '').toLowerCase().includes(term)
  );
}

/* ── Props ────────────────────────────────────────────── */

export interface LogsTabProps {
  appId: string;
}

/* ── Component ────────────────────────────────────────── */

export function LogsTab({ appId }: LogsTabProps) {
  const {
    logs,
    connected,
    error,
    paused,
    pause,
    resume,
    clearLogs,
    downloadLogs,
  } = useApplicationLogs(appId);

  const [searchTerm, setSearchTerm] = React.useState('');
  const [showTimestamps, setShowTimestamps] = React.useState(true);
  const scrollRef = React.useRef<HTMLDivElement>(null);
  const autoScrollRef = React.useRef(true);
  const prevLogsLengthRef = React.useRef(0);

  // Track whether user has scrolled up
  const handleScroll = React.useCallback(() => {
    const el = scrollRef.current;
    if (!el) return;

    const threshold = 50; // px from bottom to consider "at bottom"
    const isAtBottom =
      el.scrollHeight - el.scrollTop - el.clientHeight < threshold;
    autoScrollRef.current = isAtBottom;
  }, []);

  // Auto-scroll to bottom when new logs arrive (if user hasn't scrolled up)
  React.useEffect(() => {
    if (!autoScrollRef.current) return;
    const el = scrollRef.current;
    if (el) {
      el.scrollTop = el.scrollHeight;
    }
  }, [logs.length]);

  // Keep track of previous length for auto-scroll trigger
  React.useEffect(() => {
    prevLogsLengthRef.current = logs.length;
  }, [logs.length]);

  // Filter logs by search term
  const filteredLogs = React.useMemo(() => {
    if (!searchTerm) return logs;
    return logs.filter((log) => matchesSearch(log, searchTerm));
  }, [logs, searchTerm]);

  // ── Load state ──
  const isLoading = logs.length === 0 && connected === false && error === null;

  if (isLoading) {
    return (
      <div className="space-y-3">
        <div className="flex items-center gap-2">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-8 w-24" />
        </div>
        <Skeleton className="h-[400px] w-full rounded-lg" />
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {/* ── Toolbar ── */}
      <div className="flex flex-wrap items-center gap-2">
        {/* Search */}
        <div className="relative flex-1 min-w-[160px] max-w-sm">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground/60 pointer-events-none" />
          <Input
            type="text"
            placeholder="Filter logs..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="h-9 pl-8 text-sm"
            aria-label="Search log entries"
          />
        </div>

        {/* Connection status */}
        <Badge
          variant="outline"
          className={cn(
            'gap-1.5 select-none',
            connected
              ? 'text-emerald-400 border-emerald-500/30'
              : 'text-red-400 border-red-500/30',
          )}
        >
          <span
            className={cn(
              'h-2 w-2 rounded-full',
              connected ? 'bg-emerald-400' : 'bg-red-400',
            )}
            aria-hidden="true"
          />
          {connected ? 'Connected' : 'Disconnected'}
        </Badge>

        <div className="flex-1" />

        {/* Toggle timestamps */}
        <Button
          variant="outline"
          size="sm"
          className="h-9 gap-1.5 text-xs"
          onClick={() => setShowTimestamps(!showTimestamps)}
          aria-label={showTimestamps ? 'Hide timestamps' : 'Show timestamps'}
        >
          {showTimestamps ? (
            <EyeOff className="h-3.5 w-3.5" />
          ) : (
            <Eye className="h-3.5 w-3.5" />
          )}
          Timestamps
        </Button>

        {/* Pause / Resume */}
        <Button
          variant="outline"
          size="sm"
          className="h-9 gap-1.5 text-xs"
          onClick={paused ? resume : pause}
          aria-label={paused ? 'Resume log stream' : 'Pause log stream'}
        >
          {paused ? (
            <Play className="h-3.5 w-3.5" />
          ) : (
            <Pause className="h-3.5 w-3.5" />
          )}
          {paused ? 'Resume' : 'Pause'}
        </Button>

        {/* Download */}
        <Button
          variant="outline"
          size="sm"
          className="h-9 gap-1.5 text-xs"
          onClick={downloadLogs}
          aria-label="Download logs"
        >
          <Download className="h-3.5 w-3.5" />
          Download
        </Button>

        {/* Clear */}
        <Button
          variant="outline"
          size="sm"
          className="h-9 gap-1.5 text-xs"
          onClick={clearLogs}
          aria-label="Clear logs"
        >
          <Trash2 className="h-3.5 w-3.5" />
          Clear
        </Button>
      </div>

      {/* ── Error banner ── */}
      {error && (
        <div className="rounded-md border border-red-500/30 bg-red-950/20 px-3 py-2 text-xs text-red-400">
          {error}
        </div>
      )}

      {/* ── Log content area ── */}
      <div className="relative rounded-lg border border-border/50 bg-gray-950 overflow-hidden">
        {paused && (
          <div className="absolute top-2 right-3 z-10">
            <Badge variant="warning" className="text-[10px] px-2 py-0.5 animate-pulse">
              PAUSED
            </Badge>
          </div>
        )}

        <ScrollArea
          ref={scrollRef}
          onScroll={handleScroll}
          className="h-[500px] max-h-[60vh]"
        >
          {filteredLogs.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full py-16 text-center">
              <Terminal className="h-8 w-8 text-muted-foreground/30 mb-2" />
              <p className="text-sm text-muted-foreground/70">
                {searchTerm
                  ? 'No log entries match your search filter.'
                  : 'No log entries yet. Waiting for log output...'}
              </p>
            </div>
          ) : (
            <div className="p-3 font-mono text-xs leading-relaxed space-y-0.5">
              {filteredLogs.map((log, idx) => (
                <div
                  key={`${log.timestamp}-${idx}`}
                  className="flex items-start gap-2 hover:bg-white/[0.02] rounded px-1 py-0.5 transition-colors"
                >
                  {/* Timestamp */}
                  {showTimestamps && (
                    <span className="shrink-0 text-muted-foreground/50 w-[86px] select-none">
                      {formatTimestamp(log.timestamp)}
                    </span>
                  )}

                  {/* Level badge */}
                  <span
                    className={cn(
                      'shrink-0 w-10 text-[10px] font-semibold uppercase select-none',
                      levelColors[log.level] ?? 'text-foreground/70',
                    )}
                  >
                    {log.level}
                  </span>

                  {/* Source */}
                  {log.source && (
                    <span className="shrink-0 text-muted-foreground/50 max-w-[120px] truncate select-none">
                      [{log.source}]
                    </span>
                  )}

                  {/* Step */}
                  {log.step && (
                    <span className="shrink-0 text-muted-foreground/40 max-w-[80px] truncate select-none">
                      ({log.step})
                    </span>
                  )}

                  {/* Message */}
                  <span
                    className={cn(
                      'flex-1 break-words min-w-0',
                      log.level === 'ERROR'
                        ? 'text-red-300'
                        : log.level === 'WARN'
                          ? 'text-amber-300'
                          : log.level === 'DEBUG'
                            ? 'text-muted-foreground/60'
                            : 'text-foreground/80',
                    )}
                  >
                    {log.message}
                  </span>
                </div>
              ))}
            </div>
          )}
        </ScrollArea>
      </div>

      {/* Log count */}
      <div className="flex items-center justify-between text-xs text-muted-foreground/60">
        <span>
          {filteredLogs.length} {filteredLogs.length === 1 ? 'entry' : 'entries'}
          {searchTerm && logs.length !== filteredLogs.length
            ? ` (filtered from ${logs.length})`
            : ''}
        </span>
        <span className="font-mono text-[10px]">{appId}</span>
      </div>
    </div>
  );
}
