"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Image from "next/image";
import OnboardingShell from "@/components/shared/OnboardingShell";
import FloatingInput from "@/components/ui/FloatingInput";
import { cn } from "@/lib/utils";
import { completeOnboarding, getBoards, getGrades } from "@/lib/auth";
import { useAuthStore } from "@/store/authStore";
import type { BoardMaster, GradeMaster } from "@/types";

export default function OnboardingDetailsPage() {
  const router = useRouter();
  const { setAuth } = useAuthStore();

  const [category, setCategory] = useState<string>("student");
  const [fullName, setFullName] = useState("");
  const [grade, setGrade] = useState("");
  const [board, setBoard] = useState("");           // stores board_master.id
  const [showOtherDropdown, setShowOtherDropdown] = useState(false);
  const [subject, setSubject] = useState("");
  const [orgName, setOrgName] = useState("");
  const [orgType, setOrgType] = useState("");
  const [boards, setBoards] = useState<BoardMaster[]>([]);
  const [grades, setGrades] = useState<GradeMaster[]>([]);
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    const cat = sessionStorage.getItem("onboarding_category") || "student";
    setCategory(cat);
    getBoards().then(setBoards).catch(() => { });
    getGrades().then(setGrades).catch(() => { });
  }, []);

  const isStudent = category === "student";
  const isTeacher = category === "teacher";
  const isOrg = category === "organization";

  const cbseId = boards.find((b) => b.name === "CBSE")?.id ?? "";
  const icseId = boards.find((b) => b.name === "ICSE")?.id ?? "";
  const otherBoards = boards.filter((b) => b.name !== "CBSE" && b.name !== "ICSE");

  function validate() {
    const e: Record<string, string> = {};
    if (!fullName.trim()) e.fullName = "Full name is required";
    if (isStudent) {
      if (!grade) e.grade = "Please select a grade";
      if (!board) e.board = "Please select a board";
    }
    if (isTeacher && !subject.trim()) e.subject = "Subject is required";
    if (isOrg) {
      if (!orgName.trim()) e.orgName = "Organization name is required";
      if (!orgType.trim()) e.orgType = "Organization type is required";
    }
    return e;
  }

  async function handleSubmit() {
    const e = validate();
    if (Object.keys(e).length) { setErrors(e); return; }
    setErrors({});
    setLoading(true);
    try {
      const payload = {
        category: category as "student" | "teacher" | "organization",
        fullName: fullName.trim(),
        ...(isStudent && { grade, board }),
        ...(isTeacher && { subject: subject.trim() }),
        ...(isOrg && { organizationName: orgName.trim(), organizationType: orgType.trim() }),
      };
      const { token, user: updatedUser } = await completeOnboarding(payload);
      setAuth(token, updatedUser);
      router.replace("/dashboard");
    } catch {
      setErrors({ form: "Something went wrong. Please try again." });
    } finally {
      setLoading(false);
    }
  }

  const canContinue = fullName.trim().length > 0;

  return (
    <OnboardingShell>
      <div className="flex flex-col items-center w-full px-4 py-4" style={{ maxWidth: "800px" }}>

        {/* Card */}
        <div
          className="bg-white w-full flex flex-col items-center"
          style={{
            borderRadius: "42px",
            border: "1px solid #8E8E93",
            padding: "clamp(24px, 4vw, 36px) clamp(20px, 4vw, 42px) clamp(24px, 4vw, 32px)",
            gap: "clamp(16px, 3vw, 20px)",
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

          {errors.form && (
            <p className="w-full text-sm text-red-500 bg-red-50 rounded-xl px-4 py-2 text-center" style={{ maxWidth: "600px" }}>
              {errors.form}
            </p>
          )}

          {/* Form fields */}
          <div className="w-full flex flex-col" style={{ maxWidth: "600px", gap: "clamp(10px, 2vw, 12px)" }}>

            <FloatingInput
              label="Full Name"
              type="text"
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
              error={errors.fullName}
              autoComplete="name"
            />

            {/* Student fields */}
            {isStudent && (
              <>
                <div className="w-full relative">
                  <select
                    value={grade}
                    onChange={(e) => setGrade(e.target.value)}
                    className={cn(
                      "w-full rounded-lg border bg-white text-sm outline-none focus:ring-0 transition-colors appearance-none",
                      grade ? "text-black" : "text-gray-400",
                      errors.grade ? "border-red-400 focus:border-red-500" : "border-gray-300 focus:border-blue-600"
                    )}
                    style={{ padding: "14px 24px", height: "52px" }}
                  >
                    <option value="" disabled>Select Grade</option>
                    {grades.map((g) => (
                      <option key={g.id} value={g.id}>{g.name}</option>
                    ))}
                  </select>
                  <div className="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-gray-500">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M6 9l6 6 6-6" />
                    </svg>
                  </div>
                  {errors.grade && <p className="mt-0.5 text-sm text-red-500">{errors.grade}</p>}
                </div>

                <div className="flex flex-col gap-2">
                  <p className="text-sm text-gray-600">Select Board</p>
                  <div className="flex gap-3">
                    {/* CBSE */}
                    <button
                      onClick={() => { setBoard(cbseId); setShowOtherDropdown(false); }}
                      disabled={!cbseId}
                      className="flex-1 text-sm font-medium transition-colors py-4"
                      style={{
                        minHeight: "52px",
                        borderRadius: "50px",
                        border: board === cbseId && !showOtherDropdown ? "1px solid #BFDBFE" : "1px solid #2563EB",
                        backgroundColor: board === cbseId && !showOtherDropdown ? "#EFF6FF" : "transparent",
                        color: "#2563EB",
                      }}
                    >
                      CBSE
                    </button>
                    {/* ICSE — disabled */}
                    <button
                      disabled
                      className="flex-1 text-sm font-medium py-2 cursor-not-allowed"
                      style={{
                        minHeight: "52px",
                        borderRadius: "50px",
                        border: "1px solid #E5E7EB",
                        backgroundColor: "#F9FAFB",
                        color: "#9CA3AF",
                      }}
                    >
                      ICSE
                    </button>
                    {/* Other — disabled */}
                    <button
                      disabled
                      className="flex-1 text-sm font-medium py-2 cursor-not-allowed"
                      style={{
                        minHeight: "52px",
                        borderRadius: "50px",
                        border: "1px solid #E5E7EB",
                        backgroundColor: "#F9FAFB",
                        color: "#9CA3AF",
                      }}
                    >
                      Other
                    </button>
                  </div>
                  {errors.board && <p className="mt-0.5 text-sm text-red-500">{errors.board}</p>}
                </div>

                {showOtherDropdown && (
                  <div className="w-full relative">
                    <select
                      value={board}
                      onChange={(e) => setBoard(e.target.value)}
                      className={cn(
                        "w-full rounded-lg border bg-white text-sm outline-none focus:ring-0 transition-colors appearance-none",
                        board ? "text-black" : "text-gray-400",
                        errors.board ? "border-red-400 focus:border-red-500" : "border-gray-300 focus:border-blue-600"
                      )}
                      style={{ padding: "14px 24px", height: "52px" }}
                    >
                      <option value="" disabled>Select Board</option>
                      {otherBoards.map((b) => (
                        <option key={b.id} value={b.id}>{b.name}</option>
                      ))}
                    </select>
                    <div className="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-gray-500">
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <path d="M6 9l6 6 6-6" />
                      </svg>
                    </div>
                  </div>
                )}
              </>
            )}

            {isTeacher && (
              <FloatingInput
                label="Subject you teach"
                type="text"
                value={subject}
                onChange={(e) => setSubject(e.target.value)}
                error={errors.subject}
              />
            )}

            {isOrg && (
              <>
                <FloatingInput
                  label="Organization Name"
                  type="text"
                  value={orgName}
                  onChange={(e) => setOrgName(e.target.value)}
                  error={errors.orgName}
                />
                <FloatingInput
                  label="Organization Type (e.g. School, Institute)"
                  type="text"
                  value={orgType}
                  onChange={(e) => setOrgType(e.target.value)}
                  error={errors.orgType}
                />
              </>
            )}

            <button
              onClick={handleSubmit}
              disabled={!canContinue || loading}
              className={`flex items-center justify-center w-full font-medium text-base transition-colors ${canContinue && !loading
                ? "bg-[#2563EB] hover:bg-[#1D4ED8] text-white"
                : "bg-white text-gray-400 border border-gray-300 cursor-not-allowed"
                }`}
              style={{ height: "clamp(48px, 10vw, 52px)", borderRadius: "24px" }}
            >
              {loading ? "Saving…" : "Continue"}
            </button>
          </div>
        </div>
      </div>
    </OnboardingShell>
  );
}
