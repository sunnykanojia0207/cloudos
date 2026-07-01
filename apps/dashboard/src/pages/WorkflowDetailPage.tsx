import { useState, useMemo, useCallback, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { useWorkflow, type WorkflowDetail, type WorkflowStatus } from '@/hooks/useWorkflows';
import { usePageTitle } from '@/hooks/usePageTitle';
import type { TimelineStep as TimelineStepType } from '@/hooks/useDeployments';
import { Badge, StatusDot } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { EmptyState } from '@/components/ui/empty-state';
import { ErrorState } from '@/components/ui/error-state';
import { TimelineStep } from '@/components/ui/timeline-step';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  ArrowLeft,
  Activity,
  CheckCircle2,
  XCircle,
  Play,
  Clock,
  Timer,
  Layers,
  GitBranch,
  GitCommitHorizontal,
  Terminal,
  Server,
  RotateCcw,
  Info,
  AlertTriangle,
  ChevronDown,
  ChevronUp,
  ExternalLink,
} from 'lucide-react';
import { cn, relativeTime, truncate } from '@/lib/utils';

/* ════════════════════════════════════════════════════════════
   Execution Graph Layout (SVG)
   ════════════════════════════════════════════════════════════ */

const NODE_W = 145;
const NODE_H = 58;
const GAP_X = 60;
const GAP_Y = 90;
const PAD_X = 50;
const PAD_Y = 40;
const SVG_W = PAD_X * 2 + NODE_W * 4 + GAP_X * 3; // 860
const SVG_H = PAD_Y * 2 + NODE_H * 2 + GAP_Y;      // 286 → 300

const COL_CX = [0, 1, 2, 3].map((i) => PAD_X + NODE_W / 2 + i * (NODE_W + GAP_X));
const ROW_CY = [0, 1].map((i) => PAD_Y + NODE_H / 2 + i * (NODE_H + GAP_Y));

interface GraphNodeDef {
  key: string;
  label: string;
  cx: number;
  cy: number;
}

const GRAPH_NODES: GraphNodeDef[] = [
  { key: 'validate', label: 'Validate', cx: COL_CX[0], cy: ROW_CY[0] },
  { key: 'clone',    label: 'Clone',    cx: COL_CX[1], cy: ROW_CY[0] },
  { key: 'detect',   label: 'Detect',   cx: COL_CX[2], cy: ROW_CY[0] },
  { key: 'install',  label: 'Install',  cx: COL_CX[3], cy: ROW_CY[0] },
  { key: 'build',    label: 'Build',    cx: COL_CX[0], cy: ROW_CY[1] },
  { key: 'deploy',   label: 'Deploy',   cx: COL_CX[1], cy: ROW_CY[1] },
  { key: 'health',   label: 'Health',   cx: COL_CX[2], cy: ROW_CY[1] },
  { key: 'complete', label: 'Complete', cx: COL_CX[3], cy: ROW_CY[1] },
];

const GRAPH_EDGES: Array<{ from: string; to: string }> = [
  // Row 1 horizontal
  { from: 'validate', to: 'clone' },
  { from: 'clone', to: 'detect' },
  { from: 'detect', to: 'install' },
  // Row 1 → Row 2 vertical
  { from: 'validate', to: 'build' },
  { from: 'clone', to: 'deploy' },
  { from: 'detect', to: 'health' },
  { from: 'install', to: 'complete' },
  // Row 2 horizontal
  { from: 'build', to: 'deploy' },
  { from: 'deploy', to: 'health' },
  { from: 'health', to: 'complete' },
];

function nodeById(key: string): GraphNodeDef {
  return GRAPH_NODES.find((n) => n.key === key)!;
}

function edgePath(fromKey: string, toKey: string): string {
  const from = nodeById(fromKey);
  const to = nodeById(toKey);
  const x1 = from.cx + NODE_W / 2;
  const y1 = from.cy;
  const x2 = to.cx - NODE_W / 2;
  const y2 = to.cy;

  // Same row → horizontal line
  if (from.cy === to.cy) {
    return `M ${x1} ${y1} L ${x2} ${y2}`;
  }

  // Different row → vertical line with slight offset to avoid overlap
  const midX = to.cx; // center of target column
  return `M ${from.cx} ${from.cy + NODE_H / 2} L ${midX} ${from.cy + NODE_H / 2} L ${midX} ${to.cy - NODE_H / 2} L ${to.cx} ${to.cy - NODE_H / 2}`;
}

