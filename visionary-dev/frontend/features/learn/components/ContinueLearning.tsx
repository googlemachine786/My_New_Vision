"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ArrowRight } from "lucide-react";
import ProgressBar from "@/components/ui/ProgressBar";
import Tooltip from "@/components/ui/Tooltip";
import Badge from "@/components/ui/Badge";
import { getRecentChapters } from "@/lib/auth";
import type { RecentChapter } from "@/types";

export default function ContinueLearning() {
  const [items, setItems] = useState<RecentChapter[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getRecentChapters()
      .then(setItems)
      .catch(() => setItems([]))
      .finally(() => setLoading(false));
  }, []);

  // Hide entirely if no chapter_track data and not loading
  if (!loading && items.length === 0) return null;

  return (
    <div className="w-full">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-3xl font-semibold text-gray-900">Continue Learning</h2>
        <Link
          href="/learn/continue-learning"
          className="text-blue-500 text-sm font-medium hover:text-blue-600 transition-colors"
        >
          View All
        </Link>
      </div>

      {loading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 w-full">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="bg-gray-100 animate-pulse rounded-2xl h-[220px]" />
          ))}
        </div>
      ) : items.length === 0 ? (
        <p className="text-sm text-gray-400 py-4">No chapters found.</p>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 w-full">
          {items.map((item) => (
            <div
              key={item.chapterId}
              className="bg-white border border-gray-100 rounded-2xl p-5 flex flex-col gap-4 min-h-[220px]"
            >
              <div className="flex items-center justify-between gap-2">
                <Badge label={`${item.subjectName} – ${item.bookName}`} imageSrc={item.subjectIcon ?? "/assets/Science-icon.png"} />
                <span className="text-xs text-gray-400">Chapter {item.sortOrder}</span>
              </div>

              <Tooltip content={item.chapterName} position="top">
                <h3 className="text-lg font-bold text-gray-900 truncate w-full cursor-default">
                  {item.chapterName}
                </h3>
              </Tooltip>

              <div className="flex items-center gap-4 text-xs text-gray-400">
                <span>{item.progress}% Completed</span>
              </div>

              <ProgressBar percentage={item.progress} />

              <div className="flex justify-end mt-auto">
                <Link
                  href={`/learn/${item.subjectId}/${item.chapterId}`}
                  className="w-10 h-10 rounded-full bg-blue-50 flex items-center justify-center text-blue-500 hover:bg-blue-100 transition-colors"
                >
                  <ArrowRight size={16} />
                </Link>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
