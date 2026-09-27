"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { getMySubjects, startSubject } from "@/lib/auth";
import { useAuthStore } from "@/store/authStore";
import type { SubjectMaster } from "@/types";

export default function YourSubjects() {
    const router = useRouter();
    const setAuth = useAuthStore((s) => s.setAuth);
    const [subjects, setSubjects] = useState<SubjectMaster[]>([]);
    const [loading, setLoading] = useState(true);
    const [startingId, setStartingId] = useState<string | null>(null);

    useEffect(() => {
        getMySubjects()
            .then(setSubjects)
            .catch(() => setSubjects([]))
            .finally(() => setLoading(false));
    }, []);

    async function handleStart(id: string) {
        if (startingId) return;
        setStartingId(id);
        try {
            const { token, user } = await startSubject(id);
            setAuth(token, user);
            router.push(`/learn/${id}`);
        } catch {
            // ignore
        } finally {
            setStartingId(null);
        }
    }

    return (
        <div className="w-full mt-8">
            <h2 className="text-3xl font-medium text-gray-900 mb-6">Your Subjects</h2>

            {loading ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 w-full">
                    {Array.from({ length: 6 }).map((_, i) => (
                        <div key={i} className="bg-gray-100 animate-pulse rounded-[32px] h-[240px]" />
                    ))}
                </div>
            ) : subjects.length === 0 ? (
                <p className="text-sm text-gray-400 py-4">No subjects found for your grade and board.</p>
            ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 w-full">
                    {subjects.map((subject) => (
                        <div
                            key={subject.id}
                            className="bg-white border border-gray-100 flex flex-col"
                            style={{
                                borderRadius: 32,
                                padding: 32,
                                gap: 24,
                                minHeight: 240,
                            }}
                        >
                            {/* Icon */}
                            <div className="relative w-16 h-16">
                                <Image
                                    src={subject.icon ?? "/assets/Science-icon.png"}
                                    alt={subject.name}
                                    fill
                                    className="object-contain"
                                />
                            </div>

                            {/* Text */}
                            <div className="flex flex-col gap-1">
                                <p className="text-xl font-bold text-gray-900">{subject.name}</p>
                            </div>

                            {/* Start button */}
                            <div className="mt-auto flex justify-end">
                                <button
                                    onClick={() => handleStart(subject.id)}
                                    disabled={startingId === subject.id}
                                    className="px-10 py-3 rounded-full bg-blue-50 text-blue-500 text-sm font-medium hover:bg-blue-100 transition-colors disabled:opacity-50"
                                >
                                    {startingId === subject.id ? "..." : "Start"}
                                </button>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