/* ── Node status helpers ──────────────────────────────────── */

type NodeStepStatus = 'success' | 'failed' | 'running' | 'pending' | 'skipped';

function computeNodeStatus(
  timelineSteps: TimelineStepType[],
  buildSuccess: boolean,
  errors: string[],
): Record<string, NodeStepStatus> {
  const statuses: Record<string, NodeStepStatus> = {};
  for (const node of GRAPH_NODES) {
    const step = timelineSteps.find(
      (s) => s.action?.toLowerCase() === node.key || s.name?.toLowerCase().includes(node.key),
    );
    if (step) {
      const lower = step.status?.toLowerCase() ?? '';
      if (['success', 'succeeded'].includes(lower)) statuses[node.key] = 'success';
      else if (['failure', 'failed', 'error'].includes(lower)) statuses[node.key] = 'failed';
      else if (lower === 'running') statuses[node.key] = 'running';
      else if (lower === 'skipped') statuses[node.key] = 'skipped';
      else statuses[node.key] = 'pending';
    } else {
      // Derive from build result
      if (node.key === 'complete') statuses[node.key] = buildSuccess ? 'success' : 'failed';
      else if (node.key === 'health') statuses[node.key] = buildSuccess && errors.length === 0 ? 'success' : 'failed';
      else if (node.key === 'build') statuses[node.key] = buildSuccess ? 'success' : 'failed';
      else if (['validate', 'clone', 'detect', 'install'].includes(node.key)) statuses[node.key] = 'success';
      else statuses[node.key] = 'pending';
    }
  }
  return statuses;
}

function nodeStatusColor(status: NodeStepStatus): { fill: string; stroke: string; text: string; subtext: string; icon: string } {
  switch (status) {
    case 'success':
      return { fill: '#0D2B1A', stroke: '#2B9D5D', text: '#EDEDEF', subtext: '#2B9D5D', icon: '✓' };
    case 'failed':
      return { fill: '#2D0E0E', stroke: '#D45A5A', text: '#EDEDEF', subtext: '#D45A5A', icon: '✗' };
    case 'running':
      return { fill: '#1E1F3A', stroke: '#5E6AD2', text: '#EDEDEF', subtext: '#5E6AD2', icon: '◉' };
    case 'skipped':
      return { fill: '#1C1C1F', stroke: '#5F5F66', text: '#9D9DA3', subtext: '#5F5F66', icon: '–' };
    default:
      return { fill: '#151517', stroke: '#26262B', text: '#5F5F66', subtext: '#5F5F66', icon: '○' };
  }
}

function edgeStatusColor(fromStatus: NodeStepStatus, toStatus: NodeStepStatus): string {
  if (toStatus === 'running') return '#5E6AD2';       // accent
  if (toStatus === 'success') return '#2B9D5D';         // success
  if (toStatus === 'failed') return '#D45A5A';          // danger
  return '#26262B';                                       // muted
}

/* ── Graph SVG Sub-components ────────────────────────────── */

interface GraphNodeSvgProps {
  node: GraphNodeDef;
  status: NodeStepStatus;
  duration?: string;
  selected: boolean;
  onClick: (key: string) => void;
}

function GraphNodeSvg({ node, status, duration, selected, onClick }: GraphNodeSvgProps) {
  const colors = nodeStatusColor(status);
  const x = node.cx - NODE_W / 2;
  const y = node.cy - NODE_H / 2;

  const isRunning = status === 'running';

  return (
    <g
      className={cn(
        'cursor-pointer transition-opacity',
        isRunning && 'animate-pulse',
      )}
      onClick={() => onClick(node.key)}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onClick(node.key); } }}
      role="button"
      tabIndex={0}
      aria-label={`${node.label}: ${status}`}
      style={{ outline: selected ? '2px solid #5E6AD2' : 'none', outlineOffset: '2px', borderRadius: '8px' }}
    >
      {/* Node background */}
      <rect
        x={x}
        y={y}
        width={NODE_W}
        height={NODE_H}
        rx={8}
        ry={8}
        fill={colors.fill}
        stroke={selected ? '#5E6AD2' : colors.stroke}
        strokeWidth={selected ? 2 : 1.5}
      />

      {/* Left accent bar */}
      <rect
        x={x}
        y={y + 8}
        width={3}
        height={NODE_H - 16}
        rx={1.5}
        fill={colors.stroke}
      />

      {/* Status icon */}
      <text
        x={x + 16}
        y={node.cy + 4}
        fill={colors.stroke}
        fontSize={14}
        fontFamily="monospace"
        fontWeight={600}
      >
        {colors.icon}
      </text>

      {/* Label */}
      <text
        x={x + 36}
        y={node.cy - 2}
        fill={colors.text}
        fontSize={13}
        fontFamily="system-ui, sans-serif"
        fontWeight={500}
      >
        {node.label}
      </text>

      {/* Subtext (duration or status) */}
      <text
        x={x + 36}
        y={node.cy + 16}
        fill={colors.subtext}
        fontSize={11}
        fontFamily="system-ui, sans-serif"
        opacity={0.8}
      >
        {duration || (status === 'pending' ? 'waiting...' : status)}
      </text>
    </g>
  );
}

