import { Outlet } from '@tanstack/react-router';
import { TopNav } from './TopNav';
import { Sidebar } from './Sidebar';

export function Layout() {
  return (
    <div className="flex h-screen flex-col">
      <TopNav />
      <div className="flex flex-1 overflow-hidden">
        <Sidebar />
        <main className="flex-1 overflow-auto bg-background p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
