"use client";

import { useEffect, useRef, useState, ReactNode } from "react";
import { cn } from "@/lib/utils";

type PopoverAlign = "left" | "right" | "center";
type PopoverSide = "top" | "bottom";

interface PopoverProps {
  /** The trigger element — receives onClick and ref automatically */
  trigger: ReactNode;
  /** Popover panel content */
  children: ReactNode;
  /** Horizontal alignment relative to the trigger */
  align?: PopoverAlign;
  /** Which side of the trigger to open on */
  side?: PopoverSide;
  /** Extra classes on the panel */
  className?: string;
  /** Controlled open state (optional — uncontrolled by default) */
  open?: boolean;
  /** Called when the popover requests to close */
  onOpenChange?: (open: boolean) => void;
}

const alignClasses: Record<PopoverAlign, string> = {
  left: "left-0",
  right: "right-0",
  center: "left-1/2 -translate-x-1/2",
};

const sideClasses: Record<PopoverSide, string> = {
  bottom: "top-[calc(100%+8px)]",
  top: "bottom-[calc(100%+8px)]",
};

export default function Popover({
  trigger,
  children,
  align = "right",
  side = "bottom",
  className,
  open: controlledOpen,
  onOpenChange,
}: PopoverProps) {
  const isControlled = controlledOpen !== undefined;
  const [internalOpen, setInternalOpen] = useState(false);
  const open = isControlled ? controlledOpen : internalOpen;

  const containerRef = useRef<HTMLDivElement>(null);

  function toggle() {
    const next = !open;
    if (!isControlled) setInternalOpen(next);
    onOpenChange?.(next);
  }

  function close() {
    if (!isControlled) setInternalOpen(false);
    onOpenChange?.(false);
  }

  // Close on outside click
  useEffect(() => {
    if (!open) return;
    function onMouseDown(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        close();
      }
    }
    document.addEventListener("mousedown", onMouseDown);
    return () => document.removeEventListener("mousedown", onMouseDown);
  }, [open]);

  // Close on Escape
  useEffect(() => {
    if (!open) return;
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") close();
    }
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [open]);

  return (
    <div ref={containerRef} className="relative inline-flex">
      {/* Trigger — clone with onClick injected */}
      <div onClick={toggle} className="inline-flex cursor-pointer">
        {trigger}
      </div>

      {/* Panel */}
      {open && (
        <div
          role="dialog"
          aria-modal="false"
          className={cn(
            "absolute z-50 w-64 text-sm",
            "bg-white border border-gray-200 rounded-2xl shadow-lg",
            "transition-opacity duration-150",
            alignClasses[align],
            sideClasses[side],
            className
          )}
        >
          {children}
        </div>
      )}
    </div>
  );
}

/** Convenience sub-components for structured popovers */

interface PopoverHeaderProps {
  title: ReactNode;
  className?: string;
}

export function PopoverHeader({ title, className }: PopoverHeaderProps) {
  return (
    <div className={cn("px-4 py-2.5 border-b border-gray-100 rounded-t-2xl bg-gray-50", className)}>
      {typeof title === "string" ? (
        <h3 className="font-semibold text-sm text-gray-900">{title}</h3>
      ) : (
        title
      )}
    </div>
  );
}

interface PopoverBodyProps {
  children: ReactNode;
  className?: string;
}

export function PopoverBody({ children, className }: PopoverBodyProps) {
  return (
    <div className={cn("px-4 py-2.5", className)}>
      {children}
    </div>
  );
}

interface PopoverItemProps {
  icon?: ReactNode;
  children: ReactNode;
  onClick?: () => void;
  variant?: "default" | "danger";
  className?: string;
}

export function PopoverItem({
  icon,
  children,
  onClick,
  variant = "default",
  className,
}: PopoverItemProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "w-full flex items-center gap-3 px-4 py-2.5 text-sm transition-colors text-left",
        variant === "default" && "text-gray-700 hover:bg-gray-50",
        variant === "danger" && "text-red-500 hover:bg-red-50",
        className
      )}
    >
      {icon && <span className="shrink-0 flex items-center">{icon}</span>}
      {children}
    </button>
  );
}