/* ── Graph Arrow Sub-component ───────────────────────────── */

interface GraphArrowProps {
  fromKey: string;
  toKey: string;
  fromStatus: NodeStepStatus;
  toStatus: NodeStepStatus;
}

function GraphArrow({ fromKey, toKey, fromStatus, toStatus }: GraphArrowProps) {
  const toRunning = toStatus === 'running';
  const color = edgeStatusColor(fromStatus, toStatus);
  const path = edgePath(fromKey, toKey);

  return (
    <g>
      {/* Arrow line */}
      <path
        d={path}
        fill="none"
        stroke={color}
        strokeWidth={2}
        strokeLinecap="round"
        className={cn(toRunning && 'animate-flow-dash')}
        style={toRunning ? { strokeDasharray: '6 4' } : undefined}
      />
      {/* Arrowhead marker */}
      {(() => {
        const to = nodeById(toKey);
        // We'll just draw a small triangle at the end
        const endX = to.cx - NODE_W / 2;
        const endY = to.cy;
        return (
          <polygon
            points={`${endX - 4},${endY - 4} ${endX},${endY} ${endX - 4},${endY + 4}`}
            fill={color}
          />
        );
      })()}
    </g>
  );
}

/* ════════════════════════════════════════════════════════════
   Execution Graph (Desktop SVG)
   ════════════════════════════════════════════════════════════ */

interface ExecutionGraphProps {
  detail: WorkflowDetail;
  selectedNode: string | null;
  onSelectNode: (key: string | null) => void;
}

