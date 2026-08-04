import { useQuery } from '@tanstack/react-query';
import { projectsApi, deploymentsApi } from '../../api/endpoints';
import { useAuth } from '../auth/useAuth';
import type { Deployment, Project } from '../../types';
import {
  Rocket,
  FolderGit2,
  Activity,
  CheckCircle2,
  XCircle,
  Clock,
  Loader2,
  ExternalLink,
  Plus,
  GitCommit,
  Zap,
} from 'lucide-react';
import { Skeleton } from '../../components/ui/skeleton';
import { Link as RouterLink } from '@tanstack/react-router';
import { Card } from '../../components/ui/card';
import { Badge } from '../../components/ui/badge';

// ── helpers ──────────────────────────────────────────────────────────────────

function timeAgo(dateStr: string) {
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}

const STATUS_STYLE: Record<string, { label: string; icon: React.ReactNode; cls: string }> = {
  live: {
    label: 'Live',
    icon: <CheckCircle2 className="h-3.5 w-3.5" />,
    cls: 'text-success bg-success/10 border-success/20',
  },
  pending: {
    label: 'Pending',
    icon: <Clock className="h-3.5 w-3.5" />,
    cls: 'text-warning bg-warning/10 border-warning/20',
  },
  cloning: {
    label: 'Cloning',
    icon: <Loader2 className="h-3.5 w-3.5 animate-spin" />,
    cls: 'text-info bg-info/10 border-info/20',
  },
  building: {
    label: 'Building',
    icon: <Loader2 className="h-3.5 w-3.5 animate-spin" />,
    cls: 'text-warning bg-warning/10 border-warning/20',
  },
  running: {
    label: 'Starting',
    icon: <Loader2 className="h-3.5 w-3.5 animate-spin" />,
    cls: 'text-info bg-info/10 border-info/20',
  },
  failed: {
    label: 'Failed',
    icon: <XCircle className="h-3.5 w-3.5" />,
    cls: 'text-destructive bg-destructive/10 border-destructive/20',
  },
  stopped: {
    label: 'Stopped',
    icon: <XCircle className="h-3.5 w-3.5" />,
    cls: 'text-muted-foreground bg-muted border-border',
  },
};

// ── stat card ─────────────────────────────────────────────────────────────────

function StatCard({
  label,
  value,

  icon,
  sub,
  accent,
}: {
  label: string;
  value: string | number;
  icon: React.ReactNode;
  sub?: string;
  accent?: string;
}) {
  return (
    <Card className="hover:border-primary/40 flex flex-col gap-3 p-5 transition">
      <div className="flex items-center justify-between">
        <span className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
          {label}
        </span>
        <span className={`rounded-lg p-1.5 ${accent ?? 'bg-muted text-muted-foreground'}`}>
          {icon}
        </span>
      </div>
      <div className="text-foreground text-3xl font-bold">{value}</div>
      {sub && <div className="text-muted-foreground text-xs">{sub}</div>}
    </Card>
  );
}

// ── deployment row ────────────────────────────────────────────────────────────

function DeploymentRow({ deployment, project }: { deployment: Deployment; project?: Project }) {
  const s = STATUS_STYLE[deployment.status] ?? STATUS_STYLE.stopped;
  return (
    <RouterLink
      to="/deployments/$deploymentId"
      params={{ deploymentId: deployment.id }}
      className="hover:border-border hover:bg-accent flex items-center gap-4 rounded-lg border border-transparent px-4 py-3 transition"
    >
      <Badge variant="outline" className={s.cls}>
        {s.icon}
        {s.label}
      </Badge>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="text-foreground truncate text-sm font-medium">
            {project?.name ?? deployment.project_id.slice(0, 8)}
          </span>
          <span className="text-muted-foreground flex items-center gap-1 font-mono text-xs">
            <GitCommit className="h-3 w-3" />
            {deployment.commit_sha.slice(0, 7)}
          </span>
        </div>
        <div className="text-muted-foreground mt-0.5 truncate font-mono text-xs">
          {deployment.id.slice(0, 8)}
        </div>
      </div>
      <span className="text-muted-foreground shrink-0 text-xs">
        {timeAgo(deployment.created_at)}
      </span>
    </RouterLink>
  );
}

// ── main page ─────────────────────────────────────────────────────────────────

