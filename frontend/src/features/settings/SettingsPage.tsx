import { Settings } from 'lucide-react';

export function SettingsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-foreground text-3xl font-bold">Settings</h1>
        <p className="text-muted-foreground mt-2">Manage your account and application settings</p>
      </div>

      <div className="border-border rounded-lg border border-dashed p-12 text-center">
        <Settings className="text-muted-foreground mx-auto h-12 w-12" />
        <h3 className="text-foreground mt-4 text-lg font-semibold">Settings coming soon</h3>
        <p className="text-muted-foreground mt-2 text-sm">
          Account settings and preferences will be available here
        </p>
      </div>
    </div>
  );
}
