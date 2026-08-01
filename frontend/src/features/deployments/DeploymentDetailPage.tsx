import { useParams, useNavigate } from '@tanstack/react-router';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { deploymentsApi, projectsApi } from '../../api/endpoints';
import { useState, useCallback, useRef, useEffect } from 'react';
import { LogStreamer } from '../logs/LogStreamer';
import {
  ArrowLeft,
  ExternalLink,
  Square,
  GitCommit,
  Clock,
  Server,
  HardDrive,
  Globe,
  CheckCircle2,
  XCircle,
  Loader2,
  Terminal,
  Copy,
  ChevronDown,
  Activity,
} from 'lucide-react';
import { Link } from '@tanstack/react-router';

// --- Status config ---
const STATUS_CONFIG = {
  pending:  { label: 'Pending',  color: 'text-yellow-400', bg: 'bg-yellow-400/10 border-yellow-400/20', dot: 'bg-yellow-400', pulse: true  },
  cloning:  { label: 'Cloning',  color: 'text-blue-400',   bg: 'bg-blue-400/10 border-blue-400/20',     dot: 'bg-blue-400',   pulse: true  },
  building: { label: 'Building', color: 'text-orange-400', bg: 'bg-orange-400/10 border-orange-400/20', dot: 'bg-orange-400', pulse: true  },
  running:  { label: 'Starting', color: 'text-cyan-400',   bg: 'bg-cyan-400/10 border-cyan-400/20',     dot: 'bg-cyan-400',   pulse: true  },
  live:     { label: 'Live',     color: 'text-green-400',  bg: 'bg-green-400/10 border-green-400/20',   dot: 'bg-green-400',  pulse: false },
  stopped:  { label: 'Stopped',  color: 'text-gray-400',   bg: 'bg-gray-400/10 border-gray-400/20',     dot: 'bg-gray-400',   pulse: false },
  failed:   { label: 'Failed',   color: 'text-red-400',    bg: 'bg-red-400/10 border-red-400/20',       dot: 'bg-red-400',    pulse: false },
} as const;

// Formats a date string to "Jul 29, 2026 · 3:45 PM"
function formatDate(dateStr: string) {
  const d = new Date(dateStr);
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) +
    ' · ' + d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
}

// How long ago
function timeAgo(dateStr: string) {
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}

// --- Sub-components ---

function StatusBadge({ status }: { status: keyof typeof STATUS_CONFIG }) {
  const cfg = STATUS_CONFIG[status] ?? STATUS_CONFIG.stopped;
  return (
    <span className={`inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-semibold tracking-wide ${cfg.bg} ${cfg.color}`}>
      <span className={`relative flex h-1.5 w-1.5`}>
        {cfg.pulse && <span className={`absolute inline-flex h-full w-full animate-ping rounded-full ${cfg.dot} opacity-75`} />}
        <span className={`relative inline-flex h-1.5 w-1.5 rounded-full ${cfg.dot}`} />
      </span>
      {cfg.label}
    </span>
  );
}

function InfoCard({ label, value, mono = false, copyable = false }: { label: string; value: string; mono?: boolean; copyable?: boolean }) {
  const [copied, setCopied] = useState(false);
  const copy = () => {
    navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };
  return (
    <div className="group flex flex-col gap-1">
      <span className="text-xs font-medium uppercase tracking-wider text-gray-500">{label}</span>
      <div className="flex items-center gap-2">
        <span className={`truncate text-sm text-gray-200 ${mono ? 'font-mono' : ''}`}>{value}</span>
        {copyable && (
          <button onClick={copy} className="shrink-0 opacity-0 transition-opacity group-hover:opacity-100" title="Copy">
            {copied ? <CheckCircle2 className="h-3.5 w-3.5 text-green-400" /> : <Copy className="h-3.5 w-3.5 text-gray-500 hover:text-gray-300" />}
          </button>
        )}
      </div>
    </div>
  );
}

