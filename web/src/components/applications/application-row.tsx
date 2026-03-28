import { Link } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import type { Application, ApplicationStatus } from "@/types";

const statusStyles: Record<ApplicationStatus, string> = {
  pending: "bg-amber-100 text-amber-800",
  reviewed: "bg-blue-100 text-blue-800",
  accepted: "bg-green-100 text-green-800",
  rejected: "bg-zinc-100 text-zinc-500",
};

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

interface ApplicationRowProps {
  application: Application;
}

export function ApplicationRow({ application }: ApplicationRowProps) {
  const job = application.job;

  return (
    <Card className="flex flex-col transition-shadow duration-150 hover:shadow-md">
      <Link to={`/jobs/${application.job_id}`} className="flex flex-1 flex-col p-4 group">
        <div className="flex items-start justify-between gap-2 mb-2">
          <p className="font-display font-semibold text-card-foreground leading-snug group-hover:text-primary">
            {job?.title ?? "Job"}
          </p>
          <Badge
            className={`shrink-0 font-mono text-xs capitalize border-transparent ${statusStyles[application.status]}`}
          >
            {application.status}
          </Badge>
        </div>
        {job?.company && (
          <p className="text-sm text-muted-foreground">{job.company.name}</p>
        )}
        <p className="mt-auto pt-3 font-mono text-xs text-muted-foreground">
          Applied {formatDate(application.applied_at)}
        </p>
      </Link>
    </Card>
  );
}
