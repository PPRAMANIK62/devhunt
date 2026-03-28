import { MapPin } from "lucide-react";
import { Link } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import type { Job } from "@/types";

function formatSalary(min: number, max: number): string {
  const fmt = (n: number) =>
    n >= 1000 ? `$${(n / 1000).toFixed(0)}k` : `$${n}`;
  return `${fmt(min)} – ${fmt(max)}`;
}

interface JobCardProps {
  job: Job;
}

export function JobCard({ job }: JobCardProps) {
  return (
    <Link to={`/jobs/${job.id}`} className="group block">
      <Card className="h-[130px] transition-shadow duration-150 hover:shadow-md">
        <div className="flex h-full flex-col p-4">
          {/* Top: title + salary */}
          <div className="flex items-start justify-between gap-4">
            <h2 className="line-clamp-2 min-w-0 flex-1 font-display text-base font-semibold leading-snug text-card-foreground group-hover:text-primary">
              {job.title}
            </h2>
            <span className="shrink-0 font-mono text-xs font-medium text-muted-foreground">
              {formatSalary(job.salary_min, job.salary_max)}
            </span>
          </div>

          {/* Middle: company + tags */}
          <div className="mt-1.5 flex-1">
            {job.company && (
              <p className="text-sm text-muted-foreground">{job.company.name}</p>
            )}
            {job.tags && job.tags.length > 0 && (
              <div className="mt-2 flex flex-wrap gap-1">
                {job.tags.slice(0, 3).map((tag) => (
                  <Badge key={tag} variant="secondary" className="font-mono text-xs">
                    {tag}
                  </Badge>
                ))}
              </div>
            )}
          </div>

          {/* Bottom: location pinned */}
          <div className="mt-2 flex items-center gap-1 text-xs text-muted-foreground">
            <MapPin className="h-3 w-3 shrink-0" />
            <span className="font-mono">{job.location}</span>
          </div>
        </div>
      </Card>
    </Link>
  );
}
