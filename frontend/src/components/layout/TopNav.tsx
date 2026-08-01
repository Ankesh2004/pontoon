import { LogOut, Sun, Moon, Rocket } from 'lucide-react';
import { useAuth } from '../../features/auth/useAuth';
import { useTheme } from '../theme/ThemeProvider';
import { Link } from '@tanstack/react-router';

export function TopNav() {
  const { user, logout } = useAuth();
  const { theme, toggle } = useTheme();

  return (
    <header className="border-border bg-card border-b">
      <div className="flex h-16 items-center justify-between px-6">
        {/* Brand */}
        <Link to="/" className="group flex items-center gap-2.5">
          <div className="bg-primary/10 flex h-8 w-8 items-center justify-center rounded-lg">
            <Rocket className="text-primary h-4 w-4 transition group-hover:scale-110" />
          </div>
          <span className="text-foreground text-lg font-bold tracking-tight">Pontoon</span>
        </Link>

        {/* Right side */}
        <div className="flex items-center gap-3">
          {/* Dark/Light toggle */}
          <button
            onClick={toggle}
            title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
            className="border-border bg-background text-muted-foreground hover:border-primary/40 hover:bg-accent hover:text-foreground flex h-9 w-9 items-center justify-center rounded-lg border transition"
          >
            {theme === 'dark' ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
          </button>

          {user && (
            <>
              <div className="flex items-center gap-2.5">
                <img
                  src={user.avatar_url}
                  alt={user.github_username}
                  className="ring-border h-8 w-8 rounded-full ring-2"
                />
                <span className="text-foreground text-sm font-medium">{user.github_username}</span>
              </div>
              <button
                onClick={logout}
                title="Logout"
                className="border-border bg-background text-muted-foreground hover:border-destructive/40 hover:bg-destructive/10 hover:text-destructive flex h-9 w-9 items-center justify-center rounded-lg border transition"
              >
                <LogOut className="h-4 w-4" />
              </button>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
