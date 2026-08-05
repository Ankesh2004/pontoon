import { GitBranch } from 'lucide-react';
import { Button } from '../../components/ui/button';

export function LoginPage() {
  const handleLogin = () => {
    const apiBase = import.meta.env.VITE_API_URL || '';
    window.location.href = `${apiBase}/auth/github`;
  };

  return (
    <div className="bg-background flex min-h-screen items-center justify-center">
      <div className="border-border bg-card w-full max-w-md space-y-8 rounded-lg border p-8 shadow-sm">
        <div className="text-center">
          <h1 className="text-foreground text-3xl font-bold">Pontoon</h1>
          <p className="text-muted-foreground mt-2">Self-hosted PaaS for your applications</p>
        </div>

        <Button onClick={handleLogin} className="w-full" size="lg">
          <GitBranch className="mr-2 h-5 w-5" />
          Sign in with GitHub
        </Button>
      </div>
    </div>
  );
}
