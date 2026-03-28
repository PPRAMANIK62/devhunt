import { Building2, Pencil, Plus, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { CompanyFormDrawer } from "@/components/company/company-form-drawer";
import { JobFormDrawer } from "@/components/jobs/job-form-drawer";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button, buttonVariants } from "@/components/ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useAuth } from "@/context/auth-context";
import { useJobApplications, useUpdateApplicationStatus } from "@/hooks/use-applications";
import { useMyCompany } from "@/hooks/use-company";
import { useCompanyJobs, useDeleteJob } from "@/hooks/use-jobs";
import type { Application, ApplicationStatus, Job } from "@/types";

const statusColors: Record<ApplicationStatus, string> = {
  pending: "bg-amber-100 text-amber-800",
  reviewed: "bg-blue-100 text-blue-800",
  accepted: "bg-green-100 text-green-800",
  rejected: "bg-zinc-100 text-zinc-500",
};

function ApplicationCard({
  application,
  onStatusChange,
}: {
  application: Application;
  onStatusChange: () => void;
}) {
  const { execute: updateStatus } = useUpdateApplicationStatus();
  const statuses: ApplicationStatus[] = [
    "pending",
    "reviewed",
    "accepted",
    "rejected",
  ];

  async function handleStatus(s: ApplicationStatus) {
    try {
      await updateStatus(application.id, s);
      toast.success(`Status updated to ${s}`);
      onStatusChange();
    } catch {
      toast.error("Failed to update status");
    }
  }

  return (
    <div className="rounded-md border border-border bg-card p-4 space-y-3">
      <div className="flex items-start justify-between">
        <div>
          <p className="font-medium text-sm text-card-foreground">
            {application.user?.email ?? "Applicant"}
          </p>
          {application.cover_note && (
            <p className="mt-1 text-xs text-muted-foreground line-clamp-2">
              {application.cover_note}
            </p>
          )}
        </div>
        <Badge
          className={`font-mono text-xs capitalize shrink-0 ${statusColors[application.status]}`}
          variant="outline"
        >
          {application.status}
        </Badge>
      </div>
      <div className="flex flex-wrap gap-1.5">
        {statuses.map((s) => (
          <Button
            key={s}
            variant={application.status === s ? "default" : "outline"}
            size="sm"
            className="h-6 px-2 font-mono text-xs capitalize"
            onClick={() => handleStatus(s)}
            disabled={application.status === s}
          >
            {s}
          </Button>
        ))}
      </div>
    </div>
  );
}

function JobApplicationsPanel({
  jobId,
  onStatusChange,
}: {
  jobId: string;
  onStatusChange: () => void;
}) {
  const { data: apps, loading } = useJobApplications(jobId);

  if (loading) {
    return <p className="text-xs text-muted-foreground">Loading applicants…</p>;
  }
  if (apps.length === 0) {
    return <p className="text-xs text-muted-foreground">No applicants yet.</p>;
  }
  return (
    <div className="space-y-2">
      {apps.map((app) => (
        <ApplicationCard key={app.id} application={app} onStatusChange={onStatusChange} />
      ))}
    </div>
  );
}

