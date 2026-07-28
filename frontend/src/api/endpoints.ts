import { api } from './client';
import type { User, Project, Deployment, EnvVar, WSTicket } from '../types';

export const authApi = {
  me: () => api.get<User>('/auth/me'),
  wsTicket: () => api.post<WSTicket>('/auth/ws-ticket'),
  logout: () => api.get<void>('/auth/logout'),
};

export const projectsApi = {
  list: () => api.get<Project[]>('/api/v1/projects'),
  get: (id: string) => api.get<Project>(`/api/v1/projects/${id}`),
  create: (data: { name: string; repo_url: string; branch: string }) =>
    api.post<Project>('/api/v1/projects', data),
  update: (id: string, data: Partial<Project>) =>
    api.put<Project>(`/api/v1/projects/${id}`, data),
  delete: (id: string) => api.delete<void>(`/api/v1/projects/${id}`),
};

export const deploymentsApi = {
  list: (projectId: string) =>
    api.get<Deployment[]>(`/api/v1/projects/${projectId}/deployments`),
  get: (id: string) => api.get<Deployment>(`/api/v1/deployments/${id}`),
  trigger: (projectId: string, commitSha?: string) =>
    api.post<Deployment>(`/api/v1/projects/${projectId}/deploy`, {
      commit_sha: commitSha,
    }),
};

export const envVarsApi = {
  list: (projectId: string) =>
    api.get<EnvVar[]>(`/api/v1/projects/${projectId}/env`),
  create: (projectId: string, data: { key: string; value: string }) =>
    api.post<EnvVar>(`/api/v1/projects/${projectId}/env`, data),
  delete: (projectId: string, envVarId: string) =>
    api.delete<void>(`/api/v1/projects/${projectId}/env/${envVarId}`),
};
