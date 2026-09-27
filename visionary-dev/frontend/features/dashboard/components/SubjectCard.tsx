"use client";

import Image from "next/image";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/authStore";
import { startSubject } from "@/lib/auth";

interface SubjectCardProps {
  id: string;
  name: string;
  chapters: number;
  icon: string;
  link: string;
}

export default function SubjectCard({ id, name, chapters, icon }: SubjectCardProps) {
  const router = useRouter();
  const setAuth = useAuthStore((s) => s.setAuth);
  const [loading, setLoading] = useState(false);

  async function handleStart() {
    if (loading) return;
    setLoading(true);
    try {
      const { token, user } = await startSubject(id);
      setAuth(token, user);
      router.push(`/learn/${id}`);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }

  return (
    <div
      className="bg-white border border-gray-100 flex flex-col shrink-0"
      style={{
        width: "clamp(220px, 30vw, 333px)",
        height: "clamp(220px, 28vw, 319px)",
        borderRadius: 32,
        padding: "clamp(20px, 2.5vw, 32px)",
        gap: "clamp(16px, 2vw, 32px)",
        display: "flex",
        flexDirection: "column",
      }}
    >
      {/* Icon */}
      <div className="relative w-12 h-12 sm:w-14 sm:h-14 lg:w-16 lg:h-16">
        <Image src={icon} alt={name} fill className="object-contain" />
      </div>

      {/* Text */}
      <div className="flex flex-col gap-1">
        <p className="text-base sm:text-lg lg:text-xl font-bold text-gray-900">{name}</p>
        <p className="text-xs sm:text-sm text-gray-400">{chapters} Chapters</p>
      </div>

      {/* Start button */}
      <div className="mt-auto flex justify-end">
        <button
          onClick={handleStart}
          disabled={loading}
          className="px-6 sm:px-8 lg:px-10 py-2.5 sm:py-3 rounded-full bg-blue-50 text-blue-500 text-xs sm:text-sm font-medium hover:bg-blue-100 transition-colors disabled:opacity-50"
        >
          {loading ? "..." : "Start"}
        </button>
      </div>
    </div>
  );
}