function ExecutionGraph({ detail, selectedNode, onSelectNode }: ExecutionGraphProps) {
  const { timeline } = detail;
  const steps = timeline?.steps ?? [];
  const buildSuccess = detail.status === 'succeeded';
  const errors = timeline?.steps?.filter((s) => s.status === 'failed').map((s) => s.error || '').filter(Boolean) ?? [];

  const nodeStatuses = useMemo(
    () => computeNodeStatus(steps, buildSuccess, errors),
    [steps, buildSuccess, errors],
  );

  // Map node keys to durations from timeline
  const nodeDurations = useMemo(() => {
    const map: Record<string, string> = {};
    const timelineSteps = timeline?.steps ?? [];
    for (const node of GRAPH_NODES) {
      const step = timelineSteps.find(
        (s) => s.action?.toLowerCase() === node.key || s.name?.toLowerCase().includes(node.key),
      );
      if (step) {
        // Since timeline steps don't have a duration field, we derive it
        // In a real system, the step would have a duration
        map[node.key] = step.result?.match(/([\d.]+)s/)?.[1] ? `${step.result.match(/([\d.]+)s/)?.[1]}s` : '\u2014';
      } else {
        map[node.key] = '\u2014';
      }
    }
    return map;
  }, [timeline]);

  const svgRef = useRef<SVGSVGElement>(null);

  // Keyboard navigation
  useEffect(() => {
    const svg = svgRef.current;
    if (!svg) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onSelectNode(null);
        return;
      }

      if (!selectedNode) return;

      const idx = GRAPH_NODES.findIndex((n) => n.key === selectedNode);
      if (idx === -1) return;

      let nextIdx = idx;
      if (e.key === 'ArrowRight') nextIdx = Math.min(idx + 1, GRAPH_NODES.length - 1);
      else if (e.key === 'ArrowLeft') nextIdx = Math.max(idx - 1, 0);
      else if (e.key === 'ArrowDown') nextIdx = idx < 4 ? idx + 4 : idx;
      else if (e.key === 'ArrowUp') nextIdx = idx >= 4 ? idx - 4 : idx;
      else return;

      if (nextIdx !== idx) {
        e.preventDefault();
        onSelectNode(GRAPH_NODES[nextIdx].key);
      }
    };

    svg.addEventListener('keydown', handleKeyDown);
    return () => svg.removeEventListener('keydown', handleKeyDown);
  }, [selectedNode, onSelectNode]);

  return (
    <svg
      ref={svgRef}
      viewBox={`0 0 ${SVG_W} ${SVG_H}`}
      className="w-full h-auto max-w-full"
      role="grid"
      aria-label="Workflow execution graph"
      style={{ minHeight: '200px' }}
    >
      {/* Arrow definitions */}
      <defs>
        {/* Flowing dash animation for running edges */}
        <style>
          {`
            @keyframes flowDash {
              to { stroke-dashoffset: -20; }
            }
            .animate-flow-dash {
              animation: flowDash 0.8s linear infinite;
            }
          `}
        </style>
      </defs>

      {/* Edges (drawn first so they're behind nodes) */}
      {GRAPH_EDGES.map((edge) => (
        <GraphArrow
          key={`${edge.from}-${edge.to}`}
          fromKey={edge.from}
          toKey={edge.to}
          fromStatus={nodeStatuses[edge.from] ?? 'pending'}
          toStatus={nodeStatuses[edge.to] ?? 'pending'}
        />
      ))}

      {/* Nodes */}
      {GRAPH_NODES.map((node) => (
        <GraphNodeSvg
          key={node.key}
          node={node}
          status={nodeStatuses[node.key] ?? 'pending'}
          duration={nodeDurations[node.key]}
          selected={selectedNode === node.key}
          onClick={onSelectNode}
        />
      ))}
    </svg>
  );
}

/* ════════════════════════════════════════════════════════════
   Mobile Timeline View (replaces SVG graph on small screens)
   ════════════════════════════════════════════════════════════ */

function mapStepStatus(status: string): 'succeeded' | 'failed' | 'running' | 'pending' | 'skipped' | 'cancelled' {
  const lower = status?.toLowerCase() ?? '';
  if (['success', 'succeeded'].includes(lower)) return 'succeeded';
  if (['failure', 'failed', 'error'].includes(lower)) return 'failed';
  if (lower === 'running') return 'running';
  if (lower === 'skipped') return 'skipped';
  if (lower === 'cancelled') return 'cancelled';
  return 'pending';
}

function MobileTimeline({ detail }: { detail: WorkflowDetail }) {
  const steps = detail.timeline?.steps ?? [];
  const buildSuccess = detail.status === 'succeeded';
  const errors = steps.filter((s) => s.status === 'failed');

  // If no timeline steps, show derived steps from node statuses
  const displaySteps = steps.length > 0 ? steps : GRAPH_NODES.map((node) => ({
    id: node.key,
    name: node.label,
    action: node.key,
    status: (() => {
      if (node.key === 'complete') return buildSuccess ? 'success' : 'failed';
      if (node.key === 'health') return buildSuccess && errors.length === 0 ? 'success' : 'failed';
      if (node.key === 'build') return buildSuccess ? 'success' : 'failed';
      if (['validate', 'clone', 'detect', 'install'].includes(node.key)) return 'success';
      return 'pending';
    })(),
    result: undefined,
    error: undefined,
  }));

  return (
    <div className="md:hidden space-y-1 px-1">
      {displaySteps.map((step, index) => (
        <TimelineStep
          key={step.id ?? index}
          state={mapStepStatus(step.status)}
          title={step.name}
          detail={step.result}
          error={step.error}
          isLast={index === displaySteps.length - 1}
        />
      ))}
    </div>
  );
}

/* ════════════════════════════════════════════════════════════
   Node Detail Panel
   ════════════════════════════════════════════════════════════ */

