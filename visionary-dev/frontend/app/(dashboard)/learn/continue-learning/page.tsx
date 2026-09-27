"use client";

import SubjectNavbar from "@/features/learn/detail/components/SubjectNavbar";
import ContinueLearningAll from "@/features/learn/ContinueLearning";

const ContinueLearningPage = () => {
  return (
    <div className="w-full min-w-0 p-4 sm:p-6">
      <SubjectNavbar />
      <ContinueLearningAll />
    </div>
  );
};

export default ContinueLearningPage;
