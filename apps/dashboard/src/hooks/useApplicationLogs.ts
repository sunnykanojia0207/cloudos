import { useCallback, useEffect, useRef, useState } from 'react';

export interface LogEvent {
  timestamp: string;
  source: string;
  level: string;
  step?: string;
  message: string;
}

export function useApplicationLogs(appId: string, tail = 50) {
  const [logs, setLogs] = useState<LogEvent[]>([]);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [paused, setPaused] = useState(false);
  const eventSourceRef = useRef<EventSource | null>(null);
  const bufferRef = useRef<LogEvent[]>([]);
  const pausedRef = useRef(false);

  pausedRef.current = paused;

  // Fetch initial snapshot
  useEffect(() => {
    if (!appId) return;

    const base = import.meta.env.VITE_CLOUDOS_API_URL ?? '';
    fetch(`${base}/api/v1/applications/${encodeURIComponent(appId)}/logs?tail=${tail}`)
      .then((res) => res.json())
      .then((body) => {
        if (body.success && Array.isArray(body.data)) {
          setLogs(body.data as LogEvent[]);
        }
      })
      .catch((err) => setError(err.message));
  }, [appId, tail]);

  // Connect SSE stream
  useEffect(() => {
    if (!appId) return;

    const base = import.meta.env.VITE_CLOUDOS_API_URL ?? '';
    const url = `${base}/api/v1/applications/${encodeURIComponent(appId)}/logs/stream?tail=10`;
    const es = new EventSource(url);
    eventSourceRef.current = es;

    es.onopen = () => setConnected(true);
    es.onerror = () => {
      setConnected(false);
      setError('Log stream disconnected');
    };

    es.onmessage = (event) => {
      try {
        const logEvent = JSON.parse(event.data) as LogEvent;
        if (!pausedRef.current) {
          setLogs((prev) => {
            const next = [...prev, logEvent];
            // Cap at 1000 entries
            return next.length > 1000 ? next.slice(-1000) : next;
          });
        } else {
          bufferRef.current.push(logEvent);
        }
      } catch {
        // ignore malformed events
      }
    };

    return () => {
      es.close();
      eventSourceRef.current = null;
      setConnected(false);
    };
  }, [appId, tail]);

  const resume = useCallback(() => {
    setPaused(false);
    // Flush buffer
    if (bufferRef.current.length > 0) {
      setLogs((prev) => {
        const next = [...prev, ...bufferRef.current];
        bufferRef.current = [];
        return next.length > 1000 ? next.slice(-1000) : next;
      });
    }
  }, []);

  const pause = useCallback(() => setPaused(true), []);

  const clearLogs = useCallback(() => {
    setLogs([]);
    bufferRef.current = [];
  }, []);

  const downloadLogs = useCallback(() => {
    const base = import.meta.env.VITE_CLOUDOS_API_URL ?? '';
    window.open(`${base}/api/v1/applications/${encodeURIComponent(appId)}/logs/download`, '_blank');
  }, [appId]);

  return {
    logs,
    connected,
    error,
    paused,
    pause,
    resume,
    clearLogs,
    downloadLogs,
  };
}
