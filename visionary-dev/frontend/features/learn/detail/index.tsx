"use client";

import { useEffect, useState } from "react";
import SubjectNavbar from "./components/SubjectNavbar";
import ChapterList from "./components/ChapterList";
import { getMySubjects } from "@/lib/auth";

interface Props {
  subjectId: string;
}

export default function SubjectDetail({ subjectId }: Props) {
  const [subjectName, setSubjectName] = useState("");

  useEffect(() => {
    getMySubjects().then((subjects) => {
      const found = subjects.find((s) => s.id === subjectId);
      if (found) setSubjectName(found.name);
    }).catch(() => {});
  }, [subjectId]);

  return (
    <div className="w-full min-w-0 p-4 sm:p-6">
      <SubjectNavbar subjectName={subjectName} />
      {subjectName && (
        <h1 className="text-2xl font-bold text-gray-900 mb-4">{subjectName}</h1>
      )}
      <ChapterList subjectId={subjectId} />
    </div>
  );
}
