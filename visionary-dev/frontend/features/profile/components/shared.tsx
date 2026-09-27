"use client";

import { cn } from "@/lib/utils";

/* ─── Section card wrapper ─── */
export function SectionCard({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("bg-white border border-gray-200 rounded-2xl p-6 sm:p-8", className)}>
      {children}
    </div>
  );
}

/* ─── Pill button ─── */
export function PillButton({
  children,
  onClick,
  variant = "default",
  disabled = false,
  type = "button",
}: {
  children: React.ReactNode;
  onClick?: () => void;
  variant?: "default" | "danger";
  disabled?: boolean;
  type?: "button" | "submit";
}) {
  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      className={cn(
        "px-6 py-2.5 rounded-full text-sm font-medium transition-colors",
        variant === "danger"
          ? "bg-red-500 hover:bg-red-600 text-white"
          : "bg-gray-200 hover:bg-gray-300 text-gray-600",
        disabled ? "opacity-60 cursor-not-allowed" : "cursor-pointer"
      )}
    >
      {children}
    </button>
  );
}

/* ─── Re-export Toggle from ui components ─── */
export { default as Toggle } from "@/components/ui/Toggle";
