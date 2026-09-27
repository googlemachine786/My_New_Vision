"use client";

import { useEffect, useState } from "react";
import { useAuthStore } from "@/store/authStore";
import StatCard from "./components/StatCard";
import ContinueLearningCard from "./components/ContinueLearningCard";
import ProgressCards from "@/features/dashboard/components/ProgressCards";
import MyLessons from "./components/MyLessons";
import WeakestTopics from "./components/WeakestTopics";
import { getAllStartedChapters, getChapterStats, getRecentChapters } from "@/lib/auth";
import type { BookWithChapters, RecentChapter, StartedSubjectTrack } from "@/types";

interface Props {
  subjectId?: string;
}

export default function DashboardDetails({ subjectId }: Props) {
  const user = useAuthStore((s) => s.user);
  const [tracks, setTracks] = useState<StartedSubjectTrack[]>([]);
  const [recentChapters, setRecentChapters] = useState<RecentChapter[]>([]);
  const [chapterStats, setChapterStats] = useState({ total: 0, inProgress: 0, completed: 0 });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      getAllStartedChapters().catch(() => [] as StartedSubjectTrack[]),
      getRecentChapters().catch(() => [] as RecentChapter[]),
      getChapterStats().catch(() => ({ total: 0, inProgress: 0, completed: 0 })),
    ]).then(([t, r, s]) => {
      setTracks(t);
      setRecentChapters(r);
      setChapterStats(s);
    }).finally(() => setLoading(false));
  }, []);

  const displayTracks = subjectId
    ? tracks.filter((t) => t.subjectId === subjectId)
    : tracks;

  const stats = [
    { imagePath: "/assets/total-chapters.png", label: "Total Chapters", number: chapterStats.total },
    { imagePath: "/assets/chapter-in-progress.png", label: "Chapters in Progress", number: chapterStats.inProgress },
    { imagePath: "/assets/chapter-completed.png", label: "Chapters Completed", number: chapterStats.completed },
  ];

  // My Lessons = ALL chapters from ALL started subjects
  const allBooks: Array<BookWithChapters & { subjectName: string; subjectIcon: string | null }> = displayTracks.flatMap((t) =>
    t.books.map((b) => ({ ...b, subjectName: t.subjectName, subjectIcon: t.icon ?? null }))
  );

  return (
    <div className="w-full min-w-0">
      {/* Greeting */}
      <div className="mb-4 sm:mb-6">
        <h1 className="text-xl sm:text-2xl font-bold text-gray-900">
          Welcome back{user?.fullName ? `, ${user.fullName}` : ""}
        </h1>
        <p className="text-xs sm:text-sm text-gray-400 mt-1">
          Let&apos;s get you closer to mastery with Visionary
        </p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 mb-6 sm:mb-8">
        {stats.map((stat) => (
          <StatCard key={stat.label} {...stat} />
        ))}
      </div>

      {/* Continue Learning */}
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg sm:text-xl font-bold text-gray-900">Continue Learning</h2>
        <a href="/learn/continue-learning" className="px-4 sm:px-5 py-2 rounded-full bg-blue-50 text-blue-500 text-xs sm:text-sm font-medium hover:bg-blue-100 transition-colors">
          View All
        </a>
      </div>

      {loading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="bg-gray-100 animate-pulse rounded-2xl h-48" />
          ))}
        </div>
      ) : recentChapters.length === 0 ? (
        <p className="text-sm text-gray-400 py-4">No chapters started yet.</p>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {recentChapters.map((item) => (
            <ContinueLearningCard
              key={item.chapterId}
              subject={`${item.subjectName} – ${item.bookName}`}
              chapter={`Chapter ${item.sortOrder}`}
              title={item.chapterName}
              completedPercent={item.progress}
              sectionsCompleted={0}
              totalSections={0}
              iconPath={item.subjectIcon ?? "/assets/Science-icon.png"}
              href={`/learn/${item.subjectId}/${item.chapterId}`}
            />
          ))}
        </div>
      )}

      <ProgressCards />

      {/* <MyLessons books={allBooks} loading={loading} /> */}

      {/* <WeakestTopics /> */}
    </div>
  );
}
