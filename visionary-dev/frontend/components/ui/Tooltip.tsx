"use client";

interface TooltipProps {
  content: string;
  children: React.ReactNode;
  position?: "top" | "bottom" | "left" | "right";
}

export default function Tooltip({ content, children, position = "top" }: TooltipProps) {
  const positionClasses = {
    top: "bottom-full left-1/2 -translate-x-1/2 mb-2",
    bottom: "top-full left-1/2 -translate-x-1/2 mt-2",
    left: "right-full top-1/2 -translate-y-1/2 mr-2",
    right: "left-full top-1/2 -translate-y-1/2 ml-2",
  };

  const arrowClasses = {
    top: "top-full left-1/2 -translate-x-1/2 border-t-gray-800 border-x-transparent border-b-transparent border-4",
    bottom: "bottom-full left-1/2 -translate-x-1/2 border-b-gray-800 border-x-transparent border-t-transparent border-4",
    left: "left-full top-1/2 -translate-y-1/2 border-l-gray-800 border-y-transparent border-r-transparent border-4",
    right: "right-full top-1/2 -translate-y-1/2 border-r-gray-800 border-y-transparent border-l-transparent border-4",
  };

  return (
    <div className="relative inline-flex group">
      {children}
      <div
        className={`absolute z-50 pointer-events-none whitespace-nowrap px-3 py-2 text-xs font-medium text-white bg-gray-800 rounded-lg shadow-sm
          opacity-0 group-hover:opacity-100 transition-opacity duration-200
          ${positionClasses[position]}`}
      >
        {content}
        <span className={`absolute border ${arrowClasses[position]}`} />
      </div>
    </div>
  );
}
