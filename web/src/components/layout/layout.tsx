import { Outlet } from "react-router-dom";
import { Toaster } from "@/components/ui/sonner";
import { Header } from "./header";

export function Layout() {
  return (
    <div className="min-h-dvh bg-background">
      <Header />
      <main className="mx-auto max-w-6xl px-4 py-8">
        <Outlet />
      </main>
      <Toaster richColors />
    </div>
  );
}
