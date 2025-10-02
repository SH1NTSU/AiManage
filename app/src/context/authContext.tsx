import axios from "axios";
import { createContext, useState, useEffect, type ReactNode } from "react";
import { jwtDecode } from "jwt-decode";

interface TokenPayload { exp: number; }
interface AuthContextType {
  token: string | null;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => void;
  loading: boolean;
  error: string | null;
}

export const AuthContext = createContext<AuthContextType | null>(null);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [token, setToken] = useState<string | null>(() => {
    const t = localStorage.getItem("token");
    try { if (t) jwtDecode<TokenPayload>(t); return t; } catch { localStorage.removeItem("token"); return null; }
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 🧠 track ongoing refresh
  let refreshing = false;
  let refreshPromise: Promise<void> | null = null;

  const isTokenExpired = (token: string) => {
    try { const decoded = jwtDecode<TokenPayload>(token); return decoded.exp < Math.floor(Date.now() / 1000); }
    catch { return true; }
  };

  const refresh = async () => {
    if (refreshing && refreshPromise) return refreshPromise; // return the ongoing refresh
    refreshing = true;
    refreshPromise = (async () => {
      try {
        const res = await axios.get("http://localhost:8080/v1/refresh", { withCredentials: true });
        console.log("✅ Token refreshed");
        setToken(res.data.token);
        localStorage.setItem("token", res.data.token);
      } catch {
        console.warn("⚠️ Refresh failed, logging out...");
        logout();
      } finally {
        refreshing = false;
        refreshPromise = null;
      }
    })();
    return refreshPromise;
  };

  // Axios interceptor
  useEffect(() => {
    const reqInterceptor = axios.interceptors.request.use(async (config) => {
      const currentToken = localStorage.getItem("token");
      if (currentToken && isTokenExpired(currentToken)) {
        console.log("🔄 Token expired. Trying to refresh...");
        await refresh();
      }
      const latestToken = localStorage.getItem("token");
      if (latestToken) config.headers.Authorization = `Bearer ${latestToken}`;
      return config;
    });
    return () => axios.interceptors.request.eject(reqInterceptor);
  }, [token]);

  const login = async (email: string, password: string) => {
    setLoading(true); setError(null);
    try {
      const res = await axios.post("http://localhost:8080/v1/login", { email, password });
      setToken(res.data.token); localStorage.setItem("token", res.data.token);
    } catch (err: any) { setError(err.response?.data || "Login failed"); }
    finally { setLoading(false); }
  };

  const register = async (email: string, password: string) => {
    setLoading(true); setError(null);
    try { await axios.post("http://localhost:8080/v1/register", { email, password }); }
    catch (err: any) { setError(err.response?.data || "Register failed"); }
    finally { setLoading(false); }
  };

  const logout = () => { setToken(null); localStorage.removeItem("token"); };

  // Proactive refresh
  useEffect(() => {
    if (!token) return;
    let decoded: TokenPayload;
    try { decoded = jwtDecode<TokenPayload>(token); } catch { logout(); return; }
    const now = Date.now() / 1000;
    const timeUntilExpiry = decoded.exp - now - 60;
    if (timeUntilExpiry > 0) {
      const timer = setTimeout(refresh, timeUntilExpiry * 1000);
      return () => clearTimeout(timer);
    } else { refresh(); }
  }, [token]);

  return (
    <AuthContext.Provider value={{ token, login, register, logout, loading, error }}>
      {children}
    </AuthContext.Provider>
  );
};
