import { useAuth } from '../auth/useAuth';

export function DashboardPage() {
  const { user } = useAuth();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground">Dashboard</h1>
        <p className="text-muted-foreground">
          Welcome back, {user?.github_username}!
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <div className="rounded-lg border border-border bg-card p-6">
          <h3 className="text-lg font-semibold text-foreground">Projects</h3>
          <p className="mt-2 text-3xl font-bold text-primary">0</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Active projects
          </p>
        </div>

        <div className="rounded-lg border border-border bg-card p-6">
          <h3 className="text-lg font-semibold text-foreground">Deployments</h3>
          <p className="mt-2 text-3xl font-bold text-primary">0</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Total deployments
          </p>
        </div>

        <div className="rounded-lg border border-border bg-card p-6">
          <h3 className="text-lg font-semibold text-foreground">Status</h3>
          <p className="mt-2 text-3xl font-bold text-primary">Healthy</p>
          <p className="mt-1 text-sm text-muted-foreground">
            All systems operational
          </p>
        </div>
      </div>
    </div>
  );
}
