import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import { projectsApi, envVarsApi } from '../../api/endpoints';
import { ConfirmModal } from '../../components/ui/ConfirmModal';
import { EnvVarsManager } from '../env-vars/EnvVarsManager';
import type { Project } from '../../types';

interface ProjectSettingsTabProps {
  project: Project;
}

export function ProjectSettingsTab({ project }: ProjectSettingsTabProps) {
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  const { data: envVars } = useQuery({
    queryKey: ['envVars', project.id],
    queryFn: () => envVarsApi.list(project.id),
  });

  const portEnvVar = envVars?.find((e) => e.key === 'PORT');
  const defaultPort = portEnvVar ? portEnvVar.value : '8080';

  const [editForm, setEditForm] = useState({
    name: project.name,
    branch: project.branch,
    port: defaultPort,
  });
  const [editError, setEditError] = useState<string | null>(null);

  // Sync port state when envVars load
  useEffect(() => {
    if (portEnvVar && editForm.port !== portEnvVar.value) {
      setEditForm(prev => ({ ...prev, port: portEnvVar.value }));
    }
  }, [portEnvVar]);

  const updateMutation = useMutation({
    mutationFn: async (data: typeof editForm) => {
      // Update Project (name, branch)
      await projectsApi.update(project.id, { name: data.name, branch: data.branch });
      
      // Update or Create PORT env var
      if (portEnvVar) {
        if (portEnvVar.value !== data.port) {
          await envVarsApi.update(project.id, portEnvVar.id, { key: 'PORT', value: data.port });
        }
      } else {
        await envVarsApi.create(project.id, { key: 'PORT', value: data.port });
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', project.id] });
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      queryClient.invalidateQueries({ queryKey: ['envVars', project.id] });
      setEditError(null);
    },
    onError: (error: Error) => {
      setEditError(error.message);
    },
  });

  const handleUpdate = (e: React.FormEvent) => {
    e.preventDefault();
    if (!editForm.name.trim() || !editForm.branch.trim()) {
      setEditError('Name and branch are required');
      return;
    }
    updateMutation.mutate(editForm);
  };

  const deleteMutation = useMutation({
    mutationFn: () => projectsApi.delete(project.id),
    onSuccess: () => {
      // Navigate first to unmount the detail page
      navigate({ to: '/', replace: true });
      
      // Delay cache invalidation to prevent the unmounting component from 
      // triggering a refetch that hits a 404 and throws an error toast
      setTimeout(() => {
        queryClient.removeQueries({ queryKey: ['project', project.id] });
        queryClient.invalidateQueries({ queryKey: ['projects'] });
      }, 10);
    },
  });

  const isFormDirty = 
    editForm.name !== project.name || 
    editForm.branch !== project.branch ||
    editForm.port !== defaultPort;

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <div className="border-border bg-card rounded-lg border p-6">
        <h2 className="text-foreground mb-4 text-xl font-semibold">General Settings</h2>
        <form onSubmit={handleUpdate} className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label className="text-foreground mb-1 block text-sm font-medium">Project Name</label>
              <input
                type="text"
                value={editForm.name}
                onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
                className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-primary focus:ring-primary w-full rounded-md border px-3 py-2 text-sm focus:ring-1 focus:outline-none"
              />
            </div>
            <div>
              <label className="text-foreground mb-1 block text-sm font-medium">Branch</label>
              <input
                type="text"
                value={editForm.branch}
                onChange={(e) => setEditForm({ ...editForm, branch: e.target.value })}
                className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-primary focus:ring-primary w-full rounded-md border px-3 py-2 text-sm focus:ring-1 focus:outline-none"
              />
            </div>
          </div>
          <div>
            <label className="text-foreground mb-1 block text-sm font-medium">Application Port</label>
            <input
              type="text"
              value={editForm.port}
              onChange={(e) => setEditForm({ ...editForm, port: e.target.value })}
              className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-primary focus:ring-primary w-full rounded-md border px-3 py-2 text-sm focus:ring-1 focus:outline-none"
              placeholder="8080"
            />
            <p className="text-muted-foreground mt-1 text-xs">
              The internal port your application listens on. (Default: 8080)
            </p>
          </div>
          
          {editError && (
            <p className="text-destructive text-sm">{editError}</p>
          )}

          <div className="flex justify-end">
            <button
              type="submit"
              disabled={!isFormDirty || updateMutation.isPending}
              className="bg-primary text-primary-foreground hover:bg-primary/90 rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
            >
              {updateMutation.isPending ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>
      <div className="border-border bg-card rounded-lg border p-6">
        <h2 className="text-foreground mb-4 text-xl font-semibold">Environment Variables</h2>
        <p className="text-muted-foreground mb-4 text-sm">
          Variables configured here are automatically injected into your container during deployment.
        </p>
        <EnvVarsManager projectId={project.id} />
      </div>

      <div className="border-destructive/20 bg-destructive/5 rounded-lg border p-6">
        <h2 className="text-destructive mb-2 text-xl font-semibold">Danger Zone</h2>
        <p className="text-muted-foreground mb-4 text-sm">
          Once you delete a project, there is no going back. Please be certain.
        </p>
        <button
          onClick={() => setShowDeleteModal(true)}
          className="border-destructive text-destructive hover:bg-destructive hover:text-destructive-foreground rounded-md border px-4 py-2 text-sm font-medium transition-colors"
        >
          Delete Project
        </button>
      </div>

      {showDeleteModal && (
        <ConfirmModal
          title="Delete Project"
          description="Are you sure you want to delete this project? This will immediately stop and remove all associated containers, and permanently delete all deployment history and environment variables."
          confirmText="Yes, delete project"
          isDestructive={true}
          isLoading={deleteMutation.isPending}
          onConfirm={() => deleteMutation.mutate()}
          onCancel={() => setShowDeleteModal(false)}
        />
      )}
    </div>
  );
}
