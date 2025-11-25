import axios from "axios";
import { createContext, useState, useEffect, type ReactNode } from "react";
import { jwtDecode } from "jwt-decode";
import { useNavigate } from "react-router-dom";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8081";

interface TokenPayload { exp: number; }
interface AuthContextType {
  token: string | null;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, username: string) => Promise<{ message: string }>;
  loginWithGoogle: (code: string) => Promise<void>;
  loginWithGitHub: (code: string) => Promise<void>;
  loginWithApple: (code: string, idToken?: string) => Promise<void>;
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
  const navigate = useNavigate()
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
        const res = await axios.get(`${API_URL}/v1/refresh`, { withCredentials: true });
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
      const res = await axios.post(`${API_URL}/v1/login`, { email, password });
          setToken(res.data.token); 

	  localStorage.setItem("token", res.data.token);
      	  navigate("/"); 
    } catch (err: any) { setError(err.response?.data || "Login failed"); }
    finally { setLoading(false); }
  };

  const register = async (email: string, password: string, username: string) => {
    setLoading(true); setError(null);
    try {
      const res = await axios.post(`${API_URL}/v1/register`, { email, password, username });
      return res.data;
    }
    catch (err: any) {
      const errorMessage = err.response?.data || "Register failed";
      setError(errorMessage);
      throw new Error(errorMessage);
    }
    finally { setLoading(false); }
  };

  const loginWithGoogle = async (code: string) => {
    setLoading(true); setError(null);
    try {
      const res = await axios.post(`${API_URL}/v1/auth/google`, { code });
      setToken(res.data.token);
      localStorage.setItem("token", res.data.token);
      navigate("/");
    } catch (err: any) {
      setError(err.response?.data || "Google login failed");
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const loginWithGitHub = async (code: string) => {
    setLoading(true); setError(null);
    try {
      const res = await axios.post(`${API_URL}/v1/auth/github`, { code });
      setToken(res.data.token);
      localStorage.setItem("token", res.data.token);
      navigate("/");
    } catch (err: any) {
      setError(err.response?.data || "GitHub login failed");
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const loginWithApple = async (code: string, idToken?: string) => {
    setLoading(true); setError(null);
    try {
      const res = await axios.post(`${API_URL}/v1/auth/apple`, { code, id_token: idToken });
      setToken(res.data.token);
      localStorage.setItem("token", res.data.token);
      navigate("/");
    } catch (err: any) {
      setError(err.response?.data || "Apple login failed");
      throw err;
    } finally {
      setLoading(false);
    }
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
    <AuthContext.Provider value={{ token, login, register, loginWithGoogle, loginWithGitHub, loginWithApple, logout, loading, error }}>
      {children}
    </AuthContext.Provider>
  );
};
