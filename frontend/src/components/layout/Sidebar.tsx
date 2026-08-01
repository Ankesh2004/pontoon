import { Link, useLocation } from '@tanstack/react-router';
import { LayoutDashboard, FolderGit2, Rocket, Settings } from 'lucide-react';
import { cn } from '../../lib/utils';

const navItems = [
  { path: '/', label: 'Dashboard', icon: LayoutDashboard, exact: true },
  { path: '/projects', label: 'Projects', icon: FolderGit2 },
  { path: '/deployments', label: 'Deployments', icon: Rocket },
  { path: '/settings', label: 'Settings', icon: Settings },
];

export function Sidebar() {
  const location = useLocation();

  return (
    <aside className="border-border bg-card flex w-56 shrink-0 flex-col border-r">
      <nav className="flex flex-col gap-1 p-3 pt-4">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = item.exact
            ? location.pathname === item.path
            : location.pathname.startsWith(item.path);

          return (
            <Link
              key={item.path}
              to={item.path}
              className={cn(
                'flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-primary/10 text-primary'
                  : 'text-muted-foreground hover:bg-accent hover:text-foreground'
              )}
            >
              <Icon className={cn('h-4 w-4 shrink-0', isActive ? 'text-primary' : '')} />
              {item.label}
              {isActive && <span className="bg-primary ml-auto h-1.5 w-1.5 rounded-full" />}
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}
