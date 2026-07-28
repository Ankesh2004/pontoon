export interface User {
  id: string;
  github_id: number;
  github_username: string;
  email: string;
  avatar_url: string;
  created_at: string;
  updated_at: string;
}

export interface Project {
  id: string;
  user_id: string;
  name: string;
  repo_url: string;
  repo_owner: string;
  repo_name: string;
  branch: string;
  domain: string;
  webhook_secret: string;
  created_at: string;
  updated_at: string;
}

export interface Deployment {
  id: string;
  project_id: string;
  user_id: string;
  status: 'pending' | 'cloning' | 'building' | 'running' | 'live' | 'stopped' | 'failed';
  commit_sha: string;
  docker_image: string;
  container_id: string;
  container_name: string;
  memory_limit_mb: number;
  build_logs: string;
  triggered_by: string;
  created_at: string;
  updated_at: string;
}

export interface EnvVar {
  id: string;
  project_id: string;
  key: string;
  value: string;
  created_at: string;
}

export interface WSTicket {
  ticket: string;
  expires_in: number;
}

export interface LogMessage {
  deployment_id: string;
  line: string;
  timestamp: number;
}
