"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import Image from "next/image";
import { signupWithEmail } from "@/lib/auth";
import FloatingInput from "@/components/ui/FloatingInput";
import { HiArrowLeft } from "react-icons/hi";

export default function SignupEmailPage() {
  const router = useRouter();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [errors, setErrors] = useState<{ email?: string; password?: string; form?: string }>({});
  const [loading, setLoading] = useState(false);

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
      await signupWithEmail(email, password);
      router.replace(`/signup/otp?email=${encodeURIComponent(email)}`);
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
        "Signup failed. Please try again.";
      setErrors({ form: msg });
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen bg-white flex flex-col items-center justify-center p-4 gap-8">

      {/* Back arrow — outside the card */}
      <div className="w-full" style={{ maxWidth: "988px" }}>
        <Link
          href="/signup"
          className="flex items-center justify-center text-gray-700 hover:text-gray-900 transition-colors"
          style={{ width: "40px", height: "40px", borderRadius: "50%", border: "1px solid #E5E7EB" }}
        >
          <HiArrowLeft size={22} />        </Link>
      </div>

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
            <Image
              src="/logo.png"
              alt="Visionary"
              width={86}
              height={86}
              style={{ width: "86px", height: "86px" }}
            />
            <h1
              className="font-medium text-gray-900"
              style={{ fontSize: "36px", lineHeight: "1.2" }}
            >
              Create a Visionary<br />Account
            </h1>
          </div>
        </div>

        {/* Divider on mobile */}
        <div className="lg:hidden h-px bg-gray-200 w-full" />

        {/* Right */}
        <form onSubmit={handleSubmit} className="flex flex-col justify-center" style={{ gap: "30px" }}>
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
            autoComplete="new-password"
          />
          <button
            type="submit"
            disabled={loading}
            className={`flex items-center justify-center w-full font-medium text-base transition-colors disabled:cursor-not-allowed ${
              email && password
                ? "bg-[#2563EB] hover:bg-[#1D4ED8] text-white border-transparent"
                : "bg-white text-gray-400 border border-gray-300"
            }`}
            style={{ height: "56px", borderRadius: "50px" }}
          >
            {loading ? "Creating account…" : "Next"}
          </button>
        </form>

        {/* Bottom row */}
        <div className="lg:col-span-2 flex items-center justify-between pt-2">
          <Link href="/login" className="text-sm font-medium text-blue-600 hover:text-blue-700">
            I already have an account
          </Link>
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
