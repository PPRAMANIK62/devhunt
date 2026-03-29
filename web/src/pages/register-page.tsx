import { useState } from "react";
import { Link } from "react-router-dom";
import { toast } from "sonner";
import { MailCheck } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { api, ApiError } from "@/lib/api";
import type { User } from "@/types";

interface RegisterResponse {
  user: User;
}

export function RegisterPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<"seeker" | "company">("seeker");
  const [loading, setLoading] = useState(false);
  const [registered, setRegistered] = useState(false);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    try {
      await api.post<RegisterResponse>("/auth/register", { email, password, role });
      setRegistered(true);
    } catch (err) {
      if (err instanceof ApiError && err.status === 429) {
        toast.error("Too many attempts. Please wait a minute before trying again.");
      } else {
        toast.error(err instanceof Error ? err.message : "Registration failed");
      }
    } finally {
      setLoading(false);
    }
  }

  if (registered) {
    return (
      <div className="flex min-h-[70vh] items-center justify-center">
        <Card className="w-full max-w-sm text-center">
          <CardHeader>
            <MailCheck className="mx-auto h-9 w-9 text-primary" />
            <CardTitle className="font-display text-2xl">Check your inbox</CardTitle>
            <CardDescription>
              We sent a verification link to{" "}
              <span className="font-medium text-foreground">{email}</span>.
              Click it to activate your account.
            </CardDescription>
          </CardHeader>
          <CardFooter className="flex flex-col gap-3">
            <Button asChild className="w-full">
              <Link to="/login">Go to login</Link>
            </Button>
          </CardFooter>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex min-h-[70vh] items-center justify-center">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle className="font-display text-2xl">Sign up</CardTitle>
          <CardDescription>Create your DevHunt account.</CardDescription>
        </CardHeader>
        <form onSubmit={handleSubmit}>
          <CardContent className="flex flex-col gap-4">
            <div className="space-y-1.5">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                autoComplete="new-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                minLength={8}
                required
              />
              <p className="font-mono text-xs text-muted-foreground">
                Minimum 8 characters
              </p>
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="role">I am a</Label>
              <Select
                value={role}
                onValueChange={(v) => setRole(v as "seeker" | "company")}
              >
                <SelectTrigger id="role">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="seeker">Job Seeker</SelectItem>
                  <SelectItem value="company">Company / Hiring</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
          <CardFooter className="flex flex-col gap-3">
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "Creating account..." : "Create Account"}
            </Button>
            <p className="text-sm text-muted-foreground">
              Already have an account?{" "}
              <Link to="/login" className="text-foreground underline">
                Log in
              </Link>
            </p>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
}