// The actual terminal log panel
function LogPanel({ logs, isStreaming, connectionStatus }: {
  logs: string[];
  isStreaming: boolean;
  connectionStatus: string;
}) {
  const [autoScroll, setAutoScroll] = useState(true);
  const endRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // auto-scroll whenever new lines arrive
  useEffect(() => {
    if (autoScroll) endRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' });
  }, [logs, autoScroll]);

  // detect manual scroll up to disable autoscroll
  const onScroll = () => {
    const el = containerRef.current;
    if (!el) return;
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 40;
    setAutoScroll(atBottom);
  };

  const displayLines = logs;

  const connBadge = {
    connecting: 'bg-yellow-500/20 text-yellow-400',
    connected: 'bg-green-500/20 text-green-400',
    closed: 'bg-gray-500/20 text-gray-400',
    error: 'bg-red-500/20 text-red-400',
  }[connectionStatus] ?? 'bg-gray-500/20 text-gray-400';

  return (
    <div className="flex h-full flex-col overflow-hidden rounded-xl border border-white/[0.06] bg-[#0d1117]">
      {/* Header bar */}
      <div className="flex shrink-0 items-center justify-between border-b border-white/[0.06] px-4 py-2.5">
        <div className="flex items-center gap-3">
          <Terminal className="h-4 w-4 text-gray-500" />
          <span className="text-sm font-medium text-gray-300">Build Logs</span>
          {isStreaming && (
            <span className={`rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider ${connBadge}`}>
              {connectionStatus}
            </span>
          )}
        </div>
        <div className="flex items-center gap-3">
          {displayLines.length > 0 && (
            <span className="text-xs text-gray-600">{displayLines.length} lines</span>
          )}
          <label className="flex cursor-pointer items-center gap-1.5 text-xs text-gray-500">
            <input
              type="checkbox"
              checked={autoScroll}
              onChange={e => setAutoScroll(e.target.checked)}
              className="h-3 w-3 rounded accent-green-400"
            />
            Auto-scroll
          </label>
        </div>
      </div>

      {/* Log body */}
      <div
        ref={containerRef}
        onScroll={onScroll}
        className="flex-1 overflow-auto px-4 py-3 font-mono text-[12.5px] leading-relaxed"
      >
        {displayLines.length === 0 ? (
          <div className="flex h-full flex-col items-center justify-center gap-3 text-gray-600">
            <Terminal className="h-8 w-8 opacity-30" />
            <span className="text-sm">Waiting for logs…</span>
          </div>
        ) : (
          displayLines.map((line, i) => {
            // colorize common patterns
            const isError = /error|failed|fatal/i.test(line);
            const isWarn  = /warn/i.test(line);
            const isOk    = /success|done|complete|✓/i.test(line);
            const isStep  = /^(step|from|run|copy|add|workdir|cmd|entrypoint|arg|env|expose)\s/i.test(line) || /^\[+/.test(line);
            const cls = isError ? 'text-red-400' : isWarn ? 'text-yellow-400' : isOk ? 'text-green-400' : isStep ? 'text-cyan-300' : 'text-gray-400';
            return (
              <div key={i} className={`flex gap-3 ${cls}`}>
                <span className="w-9 shrink-0 select-none text-right text-gray-700">{i + 1}</span>
                <span className="break-all">{line || '\u00a0'}</span>
              </div>
            );
          })
        )}
        <div ref={endRef} />
      </div>

      {/* Scroll-to-bottom fab */}
      {!autoScroll && displayLines.length > 0 && (
        <button
          onClick={() => { setAutoScroll(true); endRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' }); }}
          className="absolute bottom-6 right-6 flex items-center gap-1.5 rounded-full border border-white/10 bg-gray-800 px-3 py-1.5 text-xs text-gray-300 shadow-lg transition hover:bg-gray-700"
        >
          <ChevronDown className="h-3.5 w-3.5" /> Jump to bottom
        </button>
      )}
    </div>
  );
}

// ── Main page ─────────────────────────────────────────────────────────────────

export function DeploymentDetailPage() {
  const { deploymentId } = useParams({ strict: false }) as { deploymentId: string };
  const [logs, setLogs] = useState<string[]>([]);
  const [connectionStatus, setConnectionStatus] = useState<string>('connecting');
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  const { data: deployment, isLoading } = useQuery({
    queryKey: ['deployment', deploymentId],
    queryFn: () => deploymentsApi.get(deploymentId),
    enabled: !!deploymentId,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      return status && ['pending', 'cloning', 'building', 'running'].includes(status) ? 2000 : false;
    },
  });

  const { data: project } = useQuery({
    queryKey: ['project', deployment?.project_id],
    queryFn: () => projectsApi.get(deployment!.project_id),
    enabled: !!deployment?.project_id,
  });

  const stopMutation = useMutation({
    mutationFn: () => deploymentsApi.stop(deploymentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['deployments', deployment?.project_id] });
      navigate({ to: '/projects/$projectId', params: { projectId: deployment!.project_id } });
    },
  });

  const handleLog = useCallback((line: string) => {
    setLogs(prev => [...prev, line]);
  }, []);

  const handleStatusChange = useCallback((status: string) => {
    setConnectionStatus(status);
  }, []);

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center gap-3 text-gray-500">
        <Loader2 className="h-5 w-5 animate-spin" />
        <span>Loading deployment…</span>
      </div>
    );
  }

  if (!deployment) {
    return (
      <div className="flex h-64 flex-col items-center justify-center gap-3">
        <XCircle className="h-8 w-8 text-red-400" />
        <span className="text-gray-400">Deployment not found</span>
      </div>
    );
  }

  const status = deployment.status as keyof typeof STATUS_CONFIG;
  const isActive = ['pending', 'cloning', 'building', 'running'].includes(status);
  const isLive   = status === 'live';

  // no TLS on .localhost
  const protocol = project?.domain?.includes('localhost') ? 'http' : 'https';
  const deploymentUrl = project?.domain
    ? (project.domain.startsWith('http') ? project.domain : `${protocol}://${project.domain}`)
    : undefined;

  return (
    <div className="-m-6 flex h-[calc(100vh-4rem)] flex-col gap-0 overflow-hidden" style={{ height: 'calc(100vh - 4rem)' }}>

      {/* ── Top bar ── */}
      <div className="flex shrink-0 items-center gap-4 border-b border-white/[0.06] bg-gray-950 px-6 py-3">
        <Link
          to="/projects/$projectId"
          params={{ projectId: deployment.project_id }}
          className="flex items-center gap-1.5 text-sm text-gray-500 transition-colors hover:text-gray-300"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to project
        </Link>
        <span className="text-gray-700">/</span>
        <span className="font-mono text-sm text-gray-400">
          {deployment.id.slice(0, 8)}
        </span>
        <div className="ml-auto flex items-center gap-3">
          <StatusBadge status={status} />
          {isLive && deploymentUrl && (
            <>
              <button
                onClick={() => stopMutation.mutate()}
                disabled={stopMutation.isPending}
                className="flex items-center gap-1.5 rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-1.5 text-sm font-medium text-red-400 transition hover:bg-red-500/20 disabled:opacity-50"
              >
                {stopMutation.isPending ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Square className="h-3.5 w-3.5" />}
                {stopMutation.isPending ? 'Stopping…' : 'Stop'}
              </button>
              <a
                href={deploymentUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-1.5 rounded-lg bg-indigo-600 px-3 py-1.5 text-sm font-medium text-white transition hover:bg-indigo-500"
              >
                <Globe className="h-3.5 w-3.5" />
                Open App
                <ExternalLink className="h-3 w-3 opacity-70" />
              </a>
            </>
          )}
          {status === 'failed' && (
            <div className="flex items-center gap-1.5 rounded-lg border border-red-500/20 bg-red-500/10 px-3 py-1.5 text-sm text-red-400">
              <XCircle className="h-3.5 w-3.5" /> Deployment Failed
            </div>
          )}
        </div>
      </div>

      {/* ── Body: two columns ── */}
      <div className="flex min-h-0 flex-1 gap-0">

        {/* Left sidebar — metadata */}
        <aside className="flex w-72 shrink-0 flex-col gap-4 overflow-y-auto border-r border-white/[0.06] bg-gray-950 p-5">

          {/* Live banner */}
          {isLive && deploymentUrl && (
            <div className="flex flex-col gap-2 rounded-xl border border-green-500/20 bg-green-500/5 p-4">
              <div className="flex items-center gap-2 text-sm font-semibold text-green-400">
                <Activity className="h-4 w-4" /> Live
              </div>
              <a
                href={deploymentUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="break-all font-mono text-xs text-green-300 underline underline-offset-2 opacity-80 hover:opacity-100"
              >
                {deploymentUrl}
              </a>
            </div>
          )}

          {/* Commit */}
          <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-4">
            <div className="mb-3 flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-gray-600">
              <GitCommit className="h-3.5 w-3.5" /> Commit
            </div>
            <div className="flex flex-col gap-2.5">
              <InfoCard label="SHA" value={deployment.commit_sha.slice(0, 7)} mono copyable />
              {deployment.triggered_by && (
                <InfoCard label="Triggered by" value={deployment.triggered_by} />
              )}
            </div>
          </div>

          {/* Timing */}
          <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-4">
            <div className="mb-3 flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-gray-600">
              <Clock className="h-3.5 w-3.5" /> Timing
            </div>
            <div className="flex flex-col gap-2.5">
              <InfoCard label="Created" value={formatDate(deployment.created_at)} />
              <InfoCard label="Last updated" value={timeAgo(deployment.updated_at)} />
            </div>
          </div>

          {/* Container */}
          {deployment.container_id && (
            <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-4">
              <div className="mb-3 flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-gray-600">
                <Server className="h-3.5 w-3.5" /> Container
              </div>
              <div className="flex flex-col gap-2.5">
                <InfoCard label="ID" value={deployment.container_id.slice(0, 12)} mono copyable />
                <InfoCard label="Name" value={deployment.container_name} mono copyable />
                {deployment.docker_image && (
                  <InfoCard label="Image" value={deployment.docker_image} mono copyable />
                )}
                {deployment.memory_limit_mb > 0 && (
                  <InfoCard label="Memory limit" value={`${deployment.memory_limit_mb} MB`} />
                )}
              </div>
            </div>
          )}

          {/* Deployment ID */}
          <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-4">
            <div className="mb-3 flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-gray-600">
              <HardDrive className="h-3.5 w-3.5" /> Deployment
            </div>
            <InfoCard label="ID" value={deployment.id} mono copyable />
          </div>
        </aside>

        {/* Right — log panel */}
        <main className="relative min-w-0 flex-1 bg-[#0d1117] p-4">
          <LogPanel
            logs={logs}
            isStreaming={isActive}
            connectionStatus={connectionStatus}
          />
          {/* Always mount — replays Redis history even for live/failed deployments */}
          <LogStreamer
            deploymentId={deployment.id}
            onLog={handleLog}
            onStatusChange={handleStatusChange}
          />
        </main>
      </div>
    </div>
  );
}
