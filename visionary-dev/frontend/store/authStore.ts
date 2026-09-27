import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { AuthState } from "@/types";
import { logout as backendLogout } from "@/lib/auth";

// As of DS-98 the auth_token cookie is set by the backend (httpOnly) via
// the Next.js rewrite. The frontend no longer writes document.cookie.
// `token` is kept here only for the legacy bearer fallback while existing
// localStorage sessions drain.
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      isAuthenticated: false,

      setAuth: (token, user) => {
        set({ token, user, isAuthenticated: true });
      },

      setUserFromCookie: (user) => {
        set({ token: null, user, isAuthenticated: true });
      },

      clear: async () => {
        // Wait for the backend to clear the httpOnly cookie before any
        // caller routes away — otherwise the post-logout navigation
        // re-enters middleware with the cookie still present and bounces
        // the user back into the app.
        await backendLogout();
        set({ token: null, user: null, isAuthenticated: false });
      },
    }),
    { name: "auth-storage" }
  )
);
