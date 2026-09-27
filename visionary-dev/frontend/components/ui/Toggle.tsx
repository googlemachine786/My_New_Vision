"use client";

import { useId } from "react";
import { cn } from "@/lib/utils";

interface ToggleProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  label?: string;
  className?: string;
}

export default function Toggle({ checked, onChange, label, className }: ToggleProps) {
  const id = useId();

  return (
    <label
      htmlFor={id}
      className={cn("inline-flex items-center cursor-pointer gap-3 select-none", className)}
    >
      {/* Hidden checkbox — single source of truth for state */}
      <input
        id={id}
        type="checkbox"
        className="sr-only"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
      />

      {/* Track */}
      <div
        className={cn(
          "relative w-11 h-6 rounded-full shrink-0 transition-colors duration-200",
          "bg-white",
          checked ? "border-[3px] border-blue-600" : "border-[3px] border-gray-300"
        )}
      >
        {/* Thumb */}
        <span
          className={cn(
            "absolute top-1/2 -translate-y-1/2 w-3.5 h-3.5 rounded-full transition-all duration-200",
            checked ? "bg-blue-600 left-[calc(100%-17.5px)]" : "bg-gray-300 left-[3.5px]"
          )}
        />
      </div>

      {label && (
        <span className="text-sm font-medium text-gray-700">{label}</span>
      )}
    </label>
  );
}
