"use client";

import { useState, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import Image from "next/image";
import FloatingInput from "@/components/ui/FloatingInput";
import { loginWithEmail } from "@/lib/auth";
import { useAuthStore } from "@/store/authStore";
import { HiArrowLeft } from "react-icons/hi";

export default function LoginEmailPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const setAuth = useAuthStore((s) => s.setAuth);

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [errors, setErrors] = useState<{ email?: string; password?: string; form?: string }>({});
  const [loading, setLoading] = useState(false);
  const [deletedBanner, setDeletedBanner] = useState(false);

  useEffect(() => {
    if (searchParams.get("deleted") === "1") setDeletedBanner(true);
  }, [searchParams]);

  function validate() {
    const e: typeof errors = {};
    if (!email) e.email = "Email is required";
    else if (!/\S+@\S+\.\S+/.test(email)) e.email = "Invalid email";
    if (!password) e.password = "Password is required";
    else if (password.length < 6) e.password = "Minimum 6 characters";
    return e;
  }

  async function handleSubmit(ev: React.FormEvent) {
    ev.preventDefault();
    const e = validate();
    if (Object.keys(e).length) { setErrors(e); return; }
    setErrors({});
    setLoading(true);
    try {
      const { token, user } = await loginWithEmail(email, password);
      setAuth(token, user);
      router.replace(user.onboardingCompleted ? "/dashboard" : "/onboarding");
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
        "Invalid email or password";
      if (msg === "ACCOUNT_DELETED") {
        setErrors({ form: "Your account has been deleted. All data will be permanently removed within 7 days." });
      } else {
        setErrors({ form: msg });
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen bg-white flex flex-col items-center justify-center p-4 gap-6">

      {/* Back arrow — outside the card */}
      <div className="w-full" style={{ maxWidth: "988px" }}>
        <Link
          href="/login"
          className="flex items-center justify-center text-gray-700 hover:text-gray-900 transition-colors"
          style={{ width: "40px", height: "40px", borderRadius: "50%", border: "1px solid #E5E7EB" }}
        >
          <HiArrowLeft size={22} />
        </Link>
      </div>

      <div
        className="bg-white w-full flex flex-col lg:grid"
        style={{
          maxWidth: "988px",
          borderRadius: "42px",
          border: "1px solid rgba(0,0,0,0.12)",
          padding: "clamp(32px, 5vw, 60px) clamp(20px, 4vw, 42px) clamp(28px, 4vw, 50px)",
          gap: "clamp(32px, 5vw, 64px)",
          gridTemplateColumns: "1fr 1.3fr",
        }}
      >
        {/* Left */}
        <div className="flex flex-col gap-6 lg:gap-16">
          <div className="flex flex-col items-center lg:items-start gap-3 lg:gap-4">
            <Image
              src="/logo.png"
              alt="Visionary"
              width={86}
              height={86}
              className="w-14 h-14 lg:w-[86px] lg:h-[86px]"
            />
            <h1
              className="font-medium text-gray-900 text-2xl sm:text-3xl lg:text-[36px] text-center lg:text-left"
              style={{ lineHeight: "1.2" }}
            >
              Login to Your<span className="hidden lg:inline"><br /></span>{" "}Visionary Account
            </h1>
          </div>
        </div>

        {/* Divider on mobile */}
        <div className="lg:hidden h-px bg-gray-200 w-full" />

        {/* Right */}
        <form onSubmit={handleSubmit} className="flex flex-col justify-center" style={{ gap: "24px" }}>
          {deletedBanner && (
            <div className="bg-orange-50 border border-orange-200 rounded-xl px-4 py-3">
              <p className="text-sm font-medium text-orange-700">Deletion request submitted successfully.</p>
              <p className="text-xs text-orange-500 mt-0.5">Check your email — all data will be permanently deleted in 7 days.</p>
            </div>
          )}
          {errors.form && (
            <p className="text-sm text-red-500 bg-red-50 rounded-xl px-4 py-2">{errors.form}</p>
          )}
          <FloatingInput
            label="Enter Your Email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            error={errors.email}
            autoComplete="email"
          />
          <FloatingInput
            label="Enter Password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            error={errors.password}
            autoComplete="current-password"
          />
          <button
            type="submit"
            disabled={loading}
            className={`flex items-center justify-center w-full font-medium text-base transition-colors disabled:cursor-not-allowed ${email && password
                ? "bg-[#2563EB] hover:bg-[#1D4ED8] text-white border-transparent"
                : "bg-white text-gray-400 border border-gray-300"
              }`}
            style={{ height: "56px", borderRadius: "50px" }}
          >
            {loading ? "Logging in…" : "Next"}
          </button>
        </form>

      </div>


    </div>
  );
}
