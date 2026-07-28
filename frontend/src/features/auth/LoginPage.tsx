import { GitBranch } from 'lucide-react';
import { Button } from '../../components/ui/button';

export function LoginPage() {
  const handleLogin = () => {
    window.location.href = '/auth/github';
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="w-full max-w-md space-y-8 rounded-lg border border-border bg-card p-8 shadow-sm">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-foreground">Pontoon</h1>
          <p className="mt-2 text-muted-foreground">
            Self-hosted PaaS for your applications
          </p>
        </div>

        <Button
          onClick={handleLogin}
          className="w-full"
          size="lg"
        >
          <GitBranch className="mr-2 h-5 w-5" />
          Sign in with GitHub
        </Button>
      </div>
    </div>
  );
}
