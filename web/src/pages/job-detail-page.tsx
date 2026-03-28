import { ArrowLeft, CheckCircle2, MapPin } from "lucide-react";
import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ApplyDrawer } from "@/components/applications/apply-drawer";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { useAuth } from "@/context/auth-context";
import { useMyApplications } from "@/hooks/use-applications";
import { useJob } from "@/hooks/use-jobs";

function formatSalary(min: number, max: number): string {
  const fmt = (n: number) =>
    n >= 1000 ? `$${(n / 1000).toFixed(0)}k` : `$${n}`;
  return `${fmt(min)} – ${fmt(max)} / year`;
}

export function JobDetailPage() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { data: job, loading, error } = useJob(id ?? "");
  const { isAuthenticated, role } = useAuth();
  const canTrackApplications = isAuthenticated && role === "seeker";
  const { data: myApplications, refetch: refetchApplications } = useMyApplications(canTrackApplications);
  const hasApplied = myApplications.some((a) => a.job_id === id);
  const [applyOpen, setApplyOpen] = useState(false);

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-5 w-40" />
        <Skeleton className="h-40 w-full" />
      </div>
    );
  }

  if (error || !job) {
    return (
      <div className="py-16 text-center">
        <p className="text-muted-foreground">Job not found.</p>
        <Button variant="ghost" asChild className="mt-4">
          <Link to="/jobs">Browse all jobs</Link>
        </Button>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl">
      <Button
        variant="ghost"
        size="sm"
        className="mb-6 -ml-2"
        onClick={() => navigate(-1)}
      >
        <ArrowLeft className="mr-1.5 h-4 w-4" />
        All jobs
      </Button>

      <div className="space-y-1">
        <h1 className="font-display text-3xl font-bold tracking-tight">
          {job.title}
        </h1>
        {job.company && (
          <p className="text-lg font-medium text-muted-foreground">
            {job.company.name}
          </p>
        )}
      </div>

      <div className="mt-4 flex flex-wrap items-center gap-4">
        <span className="flex items-center gap-1 font-mono text-sm text-muted-foreground">
          <MapPin className="h-3.5 w-3.5" />
          {job.location}
        </span>
        <span className="font-mono text-sm font-medium text-foreground">
          {formatSalary(job.salary_min, job.salary_max)}
        </span>
        <span className="flex items-center gap-1.5 font-mono text-xs">
          <span
            className={`inline-block h-1.5 w-1.5 rounded-full ${
              job.status === "open"
                ? "bg-green-500"
                : job.status === "draft"
                  ? "bg-amber-500"
                  : "bg-zinc-400"
            }`}
          />
          <span className="capitalize text-muted-foreground">{job.status}</span>
        </span>
      </div>

      {job.tags && job.tags.length > 0 && (
        <div className="mt-3 flex flex-wrap gap-1.5">
          {job.tags.map((tag) => (
            <Badge key={tag} variant="secondary" className="font-mono text-xs">
              {tag}
            </Badge>
          ))}
        </div>
      )}

      <Separator className="my-6" />

      <div className="prose prose-zinc max-w-none">
        <p className="whitespace-pre-wrap leading-relaxed text-foreground/90">
          {job.description}
        </p>
      </div>

      {job.status === "open" && (
        <div className="mt-8">
          {isAuthenticated && role === "seeker" ? (
            hasApplied ? (
              <div className="flex items-center gap-3 rounded-md border border-border bg-muted/40 px-4 py-3">
                <CheckCircle2 className="h-4 w-4 shrink-0 text-green-600" />
                <div className="min-w-0">
                  <p className="text-sm font-medium text-foreground">
                    You've already applied
                  </p>
                  <p className="font-mono text-xs text-muted-foreground">
                    Track your application status in{" "}
                    <Link
                      to="/applications"
                      className="underline underline-offset-2 hover:text-foreground"
                    >
                      My Applications
                    </Link>
                  </p>
                </div>
              </div>
            ) : (
              <Button size="lg" onClick={() => setApplyOpen(true)}>
                Apply for this position
              </Button>
            )
          ) : !isAuthenticated ? (
            <div className="flex items-center gap-3">
              <Button size="lg" asChild>
                <Link to="/login">Log in to apply</Link>
              </Button>
              <Button size="lg" variant="outline" asChild>
                <Link to="/register">Sign up</Link>
              </Button>
            </div>
          ) : null}
        </div>
      )}

      <ApplyDrawer
        open={applyOpen}
        onOpenChange={setApplyOpen}
        jobId={job.id}
        jobTitle={job.title}
        onSuccess={refetchApplications}
      />
    </div>
  );
}
