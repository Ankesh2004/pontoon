import { Settings } from 'lucide-react';

export function SettingsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground">Settings</h1>
        <p className="mt-2 text-muted-foreground">
          Manage your account and application settings
        </p>
      </div>

      <div className="rounded-lg border border-dashed border-border p-12 text-center">
        <Settings className="mx-auto h-12 w-12 text-muted-foreground" />
        <h3 className="mt-4 text-lg font-semibold text-foreground">
          Settings coming soon
        </h3>
        <p className="mt-2 text-sm text-muted-foreground">
          Account settings and preferences will be available here
        </p>
      </div>
    </div>
  );
}
