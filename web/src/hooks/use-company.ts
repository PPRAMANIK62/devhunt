import { useCallback, useEffect, useRef, useState } from "react";
import { ApiError, api } from "@/lib/api";
import type { Company } from "@/types";

export function useMyCompany() {
  const [data, setData] = useState<Company | null>(null);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);
  const abortRef = useRef<AbortController | null>(null);

  const fetch = useCallback(() => {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    setLoading(true);
    setNotFound(false);
    api
      .get<Company>("/companies/me", { signal: controller.signal })
      .then((c) => {
        setData(c);
        setNotFound(false);
      })
      .catch((err: unknown) => {
        if (err instanceof Error && err.name === "AbortError") return;
        if (err instanceof ApiError && err.status === 404) {
          setNotFound(true);
        }
        setData(null);
      })
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    fetch();
    return () => abortRef.current?.abort();
  }, [fetch]);

  return { data, loading, notFound, refetch: fetch };
}

export function useCompany(id: string) {
  const [data, setData] = useState<Company | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!id) return;
    const controller = new AbortController();
    setLoading(true);
    api
      .get<Company>(`/companies/${id}`, { signal: controller.signal })
      .then(setData)
      .catch((err: unknown) => {
        if (err instanceof Error && err.name === "AbortError") return;
        setData(null);
      })
      .finally(() => setLoading(false));
    return () => controller.abort();
  }, [id]);

  return { data, loading };
}

interface CompanyFormData {
  name: string;
  slug: string;
  description?: string;
  website?: string;
}

export function useCreateCompany() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const execute = useCallback(
    async (data: CompanyFormData): Promise<Company> => {
      setLoading(true);
      setError(null);
      try {
        return await api.post<Company>("/companies", data);
      } catch (err) {
        const msg =
          err instanceof Error ? err.message : "Failed to create company";
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

export function useUpdateCompany() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const execute = useCallback(
    async (data: Partial<CompanyFormData>): Promise<Company> => {
      setLoading(true);
      setError(null);
      try {
        return await api.patch<Company>("/companies/me", data);
      } catch (err) {
        const msg =
          err instanceof Error ? err.message : "Failed to update company";
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
