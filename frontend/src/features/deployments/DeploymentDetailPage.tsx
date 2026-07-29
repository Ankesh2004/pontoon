import { useParams } from '@tanstack/react-router';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { deploymentsApi, projectsApi } from '../../api/endpoints';
import { useState, useCallback } from 'react';
import { LogTerminal } from '../logs/LogTerminal';
import { LogStreamer } from '../logs/LogStreamer';
import { ArrowLeft, ExternalLink, Square } from 'lucide-react';
import { Link } from '@tanstack/react-router';

export function DeploymentDetailPage() {
  const { deploymentId } = useParams({ strict: false }) as { deploymentId: string };
  const [logs, setLogs] = useState<string[]>([]);
  const [autoScroll, setAutoScroll] = useState(true);
  const [connectionStatus, setConnectionStatus] = useState<string>('connecting');
  const queryClient = useQueryClient();

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
      queryClient.invalidateQueries({ queryKey: ['deployment', deploymentId] });
    },
  });

  const handleLog = useCallback((line: string) => {
    setLogs(prev => [...prev, line]);
  }, []);

  const handleStatusChange = useCallback((status: string) => {
    setConnectionStatus(status);
  }, []);

  if (isLoading) {
    return <div className="text-muted-foreground">Loading deployment...</div>;
  }

  if (!deployment) {
    return <div className="text-destructive">Deployment not found</div>;
  }

  const statusColors = {
    pending: 'bg-yellow-500/20 text-yellow-500',
    cloning: 'bg-blue-500/20 text-blue-500',
    building: 'bg-blue-500/20 text-blue-500',
    running: 'bg-green-500/20 text-green-500',
    live: 'bg-green-500/20 text-green-500',
    stopped: 'bg-gray-500/20 text-gray-500',
    failed: 'bg-red-500/20 text-red-500',
  };

  const isActiveStatus = (status: string) => {
    return ['pending', 'cloning', 'building', 'running'].includes(status);
  };

  // no TLS on localhost, just use plain HTTP
  const protocol = project?.domain?.includes('localhost') ? 'http' : 'https';
  const deploymentUrl = project?.domain.startsWith('http://') || project?.domain.startsWith('https://')
    ? project.domain
    : project?.domain
    ? `${protocol}://${project.domain}`
    : undefined;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link
          to="/projects/$projectId"
          params={{ projectId: deployment.project_id }}
          className="text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <div className="flex-1">
          <h1 className="text-3xl font-bold text-foreground">
            Deployment {deployment.id.slice(0, 8)}
          </h1>
          <div className="mt-2 flex items-center gap-4 text-sm text-muted-foreground">
            <span className={`relative rounded-full px-3 py-1 text-xs font-medium ${statusColors[deployment.status]}`}>
              {isActiveStatus(deployment.status) && (
                <span className={`absolute inset-0 rounded-full ${statusColors[deployment.status]} animate-pulse-ring`} />
              )}
              <span className="relative">{deployment.status}</span>
            </span>
            <span>Commit: {deployment.commit_sha.slice(0, 7)}</span>
            <span>Triggered by: {deployment.triggered_by}</span>
          </div>
        </div>
        {deployment.status === 'live' && deploymentUrl && (
          <div className="flex items-center gap-2">
            <button
              onClick={() => stopMutation.mutate()}
              disabled={stopMutation.isPending}
              className="flex items-center gap-2 rounded-md border border-input bg-background px-4 py-2 text-sm font-medium hover:bg-accent hover:text-accent-foreground disabled:opacity-50"
            >
              <Square className="h-4 w-4" />
              {stopMutation.isPending ? 'Stopping...' : 'Stop'}
            </button>
            <a
              href={deploymentUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
            >
              Open App
              <ExternalLink className="h-4 w-4" />
            </a>
          </div>
        )}
      </div>

      {deployment.status === 'live' && deploymentUrl && (
        <div className="rounded-lg border border-green-500/20 bg-green-500/10 p-4 text-sm text-green-700 dark:text-green-400">
          Deployment is live at{' '}
          <a
            href={deploymentUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="font-mono underline"
          >
            {deploymentUrl}
          </a>
        </div>
      )}

      <div className="rounded-lg border border-border bg-card">
        <div className="flex items-center justify-between border-b border-border p-4">
          <div className="flex items-center gap-3">
            <h2 className="text-lg font-semibold text-foreground">Build Logs</h2>
            <span className={`rounded-full px-2 py-0.5 text-xs ${
              connectionStatus === 'connected' 
                ? 'bg-green-500/20 text-green-500'
                : connectionStatus === 'connecting'
                ? 'bg-yellow-500/20 text-yellow-500'
                : 'bg-red-500/20 text-red-500'
            }`}>
              {connectionStatus}
            </span>
          </div>
          <label className="flex items-center gap-2 text-sm text-muted-foreground">
            <input
              type="checkbox"
              checked={autoScroll}
              onChange={(e) => setAutoScroll(e.target.checked)}
              className="rounded border-border"
            />
            Auto-scroll
          </label>
        </div>
        <div className="p-4" style={{ height: '500px' }}>
          <LogTerminal logs={logs} autoScroll={autoScroll} />
          {(deployment.status === 'pending' || deployment.status === 'cloning' || deployment.status === 'building' || deployment.status === 'running') && (
            <LogStreamer
              deploymentId={deployment.id}
              onLog={handleLog}
              onStatusChange={handleStatusChange}
            />
          )}
          {deployment.build_logs && logs.length === 0 && (
            <div className="space-y-1 font-mono text-sm">
              {deployment.build_logs.split('\n').map((line, i) => (
                <div key={i} className="text-muted-foreground">{line}</div>
              ))}
            </div>
          )}
        </div>
      </div>

      {deployment.container_id && (
        <div className="rounded-lg border border-border bg-card p-6">
          <h2 className="mb-4 text-lg font-semibold text-foreground">Container Info</h2>
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label className="text-sm text-muted-foreground">Container ID</label>
              <div className="mt-1 font-mono text-sm text-foreground">{deployment.container_id}</div>
            </div>
            <div>
              <label className="text-sm text-muted-foreground">Container Name</label>
              <div className="mt-1 font-mono text-sm text-foreground">{deployment.container_name}</div>
            </div>
            <div>
              <label className="text-sm text-muted-foreground">Docker Image</label>
              <div className="mt-1 font-mono text-sm text-foreground">{deployment.docker_image}</div>
            </div>
            <div>
              <label className="text-sm text-muted-foreground">Memory Limit</label>
              <div className="mt-1 font-mono text-sm text-foreground">{deployment.memory_limit_mb} MB</div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
