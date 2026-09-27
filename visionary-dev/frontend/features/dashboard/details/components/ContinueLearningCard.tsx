import Image from "next/image";
import Link from "next/link";
import { ArrowRight } from "lucide-react";
import ProgressBar from "@/components/ui/ProgressBar";

interface ContinueLearningCardProps {
  subject: string;
  chapter: string;
  title: string;
  completedPercent: number;
  sectionsCompleted: number;
  totalSections: number;
  iconPath: string;
  href: string;
}

export default function ContinueLearningCard({
  subject,
  chapter,
  title,
  completedPercent,
  sectionsCompleted,
  totalSections,
  iconPath,
  href,
}: ContinueLearningCardProps) {
  return (
    <div className="bg-white border border-gray-100 rounded-2xl p-5 flex flex-col gap-4 flex-1 min-w-0">
      {/* Subject badge + Chapter */}
      <div className="flex items-center gap-2 text-sm text-gray-400">
        <span className="flex items-center gap-1.5 border border-gray-200 rounded-full px-2.5 py-1">
          <span className="relative w-4 h-4 shrink-0">
            <Image src={iconPath} alt={subject} fill className="object-contain" />
          </span>
          <span className="text-xs text-gray-500">{subject}</span>
        </span>
        <span className="text-xs text-gray-400">{chapter}</span>
      </div>

      {/* Title */}
      <h3 className="text-lg font-bold text-gray-900 leading-snug line-clamp-2">{title}</h3>

      {/* Progress info */}
      <div className="flex items-center gap-4 text-xs text-gray-400">
        <span>{completedPercent}% Completed</span>
        <span>{sectionsCompleted}/{totalSections} Sections</span>
      </div>

      {/* Progress bar */}
      <ProgressBar percentage={completedPercent} />

      {/* Arrow button */}
      <div className="flex justify-end mt-auto">
        <Link href={href} className="w-10 h-10 rounded-full bg-blue-50 flex items-center justify-center text-blue-500 hover:bg-blue-100 transition-colors">
          <ArrowRight size={16} />
        </Link>
      </div>
    </div>
  );
}
