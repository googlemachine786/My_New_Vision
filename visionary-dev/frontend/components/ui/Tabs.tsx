"use client";

import { cn } from "@/lib/utils";

interface TabsProps {
  tabs: string[];
  activeTab: string;
  onChange: (tab: string) => void;
  /** "underline" = existing bottom-border style (default)
   *  "pill"      = rounded pill buttons for sidebar-style tabs */
  variant?: "underline" | "pill";
}

export default function Tabs({ tabs, activeTab, onChange, variant = "underline" }: TabsProps) {
  if (variant === "pill") {
    return (
      <div className="flex flex-col gap-0" style={{ gap: 28 }}>
        {tabs.map((tab) => (
          <button
            key={tab}
            onClick={() => onChange(tab)}
            className={cn(
              "text-sm font-medium transition-colors text-left",
              activeTab === tab
                ? "bg-blue-100 text-blue-600"
                : "text-gray-600 hover:bg-gray-50"
            )}
            style={{
              width: 190,
              height: 56,
              borderRadius: 50,
              paddingTop: 21,
              paddingBottom: 21,
              paddingLeft: 24,
              paddingRight: 24,
            }}
          >
            {tab}
          </button>
        ))}
      </div>
    );
  }

  // default: underline variant — unchanged
  return (
    <div className="flex border-b border-gray-200 w-fit">
      {tabs.map((tab) => (
        <button
          key={tab}
          onClick={() => onChange(tab)}
          className={`relative w-32 pb-3 text-sm font-medium transition-colors text-center ${
            activeTab === tab ? "text-gray-900" : "text-gray-400 hover:text-gray-600"
          }`}
        >
          {tab}
          {activeTab === tab && (
            <span className="absolute bottom-0 left-0 w-full h-0.5 bg-gray-900 rounded-full" />
          )}
        </button>
      ))}
    </div>
  );
}
