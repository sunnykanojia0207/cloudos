import { useState, useEffect, useRef, useCallback } from 'react';

/* ── Types ─────────────────────────────────────────────── */

export interface DeploymentEvent {
  type: 'init' | 'progress' | 'node_started' | 'node_completed' | 'node_failed' | 'completed' | 'failed';
  payload: {
    phase: string;
    progress: number;
    currentNode?: string;
    completedNodes?: string[];
    failedNodes?: string[];
    totalNodes: number;
    duration?: string;
    url?: string;
    result?: string;
    error?: string;
  };
}

export interface DeploymentState {
  phase: string;
  progress: number;
  currentNode: string | null;
  completedNodes: string[];
  failedNodes: string[];
  totalNodes: number;
  duration: string | null;
  url: string | null;
  error: string | null;
  connected: boolean;
  finished: boolean;
}

const INITIAL_STATE: DeploymentState = {
  phase: 'Pending',
  progress: 0,
  currentNode: null,
  completedNodes: [],
  failedNodes: [],
  totalNodes: 0,
  duration: null,
  url: null,
  error: null,
  connected: false,
  finished: false,
};

/* ── Hook ──────────────────────────────────────────────── */

export function useDeploymentEvents(workflowId: string | undefined | null) {
  const [state, setState] = useState<DeploymentState>(INITIAL_STATE);
  const eventSourceRef = useRef<EventSource | null>(null);

  const close = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
  }, []);

  useEffect(() => {
    if (!workflowId) return;

    const baseUrl = import.meta.env.VITE_CLOUDOS_API_URL ?? '';
    const url = `${baseUrl}/api/v1/workflow-executions/${encodeURIComponent(workflowId)}/events`;

    setState(INITIAL_STATE);

    const es = new EventSource(url);
    eventSourceRef.current = es;

    es.onopen = () => {
      setState((prev) => ({ ...prev, connected: true }));
    };

    es.addEventListener('execution', (event) => {
      try {
        const data = JSON.parse(event.data) as DeploymentEvent['payload'];
        const eventType = event.type as DeploymentEvent['type'];

        setState((prev) => ({
          ...prev,
          phase: data.phase,
          progress: data.progress,
          currentNode: data.currentNode ?? prev.currentNode,
          completedNodes: data.completedNodes ?? prev.completedNodes,
          failedNodes: data.failedNodes ?? prev.failedNodes,
          totalNodes: data.totalNodes,
          duration: data.duration ?? prev.duration,
          url: data.url ?? prev.url,
          error: data.error ?? prev.error,
          finished: eventType === 'completed' || eventType === 'failed' || data.phase === 'completed' || data.phase === 'failed',
        }));
      } catch {
        // Ignore parse errors
      }
    });

    es.onerror = () => {
      setState((prev) => ({ ...prev, connected: false }));
    };

    return () => {
      close();
    };
  }, [workflowId, close]);

  return state;
}
