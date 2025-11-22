// Centralized API configuration
export const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8081";

// Helper to get full URL for static files (images, models, etc.)
export const getStaticUrl = (path: string) => {
  if (!path) return '';
  if (path.startsWith('http')) return path;
  return `${API_URL}${path}`;
};

// Helper for WebSocket URLs
export const getWebSocketUrl = (path: string, token?: string) => {
  const wsProtocol = API_URL.startsWith('https') ? 'wss' : 'ws';
  const wsHost = API_URL.replace(/^https?:\/\//, '');
  const url = `${wsProtocol}://${wsHost}${path}`;
  return token ? `${url}?token=${encodeURIComponent(token)}` : url;
};