function NodeDetailPanel({
  nodeKey,
  detail,
  onClose,
}: {
  nodeKey: string;
  detail: WorkflowDetail;
  onClose: () => void;
}) {
  const nodeDef = GRAPH_NODES.find((n) => n.key === nodeKey);
  if (!nodeDef) return null;

  const step = detail.timeline?.steps?.find(
    (s) => s.action?.toLowerCase() === nodeKey || s.name?.toLowerCase().includes(nodeKey),
  );

  const status = step
    ? mapStepStatus(step.status)
    : 'pending';

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: 8 }}
      transition={{ duration: 0.15 }}
      className="rounded-md border border-border bg-surface p-4 space-y-3"
      role="region"
      aria-label={`Details for ${nodeDef.label}`}
    >
      <div className="flex items-center justify-between">
        <h3 className="text-body font-semibold text-foreground flex items-center gap-2">
          <Activity className="h-4 w-4 text-text-muted" />
          {nodeDef.label}
        </h3>
        <button
          type="button"
          onClick={onClose}
          className="text-text-muted hover:text-foreground p-0.5 rounded-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          aria-label="Close node details"
        >
          <XCircle className="h-3.5 w-3.5" />
        </button>
      </div>

      <Separator />

      <div className="grid grid-cols-2 gap-x-4 gap-y-2.5 text-small">
        <div>
          <span className="text-text-secondary">Status</span>
          <div className="mt-0.5">
            <Badge variant={
              status === 'succeeded' ? 'subtle-success' :
              status === 'failed' ? 'subtle-danger' :
              status === 'running' ? 'subtle-accent' :
              'subtle-neutral'
            } className="gap-1 text-caption">
              {status === 'succeeded' ? <CheckCircle2 className="h-3 w-3" /> :
               status === 'failed' ? <XCircle className="h-3 w-3" /> :
               status === 'running' ? <Play className="h-3 w-3" /> :
               null}
              {status.charAt(0).toUpperCase() + status.slice(1)}
            </Badge>
          </div>
        </div>
        <div>
          <span className="text-text-secondary">Action</span>
          <p className="mt-0.5 text-foreground font-mono text-small">{step?.action || nodeDef.key}</p>
        </div>
        <div>
          <span className="text-text-secondary">Output</span>
          <p className="mt-0.5 text-foreground text-small">{step?.result || '\u2014'}</p>
        </div>
        <div>
          <span className="text-text-secondary">Duration</span>
          <p className="mt-0.5 text-foreground tabular-nums text-small">{detail.duration || '\u2014'}</p>
        </div>
        <div>
          <span className="text-text-secondary">Retry Count</span>
          <p className="mt-0.5 text-foreground tabular-nums text-small">0</p>
        </div>
      </div>

      {/* Errors / Warnings */}
      {step?.error && (
        <div className="rounded-sm bg-danger-subtle border border-danger/20 p-2.5">
          <p className="text-caption font-medium text-danger mb-1 flex items-center gap-1">
            <AlertTriangle className="h-3 w-3" />
            Error
          </p>
          <p className="text-small text-danger">{step.error}</p>
        </div>
      )}
    </motion.div>
  );
}

/* ════════════════════════════════════════════════════════════
   Execution Events Log
   ════════════════════════════════════════════════════════════ */

