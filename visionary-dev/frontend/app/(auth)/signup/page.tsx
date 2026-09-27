import Link from "next/link";
import Image from "next/image";
import { getGoogleAuthUrl } from "@/lib/auth";

export default function SignupPage() {
  const googleUrl = getGoogleAuthUrl();

  return (
    <div className="min-h-screen bg-white flex items-center justify-center p-4">
      <div
        className="bg-white w-full flex flex-col lg:grid lg:grid-cols-2"
        style={{
          maxWidth: "988px",
          borderRadius: "42px",
          border: "1px solid rgba(0,0,0,0.12)",
          padding: "clamp(32px, 5vw, 100px) clamp(20px, 4vw, 42px) clamp(28px, 4vw, 80px)",
          gap: "clamp(32px, 5vw, 64px)",
        }}
      >
        {/* Left */}
        <div className="flex flex-col gap-6 lg:gap-16">
          <div className="flex flex-col items-center lg:items-start gap-3 lg:gap-0">
            <Image
              src="/logo.png"
              alt="Visionary"
              width={86}
              height={86}
              className="w-14 h-14 lg:w-[86px] lg:h-[86px] lg:mb-6"
            />
            <h1
              className="font-medium text-gray-900 text-2xl sm:text-3xl lg:text-[36px] text-center lg:text-left"
              style={{ lineHeight: "1.2" }}
            >
              Create a Visionary<span className="hidden lg:inline"><br /></span>{" "}Account
            </h1>
          </div>
        </div>

        {/* Divider on mobile */}
        <div className="lg:hidden h-px bg-gray-200 w-full" />

        {/* Right */}
        <div className="flex flex-col justify-center" style={{ gap: "10px" }}>
          <div className="flex flex-col gap-3">
            <Link
              href="/signup/email"
              className="flex items-center justify-center w-full text-white font-medium text-base transition-colors bg-[#2563EB] hover:bg-[#1D4ED8]"
              style={{ height: "56px", borderRadius: "50px" }}
            >
              Sign Up with Email &amp; Password
            </Link>

            <div className="flex items-center gap-3">
              <div className="flex-1 h-px bg-gray-200" />
              <span className="text-xs text-gray-400 uppercase tracking-wide">or</span>
              <div className="flex-1 h-px bg-gray-200" />
            </div>

            <div className="flex justify-center gap-4">
              <a
                href={googleUrl}
                className="flex items-center justify-center hover:bg-gray-50 transition-colors"
                title="Sign up with Google"
                style={{ width: "48px", height: "48px", borderRadius: "6px", border: "1px solid #E5E7EB" }}
              >
                <Image src="/Google Icon.png" alt="Google" width={20} height={20} />
              </a>
              <span
                title="Apple sign up coming soon"
                className="flex items-center justify-center opacity-40 cursor-not-allowed"
                style={{ width: "48px", height: "48px", borderRadius: "6px", border: "1px solid #E5E7EB" }}
              >
                <Image src="/apple-icon.png" alt="Apple" width={20} height={20} />
              </span>
              <span
                title="Microsoft sign up coming soon"
                className="flex items-center justify-center opacity-40 cursor-not-allowed"
                style={{ width: "48px", height: "48px", borderRadius: "6px", border: "1px solid #E5E7EB" }}
              >
                <Image src="/microsoft-icon.png" alt="Microsoft" width={20} height={20} />
              </span>
            </div>
          </div>
        </div>

        {/* Bottom row */}
        <div className="lg:col-span-2 grid grid-cols-1 lg:grid-cols-2 gap-3 pt-2">
          <div>
            <Link href="/login" className="text-sm font-medium text-blue-600 hover:text-blue-700">
              I already have an account
            </Link>
          </div>
          <div className="lg:text-right">
            <p className="text-xs text-gray-400">
              By continuing you agree to our{" "}
              <a href="#" className="text-blue-600 hover:text-blue-700">Privacy Policy</a> and{" "}
              <a href="#" className="text-blue-600 hover:text-blue-700">Terms of Service</a>.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
