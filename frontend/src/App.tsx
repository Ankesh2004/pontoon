import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import {
  RouterProvider,
  createRouter,
  createRootRoute,
  createRoute,
} from '@tanstack/react-router';
import { Layout } from './components/layout/Layout';
import { AuthGuard } from './components/layout/AuthGuard';
import { LoginPage } from './features/auth/LoginPage';
import { DashboardPage } from './features/dashboard/DashboardPage';
import { ProjectsPage } from './features/projects/ProjectsPage';
import { ProjectDetailPage } from './features/projects/ProjectDetailPage';
import { DeploymentsPage } from './features/deployments/DeploymentsPage';
import { DeploymentDetailPage } from './features/deployments/DeploymentDetailPage';
import { SettingsPage } from './features/settings/SettingsPage';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

const rootRoute = createRootRoute();

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: LoginPage,
});

const layoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'layout',
  component: Layout,
});

const indexRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/',
  component: () => (
    <AuthGuard>
      <DashboardPage />
    </AuthGuard>
  ),
});

const projectsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/projects',
  component: () => (
    <AuthGuard>
      <ProjectsPage />
    </AuthGuard>
  ),
});

const projectDetailRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/projects/$projectId',
  component: () => (
    <AuthGuard>
      <ProjectDetailPage />
    </AuthGuard>
  ),
});

const deploymentsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/deployments',
  component: () => (
    <AuthGuard>
      <DeploymentsPage />
    </AuthGuard>
  ),
});

const deploymentDetailRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/deployments/$deploymentId',
  component: () => (
    <AuthGuard>
      <DeploymentDetailPage />
    </AuthGuard>
  ),
});

const settingsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/settings',
  component: () => (
    <AuthGuard>
      <SettingsPage />
    </AuthGuard>
  ),
});

const routeTree = rootRoute.addChildren([
  loginRoute,
  layoutRoute.addChildren([
    indexRoute,
    projectsRoute,
    projectDetailRoute,
    deploymentsRoute,
    deploymentDetailRoute,
    settingsRoute,
  ]),
]);

const router = createRouter({
  routeTree,
  context: {
    queryClient,
  },
});

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  );
}
