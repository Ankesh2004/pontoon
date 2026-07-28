import { useQuery } from '@tanstack/react-query';
import { useParams } from '@tanstack/react-router';
import { projectsApi } from '../../api/endpoints';
import { EnvVarsManager } from '../env-vars/EnvVarsManager';
import { ExternalLink } from 'lucide-react';

export function ProjectDetailPage() {
  const { projectId } = useParams({ strict: false }) as { projectId: string };

  const { data: project, isLoading, error } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => projectsApi.get(projectId),
    enabled: !!projectId,
  });

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
          <h1 className="text-3xl font-bold text-foreground">{project.name}</h1>
          <div className="mt-2 flex items-center gap-4 text-sm text-muted-foreground">
            <a
              href={project.repo_url}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1 hover:text-foreground"
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
        <button className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
          Deploy Now
        </button>
      </div>

      <div className="rounded-lg border border-border bg-card p-6">
        <h2 className="mb-4 text-xl font-semibold text-foreground">
          Configuration
        </h2>
        <EnvVarsManager projectId={project.id} />
      </div>

      <div className="rounded-lg border border-border bg-card p-6">
        <h2 className="mb-4 text-xl font-semibold text-foreground">
          Webhook
        </h2>
        <div className="space-y-3">
          <div>
            <label className="mb-1 block text-sm font-medium text-foreground">
              Webhook URL
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                readOnly
                value={`${window.location.origin}/webhooks/github?project_id=${project.id}`}
                className="flex-1 rounded-md border border-input bg-secondary px-3 py-2 text-sm font-mono text-foreground"
              />
              <button
                onClick={() =>
                  navigator.clipboard.writeText(
                    `${window.location.origin}/webhooks/github?project_id=${project.id}`
                  )
                }
                className="rounded-md border border-border px-4 py-2 text-sm font-medium text-foreground hover:bg-accent"
              >
                Copy
              </button>
            </div>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-foreground">
              Webhook Secret
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                readOnly
                value={project.webhook_secret}
                className="flex-1 rounded-md border border-input bg-secondary px-3 py-2 text-sm font-mono text-foreground"
              />
              <button
                onClick={() =>
                  navigator.clipboard.writeText(project.webhook_secret)
                }
                className="rounded-md border border-border px-4 py-2 text-sm font-medium text-foreground hover:bg-accent"
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
