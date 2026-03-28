import { useCallback, useEffect, useRef, useState } from "react";
import { api } from "@/lib/api";
import type { Job, PaginatedJobs } from "@/types";

export interface JobFilters {
  search: string;
  locations: string[];
  tags: string[];
  minSalary: number;
}

export const EMPTY_JOB_FILTERS: JobFilters = { search: "", locations: [], tags: [], minSalary: 0 };

export function useJobs(page: number, pageSize = 10, filters: JobFilters = EMPTY_JOB_FILTERS) {
  const [data, setData] = useState<PaginatedJobs | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const locKey = filters.locations.join(",");
  const tagKey = filters.tags.join(",");

  useEffect(() => {
    const controller = new AbortController();
    setLoading(true);
    setError(null);
    const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
    if (filters.search) params.set("q", filters.search);
    filters.locations.forEach((l) => params.append("location", l));
    filters.tags.forEach((t) => params.append("tag", t));
    if (filters.minSalary > 0) params.set("min_salary", String(filters.minSalary));
    api
      .get<PaginatedJobs>(`/jobs?${params.toString()}`, { signal: controller.signal })
      .then(setData)
      .catch((err: unknown) => {
        if (err instanceof Error && err.name === "AbortError") return;
        setError(err instanceof Error ? err.message : "Failed to load jobs");
      })
      .finally(() => setLoading(false));
    return () => controller.abort();
  }, [page, pageSize, filters.search, locKey, tagKey, filters.minSalary]);

  return { data, loading, error };
}

export function useJob(id: string) {
  const [data, setData] = useState<Job | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    const controller = new AbortController();
    setLoading(true);
    setError(null);
    api
      .get<Job>(`/jobs/${id}`, { signal: controller.signal })
      .then(setData)
      .catch((err: unknown) => {
        if (err instanceof Error && err.name === "AbortError") return;
        setError(err instanceof Error ? err.message : "Failed to load job");
      })
      .finally(() => setLoading(false));
    return () => controller.abort();
  }, [id]);

  return { data, loading, error };
}

interface CreateJobData {
  title: string;
  description: string;
  location: string;
  salary_min: number;
  salary_max: number;
  tags?: string[];
}

interface UpdateJobData {
  title?: string;
  description?: string;
  location?: string;
  salary_min?: number;
  salary_max?: number;
  status?: string;
  tags?: string[];
}

export function useCreateJob() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const execute = useCallback(async (data: CreateJobData): Promise<Job> => {
    setLoading(true);
    setError(null);
    try {
      return await api.post<Job>("/jobs", data);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to create job";
      setError(msg);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  return { execute, loading, error };
}

export function useUpdateJob() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const execute = useCallback(
    async (id: string, data: UpdateJobData): Promise<Job> => {
      setLoading(true);
      setError(null);
      try {
        return await api.patch<Job>(`/jobs/${id}`, data);
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Failed to update job";
        setError(msg);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  return { execute, loading, error };
}

export function useDeleteJob() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const execute = useCallback(async (id: string): Promise<void> => {
    setLoading(true);
    setError(null);
    try {
      await api.delete(`/jobs/${id}`);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to delete job";
      setError(msg);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  return { execute, loading, error };
}

interface FilterOptions {
  locations: string[];
  tags: string[];
}

export function useJobFilterOptions() {
  const [data, setData] = useState<FilterOptions>({ locations: [], tags: [] });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const controller = new AbortController();
    api
      .get<FilterOptions>("/jobs/filters", { signal: controller.signal })
      .then((res) => setData(res ?? { locations: [], tags: [] }))
      .catch((err: unknown) => {
        if (err instanceof Error && err.name === "AbortError") return;
      })
      .finally(() => setLoading(false));
    return () => controller.abort();
  }, []);

  return { data, loading };
}

export function useCompanyJobs(status = "") {
  const [data, setData] = useState<Job[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const fetch = useCallback(() => {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    setLoading(true);
    setError(null);
    const query = status ? `?status=${status}` : "";
    api
      .get<Job[]>(`/companies/me/jobs${query}`, { signal: controller.signal })
      .then((res) => setData(res ?? []))
      .catch((err: unknown) => {
        if (err instanceof Error && err.name === "AbortError") return;
        setError(err instanceof Error ? err.message : "Failed to load jobs");
      })
      .finally(() => setLoading(false));
  }, [status]);

  useEffect(() => {
    fetch();
    return () => abortRef.current?.abort();
  }, [fetch]);

  return { data, loading, error, refetch: fetch };
}
