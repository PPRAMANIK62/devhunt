import { ArrowRight, Building2, Zap } from "lucide-react";
import { Link } from "react-router-dom";
import { JobCard } from "@/components/jobs/job-card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useAuth } from "@/context/auth-context";
import { useJobFilterOptions, useJobs, EMPTY_JOB_FILTERS } from "@/hooks/use-jobs";

export function HomePage() {
  const { role } = useAuth();
  const { data, loading } = useJobs(1, 3, EMPTY_JOB_FILTERS);
  const { data: filterOptions } = useJobFilterOptions();
  const total = data?.total ?? 0;

  return (
    <div className="space-y-16">
      {/* ── Hero ──────────────────────────────────────────────── */}
      <section className="hero-grid relative overflow-hidden rounded-lg border border-border bg-card px-8 py-16 sm:px-12 sm:py-24">
        <div className="absolute top-0 left-0 h-1 w-24 bg-primary" />

        <h1
          className="animate-fade-up mb-6 border-l-4 border-primary pl-5 font-display text-5xl font-bold leading-tight tracking-tight text-foreground sm:text-6xl"
          style={{ animationDelay: "0ms" }}
        >
          The job board
          <br />
          for developers
          <br />
          <span className="text-primary">who ship.</span>
        </h1>

        <p
          className="animate-fade-up mb-10 max-w-md text-lg text-muted-foreground"
          style={{ animationDelay: "80ms" }}
        >
          Curated roles from companies building what comes next. No noise, just
          signal.
        </p>

        <div
          className="animate-fade-up flex flex-wrap items-center gap-3"
          style={{ animationDelay: "160ms" }}
        >
          {role !== "company" && (
            <Button size="lg" asChild>
              <Link to="/jobs">
                Browse open roles
                <ArrowRight className="ml-2 h-4 w-4" />
              </Link>
            </Button>
          )}
          {!role && (
            <Button size="lg" variant="outline" asChild>
              <Link to="/register">Post a job</Link>
            </Button>
          )}
          {role === "company" && (
            <Button size="lg" asChild>
              <Link to="/dashboard">Go to Dashboard</Link>
            </Button>
          )}
        </div>
      </section>

      {/* ── Stats strip ───────────────────────────────────────── */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
        <div className="rounded-lg border border-border bg-card p-6">
          <p className="font-display text-4xl font-bold text-foreground">
            {loading ? "—" : total}
          </p>
          <p className="mt-1 font-mono text-sm text-muted-foreground">
            open positions
          </p>
        </div>
        <div className="rounded-lg border border-border bg-card p-6">
          <p className="font-display text-4xl font-bold text-foreground">
            {loading ? "—" : filterOptions.locations.length}
          </p>
          <p className="mt-1 font-mono text-sm text-muted-foreground">
            cities &amp; regions
          </p>
        </div>
        <div className="col-span-2 rounded-lg border border-border bg-card p-6 sm:col-span-1">
          <p className="font-display text-4xl font-bold text-foreground">
            {loading ? "—" : filterOptions.tags.length}
          </p>
          <p className="mt-1 font-mono text-sm text-muted-foreground">
            tech stacks
          </p>
        </div>
      </div>

      {/* ── Featured jobs ─────────────────────────────────────── */}
      <div>
        <div className="mb-5 flex items-center justify-between">
          <h2 className="font-display text-xl font-semibold text-foreground">
            Latest roles
          </h2>
          <Button variant="ghost" size="sm" asChild className="font-mono text-xs">
            <Link to="/jobs">
              View all <ArrowRight className="ml-1 h-3 w-3" />
            </Link>
          </Button>
        </div>

        {loading ? (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-[130px] w-full rounded-lg" />
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {(data?.jobs ?? []).map((job) => (
              <JobCard key={job.id} job={job} />
            ))}
          </div>
        )}
      </div>

      {/* ── For companies ─────────────────────────────────────── */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div className="rounded-lg border border-border bg-card p-6">
          <div className="mb-4 flex h-9 w-9 items-center justify-center rounded-md bg-primary/10">
            <Zap className="h-4 w-4 text-primary" />
          </div>
          <h3 className="font-display text-lg font-semibold text-foreground">
            For job seekers
          </h3>
          <p className="mt-2 text-sm text-muted-foreground">
            Browse curated engineering, design, and product roles. Filter by
            location, stack, and salary. Apply in under a minute.
          </p>
          {!role && (
            <Button variant="outline" size="sm" asChild className="mt-4">
              <Link to="/register">Create an account</Link>
            </Button>
          )}
        </div>

        <div className="rounded-lg border border-border bg-card p-6">
          <div className="mb-4 flex h-9 w-9 items-center justify-center rounded-md bg-primary/10">
            <Building2 className="h-4 w-4 text-primary" />
          </div>
          <h3 className="font-display text-lg font-semibold text-foreground">
            For companies
          </h3>
          <p className="mt-2 text-sm text-muted-foreground">
            Post jobs, manage applications, and find developers who actually
            ship. No recruiter spam, no noise — just qualified applicants.
          </p>
          {!role && (
            <Button variant="outline" size="sm" asChild className="mt-4">
              <Link to="/register">Post a job</Link>
            </Button>
          )}
          {role === "company" && (
            <Button variant="outline" size="sm" asChild className="mt-4">
              <Link to="/dashboard">Go to Dashboard</Link>
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}
