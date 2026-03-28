import { useCallback, useEffect, useRef, useState } from "react";
import { api } from "@/lib/api";
import type { Application, ApplicationStatus } from "@/types";

export function useJobApplications(jobId: string) {
  const [data, setData] = useState<Application[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const fetch = useCallback(() => {
    if (!jobId) return;
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    setLoading(true);
    setError(null);
    api
      .get<Application[]>(`/jobs/${jobId}/applications`, { signal: controller.signal })
      .then((d) => setData(d ?? []))
      .catch((err: unknown) => {
        if (err instanceof Error && err.name === "AbortError") return;
        setError(err instanceof Error ? err.message : "Failed to load applicants");
      })
      .finally(() => setLoading(false));
  }, [jobId]);

  useEffect(() => {
    fetch();
    return () => abortRef.current?.abort();
  }, [fetch]);

  return { data, loading, error, refetch: fetch };
}

export function useMyApplications(enabled = true) {
  const [data, setData] = useState<Application[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const fetch = useCallback(() => {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    setLoading(true);
    setError(null);
    api
      .get<Application[]>("/applications", { signal: controller.signal })
      .then((d) => setData(d ?? []))
      .catch((err: unknown) => {
        if (err instanceof Error && err.name === "AbortError") return;
        setError(err instanceof Error ? err.message : "Failed to load applications");
      })
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (!enabled) return;
    fetch();
    return () => abortRef.current?.abort();
  }, [fetch, enabled]);

  return { data, loading, error, refetch: fetch };
}

export function useApply() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const execute = useCallback(
    async (jobId: string, coverNote?: string): Promise<Application> => {
      setLoading(true);
      setError(null);
      try {
        return await api.post<Application>(`/jobs/${jobId}/applications`, {
          cover_note: coverNote ?? "",
        });
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Failed to apply";
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

export function useUpdateApplicationStatus() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const execute = useCallback(
    async (id: string, status: ApplicationStatus): Promise<Application> => {
      setLoading(true);
      setError(null);
      try {
        return await api.patch<Application>(`/applications/${id}/status`, { status });
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Failed to update status";
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
