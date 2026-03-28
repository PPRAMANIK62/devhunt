import { ChevronLeft, ChevronRight } from "lucide-react";
import { useSearchParams } from "react-router-dom";
import { JobFiltersSidebar } from "@/components/jobs/job-filters";
import { JobCard } from "@/components/jobs/job-card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useJobs, type JobFilters } from "@/hooks/use-jobs";

const PAGE_SIZE = 9;

export function JobsPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Math.max(1, Number(searchParams.get("page") ?? "1"));
  const filters: JobFilters = {
    search: searchParams.get("q") ?? "",
    locations: searchParams.getAll("location"),
    tags: searchParams.getAll("tag"),
    minSalary: Number(searchParams.get("min_salary") ?? "0"),
  };
  const { data, loading, error } = useJobs(page, PAGE_SIZE, filters);

  function buildParams(f: JobFilters, p: number): URLSearchParams {
    const params = new URLSearchParams({ page: String(p) });
    if (f.search) params.set("q", f.search);
    f.locations.forEach((l) => params.append("location", l));
    f.tags.forEach((t) => params.append("tag", t));
    if (f.minSalary > 0) params.set("min_salary", String(f.minSalary));
    return params;
  }

  function handleFiltersChange(next: JobFilters) {
    setSearchParams(buildParams(next, 1));
  }

  function setPage(p: number) {
    setSearchParams(buildParams(filters, p));
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  const totalPages = data ? Math.ceil(data.total / PAGE_SIZE) : 1;
  const total = data?.total ?? 0;

  return (
    <div>
      <div className="mb-8">
        <h1 className="font-display text-3xl font-bold tracking-tight">
          Open roles
        </h1>
        <p className="mt-1 font-mono text-sm text-muted-foreground">
          {loading ? "—" : `${total} position${total !== 1 ? "s" : ""}`}
        </p>
      </div>

      <div className="flex flex-col gap-8 lg:flex-row lg:items-start">
        <JobFiltersSidebar filters={filters} onChange={handleFiltersChange} />

        <div className="min-w-0 flex-1">
          {error && (
            <p className="mb-4 rounded-md border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">
              {error}
            </p>
          )}

          {loading ? (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
              {Array.from({ length: 9 }).map((_, i) => (
                <Skeleton key={i} className="h-[130px] w-full rounded-lg" />
              ))}
            </div>
          ) : data && (data.jobs ?? []).length > 0 ? (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
              {(data.jobs ?? []).map((job) => (
                <JobCard key={job.id} job={job} />
              ))}
            </div>
          ) : (
            <p className="py-16 text-center text-muted-foreground">
              {filters.search || filters.locations.length > 0 || filters.tags.length > 0
                ? "No jobs match your filters."
                : "No jobs available right now."}
            </p>
          )}

          {data && data.total > PAGE_SIZE && (data.jobs ?? []).length > 0 && (
            <div className="mt-8 flex items-center justify-between">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(page - 1)}
                disabled={page <= 1}
              >
                <ChevronLeft className="mr-1 h-4 w-4" />
                Previous
              </Button>
              <span className="font-mono text-sm text-muted-foreground">
                {page} / {totalPages}
              </span>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(page + 1)}
                disabled={page >= totalPages}
              >
                Next
                <ChevronRight className="ml-1 h-4 w-4" />
              </Button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
