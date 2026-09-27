"use client";

import { useEffect, useRef, useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { useAuthStore } from "@/store/authStore";
import SubjectCard from "@/features/dashboard/components/SubjectCard";
import ProgressCards from "@/features/dashboard/components/ProgressCards";
import DashboardDetails from "@/features/dashboard/details";
import { getMySubjects } from "@/lib/auth";
import type { SubjectMaster } from "@/types";

export default function DashboardPage() {
  const user = useAuthStore((s) => s.user);
  const scrollRef = useRef<HTMLDivElement>(null);
  const [subjects, setSubjects] = useState<SubjectMaster[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!user?.isStarted) {
      getMySubjects()
        .then(setSubjects)
        .catch(() => setSubjects([]))
        .finally(() => setLoading(false));
    }
  }, [user?.isStarted]);

  function scroll(dir: "left" | "right") {
    if (!scrollRef.current) return;
    scrollRef.current.scrollBy({ left: dir === "left" ? -300 : 300, behavior: "smooth" });
  }

  // Once isStarted = true, show the detail view directly at /dashboard
  if (user?.isStarted) {
    return <DashboardDetails />;
  }

  return (
    <div className="w-full min-w-0">
      {/* Greeting */}
      <div className="mb-4 sm:mb-6">
        <h1 className="text-xl sm:text-2xl font-bold text-gray-900">
          Welcome to Visionary{user?.fullName ? `, ${user.fullName}` : ""}
        </h1>
        <p className="text-xs sm:text-sm text-gray-400 mt-1">
          Let&apos;s get you closer to mastery with Visionary
        </p>
      </div>

      {/* Section header */}
      <div className="mb-3 sm:mb-4 flex items-center justify-between">
        <h2 className="text-base sm:text-lg font-bold text-gray-900">Start Learning</h2>
        <div className="flex items-center gap-2">
          <button
            onClick={() => scroll("left")}
            className="w-8 h-8 sm:w-9 sm:h-9 rounded-full border border-gray-200 bg-white flex items-center justify-center text-gray-500 hover:bg-gray-50 transition-colors"
          >
            <ChevronLeft size={14} />
          </button>
          <button
            onClick={() => scroll("right")}
            className="w-8 h-8 sm:w-9 sm:h-9 rounded-full border border-gray-200 bg-white flex items-center justify-center text-gray-500 hover:bg-gray-50 transition-colors"
          >
            <ChevronRight size={14} />
          </button>
        </div>
      </div>

      {/* Subject cards */}
      <div
        ref={scrollRef}
        className="flex gap-3 sm:gap-4 overflow-x-auto pb-2"
        style={{ scrollbarWidth: "none", msOverflowStyle: "none" }}
      >
        {loading ? (
          Array.from({ length: 4 }).map((_, i) => (
            <div
              key={i}
              className="shrink-0 bg-gray-100 animate-pulse"
              style={{
                width: "clamp(220px, 30vw, 333px)",
                height: "clamp(220px, 28vw, 319px)",
                borderRadius: 32,
              }}
            />
          ))
        ) : subjects.length === 0 ? (
          <p className="text-sm text-gray-400 py-8">No subjects found for your grade and board.</p>
        ) : (
          subjects.map((subject) => (
            <SubjectCard
              key={subject.id}
              id={subject.id}
              name={subject.name}
              chapters={10}
              icon={subject.icon ?? "/assets/Science-icon.png"}
              link={`/dashboard/${subject.id}`}
            />
          ))
        )}
      </div>

      <ProgressCards />
    </div>
  );
}
