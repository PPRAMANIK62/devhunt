import { getToken } from "@/lib/auth";

const BASE = "/api/v1";

export class ApiError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const res = await fetch(`${BASE}${path}`, { ...options, headers });

  if (res.status === 204) return undefined as T;

  const body = await res.json();

  if (!res.ok) {
    throw new ApiError(
      res.status,
      (body as { code?: string }).code ?? "UNKNOWN",
      (body as { error?: string }).error ?? "Request failed",
    );
  }

  return (body as { data: T }).data;
}

export const api = {
  get: <T>(path: string, opts?: RequestInit) => request<T>(path, opts),
  post: <T>(path: string, data: unknown) =>
    request<T>(path, { method: "POST", body: JSON.stringify(data) }),
  patch: <T>(path: string, data: unknown) =>
    request<T>(path, { method: "PATCH", body: JSON.stringify(data) }),
  delete: <T>(path: string) => request<T>(path, { method: "DELETE" }),
};
