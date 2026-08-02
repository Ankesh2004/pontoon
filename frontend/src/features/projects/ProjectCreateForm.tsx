import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { projectsApi } from '../../api/endpoints';
import { X } from 'lucide-react';

interface ProjectCreateFormProps {
  onClose: () => void;
}

export function ProjectCreateForm({ onClose }: ProjectCreateFormProps) {
  const queryClient = useQueryClient();
  const [formData, setFormData] = useState({
    name: '',
    repo_url: '',
    branch: 'main',
    port: '8080',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const createMutation = useMutation({
    mutationFn: projectsApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      onClose();
    },
    onError: (error: Error) => {
      setErrors({ submit: error.message });
    },
  });

  const validate = () => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Project name is required';
    } else if (!/^[a-z0-9-]+$/.test(formData.name)) {
      newErrors.name = 'Project name must contain only lowercase letters, numbers, and hyphens';
    }

    if (!formData.repo_url.trim()) {
      newErrors.repo_url = 'Repository URL is required';
    } else if (!/^https:\/\/github\.com\/[^/]+\/[^/]+/.test(formData.repo_url)) {
      newErrors.repo_url = 'Invalid GitHub repository URL';
    }

    if (!formData.branch.trim()) {
      newErrors.branch = 'Branch is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (validate()) {
      createMutation.mutate(formData);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="border-border bg-card w-full max-w-md rounded-lg border p-6 shadow-lg">
        <div className="mb-6 flex items-center justify-between">
          <h2 className="text-foreground text-2xl font-bold">Create Project</h2>
          <button onClick={onClose} className="text-muted-foreground hover:text-foreground">
            <X className="h-5 w-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="text-foreground mb-2 block text-sm font-medium">Project Name</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-primary focus:ring-primary w-full rounded-md border px-3 py-2 text-sm focus:ring-1 focus:outline-none"
              placeholder="my-project"
            />
            {errors.name && <p className="text-destructive mt-1 text-sm">{errors.name}</p>}
          </div>

          <div>
            <label className="text-foreground mb-2 block text-sm font-medium">
              GitHub Repository URL
            </label>
            <input
              type="text"
              value={formData.repo_url}
              onChange={(e) => setFormData({ ...formData, repo_url: e.target.value })}
              className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-primary focus:ring-primary w-full rounded-md border px-3 py-2 text-sm focus:ring-1 focus:outline-none"
              placeholder="https://github.com/username/repo"
            />
            {errors.repo_url && <p className="text-destructive mt-1 text-sm">{errors.repo_url}</p>}
          </div>

          <div>
            <label className="text-foreground mb-2 block text-sm font-medium">Branch</label>
            <input
              type="text"
              value={formData.branch}
              onChange={(e) => setFormData({ ...formData, branch: e.target.value })}
              className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-primary focus:ring-primary w-full rounded-md border px-3 py-2 text-sm focus:ring-1 focus:outline-none"
              placeholder="main"
            />
            {errors.branch && <p className="text-destructive mt-1 text-sm">{errors.branch}</p>}
          </div>

          <div>
            <label className="text-foreground mb-2 block text-sm font-medium">
              Port
              <span className="text-muted-foreground ml-2 font-normal text-xs" title="The port your application listens on. We will route external traffic to this port and inject it into your container as $PORT. Defaults to 8080.">
                (Hover for info)
              </span>
            </label>
            <input
              type="text"
              value={formData.port}
              onChange={(e) => setFormData({ ...formData, port: e.target.value })}
              className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-primary focus:ring-primary w-full rounded-md border px-3 py-2 text-sm focus:ring-1 focus:outline-none"
              placeholder="8080"
            />
          </div>

          {errors.submit && (
            <div className="bg-destructive/10 text-destructive rounded-md p-3 text-sm">
              {errors.submit}
            </div>
          )}

          <div className="flex gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="border-border text-foreground hover:bg-accent flex-1 rounded-md border px-4 py-2 text-sm font-medium"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={createMutation.isPending}
              className="bg-primary text-primary-foreground hover:bg-primary/90 flex-1 rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
            >
              {createMutation.isPending ? 'Creating...' : 'Create Project'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
