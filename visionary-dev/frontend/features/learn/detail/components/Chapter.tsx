"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { ArrowRight, RotateCcw } from "lucide-react";
import ProgressBar from "@/components/ui/ProgressBar";
import Badge from "@/components/ui/Badge";
import Tooltip from "@/components/ui/Tooltip";
import { startChapter } from "@/lib/auth";
import { useAuthStore } from "@/store/authStore";

export interface ChapterProps {
  id: string;
  number: number;
  title: string;
  progress: number;
  status: "Not Started" | "In Progress" | "Completed";
  sections: number;
  sectionsCompleted: number;
  link: string;
}

export default function Chapter({ id, number, title, progress, status, sections, sectionsCompleted, link }: ChapterProps) {
  const router = useRouter();
  const setAuth = useAuthStore((s) => s.setAuth);
  const [loading, setLoading] = useState(false);

  async function handleClick() {
    if (loading) return;
    setLoading(true);
    try {
      const { token, user } = await startChapter(id);
      setAuth(token, user);
    } catch {
      // ignore — still navigate
    } finally {
      setLoading(false);
    }
    router.push(link);
  }

  return (
    <div className="bg-white border border-gray-100 rounded-2xl p-5 flex flex-col gap-4">
      {/* Chapter badge */}
      <Badge label={`Chapter ${number}`} variant="gray" className="self-start" />

      {/* Title */}
      <Tooltip content={title} position="top">
        <h3 className="text-base font-bold text-gray-900 truncate w-full cursor-default">
          {title}
        </h3>
      </Tooltip>

      {/* Progress info */}
      <div className="flex items-center gap-4 text-xs text-gray-400">
        {status === "Completed" ? (
          <span className="text-green-500 font-semibold">100% Complete</span>
        ) : (
          <span>{status === "Not Started" ? "Not Started" : `${progress}% Completed`}</span>
        )}
        <span>
          {status === "In Progress" ? `${sectionsCompleted}/${sections}` : sections} Sections
        </span>
      </div>

      {/* Progress bar */}
      <ProgressBar percentage={progress} />

      {/* Action button */}
      <div className="flex justify-end mt-auto">
        <button
          onClick={handleClick}
          disabled={loading}
          className="w-10 h-10 rounded-full bg-blue-50 flex items-center justify-center text-blue-500 hover:bg-blue-100 transition-colors disabled:opacity-50"
        >
          {status === "Completed" ? <RotateCcw size={15} /> : <ArrowRight size={16} />}
        </button>
      </div>
    </div>
  );
}
