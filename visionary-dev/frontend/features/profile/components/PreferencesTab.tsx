"use client";

import { useState } from "react";
import { cn } from "@/lib/utils";
import { SectionCard, Toggle } from "./shared";

/* ─── Appearance mode ─── */
type ColorMode = "Dark" | "Light" | "Auto";
const COLOR_MODES: ColorMode[] = ["Dark", "Light", "Auto"];

/* ─── Notification toggle row ─── */
function NotifRow({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (v: boolean) => void;
}) {
  return (
    <div className="flex items-center justify-between py-4 last:pb-0 first:pt-0">
      <span className="text-sm text-gray-700">{label}</span>
      <Toggle checked={checked} onChange={onChange} />
    </div>
  );
}

/* ─── Notification section card ─── */
function NotifSection({
  title,
  rows,
  values,
  onChange,
}: {
  title: string;
  rows: string[];
  values: boolean[];
  onChange: (index: number, v: boolean) => void;
}) {
  return (
    <div className="bg-white border border-gray-200 rounded-2xl px-6 py-6">
      <h3 className="text-sm font-semibold text-gray-900 mb-2">{title}</h3>
      <div className="divide-y divide-gray-100">
        {rows.map((label, i) => (
          <NotifRow
            key={label}
            label={label}
            checked={values[i]}
            onChange={(v) => onChange(i, v)}
          />
        ))}
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════════════════════════
   PREFERENCES TAB
═══════════════════════════════════════════════════════════ */
export default function PreferencesTab() {
  const [colorMode, setColorMode] = useState<ColorMode>("Light");

  /* Email notification toggles */
  const [streaks, setStreaks] = useState([true]);
  const [reminders, setReminders] = useState([true]);
  const [news, setNews] = useState([true, true]);

  function updateAt<T>(arr: T[], i: number, v: T): T[] {
    return arr.map((item, idx) => (idx === i ? v : item));
  }

  return (
    <div className="flex flex-col gap-5">
      {/* ── Appearance ── */}
      <SectionCard>
        <h2 className="text-lg font-semibold text-gray-900 mb-1">Appearance</h2>
        <p className="text-sm text-gray-500 mb-5">Choose your preferred color mode</p>

        {/* Segmented control */}
        <div className="inline-flex items-center bg-gray-100 rounded-full p-1 gap-1">
          {COLOR_MODES.map((mode) => (
            <button
              key={mode}
              type="button"
              onClick={() => setColorMode(mode)}
              className={cn(
                "px-5 py-1.5 rounded-full text-sm font-medium transition-all duration-150",
                colorMode === mode
                  ? "bg-white text-blue-600 shadow-sm"
                  : "text-gray-500 hover:text-gray-700"
              )}
            >
              {mode}
            </button>
          ))}
        </div>
      </SectionCard>

      {/* ── Email notifications heading ── */}
      <h2 className="text-lg font-semibold text-gray-900 -mb-1 px-1">
        Email notifications
      </h2>

      {/* ── Streaks ── */}
      <NotifSection
        title="Streaks"
        rows={["Streaks daily reminder"]}
        values={streaks}
        onChange={(i, v) => setStreaks(updateAt(streaks, i, v))}
      />

      {/* ── Learning reminders ── */}
      <NotifSection
        title="Learning reminders"
        rows={["Daily Practice"]}
        values={reminders}
        onChange={(i, v) => setReminders(updateAt(reminders, i, v))}
      />

      {/* ── News & Announcements ── */}
      <NotifSection
        title="News & Announcements"
        rows={["Product launches and updates", "Offers & Promotions"]}
        values={news}
        onChange={(i, v) => setNews(updateAt(news, i, v))}
      />
    </div>
  );
}
