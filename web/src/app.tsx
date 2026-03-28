import { Navigate, Route, Routes } from "react-router-dom";
import { Layout } from "@/components/layout/layout";
import { useAuth } from "@/context/auth-context";
import { ApplicationsPage } from "@/pages/applications-page";
import { DashboardPage } from "@/pages/dashboard-page";
import { HomePage } from "@/pages/home-page";
import { JobDetailPage } from "@/pages/job-detail-page";
import { JobsPage } from "@/pages/jobs-page";
import { LoginPage } from "@/pages/login-page";
import { RegisterPage } from "@/pages/register-page";

function GuestRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();
  return isAuthenticated ? <Navigate to="/" replace /> : <>{children}</>;
}

function AuthRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />;
}

function CompanyRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, role } = useAuth();
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  if (role !== "company") return <Navigate to="/" replace />;
  return <>{children}</>;
}

function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<HomePage />} />
        <Route path="/jobs" element={<JobsPage />} />
        <Route path="/jobs/:id" element={<JobDetailPage />} />
        <Route
          path="/login"
          element={
            <GuestRoute>
              <LoginPage />
            </GuestRoute>
          }
        />
        <Route
          path="/register"
          element={
            <GuestRoute>
              <RegisterPage />
            </GuestRoute>
          }
        />
        <Route
          path="/dashboard"
          element={
            <CompanyRoute>
              <DashboardPage />
            </CompanyRoute>
          }
        />
        <Route
          path="/applications"
          element={
            <AuthRoute>
              <ApplicationsPage />
            </AuthRoute>
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  );
}

export default App;
