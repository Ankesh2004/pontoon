import { toast } from 'sonner';

const API_BASE = '';

let csrfToken = '';

export async function fetchCsrfToken() {
  try {
    const response = await fetch(`${API_BASE}/auth/csrf`, { credentials: 'include' });
    if (response.ok) {
      const data = await response.json();
      csrfToken = data.token;
    }
  } catch (error) {
    console.error('Failed to fetch CSRF token:', error);
  }
}

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    credentials: 'include',
    headers: {
      'Content-Type': options.headers && 'Content-Type' in options.headers 
        ? (options.headers as any)['Content-Type'] 
        : 'application/json',
      'X-CSRF-Token': csrfToken,
      ...options.headers,
    },
  });

  if (!response.ok) {
    if (response.status === 401 && window.location.pathname !== '/login') {
      window.location.href = `/login?redirect=${encodeURIComponent(window.location.pathname + window.location.search)}`;
      // Halt execution so no errors are thrown while the page unloads
      await new Promise(() => {});
    }
    const error = await response.text();
    toast.error(`Error: ${response.status}`, { description: error || 'An unexpected error occurred.' });
    throw new Error(error || `HTTP ${response.status}`);
  }

  if (response.status === 204) {
    return {} as T;
  }

  return response.json();
}

export const api = {
  get: <T>(endpoint: string) => request<T>(endpoint),
  getText: async (endpoint: string): Promise<string> => {
    const response = await fetch(`${API_BASE}${endpoint}`, {
      credentials: 'include',
      headers: { 'X-CSRF-Token': csrfToken },
    });
    if (!response.ok) {
      if (response.status === 401 && window.location.pathname !== '/login') {
        window.location.href = `/login?redirect=${encodeURIComponent(window.location.pathname + window.location.search)}`;
        await new Promise(() => {});
      }
      const error = await response.text();
      toast.error(`Error: ${response.status}`, { description: error || 'An unexpected error occurred.' });
      throw new Error(error || `HTTP ${response.status}`);
    }
    return response.text();
  },

  post: <T>(endpoint: string, data?: unknown) =>
    request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    }),

  put: <T>(endpoint: string, data?: unknown) =>
    request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    }),

  delete: <T>(endpoint: string) => request<T>(endpoint, { method: 'DELETE' }),
};
