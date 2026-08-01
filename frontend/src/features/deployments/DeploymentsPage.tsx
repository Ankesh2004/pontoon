import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { deploymentsApi } from '../../api/endpoints';
import type { Deployment } from '../../types';
import { Rocket, Clock, CheckCircle, XCircle, Loader2 } from 'lucide-react';

export function DeploymentsPage() {
  // Adaptive polling: check every 2s if any deployment is active, otherwise 30s
  const { data: deployments, isLoading } = useQuery({
    queryKey: ['deployments'],
    queryFn: async () => {
      // Fetch all projects first, then get deployments for each
      const response = await fetch('/api/v1/projects', { credentials: 'include' });
      const projects = await response.json();

      const allDeployments: Deployment[] = [];
      for (const project of projects) {
        const projectDeployments = await deploymentsApi.list(project.id);
        allDeployments.push(...projectDeployments);
      }

      return allDeployments.sort(
        (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      );
    },
    refetchInterval: (query) => {
      const deployments = query.state.data;
      if (!deployments) return 30000;

      const hasActive = deployments.some((d) =>
        ['pending', 'cloning', 'building', 'running'].includes(d.status)
      );

      return hasActive ? 2000 : 30000;
    },
  });

  if (isLoading) {
    return <div className="text-muted-foreground">Loading deployments...</div>;
  }

  if (!deployments || deployments.length === 0) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-foreground text-3xl font-bold">Deployments</h1>
          <p className="text-muted-foreground mt-2">
            View and manage all deployments across your projects
          </p>
        </div>

        <div className="border-border rounded-lg border border-dashed p-12 text-center">
          <Rocket className="text-muted-foreground mx-auto h-12 w-12" />
          <h3 className="text-foreground mt-4 text-lg font-semibold">No deployments yet</h3>
          <p className="text-muted-foreground mt-2 text-sm">
            Deployments will appear here once you trigger them from your projects
          </p>
        </div>
      </div>
    );
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'pending':
        return <Clock className="h-4 w-4 text-yellow-500" />;
      case 'cloning':
      case 'building':
      case 'running':
        return <Loader2 className="h-4 w-4 animate-spin text-blue-500" />;
      case 'live':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-500" />;
      default:
        return <Clock className="h-4 w-4 text-gray-500" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending':
        return 'bg-yellow-500/20 text-yellow-500';
      case 'cloning':
      case 'building':
      case 'running':
        return 'bg-blue-500/20 text-blue-500';
      case 'live':
        return 'bg-green-500/20 text-green-500';
      case 'failed':
        return 'bg-red-500/20 text-red-500';
      default:
        return 'bg-gray-500/20 text-gray-500';
    }
  };

  const isActiveStatus = (status: string) => {
    return ['pending', 'cloning', 'building', 'running'].includes(status);
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-foreground text-3xl font-bold">Deployments</h1>
        <p className="text-muted-foreground mt-2">
          View and manage all deployments across your projects
        </p>
      </div>

      <div className="space-y-3">
        {deployments.map((deployment) => (
          <Link
            key={deployment.id}
            to="/deployments/$deploymentId"
            params={{ deploymentId: deployment.id }}
            className="border-border bg-card hover:border-primary/50 block rounded-lg border p-4 transition-colors"
          >
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-3">
                  <h3 className="text-foreground font-semibold">{deployment.id.slice(0, 8)}</h3>
                  <span
                    className={`relative flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ${getStatusColor(deployment.status)}`}
                  >
                    {isActiveStatus(deployment.status) && (
                      <span
                        className={`absolute inset-0 rounded-full ${getStatusColor(deployment.status)} animate-pulse-ring`}
                      />
                    )}
                    <span className="relative flex items-center gap-1.5">
                      {getStatusIcon(deployment.status)}
                      {deployment.status}
                    </span>
                  </span>
                </div>
                <div className="text-muted-foreground mt-2 flex items-center gap-4 text-sm">
                  <span>Project: {deployment.project_id.slice(0, 8)}</span>
                  <span>Commit: {deployment.commit_sha.slice(0, 7)}</span>
                  <span>Triggered by: {deployment.triggered_by}</span>
                </div>
                <div className="text-muted-foreground mt-1 text-xs">
                  {new Date(deployment.created_at).toLocaleString()}
                </div>
              </div>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
