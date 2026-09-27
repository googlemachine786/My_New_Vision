"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { Home, ChevronRight } from "lucide-react";

export interface BreadcrumbItem {
  label: string;
  href?: string;
}

interface BreadcrumbProps {
  items: BreadcrumbItem[];
  className?: string;
}

export default function Breadcrumb({ items, className }: BreadcrumbProps) {
  const router = useRouter();

  return (
    <nav aria-label="Breadcrumb" className={className}>
      <ol className="flex items-center gap-1.5 flex-wrap">
        {/* Home icon */}
        <li>
          <button
            onClick={() => router.push("/dashboard")}
            className="text-gray-500 hover:text-gray-700 transition-colors"
            aria-label="Home"
          >
            <Home size={15} />
          </button>
        </li>

        {items.map((item, i) => {
          const isLast = i === items.length - 1;
          return (
            <li key={i} aria-current={isLast ? "page" : undefined} className="flex items-center gap-1.5">
              <ChevronRight size={14} className="text-gray-400" />
              {isLast || !item.href ? (
                <span className="text-sm font-medium text-gray-800">{item.label}</span>
              ) : (
                <Link
                  href={item.href}
                  className="text-sm text-gray-500 hover:text-gray-700 transition-colors"
                >
                  {item.label}
                </Link>
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
