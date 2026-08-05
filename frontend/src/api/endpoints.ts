import { api, fetchCsrfToken } from './client';
import type { User, Project, Deployment, EnvVar, WSTicket } from '../types';

export const authApi = {
  me: async () => {
    await fetchCsrfToken();
    return api.get<User>('/auth/me');
  },
  wsTicket: () => api.post<WSTicket>('/auth/ws-ticket'),
  logout: () => api.get<void>('/auth/logout'),
};

export const projectsApi = {
  list: () => api.get<Project[]>('/api/v1/projects'),
  get: (id: string) => api.get<Project>(`/api/v1/projects/${id}`),
  create: (data: { name: string; repo_url: string; branch: string }) =>
    api.post<Project>('/api/v1/projects', data),
  update: (id: string, data: Partial<Project>) => api.put<Project>(`/api/v1/projects/${id}`, data),
  delete: (id: string) => api.delete<void>(`/api/v1/projects/${id}`),
};

export const deploymentsApi = {
  list: (projectId: string) => api.get<Deployment[]>(`/api/v1/projects/${projectId}/deployments`),
  get: (id: string) => api.get<Deployment>(`/api/v1/deployments/${id}`),
  getLogs: (id: string) => api.getText(`/api/v1/deployments/${id}/logs`),
  trigger: (projectId: string, commitSha?: string) =>
    api.post<Deployment>(`/api/v1/projects/${projectId}/deploy`, {
      commit_sha: commitSha,
    }),
  stop: (id: string) => api.post<{ status: string }>(`/api/v1/deployments/${id}/stop`),
  delete: (id: string) => api.delete<void>(`/api/v1/deployments/${id}`),
};

export const envVarsApi = {
  list: (projectId: string) => api.get<EnvVar[]>(`/api/v1/projects/${projectId}/env`),
  create: (projectId: string, data: { key: string; value: string }) =>
    api.post<EnvVar>(`/api/v1/projects/${projectId}/env`, data),
  update: (projectId: string, envVarId: string, data: { key: string; value: string }) =>
    api.put<EnvVar>(`/api/v1/projects/${projectId}/env/${envVarId}`, data),
  delete: (projectId: string, envVarId: string) =>
    api.delete<void>(`/api/v1/projects/${projectId}/env/${envVarId}`),
};

export interface AIPipelineContext {
  pipeline_id: string;
  project_id: string;
  deployment_id: string;
  parsed_error?: string;
  root_cause?: string;
  proposed_patch?: string;
  security_passed: boolean;
  confidence_score?: number;
}

export const aiPipelinesApi = {
  recover: (deploymentId: string, projectId: string, rawLogs: string) => 
    api.post<{ pipeline_id: string; status: string }>(`/api/v1/deployments/${deploymentId}/recover`, {
      project_id: projectId,
      raw_logs: rawLogs,
    }),
  get: (pipelineId: string) => api.get<AIPipelineContext>(`/api/v1/pipelines/${pipelineId}`),
  approve: (pipelineId: string) => api.post<{ status: string }>(`/api/v1/pipelines/${pipelineId}/approve`),
  reject: (pipelineId: string) => api.post<{ status: string }>(`/api/v1/pipelines/${pipelineId}/reject`),
};
