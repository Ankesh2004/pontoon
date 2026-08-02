import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { envVarsApi } from '../../api/endpoints';
import type { EnvVar } from '../../types';
import { Eye, EyeOff, Plus, Trash2, Pencil, Check, X } from 'lucide-react';

interface EnvVarsManagerProps {
  projectId: string;
}

function EnvVarRow({ envVar, deleteMutation, updateMutation }: { 
  envVar: EnvVar, 
  deleteMutation: any, 
  updateMutation: any 
}) {
  const [isEditing, setIsEditing] = useState(false);
  const [showValue, setShowValue] = useState(false);
  const [editKey, setEditKey] = useState(envVar.key);
  const [editValue, setEditValue] = useState(envVar.value);

  const handleUpdate = () => {
    if (editKey.trim() && editValue.trim()) {
      updateMutation.mutate({ envVarId: envVar.id, key: editKey, value: editValue }, {
        onSuccess: () => setIsEditing(false)
      });
    }
  };

  const handleCancel = () => {
    setEditKey(envVar.key);
    setEditValue(envVar.value);
    setIsEditing(false);
  };

  if (isEditing) {
    return (
      <div className="border-border bg-card flex items-center gap-2 rounded-lg border p-3">
        <input
          type="text"
          value={editKey}
          onChange={(e) => setEditKey(e.target.value.toUpperCase())}
          className="border-input bg-background text-foreground focus:border-primary focus:ring-primary w-1/3 rounded-md border px-2 py-1 font-mono text-sm focus:ring-1 focus:outline-none"
        />
        <span className="text-muted-foreground">=</span>
        <input
          type="text"
          value={editValue}
          onChange={(e) => setEditValue(e.target.value)}
          className="border-input bg-background text-foreground focus:border-primary focus:ring-primary flex-1 rounded-md border px-2 py-1 font-mono text-sm focus:ring-1 focus:outline-none"
        />
        <button
          onClick={handleUpdate}
          disabled={updateMutation.isPending}
          className="text-primary hover:text-primary/80 disabled:opacity-50"
          title="Save"
        >
          <Check className="h-4 w-4" />
        </button>
        <button
          onClick={handleCancel}
          disabled={updateMutation.isPending}
          className="text-muted-foreground hover:text-foreground disabled:opacity-50"
          title="Cancel"
        >
          <X className="h-4 w-4" />
        </button>
      </div>
    );
  }

  return (
    <div className="border-border bg-card flex items-center gap-3 rounded-lg border p-3">
      <div className="flex-1 font-mono text-sm">
        <span className="text-foreground font-semibold">{envVar.key}</span>
        <span className="text-muted-foreground mx-2">=</span>
        <span className="text-muted-foreground">
          {showValue ? envVar.value : '••••••••••••••••'}
        </span>
      </div>
      <button
        onClick={() => setShowValue(!showValue)}
        className="text-muted-foreground hover:text-foreground"
        title={showValue ? 'Hide value' : 'Show value'}
      >
        {showValue ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
      </button>
      <button
        onClick={() => setIsEditing(true)}
        className="text-muted-foreground hover:text-primary"
        title="Edit variable"
      >
        <Pencil className="h-4 w-4" />
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
  );
}

export function EnvVarsManager({ projectId }: EnvVarsManagerProps) {
  const queryClient = useQueryClient();
  const [newKey, setNewKey] = useState('');
  const [newValue, setNewValue] = useState('');
  const [showAddForm, setShowAddForm] = useState(false);

  const { data: envVars, isLoading } = useQuery({
    queryKey: ['envVars', projectId],
    queryFn: () => envVarsApi.list(projectId),
  });

  const createMutation = useMutation({
    mutationFn: (data: { key: string; value: string }) => envVarsApi.create(projectId, data),
    onMutate: async (newEnvVar) => {
      await queryClient.cancelQueries({ queryKey: ['envVars', projectId] });
      const previousEnvVars = queryClient.getQueryData<EnvVar[]>(['envVars', projectId]);

      const optimisticEnvVar: EnvVar = {
        id: `temp-${Date.now()}`,
        project_id: projectId,
        key: newEnvVar.key,
        value: newEnvVar.value,
        created_at: new Date().toISOString(),
      };

      queryClient.setQueryData(['envVars', projectId], (old: EnvVar[] = []) => [
        ...old,
        optimisticEnvVar,
      ]);

      setNewKey('');
      setNewValue('');
      setShowAddForm(false);

      return { previousEnvVars };
    },
    onError: (_err, _newEnvVar, context) => {
      if (context?.previousEnvVars) {
        queryClient.setQueryData(['envVars', projectId], context.previousEnvVars);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['envVars', projectId] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (envVarId: string) => envVarsApi.delete(projectId, envVarId),
    onMutate: async (deletedId) => {
      await queryClient.cancelQueries({ queryKey: ['envVars', projectId] });
      const previousEnvVars = queryClient.getQueryData<EnvVar[]>(['envVars', projectId]);

      queryClient.setQueryData(['envVars', projectId], (old: EnvVar[] = []) =>
        old.filter((env) => env.id !== deletedId)
      );

      return { previousEnvVars };
    },
    onError: (_err, _deletedId, context) => {
      if (context?.previousEnvVars) {
        queryClient.setQueryData(['envVars', projectId], context.previousEnvVars);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['envVars', projectId] });
    },
  });

  const updateMutation = useMutation({
    mutationFn: (data: { envVarId: string; key: string; value: string }) =>
      envVarsApi.update(projectId, data.envVarId, { key: data.key, value: data.value }),
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

  if (isLoading) {
    return <div className="text-muted-foreground">Loading environment variables...</div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-foreground text-lg font-semibold">Environment Variables</h3>
        <button
          onClick={() => setShowAddForm(!showAddForm)}
          className="bg-primary text-primary-foreground hover:bg-primary/90 flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium"
        >
          <Plus className="h-4 w-4" />
          Add Variable
        </button>
      </div>

      {showAddForm && (
        <form onSubmit={handleAdd} className="border-border bg-card rounded-lg border p-4">
          <div className="grid gap-3 md:grid-cols-2">
            <div>
              <label className="text-foreground mb-1 block text-sm font-medium">Key</label>
              <input
                type="text"
                value={newKey}
                onChange={(e) => setNewKey(e.target.value.toUpperCase())}
                className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-primary focus:ring-primary w-full rounded-md border px-3 py-2 font-mono text-sm focus:ring-1 focus:outline-none"
                placeholder="DATABASE_URL"
              />
            </div>
            <div>
              <label className="text-foreground mb-1 block text-sm font-medium">Value</label>
              <input
                type="text"
                value={newValue}
                onChange={(e) => setNewValue(e.target.value)}
                className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-primary focus:ring-primary w-full rounded-md border px-3 py-2 font-mono text-sm focus:ring-1 focus:outline-none"
                placeholder="postgres://..."
              />
            </div>
          </div>
          <div className="mt-3 flex gap-2">
            <button
              type="submit"
              disabled={createMutation.isPending}
              className="bg-primary text-primary-foreground hover:bg-primary/90 rounded-md px-4 py-1.5 text-sm font-medium disabled:opacity-50"
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
              className="border-border text-foreground hover:bg-accent rounded-md border px-4 py-1.5 text-sm font-medium"
            >
              Cancel
            </button>
          </div>
        </form>
      )}

      {envVars && envVars.filter(e => e.key !== 'PORT').length > 0 ? (
        <div className="space-y-2">
          {envVars.filter(e => e.key !== 'PORT').map((envVar: EnvVar) => (
            <EnvVarRow 
              key={envVar.id} 
              envVar={envVar} 
              deleteMutation={deleteMutation} 
              updateMutation={updateMutation} 
            />
          ))}
        </div>
      ) : (
        <div className="border-border rounded-lg border border-dashed p-8 text-center">
          <p className="text-muted-foreground text-sm">No environment variables configured</p>
        </div>
      )}
    </div>
  );
}
