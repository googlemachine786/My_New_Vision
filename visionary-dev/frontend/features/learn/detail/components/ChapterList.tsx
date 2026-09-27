"use client";

import { useEffect, useState } from "react";
import Tabs from "@/components/ui/Tabs";
import Chapter from "./Chapter";
import { getSubjectChapters } from "@/lib/auth";
import type { BookWithChapters } from "@/types";

const PART_LABELS: Record<number, string> = { 1: "Part I", 2: "Part II", 3: "Part III", 4: "Part IV" };

export default function ChapterList({ subjectId }: { subjectId: string }) {
    const [book, setBook] = useState<BookWithChapters | null>(null);
    const [activePart, setActivePart] = useState<number>(1);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        getSubjectChapters(subjectId)
            .then((data) => {
                // subjects have a single book now; take the first
                setBook(data[0] ?? null);
                setActivePart(1);
            })
            .catch(() => setBook(null))
            .finally(() => setLoading(false));
    }, [subjectId]);

    if (loading) {
        return (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-6">
                {Array.from({ length: 6 }).map((_, i) => (
                    <div key={i} className="bg-gray-100 animate-pulse rounded-2xl h-48" />
                ))}
            </div>
        );
    }

    if (!book) return <p className="text-sm text-gray-400 py-4">No chapters found.</p>;

    const hasParts = book.totalParts != null && book.totalParts > 1;
    const totalParts = book.totalParts ?? 1;

    // Build tab labels
    const tabs = hasParts
        ? Array.from({ length: totalParts }, (_, i) => PART_LABELS[i + 1] ?? `Part ${i + 1}`)
        : [];

    // Filter chapters by active part (or show all if no parts)
    const chapters = hasParts
        ? book.chapters.filter((c) => c.part === activePart)
        : book.chapters;

    return (
        <div className="w-full">
            {hasParts && (
                <div className="mb-6">
                    <Tabs
                        tabs={tabs}
                        activeTab={PART_LABELS[activePart] ?? `Part ${activePart}`}
                        onChange={(label) => {
                            const idx = tabs.indexOf(label) + 1;
                            setActivePart(idx);
                        }}
                    />
                </div>
            )}

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {chapters.map((chapter) => (
                    <Chapter
                        key={chapter.id}
                        id={chapter.id}
                        number={chapter.sortOrder}
                        title={chapter.name}
                        progress={0}
                        status="Not Started"
                        sections={0}
                        sectionsCompleted={0}
                        link={`/learn/${subjectId}/${chapter.id}`}
                    />
                ))}
            </div>

            {chapters.length === 0 && (
                <p className="text-sm text-gray-400 py-4">No chapters in this part.</p>
            )}
        </div>
    );
}
