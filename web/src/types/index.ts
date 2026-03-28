export type UserRole = "seeker" | "company" | "admin";

export interface User {
  id: string;
  email: string;
  role: UserRole;
  created_at: string;
  updated_at: string;
}

export type JobStatus = "open" | "closed" | "draft";

export interface Job {
  id: string;
  company_id: string;
  title: string;
  description: string;
  location: string;
  salary_min: number;
  salary_max: number;
  status: JobStatus;
  tags: string[] | null;
  created_at: string;
  updated_at: string;
  company?: Company;
}

export interface PaginatedJobs {
  jobs: Job[];
  total: number;
  page: number;
  page_size: number;
}

export interface Company {
  id: string;
  user_id: string;
  name: string;
  slug: string;
  description?: string;
  website?: string;
  created_at: string;
  updated_at: string;
}

export type ApplicationStatus =
  | "pending"
  | "reviewed"
  | "rejected"
  | "accepted";

export interface Application {
  id: string;
  job_id: string;
  user_id: string;
  status: ApplicationStatus;
  cover_note?: string;
  applied_at: string;
  updated_at: string;
  job?: Job;
  user?: User;
}

export interface ApiError {
  error: string;
  code: string;
}
