import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { projectsApi } from '../../api/endpoints';
import type { Project } from '../../types';
import { FolderGit2, ExternalLink } from 'lucide-react';
import { ProjectCreateForm } from './ProjectCreateForm';

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
          <Link
            key={project.id}
            to="/projects/$projectId"
            params={{ projectId: project.id }}
            className="block rounded-lg border border-border bg-card p-6 transition-colors hover:border-primary/50"
          >
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <h3 className="font-semibold text-foreground">
                  {project.name}
                </h3>
                <p className="mt-1 text-sm text-muted-foreground">
                  {project.repo_owner}/{project.repo_name}
                </p>
              </div>
              <a
                href={project.repo_url}
                target="_blank"
                rel="noopener noreferrer"
                onClick={(e) => e.stopPropagation()}
                className="text-muted-foreground hover:text-foreground"
              >
                <ExternalLink className="h-4 w-4" />
              </a>
            </div>

            <div className="mt-4 space-y-2">
              <div className="flex items-center gap-2 text-sm">
                <span className="text-muted-foreground">Branch:</span>
                <span className="font-mono text-xs rounded bg-secondary px-2 py-0.5">
                  {project.branch}
                </span>
              </div>
              <div className="flex items-center gap-2 text-sm">
                <span className="text-muted-foreground">Domain:</span>
                <span className="font-mono text-xs text-foreground">
                  {project.domain}
                </span>
              </div>
            </div>

            <div className="mt-4 flex gap-2">
              <button
                onClick={(e) => e.preventDefault()}
                className="flex-1 rounded-md bg-secondary px-3 py-1.5 text-xs font-medium text-secondary-foreground hover:bg-secondary/80"
              >
                Deploy
              </button>
              <button
                onClick={(e) => e.preventDefault()}
                className="flex-1 rounded-md border border-border px-3 py-1.5 text-xs font-medium text-foreground hover:bg-accent"
              >
                Settings
              </button>
            </div>
          </Link>
        ))}
      </div>

      {showCreateForm && (
        <ProjectCreateForm onClose={() => setShowCreateForm(false)} />
      )}
    </div>
  );
}
