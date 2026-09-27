import Image from "next/image";
import Link from "next/link";
import ProgressBar from "@/components/ui/ProgressBar";
import { ScrollArea } from "@/components/ui/ScrollArea";
import type { BookWithChapters } from "@/types";

interface Props {
  books: Array<BookWithChapters & { subjectName?: string; subjectIcon?: string | null }>;
  loading: boolean;
}

const statusLabel: Record<string, string> = {
  NOT_STARTED: "Not Started",
  IN_PROGRESS: "In Progress",
  COMPLETED: "Completed",
};

export default function MyLessons({ books, loading }: Props) {
  // Flatten chapters with book name
  const lessons = books.flatMap((book) =>
    book.chapters.map((ch) => ({
      id: ch.id,
      subject: book.subjectName ? `${book.subjectName} – ${book.name}` : book.name,
      icon: book.subjectIcon ?? "/assets/Science-icon.png",
      chapterName: ch.name,
      progress: ch.progress,
      status: ch.status,
    }))
  );

  return (
    <div className="bg-white border border-gray-100 rounded-3xl p-5 sm:p-8 mt-6 w-full overflow-x-auto">
      <h2 className="text-xl sm:text-2xl font-bold text-[#2D2D7B] mb-6">My Lessons</h2>

      {loading ? (
        <div className="flex flex-col gap-3">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="h-16 bg-gray-100 animate-pulse rounded-xl" />
          ))}
        </div>
      ) : lessons.length === 0 ? (
        <p className="text-sm text-gray-400">No chapters available.</p>
      ) : (
        <>
          {/* Desktop table */}
          <div className="hidden md:block min-w-0">
            <div className="grid grid-cols-[1fr_1fr_1fr_120px] gap-4 px-2 mb-2">
              <span className="text-sm text-gray-400">Subject/Book</span>
              <span className="text-sm text-gray-400">Progress</span>
              <span className="text-sm text-gray-400 text-center">Status</span>
              <span />
            </div>
            <ScrollArea className="h-[420px]">
              <div className="flex flex-col divide-y divide-gray-100 pr-3">
                {lessons.map((lesson) => (
                  <div key={lesson.id} className="grid grid-cols-[1fr_1fr_1fr_120px] gap-4 items-center py-4 px-2">
                    <div className="flex flex-col gap-1">
                      <div className="flex items-center gap-1.5">
                        <span className="relative w-4 h-4 shrink-0">
                          <Image src={lesson.icon} alt={lesson.subject} fill className="object-contain" />
                        </span>
                        <span className="text-xs text-gray-400">{lesson.subject}</span>
                      </div>
                      <span className="text-base font-semibold text-gray-900">{lesson.chapterName}</span>
                    </div>
                    <div className="flex flex-col gap-1.5">
                      <span className="text-sm text-gray-700">{lesson.progress}%</span>
                      <ProgressBar percentage={lesson.progress} />
                    </div>
                    <div className="flex justify-center">
                      <span className={`inline-flex items-center px-4 py-1.5 rounded-full border text-sm ${
                        lesson.status === "COMPLETED"
                          ? "border-green-200 text-green-600 bg-green-50"
                          : "border-gray-200 text-gray-500 bg-white"
                      }`}>
                        {statusLabel[lesson.status] ?? lesson.status}
                      </span>
                    </div>
                    <Link
                      href={`/lessons/${lesson.id}`}
                      className="px-6 py-2.5 rounded-full bg-blue-50 text-blue-500 text-sm font-semibold hover:bg-blue-100 transition-colors whitespace-nowrap text-center"
                    >
                      Continue
                    </Link>
                  </div>
                ))}
              </div>
            </ScrollArea>
          </div>

          {/* Mobile cards */}
          <ScrollArea className="h-[420px] md:hidden">
            <div className="flex flex-col gap-3 pr-3">
              {lessons.map((lesson) => (
                <div key={lesson.id} className="border border-gray-100 rounded-2xl p-4 flex flex-col gap-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-1.5">
                      <span className="relative w-4 h-4 shrink-0">
                        <Image src={lesson.icon} alt={lesson.subject} fill className="object-contain" />
                      </span>
                      <span className="text-xs text-gray-400">{lesson.subject}</span>
                    </div>
                    <span className={`inline-flex items-center px-3 py-1 rounded-full border text-xs ${
                      lesson.status === "COMPLETED"
                        ? "border-green-200 text-green-600 bg-green-50"
                        : "border-gray-200 text-gray-500 bg-white"
                    }`}>
                      {statusLabel[lesson.status] ?? lesson.status}
                    </span>
                  </div>
                  <span className="text-sm font-semibold text-gray-900">{lesson.chapterName}</span>
                  <div className="flex flex-col gap-1">
                    <span className="text-xs text-gray-500">{lesson.progress}%</span>
                    <ProgressBar percentage={lesson.progress} />
                  </div>
                  <Link
                    href={`/lessons/${lesson.id}`}
                    className="w-full text-center py-2 rounded-full bg-blue-50 text-blue-500 text-sm font-semibold hover:bg-blue-100 transition-colors"
                  >
                    Continue
                  </Link>
                </div>
              ))}
            </div>
          </ScrollArea>
        </>
      )}
    </div>
  );
}
