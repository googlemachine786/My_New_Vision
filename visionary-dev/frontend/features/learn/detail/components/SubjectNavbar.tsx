"use client";

import { usePathname } from "next/navigation";
import Breadcrumb from "@/components/ui/Breadcrumb";

function formatLabel(segment: string) {
  return segment
    .split("-")
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(" ");
}

export default function SubjectNavbar({ subjectName }: { subjectName?: string }) {
  const pathname = usePathname();

  const segments = pathname.split("/").filter(Boolean);

  const items = segments.map((segment, i) => {
    const href = "/" + segments.slice(0, i + 1).join("/");
    const isLast = i === segments.length - 1;
    return {
      label: isLast && subjectName ? subjectName : formatLabel(segment),
      href: isLast ? undefined : href,
    };
  });

  return (
    <div className="w-full py-3 mb-4">
      <Breadcrumb items={items} />
    </div>
  );
}
