"use client";

import BarChart from "@/components/ui/BarChart";
import DonutChart from "@/components/ui/DonutChart";

const curriculumSeries = [
  {
    name: "Progress",
    color: "#2563EB",
    data: [
      { x: "Physics", y: 0 },
      { x: "Maths", y: 0 },
      { x: "Chemistry", y: 0 },
      { x: "Biology", y: 0 },
      { x: "English", y: 0 },
      { x: "Hindi", y: 0 },
    ],
  },
];

export default function ProgressCards() {
  return (
    <div className="flex flex-col lg:flex-row gap-4 mt-6 w-full min-w-0">
      {/* Curriculum Progress — grows to fill remaining space */}
      <div
        className="bg-white border border-gray-100 flex flex-col min-w-0 flex-1"
        style={{ borderRadius: 32, padding: 32, gap: 16, minHeight: 320 }}
      >
        <div className="flex items-center justify-between">
          <h2 className="text-base sm:text-lg font-bold text-[#2D2D7B]">Curriculum Progress</h2>
          <button className="flex items-center gap-1.5 border border-gray-200 rounded-full px-3 sm:px-4 py-1.5 text-xs sm:text-sm text-gray-500 hover:bg-gray-50 transition-colors">
            All Time
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="6 9 12 15 18 9" />
            </svg>
          </button>
        </div>
        <div className="flex-1 min-h-0">
          <BarChart series={curriculumSeries} height={240} />
        </div>
      </div>

      {/* Overall Progress — fixed width on desktop, full width on mobile */}
      <div
        className="bg-white border border-gray-100 flex flex-col w-full lg:w-[378px] lg:shrink-0"
        style={{ borderRadius: 32, padding: 32, gap: 24, minHeight: 320 }}
      >
        <h2 className="text-base sm:text-lg font-bold text-[#2D2D7B]">Overall Progress</h2>
        <div className="flex-1 flex items-center justify-center">
          <DonutChart
            series={[0, 0, 0, 0]}
            labels={["Physics", "Maths", "Chemistry", "Biology"]}
            colors={["#2563EB", "#93C5FD", "#1D4ED8", "#BFDBFE"]}
            centerLabel="Overall Progress"
            height={240}
          />
        </div>
      </div>
    </div>
  );
}
