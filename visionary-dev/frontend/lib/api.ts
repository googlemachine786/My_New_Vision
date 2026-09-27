import axios from "axios";

// All API calls go to a same-origin path; Next.js rewrites /api/* to the
// backend so the browser sees one origin. This lets the backend set
// httpOnly cookies that the frontend can use end-to-end.
const api = axios.create({
  baseURL: "/api",
  withCredentials: true,
});

api.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    try {
      const raw = localStorage.getItem("auth-storage");
      if (raw) {
        const parsed = JSON.parse(raw);
        const token = parsed?.state?.token;
        if (token) {
          // Legacy bearer fallback for sessions issued before the cookie
          // migration. The backend accepts cookie OR bearer; cookie wins.
          config.headers.Authorization = `Bearer ${token}`;
        }
      }
    } catch {
      // ignore
    }
  }
  return config;
});

api.interceptors.response.use(
  (res) => res,
  async (error) => {
    if (error.response?.status === 401 && typeof window !== "undefined") {
      // Only redirect if the user was already authenticated.
      // Don't redirect on login/auth endpoints — let the form handle the error.
      const authPaths = [
        "/auth/login",
        "/auth/register",
        "/auth/verify-otp",
        "/auth/google",
        "/auth/logout",
      ];
      const requestUrl: string = error.config?.url || "";
      const isAuthRequest = authPaths.some((p) => requestUrl.includes(p));
      if (!isAuthRequest) {
        localStorage.removeItem("auth-storage");
        // The auth_token cookie is httpOnly, so JS cannot clear it. Ask the
        // backend to clear it server-side; ignore failures so we still
        // redirect and let the user re-authenticate.
        try {
          await api.post("/auth/logout");
        } catch {
          // ignore
        }
        window.location.href = "/login";
      }
    }
    return Promise.reject(error);
  }
);

export default api;
