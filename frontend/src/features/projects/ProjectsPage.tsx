import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate } from '@tanstack/react-router';
import { deploymentsApi, projectsApi } from '../../api/endpoints';
import type { Project } from '../../types';
import { FolderGit2, ExternalLink, Loader2, Rocket } from 'lucide-react';
import { Skeleton } from '../../components/ui/skeleton';
import { ProjectCreateForm } from './ProjectCreateForm';

function deploymentUrl(domain: string) {
  if (domain.startsWith('http://') || domain.startsWith('https://')) {
    return domain;
  }

  // no TLS on localhost, just use plain HTTP
  const protocol = domain.includes('localhost') ? 'http' : 'https';
  return `${protocol}://${domain}`;
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
    <div className="border-border bg-card hover:border-primary/50 rounded-lg border p-6 transition-colors">
      <div className="flex items-start justify-between gap-4">
        <Link
          to="/projects/$projectId"
          params={{ projectId: project.id }}
          className="min-w-0 flex-1"
        >
          <h3 className="text-foreground hover:text-primary font-semibold">{project.name}</h3>
          <p className="text-muted-foreground mt-1 truncate text-sm">
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
          <span className="bg-secondary rounded px-2 py-0.5 font-mono text-xs">
            {project.branch}
          </span>
        </div>
        <div className="flex items-center gap-2 text-sm">
          <span className="text-muted-foreground">Domain:</span>
          <a
            href={deploymentUrl(project.domain)}
            target="_blank"
            rel="noopener noreferrer"
            className="text-foreground hover:text-primary truncate font-mono text-xs hover:underline"
          >
            {project.domain}
          </a>
        </div>
      </div>

      {deployError && (
        <div className="bg-destructive/10 border-destructive/20 text-destructive mt-4 rounded-md border p-3 text-xs">
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
          className="bg-secondary text-secondary-foreground hover:bg-secondary/80 flex flex-1 items-center justify-center gap-2 rounded-md px-3 py-1.5 text-xs font-medium disabled:cursor-not-allowed disabled:opacity-50"
        >
          {deployMutation.isPending && <Loader2 className="h-3 w-3 animate-spin" />}
          {deployMutation.isPending ? 'Deploying...' : 'Deploy'}
        </button>
        <Link
          to="/projects/$projectId"
          params={{ projectId: project.id }}
          className="border-border text-foreground hover:bg-accent flex-1 rounded-md border px-3 py-1.5 text-center text-xs font-medium"
        >
          Settings
        </Link>
      </div>
    </div>
  );
}

export function ProjectsPage() {
  const [showCreateForm, setShowCreateForm] = useState(false);

  const {
    data: projects,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['projects'],
    queryFn: projectsApi.list,
  });

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <Skeleton className="h-10 w-48" />
          <Skeleton className="h-10 w-32" />
        </div>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {[...Array(3)].map((_, i) => (
            <Skeleton key={i} className="h-48 w-full rounded-xl" />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return <div className="text-destructive">Error loading projects</div>;
  }

  if (!projects || projects.length === 0) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-foreground text-3xl font-bold">Projects</h1>
          <button
            onClick={() => setShowCreateForm(true)}
            className="bg-primary text-primary-foreground hover:bg-primary/90 rounded-md px-4 py-2 text-sm font-medium"
          >
            New Project
          </button>
        </div>
        <div className="border-border rounded-lg border border-dashed p-12 text-center">
          <FolderGit2 className="text-muted-foreground mx-auto h-12 w-12" />
          <h3 className="text-foreground mt-4 text-lg font-semibold">No projects yet</h3>
          <p className="text-muted-foreground mt-2 text-sm">
            Create your first project to get started
          </p>
        </div>

        {showCreateForm && <ProjectCreateForm onClose={() => setShowCreateForm(false)} />}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-foreground text-3xl font-bold">Projects</h1>
        <button
          onClick={() => setShowCreateForm(true)}
          className="bg-primary text-primary-foreground hover:bg-primary/90 rounded-md px-4 py-2 text-sm font-medium"
        >
          New Project
        </button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {projects.map((project: Project) => (
          <ProjectCard key={project.id} project={project} />
        ))}
      </div>

      {showCreateForm && <ProjectCreateForm onClose={() => setShowCreateForm(false)} />}
    </div>
  );
}
