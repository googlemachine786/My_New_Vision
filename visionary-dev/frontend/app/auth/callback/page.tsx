"use client";

import { useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense } from "react";
import { useAuthStore } from "@/store/authStore";
import { getMe } from "@/lib/auth";

function CallbackHandler() {
  const router = useRouter();
  const params = useSearchParams();
  const setUserFromCookie = useAuthStore((s) => s.setUserFromCookie);

  useEffect(() => {
    let cancelled = false;
    const onboardingCompleted = params.get("onboardingCompleted") === "true";

    (async () => {
      try {
        // Cookie is already set by the backend on the OAuth redirect; we
        // just ask the server who we are now and hydrate UI state.
        const user = await getMe();
        if (cancelled) return;
        setUserFromCookie(user);
        router.replace(onboardingCompleted ? "/dashboard" : "/onboarding");
      } catch {
        if (cancelled) return;
        router.replace("/login");
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [params, router, setUserFromCookie]);

  return (
    <div className="min-h-screen flex items-center justify-center">
      <p className="text-gray-500 text-sm">Signing you in…</p>
    </div>
  );
}

export default function AuthCallbackPage() {
  return (
    <Suspense>
      <CallbackHandler />
    </Suspense>
  );
}
