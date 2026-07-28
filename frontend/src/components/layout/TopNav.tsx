import { LogOut } from 'lucide-react';
import { useAuth } from '../../features/auth/useAuth';
import { Button } from '../ui/button';

export function TopNav() {
  const { user, logout } = useAuth();

  return (
    <header className="border-b border-border bg-card">
      <div className="flex h-16 items-center justify-between px-6">
        <div className="flex items-center gap-2">
          <h1 className="text-xl font-bold text-foreground">Pontoon</h1>
        </div>

        {user && (
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <img
                src={user.avatar_url}
                alt={user.github_username}
                className="h-8 w-8 rounded-full"
              />
              <span className="text-sm font-medium text-foreground">
                {user.github_username}
              </span>
            </div>
            <Button
              variant="ghost"
              size="icon"
              onClick={logout}
              title="Logout"
            >
              <LogOut className="h-4 w-4" />
            </Button>
          </div>
        )}
      </div>
    </header>
  );
}
