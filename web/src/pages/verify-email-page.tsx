import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { CheckCircle2, Loader2, XCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { api, ApiError } from "@/lib/api";

type State = "loading" | "success" | "expired" | "error";

export function VerifyEmailPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const [state, setState] = useState<State>("loading");

  useEffect(() => {
    if (!token) {
      setState("error");
      return;
    }

    const controller = new AbortController();

    api
      .get(`/auth/verify-email?token=${encodeURIComponent(token)}`, {
        signal: controller.signal,
      })
      .then(() => setState("success"))
      .catch((err) => {
        if (err instanceof ApiError && err.status === 410) {
          setState("expired");
        } else {
          setState("error");
        }
      });

    return () => controller.abort();
  }, [token]);

  return (
    <div className="flex min-h-[70vh] items-center justify-center px-4">
      <div className="w-full max-w-sm text-center">
        {state === "loading" && (
          <>
            <Loader2 className="mx-auto h-10 w-10 animate-spin text-primary" />
            <h1 className="mt-5 font-display text-2xl font-bold tracking-tight">
              Verifying your email…
            </h1>
            <p className="mt-2 text-sm text-muted-foreground">
              Just a moment.
            </p>
          </>
        )}

        {state === "success" && (
          <>
            <CheckCircle2 className="mx-auto h-10 w-10 text-green-600" />
            <h1 className="mt-5 font-display text-2xl font-bold tracking-tight">
              Email verified
            </h1>
            <p className="mt-2 text-sm text-muted-foreground">
              Your account is ready. You can now log in.
            </p>
            <Button asChild className="mt-6 w-full">
              <Link to="/login">Log in</Link>
            </Button>
          </>
        )}

        {state === "expired" && (
          <>
            <XCircle className="mx-auto h-10 w-10 text-amber-500" />
            <h1 className="mt-5 font-display text-2xl font-bold tracking-tight">
              Link expired
            </h1>
            <p className="mt-2 text-sm text-muted-foreground">
              This verification link has expired. Request a new one from the login page.
            </p>
            <Button asChild variant="outline" className="mt-6 w-full">
              <Link to="/login">Back to login</Link>
            </Button>
          </>
        )}

        {state === "error" && (
          <>
            <XCircle className="mx-auto h-10 w-10 text-destructive" />
            <h1 className="mt-5 font-display text-2xl font-bold tracking-tight">
              Invalid link
            </h1>
            <p className="mt-2 text-sm text-muted-foreground">
              This verification link is invalid or has already been used.
            </p>
            <Button asChild variant="outline" className="mt-6 w-full">
              <Link to="/login">Back to login</Link>
            </Button>
          </>
        )}
      </div>
    </div>
  );
}
