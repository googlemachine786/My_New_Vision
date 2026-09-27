import Image from "next/image";
import { ArrowUpRight } from "lucide-react";

interface WeakTopic {
  id: string;
  subject: string;
  chapterName: string;
  score: number;
  iconPath: string;
}

const weakTopics: WeakTopic[] = [
  { id: "physics-ch1", subject: "Physics Part-I", chapterName: "Chapter Name Example", score: 45, iconPath: "/assets/physics.png" },
  { id: "physics-ch2", subject: "Physics Part-I", chapterName: "Chapter Name Example", score: 45, iconPath: "/assets/physics.png" },
];

export default function WeakestTopics() {
  return (
    <div className="mt-6 w-full">
      <h2 className="text-2xl font-bold text-gray-900 mb-4">Weakest Topics</h2>

      <div className="flex flex-col sm:flex-row gap-4">
        {weakTopics.map((topic) => (
          <div
            key={topic.id}
            className="bg-white border border-gray-200 rounded-2xl p-4 flex flex-col gap-2 w-full sm:w-64"
          >
            {/* Subject + arrow */}
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5">
                <span className="relative w-4 h-4 shrink-0">
                  <Image src={topic.iconPath} alt={topic.subject} fill className="object-contain" />
                </span>
                <span className="text-xs text-gray-400">{topic.subject}</span>
              </div>
              <button className="text-gray-400 hover:text-gray-600 transition-colors">
                <ArrowUpRight size={16} />
              </button>
            </div>

            {/* Chapter name */}
            <p className="text-base font-semibold text-gray-900">{topic.chapterName}</p>

            {/* Score */}
            <p className="text-sm font-semibold text-red-500">Score {topic.score}%</p>
          </div>
        ))}
      </div>
    </div>
  );
}
