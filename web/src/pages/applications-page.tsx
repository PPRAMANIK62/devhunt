import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { ApplicationRow } from "@/components/applications/application-row";
import { Skeleton } from "@/components/ui/skeleton";
import { useAuth } from "@/context/auth-context";
import { useMyApplications } from "@/hooks/use-applications";

export function ApplicationsPage() {
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const { data, loading, error } = useMyApplications();

  useEffect(() => {
    if (!isAuthenticated) navigate("/login");
  }, [isAuthenticated, navigate]);

  return (
    <div>
      <div className="mb-8">
        <h1 className="font-display text-3xl font-bold tracking-tight">
          My Applications
        </h1>
        <p className="mt-1 text-muted-foreground">
          Track the status of your job applications.
        </p>
      </div>

      {error && (
        <p className="rounded-md border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">
          {error}
        </p>
      )}

      {loading ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-36 w-full rounded-lg" />
          ))}
        </div>
      ) : data.length === 0 ? (
        <p className="py-16 text-center text-muted-foreground">
          No applications yet. Browse jobs and apply!
        </p>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {data.map((app) => (
            <ApplicationRow key={app.id} application={app} />
          ))}
        </div>
      )}
    </div>
  );
}
