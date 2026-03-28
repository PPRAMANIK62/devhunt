import { Building2, FileText, Pencil, Plus, Trash2 } from "lucide-react";
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
import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
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

const statusLabels: ApplicationStatus[] = ["pending", "reviewed", "accepted", "rejected"];

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function ApplicantsDrawer({
  job,
  open,
  onOpenChange,
}: {
  job: Job | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const { data: apps, loading, refetch } = useJobApplications(job?.id ?? "");
  const { execute: updateStatus } = useUpdateApplicationStatus();
  const [selected, setSelected] = useState<Application | null>(null);

  // Auto-select first applicant when list loads
  useEffect(() => {
    if (apps.length > 0 && !selected) setSelected(apps[0]);
  }, [apps, selected]);

  // Keep selected in sync after a status update
  useEffect(() => {
    if (!selected) return;
    const fresh = apps.find((a) => a.id === selected.id);
    if (fresh) setSelected(fresh);
  }, [apps]); // eslint-disable-line react-hooks/exhaustive-deps

  async function handleStatusChange(status: ApplicationStatus) {
    if (!selected) return;
    try {
      await updateStatus(selected.id, status);
      toast.success(`Status updated to ${status}`);
      refetch();
    } catch {
      toast.error("Failed to update status");
    }
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange} direction="right" shouldScaleBackground={false}>
      <DrawerContent className="inset-y-0 right-0 left-auto mt-0 h-full w-full max-w-2xl rounded-none p-0 flex flex-col [&>div:first-child]:hidden">
        <DrawerHeader className="px-6 py-4 border-b border-border shrink-0">
          <DrawerTitle className="font-display">
            {job?.title ?? "Applicants"}
          </DrawerTitle>
        </DrawerHeader>

        {loading ? (
          <div className="p-6 space-y-3">
            {[1, 2, 3].map((i) => <Skeleton key={i} className="h-12 w-full" />)}
          </div>
        ) : apps.length === 0 ? (
          <div className="flex flex-1 items-center justify-center">
            <p className="text-sm text-muted-foreground">No applicants yet.</p>
          </div>
        ) : (
          <div className="flex flex-1 min-h-0">
            {/* Left: applicant list */}
            <div className="w-56 shrink-0 border-r border-border overflow-y-auto">
              {apps.map((app) => (
                <button
                  key={app.id}
                  onClick={() => setSelected(app)}
                  className={`w-full text-left px-4 py-3 border-b border-border transition-colors hover:bg-muted/50 ${
                    selected?.id === app.id ? "bg-muted" : ""
                  }`}
                >
                  <p className="text-sm font-medium truncate text-foreground">
                    {app.user?.email ?? "Applicant"}
                  </p>
                  <div className="mt-1 flex items-center justify-between gap-2">
                    <span className="font-mono text-xs text-muted-foreground">
                      {formatDate(app.applied_at)}
                    </span>
                    <span
                      className={`font-mono text-xs capitalize rounded px-1.5 py-0.5 ${statusColors[app.status]}`}
                    >
                      {app.status}
                    </span>
                  </div>
                </button>
              ))}
            </div>

            {/* Right: detail pane */}
            {selected ? (
              <div className="flex-1 overflow-y-auto p-6 space-y-6">
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <p className="font-semibold text-foreground">
                      {selected.user?.email ?? "Applicant"}
                    </p>
                    <p className="font-mono text-xs text-muted-foreground mt-0.5">
                      Applied {formatDate(selected.applied_at)}
                    </p>
                  </div>
                  <Select
                    value={selected.status}
                    onValueChange={(v) => handleStatusChange(v as ApplicationStatus)}
                  >
                    <SelectTrigger className="w-36 font-mono text-xs">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {statusLabels.map((s) => (
                        <SelectItem key={s} value={s} className="font-mono text-xs capitalize">
                          {s}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <Separator />

                <div>
                  <div className="flex items-center gap-1.5 mb-3">
                    <FileText className="h-3.5 w-3.5 text-muted-foreground" />
                    <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                      Cover Letter
                    </p>
                  </div>
                  {selected.cover_note ? (
                    <p className="text-sm leading-relaxed text-foreground/90 whitespace-pre-wrap">
                      {selected.cover_note}
                    </p>
                  ) : (
                    <p className="text-sm text-muted-foreground italic">
                      No cover letter provided.
                    </p>
                  )}
                </div>
              </div>
            ) : null}
          </div>
        )}
      </DrawerContent>
    </Drawer>
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
  const [applicantsJob, setApplicantsJob] = useState<Job | null>(null);
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
              {jobs.map((job) => (
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
                        onClick={() => setApplicantsJob(job)}
                      >
                        View applicants
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
                </Card>
              ))}
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

      <ApplicantsDrawer
        job={applicantsJob}
        open={applicantsJob !== null}
        onOpenChange={(open: boolean) => { if (!open) setApplicantsJob(null); }}
      />
    </div>
  );
}