export function DashboardPage() {
  const { user } = useAuth();

  const { data: projects = [], isLoading: loadingProjects } = useQuery({
    queryKey: ['projects'],
    queryFn: projectsApi.list,
  });

  // fetch deployments for all projects
  const { data: allDeployments = [], isLoading: loadingDeployments } = useQuery({
    queryKey: ['dashboard-deployments', projects.map((p) => p.id).join(',')],
    queryFn: async () => {
      const results = await Promise.all(projects.map((p) => deploymentsApi.list(p.id)));
      return results
        .flat()
        .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
    },
    enabled: projects.length > 0,
    refetchInterval: 10000,
  });

  const projectMap = Object.fromEntries(projects.map((p) => [p.id, p]));

  const liveCount = allDeployments.filter((d) => d.status === 'live').length;
  const activeCount = allDeployments.filter((d) =>
    ['pending', 'cloning', 'building', 'running'].includes(d.status)
  ).length;
  const failedCount = allDeployments.filter((d) => d.status === 'failed').length;
  const recentDeps = allDeployments.slice(0, 8);

  const isLoading = loadingProjects || loadingDeployments;

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-foreground text-2xl font-bold">
            Good{' '}
            {new Date().getHours() < 12
              ? 'morning'
              : new Date().getHours() < 18
                ? 'afternoon'
                : 'evening'}
            , <span className="text-primary">{user?.github_username}</span> 👋
          </h1>
          <p className="text-muted-foreground mt-1 text-sm">
            Here's what's happening across your projects.
          </p>
        </div>
        <RouterLink
          to="/projects"
          className="bg-primary text-primary-foreground flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition hover:opacity-90"
        >
          <Plus className="h-4 w-4" />
          New Project
        </RouterLink>
      </div>

      {/* Stat cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          label="Projects"
          value={isLoading ? '—' : projects.length}
          icon={<FolderGit2 className="h-4 w-4" />}
          sub="Total repositories"
          accent="bg-primary/10 text-primary"
        />
        <StatCard
          label="Live Containers"
          value={isLoading ? '—' : liveCount}
          icon={<Activity className="h-4 w-4" />}
          sub="Serving traffic right now"
          accent="bg-success/10 text-success"
        />
        <StatCard
          label="In Progress"
          value={isLoading ? '—' : activeCount}
          icon={<Zap className="h-4 w-4" />}
          sub={activeCount > 0 ? 'Building or starting…' : 'No active builds'}
          accent="bg-warning/10 text-warning"
        />
        <StatCard
          label="Total Deployments"
          value={isLoading ? '—' : allDeployments.length}
          icon={<Rocket className="h-4 w-4" />}
          sub={failedCount > 0 ? `${failedCount} failed` : 'All time'}
          accent="bg-info/10 text-info"
        />
      </div>

      {/* Main grid */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Recent deployments */}
        <Card className="lg:col-span-2">
          <div className="border-border flex items-center justify-between border-b px-5 py-4">
            <h2 className="text-foreground text-sm font-semibold">Recent Deployments</h2>
            <RouterLink to="/deployments" className="text-primary text-xs hover:opacity-80">
              View all →
            </RouterLink>
          </div>

          {isLoading ? (
            <div className="divide-border divide-y px-2 py-2">
              {[...Array(4)].map((_, i) => (
                <div key={i} className="flex items-center gap-3 px-3 py-3">
                  <div className="flex-1 space-y-2">
                    <div className="flex items-center gap-2">
                      <Skeleton className="h-4 w-20" />
                      <Skeleton className="h-4 w-16 rounded-full" />
                    </div>
                    <Skeleton className="h-3 w-48" />
                  </div>
                  <Skeleton className="h-4 w-24" />
                </div>
              ))}
            </div>
          ) : recentDeps.length === 0 ? (
            <div className="text-muted-foreground flex h-48 flex-col items-center justify-center gap-3">
              <Rocket className="h-8 w-8 opacity-30" />
              <p className="text-sm">No deployments yet.</p>
              {projects.length === 0 ? (
                <RouterLink to="/projects" className="text-primary text-xs hover:underline">
                  Create your first project →
                </RouterLink>
              ) : (
                <RouterLink to="/projects" className="text-primary text-xs hover:underline">
                  Deploy a project →
                </RouterLink>
              )}
            </div>
          ) : (
            <div className="divide-border divide-y px-2 py-2">
              {recentDeps.map((d) => (
                <DeploymentRow key={d.id} deployment={d} project={projectMap[d.project_id]} />
              ))}
            </div>
          )}
        </Card>

        {/* Projects list */}
        <Card>
          <div className="border-border flex items-center justify-between border-b px-5 py-4">
            <h2 className="text-foreground text-sm font-semibold">Projects</h2>
            <RouterLink to="/projects" className="text-primary text-xs hover:opacity-80">
              Manage →
            </RouterLink>
          </div>

          {isLoading ? (
            <div className="divide-border divide-y py-2">
              {[...Array(3)].map((_, i) => (
                <div key={i} className="flex items-start gap-3 px-5 py-3">
                  <Skeleton className="mt-1 h-8 w-8 rounded-md" />
                  <div className="flex-1 space-y-2">
                    <Skeleton className="h-4 w-32" />
                    <Skeleton className="h-3 w-48" />
                  </div>
                </div>
              ))}
            </div>
          ) : projects.length === 0 ? (
            <div className="text-muted-foreground flex h-48 flex-col items-center justify-center gap-3">
              <FolderGit2 className="h-8 w-8 opacity-30" />
              <p className="text-sm">No projects yet.</p>
            </div>
          ) : (
            <div className="divide-border divide-y py-2">
              {projects.map((p) => {
                const deps = allDeployments.filter((d) => d.project_id === p.id);
                const live = deps.filter((d) => d.status === 'live').length;
                return (
                  <RouterLink
                    key={p.id}
                    to="/projects/$projectId"
                    params={{ projectId: p.id }}
                    className="hover:bg-accent flex items-start gap-3 px-5 py-3 transition"
                  >
                    <div className="bg-primary/10 text-primary mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-md">
                      <FolderGit2 className="h-3.5 w-3.5" />
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <span className="text-foreground truncate text-sm font-medium">
                          {p.name}
                        </span>
                        {live > 0 && (
                          <Badge variant="success" className="px-1.5 py-0.5 text-xs">
                            <span className="bg-success h-1 w-1 animate-pulse rounded-full" />
                            {live} live
                          </Badge>
                        )}
                      </div>
                      <div className="text-muted-foreground mt-0.5 flex items-center gap-1 text-xs">
                        <a
                          href={p.repo_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          onClick={(e) => e.stopPropagation()}
                          className="hover:text-foreground flex items-center gap-0.5"
                        >
                          {p.repo_owner}/{p.repo_name}
                          <ExternalLink className="h-2.5 w-2.5" />
                        </a>
                        <span>·</span>
                        <span>{p.branch}</span>
                      </div>
                    </div>
                    <span className="text-muted-foreground mt-1 shrink-0 text-xs">
                      {deps.length} deploy{deps.length !== 1 ? 's' : ''}
                    </span>
                  </RouterLink>
                );
              })}
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}
