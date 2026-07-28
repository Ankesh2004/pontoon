import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate } from '@tanstack/react-router';
import { deploymentsApi, projectsApi } from '../../api/endpoints';
import type { Project } from '../../types';
import { FolderGit2, ExternalLink, Loader2, Rocket } from 'lucide-react';
import { ProjectCreateForm } from './ProjectCreateForm';

function deploymentUrl(domain: string) {
  if (domain.startsWith('http://') || domain.startsWith('https://')) {
    return domain;
  }

  return `https://${domain}`;
}

function ProjectCard({ project }: { project: Project }) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [deployError, setDeployError] = useState<string | null>(null);

  const deployMutation = useMutation({
    mutationFn: () => deploymentsApi.trigger(project.id),
    onSuccess: (deployment) => {
      queryClient.invalidateQueries({ queryKey: ['deployments'] });
      queryClient.invalidateQueries({ queryKey: ['project-deployments', project.id] });
      navigate({ to: '/deployments/$deploymentId', params: { deploymentId: deployment.id } });
    },
    onError: (error: Error) => {
      setDeployError(error.message);
    },
  });

  return (
    <div className="rounded-lg border border-border bg-card p-6 transition-colors hover:border-primary/50">
      <div className="flex items-start justify-between gap-4">
        <Link
          to="/projects/$projectId"
          params={{ projectId: project.id }}
          className="min-w-0 flex-1"
        >
          <h3 className="font-semibold text-foreground hover:text-primary">
            {project.name}
          </h3>
          <p className="mt-1 truncate text-sm text-muted-foreground">
            {project.repo_owner}/{project.repo_name}
          </p>
        </Link>
        <div className="flex items-center gap-2">
          <a
            href={deploymentUrl(project.domain)}
            target="_blank"
            rel="noopener noreferrer"
            title="Open deployed app"
            className="text-muted-foreground hover:text-foreground"
          >
            <Rocket className="h-4 w-4" />
          </a>
          <a
            href={project.repo_url}
            target="_blank"
            rel="noopener noreferrer"
            title="Open repository"
            className="text-muted-foreground hover:text-foreground"
          >
            <ExternalLink className="h-4 w-4" />
          </a>
        </div>
      </div>

      <div className="mt-4 space-y-2">
        <div className="flex items-center gap-2 text-sm">
          <span className="text-muted-foreground">Branch:</span>
          <span className="rounded bg-secondary px-2 py-0.5 font-mono text-xs">
            {project.branch}
          </span>
        </div>
        <div className="flex items-center gap-2 text-sm">
          <span className="text-muted-foreground">Domain:</span>
          <a
            href={deploymentUrl(project.domain)}
            target="_blank"
            rel="noopener noreferrer"
            className="truncate font-mono text-xs text-foreground hover:text-primary hover:underline"
          >
            {project.domain}
          </a>
        </div>
      </div>

      {deployError && (
        <div className="mt-4 rounded-md bg-destructive/10 border border-destructive/20 p-3 text-xs text-destructive">
          {deployError}
        </div>
      )}

      <div className="mt-4 flex gap-2">
        <button
          onClick={() => {
            setDeployError(null);
            deployMutation.mutate();
          }}
          disabled={deployMutation.isPending}
          className="flex flex-1 items-center justify-center gap-2 rounded-md bg-secondary px-3 py-1.5 text-xs font-medium text-secondary-foreground hover:bg-secondary/80 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {deployMutation.isPending && <Loader2 className="h-3 w-3 animate-spin" />}
          {deployMutation.isPending ? 'Deploying...' : 'Deploy'}
        </button>
        <Link
          to="/projects/$projectId"
          params={{ projectId: project.id }}
          className="flex-1 rounded-md border border-border px-3 py-1.5 text-center text-xs font-medium text-foreground hover:bg-accent"
        >
          Settings
        </Link>
      </div>
    </div>
  );
}

export function ProjectsPage() {
  const [showCreateForm, setShowCreateForm] = useState(false);

  const { data: projects, isLoading, error } = useQuery({
    queryKey: ['projects'],
    queryFn: projectsApi.list,
  });

  if (isLoading) {
    return <div className="text-muted-foreground">Loading projects...</div>;
  }

  if (error) {
    return <div className="text-destructive">Error loading projects</div>;
  }

  if (!projects || projects.length === 0) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-foreground">Projects</h1>
          <button
            onClick={() => setShowCreateForm(true)}
            className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            New Project
          </button>
        </div>
        <div className="rounded-lg border border-dashed border-border p-12 text-center">
          <FolderGit2 className="mx-auto h-12 w-12 text-muted-foreground" />
          <h3 className="mt-4 text-lg font-semibold text-foreground">
            No projects yet
          </h3>
          <p className="mt-2 text-sm text-muted-foreground">
            Create your first project to get started
          </p>
        </div>

        {showCreateForm && (
          <ProjectCreateForm onClose={() => setShowCreateForm(false)} />
        )}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-foreground">Projects</h1>
        <button
          onClick={() => setShowCreateForm(true)}
          className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          New Project
        </button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {projects.map((project: Project) => (
          <ProjectCard key={project.id} project={project} />
        ))}
      </div>

      {showCreateForm && (
        <ProjectCreateForm onClose={() => setShowCreateForm(false)} />
      )}
    </div>
  );
}
