import * as React from 'react';
import { useApplicationLogs, type LogEvent } from '@/hooks/useApplicationLogs';
import { cn } from '@/lib/utils';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { TerminalStatus } from '@/components/ui/terminal';
import {
  Pause,
  Play,
  Download,
  Trash2,
  Search,
  Eye,
  EyeOff,
  Terminal as TerminalIcon,
} from 'lucide-react';

/* ── Helpers ──────────────────────────────────────────── */

const levelColors: Record<string, string> = {
  ERROR: 'text-danger',
  WARN: 'text-warning',
  INFO: 'text-foreground/90',
  DEBUG: 'text-text-muted/60',
  TRACE: 'text-text-muted/40',
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

export interface LogViewerProps {
  appId: string;
  footerLabel?: string;
}

/* ── Component ────────────────────────────────────────── */

export function LogViewer({ appId, footerLabel }: LogViewerProps) {
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

  const handleScroll = React.useCallback(() => {
    const el = scrollRef.current;
    if (!el) return;
    const threshold = 50;
    const isAtBottom = el.scrollHeight - el.scrollTop - el.clientHeight < threshold;
    autoScrollRef.current = isAtBottom;
  }, []);

  React.useEffect(() => {
    if (!autoScrollRef.current) return;
    const el = scrollRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [logs.length]);

  const filteredLogs = React.useMemo(() => {
    if (!searchTerm) return logs;
    return logs.filter((log) => matchesSearch(log, searchTerm));
  }, [logs, searchTerm]);

  const isLoading = logs.length === 0 && connected === false && error === null;

  if (isLoading) {
    return (
      <div className="space-y-3">
        <div className="flex items-center gap-2">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-8 w-24" />
        </div>
        <Skeleton className="h-[400px] w-full rounded-md" />
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {/* Toolbar */}
      <div className="flex flex-wrap items-center gap-2">
        <div className="relative flex-1 min-w-[160px] max-w-sm">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-text-muted pointer-events-none" />
          <Input
            type="text"
            placeholder="Filter logs..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="h-8 pl-8 text-small"
            aria-label="Search log entries"
          />
        </div>

        <Badge
          variant={connected ? 'subtle-success' : 'subtle-danger'}
          className="gap-1.5 select-none text-caption"
        >
          <span className={cn('h-1.5 w-1.5 rounded-full', connected ? 'bg-success' : 'bg-danger')} aria-hidden="true" />
          {connected ? 'Connected' : 'Disconnected'}
        </Badge>

        <div className="flex-1" />

        <Button variant="secondary" size="sm" className="h-8 gap-1.5 text-small"
          onClick={() => setShowTimestamps(!showTimestamps)}
          aria-label={showTimestamps ? 'Hide timestamps' : 'Show timestamps'}>
          {showTimestamps ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
          Time
        </Button>

        <Button variant="secondary" size="sm" className="h-8 gap-1.5 text-small"
          onClick={paused ? resume : pause}
          aria-label={paused ? 'Resume log stream' : 'Pause log stream'}>
          {paused ? <Play className="h-3.5 w-3.5" /> : <Pause className="h-3.5 w-3.5" />}
          {paused ? 'Resume' : 'Pause'}
        </Button>

        <Button variant="secondary" size="sm" className="h-8 gap-1.5 text-small"
          onClick={downloadLogs} aria-label="Download logs">
          <Download className="h-3.5 w-3.5" />
        </Button>

        <Button variant="secondary" size="sm" className="h-8 gap-1.5 text-small"
          onClick={clearLogs} aria-label="Clear logs">
          <Trash2 className="h-3.5 w-3.5" />
        </Button>
      </div>

      {/* Error banner */}
      {error && (
        <div className="rounded-md border border-danger/30 bg-danger-subtle px-3 py-2 text-small text-danger">
          {error}
        </div>
      )}

      {/* Log content */}
      <div className="relative rounded-md border border-border bg-terminal overflow-hidden">
        {paused && (
          <div className="absolute top-2 right-3 z-10">
            <Badge variant="subtle-warning" className="text-caption px-2 py-0.5 animate-pulse">PAUSED</Badge>
          </div>
        )}

        <ScrollArea ref={scrollRef} onScroll={handleScroll} className="h-[500px] max-h-[60vh]">
          {filteredLogs.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full py-16 text-center">
              <TerminalIcon className="h-8 w-8 text-text-muted/30 mb-2" />
              <p className="text-small text-text-muted">
                {searchTerm ? 'No log entries match your search filter.' : 'No log entries yet. Waiting for log output...'}
              </p>
            </div>
          ) : (
            <div className="p-3 font-mono text-code leading-relaxed space-y-0.5">
              {filteredLogs.map((log, idx) => (
                <div key={`${log.timestamp}-${idx}`}
                  className="flex items-start gap-2 hover:bg-white/[0.02] rounded px-1 py-0.5 transition-colors">
                  {showTimestamps && (
                    <span className="shrink-0 text-text-muted w-[86px] select-none text-code-sm">
                      {formatTimestamp(log.timestamp)}
                    </span>
                  )}
                  <span className={cn('shrink-0 w-10 text-code-sm font-semibold uppercase select-none',
                    levelColors[log.level] ?? 'text-foreground/70')}>
                    {log.level}
                  </span>
                  {log.source && (
                    <span className="shrink-0 text-text-muted max-w-[120px] truncate select-none text-code-sm">
                      [{log.source}]
                    </span>
                  )}
                  {log.step && (
                    <span className="shrink-0 text-text-muted max-w-[80px] truncate select-none text-code-sm">
                      ({log.step})
                    </span>
                  )}
                  <span className={cn('flex-1 break-words min-w-0 text-terminal-fg',
                    log.level === 'ERROR' ? 'text-danger' :
                    log.level === 'WARN' ? 'text-warning' :
                    log.level === 'DEBUG' ? 'text-text-muted/60' :
                    'text-foreground/80')}>
                    {log.message}
                  </span>
                </div>
              ))}
            </div>
          )}
        </ScrollArea>

        <div className="border-t border-border/30 px-3 py-1.5">
          <TerminalStatus streaming={connected && !error} paused={paused} error={!!error} />
        </div>
      </div>

      {/* Log count footer */}
      <div className="flex items-center justify-between text-caption text-text-muted">
        <span>
          {filteredLogs.length} {filteredLogs.length === 1 ? 'entry' : 'entries'}
          {searchTerm && logs.length !== filteredLogs.length ? ` (filtered from ${logs.length})` : ''}
        </span>
        {footerLabel && (
          <span className="font-mono text-code-sm">{footerLabel}</span>
        )}
      </div>
    </div>
  );
}