function ExecutionEvents({ detail }: { detail: WorkflowDetail }) {
  const [expanded, setExpanded] = useState(true);
  const steps = detail.timeline?.steps ?? [];

  // Build synthetic events from timeline steps
  const events = useMemo(() => {
    const evts: Array<{ time: string; message: string; level: 'info' | 'success' | 'warn' | 'error' }> = [];

    if (detail.timeline?.startedAt) {
      evts.push({
        time: new Date(detail.timeline.startedAt).toLocaleTimeString(),
        message: 'Workflow execution started',
        level: 'info',
      });
    }

    for (const step of steps) {
      const time = ''; // We don't have per-step timestamps
      const statusLower = step.status?.toLowerCase() ?? '';
      if (['success', 'succeeded'].includes(statusLower)) {
        evts.push({ time: '', message: `${step.name}: ${step.result || 'Completed'}`, level: 'success' });
      } else if (['failure', 'failed', 'error'].includes(statusLower)) {
        evts.push({ time: '', message: `${step.name}: ${step.error || 'Failed'}`, level: 'error' });
      } else if (statusLower === 'running') {
        evts.push({ time: '', message: `${step.name}: Running...`, level: 'info' });
      } else {
        evts.push({ time: '', message: `${step.name}: ${step.result || 'Pending'}`, level: 'info' });
      }
    }

    if (detail.timeline?.completedAt) {
      evts.push({
        time: new Date(detail.timeline.completedAt).toLocaleTimeString(),
        message: `Workflow ${detail.status === 'succeeded' ? 'completed successfully' : 'failed'}`,
        level: detail.status === 'succeeded' ? 'success' : 'error',
      });
    }

    // Reverse so newest first
    evts.reverse();

    return evts;
  }, [steps, detail.timeline, detail.status]);

  if (events.length === 0) return null;

  return (
    <Card>
      <CardHeader className="pb-2">
        <button
          type="button"
          onClick={() => setExpanded(!expanded)}
          className="flex items-center justify-between w-full text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded-sm"
          aria-expanded={expanded}
          aria-label={expanded ? 'Collapse execution events' : 'Expand execution events'}
        >
          <CardTitle className="text-body font-semibold flex items-center gap-2">
            <Terminal className="h-4 w-4 text-text-muted" />
            Execution Events
            <span className="text-caption font-normal text-text-muted tabular-nums">
              ({events.length})
            </span>
          </CardTitle>
          {expanded ? <ChevronUp className="h-4 w-4 text-text-muted" /> : <ChevronDown className="h-4 w-4 text-text-muted" />}
        </button>
      </CardHeader>
      <AnimatePresence>
        {expanded && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.15 }}
          >
            <CardContent className="pt-0">
              <div className="rounded-sm bg-terminal border border-border p-2.5 max-h-[240px] overflow-y-auto font-mono text-code-sm leading-relaxed">
                {events.map((evt, i) => (
                  <div key={i} className="flex items-start gap-2 text-terminal-fg">
                    {evt.time && (
                      <span className="shrink-0 text-text-muted tabular-nums">{evt.time}</span>
                    )}
                    <span className={cn(
                      'shrink-0 w-3 text-center',
                      evt.level === 'success' ? 'text-success' :
                      evt.level === 'error' ? 'text-danger' :
                      evt.level === 'warn' ? 'text-warning' :
                      'text-text-muted',
                    )}>
                      {evt.level === 'success' ? '✓' : evt.level === 'error' ? '✗' : evt.level === 'warn' ? '⚠' : '•'}
                    </span>
                    <span className="flex-1 break-all">{evt.message}</span>
                  </div>
                ))}
              </div>
            </CardContent>
          </motion.div>
        )}
      </AnimatePresence>
    </Card>
  );
}

/* ════════════════════════════════════════════════════════════
   Loading / Error / Empty
   ════════════════════════════════════════════════════════════ */

function DetailSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      {/* Back + header */}
      <div className="flex items-center gap-3">
        <Skeleton className="h-7 w-7 rounded-sm" />
        <Skeleton className="h-5 w-48" />
        <Skeleton className="h-5 w-24 rounded-sm" />
      </div>

      {/* Two columns */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
        <div className="lg:col-span-2">
          <Skeleton className="h-[300px] w-full rounded-md" shimmer />
        </div>
        <div>
          <div className="rounded-md border border-border bg-surface p-4 space-y-3">
            <Skeleton className="h-5 w-36" />
            <Skeleton className="h-px w-full" />
            <div className="space-y-2.5">
              {Array.from({ length: 6 }).map((_, i) => (
                <div key={i} className="flex justify-between">
                  <Skeleton className="h-4 w-24" />
                  <Skeleton className="h-4 w-20" />
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

/* ════════════════════════════════════════════════════════════
   Main Page Component
   ════════════════════════════════════════════════════════════ */

export default function WorkflowDetailPage() {
  const { workflowId } = useParams<{ workflowId: string }>();
  const navigate = useNavigate();
  const { data: detail, isLoading, error, refetch } = useWorkflow(workflowId ?? '');

  usePageTitle(detail ? `Workflow ${truncate(detail.id, 20)}` : 'Workflow Detail');
  const [selectedNode, setSelectedNode] = useState<string | null>(null);

  // Reset selected node when detail changes
  useEffect(() => {
    setSelectedNode(null);
  }, [workflowId]);

  const handleSelectNode = useCallback((key: string | null) => {
    setSelectedNode((prev) => (prev === key ? null : key));
  }, []);

  // ── Error ──
  if (!isLoading && error) {
    return (
      <ErrorState
        title="Workflow not found"
        message={(error as Error)?.message || 'Could not load this workflow execution.'}
        onRetry={() => refetch()}
      />
    );
  }

  // ── Loading ──
  if (isLoading || !detail) {
    return <DetailSkeleton />;
  }

  const statusLabel: Record<WorkflowStatus, string> = {
    succeeded: 'Succeeded',
    failed: 'Failed',
    running: 'Running',
    pending: 'Pending',
    cancelled: 'Cancelled',
  };

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      className="flex flex-col gap-6"
    >
      {/* ══════════ HEADER ══════════ */}
      <div className="flex flex-col gap-2">
        {/* Back */}
        <button
          type="button"
          onClick={() => navigate('/workflows')}
          className="inline-flex items-center gap-1.5 text-small text-text-secondary hover:text-foreground transition-colors w-fit focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded-sm"
          aria-label="Back to all workflows"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          All Workflows
        </button>

        {/* Title bar */}
        <div className="flex flex-wrap items-center gap-3">
          <h1 className="text-h1 text-foreground font-mono text-base md:text-h1">
            {truncate(detail.id, 32)}
          </h1>
          <Badge variant={
            detail.status === 'succeeded' ? 'subtle-success' :
            detail.status === 'failed' ? 'subtle-danger' :
            detail.status === 'running' ? 'subtle-accent' :
            'subtle-neutral'
          } className="gap-1.5">
            <StatusDot status={
              detail.status === 'succeeded' ? 'success' :
              detail.status === 'failed' ? 'danger' :
              detail.status === 'running' ? 'deploying' :
              'pending'
            } pulsing={detail.status === 'running'} />
            {statusLabel[detail.status]}
          </Badge>
        </div>

        {/* Subtitle */}
        <p className="text-small text-text-secondary">
          {detail.appName} &middot; Deployment #{detail.deploymentNumber}
          {detail.branch && <> &middot; {detail.branch}</>}
        </p>
      </div>

      {/* ══════════ MAIN CONTENT (Two columns on desktop) ══════════ */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
        {/* ── Left: Graph (desktop) / Timeline (mobile) ── */}
        <div className="lg:col-span-2 space-y-4">
          {/* Desktop SVG graph (hidden on mobile) */}
          <div className="hidden md:block">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-body font-semibold flex items-center gap-2">
                  <Activity className="h-4 w-4 text-text-muted" />
                  Execution Graph
                </CardTitle>
              </CardHeader>
              <CardContent className="pt-1">
                <ExecutionGraph
                  detail={detail}
                  selectedNode={selectedNode}
                  onSelectNode={handleSelectNode}
                />
                {/* Legend */}
                <div className="flex items-center gap-4 mt-2 text-caption text-text-muted">
                  <span className="flex items-center gap-1">
                    <span className="inline-block h-2 w-2 rounded-full bg-success" />
                    Succeeded
                  </span>
                  <span className="flex items-center gap-1">
                    <span className="inline-block h-2 w-2 rounded-full bg-danger" />
                    Failed
                  </span>
                  <span className="flex items-center gap-1">
                    <span className="inline-block h-2 w-2 rounded-full bg-accent animate-pulse" />
                    Running
                  </span>
                  <span className="flex items-center gap-1">
                    <span className="inline-block h-2 w-2 rounded-full bg-border" />
                    Pending
                  </span>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Mobile vertical timeline (hidden on desktop) */}
          <div className="md:hidden">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-body font-semibold flex items-center gap-2">
                  <Activity className="h-4 w-4 text-text-muted" />
                  Execution Steps
                </CardTitle>
              </CardHeader>
              <CardContent>
                <MobileTimeline detail={detail} />
              </CardContent>
            </Card>
          </div>

          {/* ── Node Detail Panel (when a node is selected) ── */}
          <AnimatePresence mode="wait">
            {selectedNode && (
              <NodeDetailPanel
                key={selectedNode}
                nodeKey={selectedNode}
                detail={detail}
                onClose={() => setSelectedNode(null)}
              />
            )}
          </AnimatePresence>

          {/* ── Execution Events ── */}
          <ExecutionEvents detail={detail} />
        </div>

        {/* ── Right: Workflow Summary (sticky) ── */}
        <div className="lg:col-span-1">
          <div className="lg:sticky lg:top-6 space-y-4">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-body font-semibold flex items-center gap-2">
                  <Info className="h-4 w-4 text-text-muted" />
                  Workflow Summary
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {/* Execution ID */}
                <div>
                  <span className="text-caption text-text-secondary">Execution ID</span>
                  <p className="text-small text-foreground font-mono break-all mt-0.5">{detail.id}</p>
                </div>

                <Separator />

                {/* Application */}
                <div className="flex items-center justify-between">
                  <span className="text-small text-text-secondary">Application</span>
                  <button
                    type="button"
                    onClick={() => navigate(`/applications/${detail.appId}`)}
                    className="text-small text-accent hover:text-accent-hover focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded-sm"
                  >
                    {detail.appName}
                  </button>
                </div>

                {/* Deployment */}
                <div className="flex items-center justify-between">
                  <span className="text-small text-text-secondary">Deployment</span>
                  <button
                    type="button"
                    onClick={() => navigate(`/applications/${detail.appId}/deployments/${detail.deploymentNumber}`)}
                    className="text-small text-accent hover:text-accent-hover focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded-sm"
                  >
                    #{detail.deploymentNumber}
                  </button>
                </div>

                {/* Runtime */}
                <div className="flex items-center justify-between">
                  <span className="text-small text-text-secondary">Runtime</span>
                  <span className="text-small text-foreground">{detail.runtime || '\u2014'}</span>
                </div>

                <Separator />

                {/* Started */}
                <div className="flex items-center justify-between">
                  <span className="text-small text-text-secondary">Started</span>
                  <span className="text-small text-foreground tabular-nums">{relativeTime(detail.startedAt)}</span>
                </div>

                {/* Finished */}
                <div className="flex items-center justify-between">
                  <span className="text-small text-text-secondary">Finished</span>
                  <span className="text-small text-foreground tabular-nums">
                    {detail.completedAt ? relativeTime(detail.completedAt) : '\u2014'}
                  </span>
                </div>

                {/* Duration */}
                <div className="flex items-center justify-between">
                  <span className="text-small text-text-secondary">Duration</span>
                  <span className="text-small text-foreground tabular-nums font-medium">{detail.duration || '\u2014'}</span>
                </div>

                {/* Retries */}
                <div className="flex items-center justify-between">
                  <span className="text-small text-text-secondary">Retries</span>
                  <span className="text-small text-foreground tabular-nums">0</span>
                </div>

                {/* Queue Time */}
                <div className="flex items-center justify-between">
                  <span className="text-small text-text-secondary">Queue Time</span>
                  <span className="text-small text-foreground tabular-nums">{detail.queueTime || '\u2014'}</span>
                </div>

                {/* Nodes */}
                <Separator />

                <div className="flex items-center justify-between">
                  <span className="text-small text-text-secondary">Total Nodes</span>
                  <span className="text-small text-foreground tabular-nums">{detail.nodeCount || 8}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-small text-text-secondary">Completed</span>
                  <span className="text-small text-success tabular-nums">{detail.completedNodes}</span>
                </div>
                {detail.failedNodes > 0 && (
                  <div className="flex items-center justify-between">
                    <span className="text-small text-text-secondary">Failed</span>
                    <span className="text-small text-danger tabular-nums">{detail.failedNodes}</span>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Quick Actions */}
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-body font-semibold flex items-center gap-2">
                  <ExternalLink className="h-4 w-4 text-text-muted" />
                  Quick Actions
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <Button
                  variant="secondary"
                  size="sm"
                  className="w-full justify-start gap-2"
                  onClick={() => navigate(`/applications/${detail.appId}`)}
                >
                  <Activity className="h-3.5 w-3.5" />
                  View Application
                </Button>
                <Button
                  variant="secondary"
                  size="sm"
                  className="w-full justify-start gap-2"
                  onClick={() => navigate(`/applications/${detail.appId}/deployments/${detail.deploymentNumber}/timeline`)}
                >
                  <Terminal className="h-3.5 w-3.5" />
                  View Timeline
                </Button>
                <Button
                  variant="secondary"
                  size="sm"
                  className="w-full justify-start gap-2"
                  onClick={() => navigate(`/applications/${detail.appId}/deployments/${detail.deploymentNumber}/compare`)}
                  disabled={detail.deploymentNumber <= 1}
                >
                  <GitCommitHorizontal className="h-3.5 w-3.5" />
                  Compare Deployments
                </Button>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </motion.div>
  );
}
