import axios from "axios";
import { createContext, useState, useEffect, type ReactNode } from "react";

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
  const [token, setToken] = useState<string | null>(localStorage.getItem("token"));
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Axios interceptor to attach token
  useEffect(() => {
    const reqInterceptor = axios.interceptors.request.use(config => {
      if (token) config.headers.Authorization = `Bearer ${token}`;
      return config;
    });
    return () => {
      axios.interceptors.request.eject(reqInterceptor);
    };
  }, [token]);

  const login = async (email: string, password: string) => {
    setLoading(true);
    setError(null);
    try {
      const res = await axios.post("http://localhost:8080/v1/login", { email, password });
      setToken(res.data.token);
      localStorage.setItem("token", res.data.token);
    } catch (err: any) {
      setError(err.response?.data || "Login failed");
    } finally {
      setLoading(false);
    }
  };

  const register = async (email: string, password: string) => {
    setLoading(true);
    setError(null);
    try {
      await axios.post("http://localhost:8080/v1/register", { email, password });
    } catch (err: any) {
      setError(err.response?.data || "Register failed");
    } finally {
      setLoading(false);
    }
  };

  const logout = () => {
    setToken(null);
    localStorage.removeItem("token");
  };

  // Optional: refresh token logic
//   const refresh = async () => {
//     try {
//       const res = await axios.get("http://localhost:8080/v1/refresh", { withCredentials: true });
//       setToken(res.data.token);
//       localStorage.setItem("token", res.data.token);
//     } catch {}
//   };

  return (
    <AuthContext.Provider value={{ token, login, register, logout, loading, error }}>
      {children}
    </AuthContext.Provider>
  );
};