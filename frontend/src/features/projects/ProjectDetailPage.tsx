import { useQuery, useMutation } from '@tanstack/react-query';
import { useParams, useNavigate } from '@tanstack/react-router';
import { projectsApi, deploymentsApi } from '../../api/endpoints';
import { EnvVarsManager } from '../env-vars/EnvVarsManager';
import { ExternalLink, Loader2 } from 'lucide-react';
import { useState } from 'react';

export function ProjectDetailPage() {
  const { projectId } = useParams({ strict: false }) as { projectId: string };
  const navigate = useNavigate();
  const [deployError, setDeployError] = useState<string | null>(null);

  const {
    data: project,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => projectsApi.get(projectId),
    enabled: !!projectId,
  });

  const deployMutation = useMutation({
    mutationFn: () => deploymentsApi.trigger(projectId),
    onSuccess: (deployment) => {
      navigate({ to: `/deployments/${deployment.id}` });
    },
    onError: (error: Error) => {
      setDeployError(error.message);
    },
  });

  const handleDeploy = () => {
    setDeployError(null);
    deployMutation.mutate();
  };

  if (isLoading) {
    return <div className="text-muted-foreground">Loading project...</div>;
  }

  if (error || !project) {
    return <div className="text-destructive">Error loading project</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-foreground text-3xl font-bold">{project.name}</h1>
          <div className="text-muted-foreground mt-2 flex items-center gap-4 text-sm">
            <a
              href={project.repo_url}
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-foreground flex items-center gap-1"
            >
              {project.repo_owner}/{project.repo_name}
              <ExternalLink className="h-3 w-3" />
            </a>
            <span>•</span>
            <span>Branch: {project.branch}</span>
            <span>•</span>
            <span className="font-mono">{project.domain}</span>
          </div>
        </div>
        <button
          onClick={handleDeploy}
          disabled={deployMutation.isPending}
          className="bg-primary text-primary-foreground hover:bg-primary/90 flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium disabled:cursor-not-allowed disabled:opacity-50"
        >
          {deployMutation.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
          {deployMutation.isPending ? 'Deploying...' : 'Deploy Now'}
        </button>
      </div>

      {deployError && (
        <div className="bg-destructive/10 border-destructive/20 text-destructive rounded-md border p-4 text-sm">
          {deployError}
        </div>
      )}

      <div className="border-border bg-card rounded-lg border p-6">
        <h2 className="text-foreground mb-4 text-xl font-semibold">Configuration</h2>
        <EnvVarsManager projectId={project.id} />
      </div>

      <div className="border-border bg-card rounded-lg border p-6">
        <h2 className="text-foreground mb-4 text-xl font-semibold">Webhook</h2>
        <div className="space-y-3">
          <div>
            <label className="text-foreground mb-1 block text-sm font-medium">Webhook URL</label>
            <div className="flex gap-2">
              <input
                type="text"
                readOnly
                value={`${window.location.origin}/webhooks/github?project_id=${project.id}`}
                className="border-input bg-secondary text-foreground flex-1 rounded-md border px-3 py-2 font-mono text-sm"
              />
              <button
                onClick={() =>
                  navigator.clipboard.writeText(
                    `${window.location.origin}/webhooks/github?project_id=${project.id}`
                  )
                }
                className="border-border text-foreground hover:bg-accent rounded-md border px-4 py-2 text-sm font-medium"
              >
                Copy
              </button>
            </div>
          </div>
          <div>
            <label className="text-foreground mb-1 block text-sm font-medium">Webhook Secret</label>
            <div className="flex gap-2">
              <input
                type="text"
                readOnly
                value={project.webhook_secret}
                className="border-input bg-secondary text-foreground flex-1 rounded-md border px-3 py-2 font-mono text-sm"
              />
              <button
                onClick={() => navigator.clipboard.writeText(project.webhook_secret)}
                className="border-border text-foreground hover:bg-accent rounded-md border px-4 py-2 text-sm font-medium"
              >
                Copy
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
