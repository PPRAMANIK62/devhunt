import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
} from "react";
import { clearToken, decodeToken, getToken, setToken } from "@/lib/auth";
import type { UserRole } from "@/types";

interface AuthState {
  token: string | null;
  userId: string | null;
  role: UserRole | null;
}

interface AuthContextValue extends AuthState {
  login: (token: string) => void;
  logout: () => void;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextValue | null>(null);

function resolveState(): AuthState {
  const token = getToken();
  if (!token) return { token: null, userId: null, role: null };
  const payload = decodeToken(token);
  if (!payload) return { token: null, userId: null, role: null };
  return { token, userId: payload.user_id, role: payload.role as UserRole };
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<AuthState>(resolveState);

  const login = useCallback((token: string) => {
    setToken(token);
    const payload = decodeToken(token);
    if (payload) {
      setState({
        token,
        userId: payload.user_id,
        role: payload.role as UserRole,
      });
    }
  }, []);

  const logout = useCallback(() => {
    clearToken();
    setState({ token: null, userId: null, role: null });
  }, []);

  const value = useMemo(
    () => ({ ...state, login, logout, isAuthenticated: !!state.token }),
    [state, login, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
