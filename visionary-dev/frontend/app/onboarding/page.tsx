"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Image from "next/image";
import OnboardingShell from "@/components/shared/OnboardingShell";
import { cn } from "@/lib/utils";

const categories = [
  { value: "student", label: "Student" },
  { value: "teacher", label: "Teacher" },
  { value: "organization", label: "Organization" },
] as const;

type Category = (typeof categories)[number]["value"];

export default function OnboardingCategoryPage() {
  const router = useRouter();
  const [selected, setSelected] = useState<Category>("student");

  function handleContinue() {
    sessionStorage.setItem("onboarding_category", selected);
    router.push("/onboarding/details");
  }

  return (
    <OnboardingShell>
      <div
        className="bg-white w-full flex flex-col items-center"
        style={{
          maxWidth: "800px",
          width: "100%",
          borderRadius: "42px",
          border: "1px solid #8E8E93",
          padding: "clamp(32px, 5vw, 72px) clamp(20px, 4vw, 42px) clamp(32px, 5vw, 64px)",
          gap: "clamp(24px, 4vw, 40px)",
        }}
      >
        <Image
          src="/logo.png"
          alt="Visionary"
          width={56}
          height={56}
          className="w-12 h-12 sm:w-14 sm:h-14"
        />

        <h1
          className="font-medium text-gray-900 text-center text-xl sm:text-2xl lg:text-[28px]"
          style={{ lineHeight: "130%", letterSpacing: "-0.005em" }}
        >
          Welcome to Visionary
        </h1>

        <div className="w-full flex flex-col gap-3" style={{ maxWidth: "600px" }}>
          <p className="text-sm text-gray-600">Select your category:</p>
          <div className="flex flex-col sm:flex-row gap-3 sm:gap-4">
            {categories.map(({ value, label }) => {
              const isDisabled = value === "teacher" || value === "organization";
              return (
                <button
                  key={value}
                  onClick={() => !isDisabled && setSelected(value)}
                  disabled={isDisabled}
                  className={cn(
                    "flex-1 text-sm font-medium transition-colors py-4",
                    isDisabled
                      ? "text-[#9CA3AF] cursor-not-allowed"
                      : selected === value
                      ? "text-[#2563EB]"
                      : "text-[#2563EB] hover:opacity-80"
                  )}
                  style={{
                    minHeight: "56px",
                    borderRadius: "50px",
                    border: isDisabled
                      ? "1px solid #E5E7EB"
                      : selected === value
                      ? "1px solid #BFDBFE"
                      : "1px solid #2563EB",
                    backgroundColor: isDisabled
                      ? "#F9FAFB"
                      : selected === value
                      ? "#EFF6FF"
                      : "transparent",
                  }}
                >
                  {label}
                </button>
              );
            })}
          </div>
        </div>

        <button
          onClick={handleContinue}
          className="text-white font-medium text-base transition-colors hover:bg-[#1D4ED8]"
          style={{
            width: "100%",
            maxWidth: "600px",
            height: "56px",
            borderRadius: "24px",
            backgroundColor: "#2563EB",
          }}
        >
          Continue
        </button>
      </div>
    </OnboardingShell>
  );
}
