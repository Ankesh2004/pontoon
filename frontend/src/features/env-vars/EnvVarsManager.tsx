import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { envVarsApi } from '../../api/endpoints';
import type { EnvVar } from '../../types';
import { Eye, EyeOff, Plus, Trash2 } from 'lucide-react';

interface EnvVarsManagerProps {
  projectId: string;
}

export function EnvVarsManager({ projectId }: EnvVarsManagerProps) {
  const queryClient = useQueryClient();
  const [showValues, setShowValues] = useState<Record<string, boolean>>({});
  const [newKey, setNewKey] = useState('');
  const [newValue, setNewValue] = useState('');
  const [showAddForm, setShowAddForm] = useState(false);

  const { data: envVars, isLoading } = useQuery({
    queryKey: ['envVars', projectId],
    queryFn: () => envVarsApi.list(projectId),
  });

  const createMutation = useMutation({
    mutationFn: (data: { key: string; value: string }) =>
      envVarsApi.create(projectId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['envVars', projectId] });
      setNewKey('');
      setNewValue('');
      setShowAddForm(false);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (envVarId: string) => envVarsApi.delete(projectId, envVarId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['envVars', projectId] });
    },
  });

  const handleAdd = (e: React.FormEvent) => {
    e.preventDefault();
    if (newKey.trim() && newValue.trim()) {
      createMutation.mutate({ key: newKey, value: newValue });
    }
  };

  const toggleValueVisibility = (id: string) => {
    setShowValues((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  if (isLoading) {
    return <div className="text-muted-foreground">Loading environment variables...</div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-foreground">
          Environment Variables
        </h3>
        <button
          onClick={() => setShowAddForm(!showAddForm)}
          className="flex items-center gap-2 rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          <Plus className="h-4 w-4" />
          Add Variable
        </button>
      </div>

      {showAddForm && (
        <form
          onSubmit={handleAdd}
          className="rounded-lg border border-border bg-card p-4"
        >
          <div className="grid gap-3 md:grid-cols-2">
            <div>
              <label className="mb-1 block text-sm font-medium text-foreground">
                Key
              </label>
              <input
                type="text"
                value={newKey}
                onChange={(e) => setNewKey(e.target.value.toUpperCase())}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                placeholder="DATABASE_URL"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-foreground">
                Value
              </label>
              <input
                type="text"
                value={newValue}
                onChange={(e) => setNewValue(e.target.value)}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                placeholder="postgres://..."
              />
            </div>
          </div>
          <div className="mt-3 flex gap-2">
            <button
              type="submit"
              disabled={createMutation.isPending}
              className="rounded-md bg-primary px-4 py-1.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
            >
              {createMutation.isPending ? 'Adding...' : 'Add'}
            </button>
            <button
              type="button"
              onClick={() => {
                setShowAddForm(false);
                setNewKey('');
                setNewValue('');
              }}
              className="rounded-md border border-border px-4 py-1.5 text-sm font-medium text-foreground hover:bg-accent"
            >
              Cancel
            </button>
          </div>
        </form>
      )}

      {envVars && envVars.length > 0 ? (
        <div className="space-y-2">
          {envVars.map((envVar: EnvVar) => (
            <div
              key={envVar.id}
              className="flex items-center gap-3 rounded-lg border border-border bg-card p-3"
            >
              <div className="flex-1 font-mono text-sm">
                <span className="font-semibold text-foreground">
                  {envVar.key}
                </span>
                <span className="mx-2 text-muted-foreground">=</span>
                <span className="text-muted-foreground">
                  {showValues[envVar.id]
                    ? envVar.value
                    : '••••••••••••••••'}
                </span>
              </div>
              <button
                onClick={() => toggleValueVisibility(envVar.id)}
                className="text-muted-foreground hover:text-foreground"
                title={showValues[envVar.id] ? 'Hide value' : 'Show value'}
              >
                {showValues[envVar.id] ? (
                  <EyeOff className="h-4 w-4" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
              </button>
              <button
                onClick={() => deleteMutation.mutate(envVar.id)}
                disabled={deleteMutation.isPending}
                className="text-muted-foreground hover:text-destructive disabled:opacity-50"
                title="Delete variable"
              >
                <Trash2 className="h-4 w-4" />
              </button>
            </div>
          ))}
        </div>
      ) : (
        <div className="rounded-lg border border-dashed border-border p-8 text-center">
          <p className="text-sm text-muted-foreground">
            No environment variables configured
          </p>
        </div>
      )}
    </div>
  );
}
