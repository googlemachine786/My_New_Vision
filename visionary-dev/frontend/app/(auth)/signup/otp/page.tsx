"use client";

import { useEffect, useRef, useState, useCallback, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Image from "next/image";
import { HiArrowLeft } from "react-icons/hi";
import { verifyOtp, resendOtp } from "@/lib/auth";
import { useAuthStore } from "@/store/authStore";

const OTP_LENGTH = 6;
const RESEND_SECONDS = 60;

export default function OtpPage() {
  return (
    <Suspense>
      <OtpPageInner />
    </Suspense>
  );
}

function OtpPageInner() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const email = searchParams.get("email") ?? "";
  const setAuth = useAuthStore((s) => s.setAuth);

  const [digits, setDigits] = useState<string[]>(Array(OTP_LENGTH).fill(""));
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [countdown, setCountdown] = useState(RESEND_SECONDS);
  const [resending, setResending] = useState(false);
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  // Countdown timer
  useEffect(() => {
    if (countdown <= 0) return;
    const t = setTimeout(() => setCountdown((c) => c - 1), 1000);
    return () => clearTimeout(t);
  }, [countdown]);

  function handleChange(index: number, value: string) {
    const char = value.replace(/\D/g, "").slice(-1); // digits only
    const next = [...digits];
    next[index] = char;
    setDigits(next);
    setError("");
    if (char && index < OTP_LENGTH - 1) {
      inputRefs.current[index + 1]?.focus();
    }
  }

  function handleKeyDown(index: number, e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === "Backspace" && !digits[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  }

  function handlePaste(e: React.ClipboardEvent) {
    e.preventDefault();
    const pasted = e.clipboardData.getData("text").replace(/\D/g, "").slice(0, OTP_LENGTH);
    const next = Array(OTP_LENGTH).fill("");
    pasted.split("").forEach((c, i) => { next[i] = c; });
    setDigits(next);
    inputRefs.current[Math.min(pasted.length, OTP_LENGTH - 1)]?.focus();
  }

  const handleSubmit = useCallback(async () => {
    const otp = digits.join("");
    if (otp.length < OTP_LENGTH) { setError("Please enter the complete OTP"); return; }
    setError("");
    setLoading(true);
    try {
      const { token, user } = await verifyOtp(email, otp);
      setAuth(token, user);
      router.replace(user.onboardingCompleted ? "/dashboard" : "/onboarding");
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
        "Invalid or expired OTP";
      setError(msg);
      setDigits(Array(OTP_LENGTH).fill(""));
      inputRefs.current[0]?.focus();
    } finally {
      setLoading(false);
    }
  }, [digits, email]);

  async function handleResend() {
    if (countdown > 0 || resending) return;
    setResending(true);
    setError("");
    try {
      await resendOtp(email);
      setCountdown(RESEND_SECONDS);
      setDigits(Array(OTP_LENGTH).fill(""));
      inputRefs.current[0]?.focus();
    } catch {
      setError("Failed to resend OTP. Try again.");
    } finally {
      setResending(false);
    }
  }

  const isFilled = digits.every((d) => d !== "");

  return (
    <div className="min-h-screen bg-white flex flex-col items-center justify-center p-4 gap-8">

      {/* Back arrow */}
      <div className="w-full" style={{ maxWidth: "988px" }}>
        <button
          onClick={() => router.back()}
          className="flex items-center justify-center text-gray-700 hover:text-gray-900 transition-colors"
          style={{ width: "40px", height: "40px", borderRadius: "50%", border: "1px solid #E5E7EB" }}
        >
          <HiArrowLeft size={22} />
        </button>
      </div>

      {/* Card */}
      <div
        className="bg-white w-full flex flex-col lg:grid"
        style={{
          maxWidth: "988px",
          borderRadius: "42px",
          border: "1px solid rgba(0,0,0,0.12)",
          padding: "60px 42px 50px",
          minHeight: "382px",
          gap: "64px",
          gridTemplateColumns: "1fr 1.3fr",
        }}
      >
        {/* Left */}
        <div className="flex flex-col gap-16">
          <div className="flex flex-col gap-4">
            <Image src="/logo.png" alt="Visionary" width={86} height={86} style={{ width: "86px", height: "86px" }} />
            <h1 className="font-medium text-gray-900" style={{ fontSize: "36px", lineHeight: "1.2" }}>
              Create a Visionary<br />Account
            </h1>
          </div>
        </div>

        {/* Divider on mobile */}
        <div className="lg:hidden h-px bg-gray-200 w-full" />

        {/* Right */}
        <div className="flex flex-col justify-center" style={{ gap: "30px" }}>
          <div>
            <p className="text-sm text-gray-500 mb-6">
              Please enter the OTP received on <span className="font-medium text-gray-800">{email}</span>
            </p>

            {/* OTP boxes */}
            <div className="flex gap-3 mb-2" onPaste={handlePaste}>
              {digits.map((digit, i) => (
                <input
                  key={i}
                  ref={(el) => { inputRefs.current[i] = el; }}
                  type="text"
                  inputMode="numeric"
                  maxLength={1}
                  value={digit}
                  onChange={(e) => handleChange(i, e.target.value)}
                  onKeyDown={(e) => handleKeyDown(i, e)}
                  className="text-center text-xl font-semibold text-gray-900 border border-gray-300 rounded-xl focus:outline-none focus:border-blue-500 transition-colors"
                  style={{ width: "64px", height: "64px" }}
                />
              ))}
            </div>

            {error && <p className="text-sm text-red-500 mt-2">{error}</p>}
          </div>

          {/* Next button */}
          <button
            onClick={handleSubmit}
            disabled={loading || !isFilled}
            className={`flex items-center justify-center w-full font-medium text-base transition-colors disabled:cursor-not-allowed ${
              isFilled
                ? "bg-[#2563EB] hover:bg-[#1D4ED8] text-white"
                : "bg-white text-gray-400 border border-gray-300"
            }`}
            style={{ height: "56px", borderRadius: "50px" }}
          >
            {loading ? "Verifying…" : "Next"}
          </button>

          {/* Resend */}
          <div className="flex items-center gap-2 text-sm">
            <span className="text-gray-400">Didn&apos;t receive the code?</span>
            {countdown > 0 ? (
              <span className="text-gray-400">
                Resend in <span className="text-blue-500 font-medium">{countdown}s</span>
              </span>
            ) : (
              <button
                onClick={handleResend}
                disabled={resending}
                className="text-blue-600 font-medium hover:text-blue-700 disabled:opacity-50"
              >
                {resending ? "Sending…" : "Resend OTP"}
              </button>
            )}
          </div>
        </div>

        {/* Bottom row */}
        <div className="lg:col-span-2 flex items-center justify-between pt-2">
          <p className="text-xs text-gray-400">
            OTP expires in 10 minutes
          </p>
          <p className="text-xs text-gray-400">
            By continuing you agree to our{" "}
            <a href="#" className="text-blue-600 hover:text-blue-700">Privacy Policy</a> and{" "}
            <a href="#" className="text-blue-600 hover:text-blue-700">Terms of Service</a>.
          </p>
        </div>
      </div>
    </div>
  );
}