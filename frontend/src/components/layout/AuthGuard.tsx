import { Navigate, useLocation } from '@tanstack/react-router';
import { useAuth } from '../../features/auth/useAuth';
import { Loader2 } from 'lucide-react';

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-muted-foreground flex flex-col items-center gap-2">
          <Loader2 className="h-8 w-8 animate-spin" />
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" search={{ redirect: location.pathname }} />;
  }

  return <>{children}</>;
}