export function DashboardPage() {
  const { role } = useAuth();
  const navigate = useNavigate();
  const {
    data: company,
    loading: companyLoading,
    notFound,
    refetch: refetchCompany,
  } = useMyCompany();
  const [statusFilter, setStatusFilter] = useState("");
  const {
    data: jobs,
    loading: jobsLoading,
    refetch: refetchJobs,
  } = useCompanyJobs(statusFilter);
  const { execute: deleteJob } = useDeleteJob();

  const [companyDrawerOpen, setCompanyDrawerOpen] = useState(false);
  const [jobDrawerOpen, setJobDrawerOpen] = useState(false);
  const [editingJob, setEditingJob] = useState<Job | null>(null);
  const [expandedJob, setExpandedJob] = useState<string | null>(null);
  const [jobToDelete, setJobToDelete] = useState<string | null>(null);

  useEffect(() => {
    if (role !== "company") navigate("/");
  }, [role, navigate]);

  useEffect(() => {
    if (notFound) setCompanyDrawerOpen(true);
  }, [notFound]);

  function openCreateJob() {
    setEditingJob(null);
    setJobDrawerOpen(true);
  }

  function openEditJob(job: Job) {
    setEditingJob(job);
    setJobDrawerOpen(true);
  }

  async function handleDeleteJob() {
    if (!jobToDelete) return;
    try {
      await deleteJob(jobToDelete);
      toast.success("Job deleted");
      refetchJobs();
    } catch {
      toast.error("Failed to delete job");
    } finally {
      setJobToDelete(null);
    }
  }

  if (companyLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-20 w-full" />
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Company header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="font-display text-3xl font-bold">Dashboard</h1>
          {company && (
            <p className="mt-1 text-muted-foreground">{company.name}</p>
          )}
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setCompanyDrawerOpen(true)}
          >
            <Building2 className="mr-1.5 h-4 w-4" />
            {company ? "Edit Profile" : "Create Profile"}
          </Button>
          {company && (
            <Button size="sm" onClick={openCreateJob}>
              <Plus className="mr-1.5 h-4 w-4" />
              Post Job
            </Button>
          )}
        </div>
      </div>

      {!company && !notFound && (
        <Card>
          <CardHeader>
            <CardTitle className="font-display">
              No company profile yet
            </CardTitle>
            <CardDescription>
              Create a company profile to start posting jobs.
            </CardDescription>
          </CardHeader>
        </Card>
      )}

      {/* Jobs */}
      {company && (
        <div>
          <div className="mb-4 flex items-center justify-between gap-4">
            <h2 className="font-display text-xl font-semibold">
              Your Listings
            </h2>
            <div className="flex items-center gap-1">
              {(["", "open", "draft", "closed"] as const).map((s) => (
                <Button
                  key={s}
                  variant={statusFilter === s ? "default" : "ghost"}
                  size="sm"
                  className="h-7 px-3 font-mono text-xs capitalize"
                  onClick={() => setStatusFilter(s)}
                >
                  {s === "" ? "All" : s}
                </Button>
              ))}
            </div>
          </div>

          {jobsLoading ? (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-40 w-full rounded-lg" />
              ))}
            </div>
          ) : jobs.length === 0 ? (
            <p className="py-8 text-center text-sm text-muted-foreground">
              No jobs posted yet.
            </p>
          ) : (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {jobs.map((job) => {
                const isExpanded = expandedJob === job.id;
                return (
                  <Card key={job.id} className="flex flex-col">
                    <div className="flex flex-1 flex-col p-4">
                      <div className="flex items-start justify-between gap-2 mb-3">
                        <span className="font-display font-semibold text-sm text-card-foreground leading-snug">
                          {job.title}
                        </span>
                        <Badge
                          className={`shrink-0 font-mono text-xs capitalize border-transparent ${
                            job.status === "open"
                              ? "bg-green-100 text-green-800"
                              : "bg-zinc-100 text-zinc-500"
                          }`}
                        >
                          {job.status}
                        </Badge>
                      </div>
                      <div className="mt-auto flex items-center gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 px-2 font-mono text-xs"
                          onClick={() =>
                            setExpandedJob(isExpanded ? null : job.id)
                          }
                        >
                          {isExpanded ? "Hide" : "View"} applicants
                        </Button>
                        <div className="ml-auto flex items-center gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7"
                            onClick={() => openEditJob(job)}
                          >
                            <Pencil className="h-3.5 w-3.5" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7 text-destructive hover:text-destructive"
                            onClick={() => setJobToDelete(job.id)}
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        </div>
                      </div>
                    </div>
                    {isExpanded && (
                      <div className="border-t border-border p-4">
                        <JobApplicationsPanel
                          jobId={job.id}
                          onStatusChange={() => {}}
                        />
                      </div>
                    )}
                  </Card>
                );
              })}
            </div>
          )}
        </div>
      )}

      <AlertDialog
        open={jobToDelete !== null}
        onOpenChange={(open) => { if (!open) setJobToDelete(null); }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete job listing?</AlertDialogTitle>
            <AlertDialogDescription>
              This will permanently remove the listing and cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              className={buttonVariants({ variant: "destructive" })}
              onClick={handleDeleteJob}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <CompanyFormDrawer
        open={companyDrawerOpen}
        onOpenChange={setCompanyDrawerOpen}
        company={company}
        onSuccess={() => refetchCompany()}
      />

      <JobFormDrawer
        open={jobDrawerOpen}
        onOpenChange={setJobDrawerOpen}
        job={editingJob}
        onSuccess={refetchJobs}
      />
    </div>
  );
}
