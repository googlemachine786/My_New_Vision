"use client";

import React, { useRef, useState, useEffect } from "react";

interface ScrollAreaProps {
  children: React.ReactNode;
  className?: string;
}

export const ScrollArea: React.FC<ScrollAreaProps> = ({ children, className = "" }) => {
  const viewportRef = useRef<HTMLDivElement>(null);
  const [scrollTop, setScrollTop] = useState(0);
  const [scrollHeight, setScrollHeight] = useState(1);
  const [clientHeight, setClientHeight] = useState(1);

  useEffect(() => {
    const el = viewportRef.current;
    if (!el) return;
    const handleScroll = () => {
      setScrollTop(el.scrollTop);
      setScrollHeight(el.scrollHeight);
      setClientHeight(el.clientHeight);
    };
    handleScroll();
    el.addEventListener("scroll", handleScroll);
    window.addEventListener("resize", handleScroll);
    return () => {
      el.removeEventListener("scroll", handleScroll);
      window.removeEventListener("resize", handleScroll);
    };
  }, []);

  const thumbHeight = Math.max((clientHeight / scrollHeight) * clientHeight, 20);
  const thumbTop =
    (scrollTop / (scrollHeight - clientHeight)) * (clientHeight - thumbHeight);

  return (
    <div className={`relative overflow-hidden ${className}`}>
      <div
        ref={viewportRef}
        className="h-full w-full overflow-auto scrollbar-none"
        style={{ scrollbarWidth: "none", msOverflowStyle: "none" } as React.CSSProperties}
      >
        {children}
      </div>
      <div className="absolute right-1 top-1 bottom-1 w-2 rounded bg-transparent">
        <div
          className="absolute w-full rounded bg-gray-400 hover:bg-gray-500 transition-colors"
          style={{ height: `${thumbHeight}px`, transform: `translateY(${thumbTop}px)` }}
        />
      </div>
    </div>
  );
};
