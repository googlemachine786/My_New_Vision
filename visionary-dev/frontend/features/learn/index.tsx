"use client";

import ContinueLearning from "./components/ContinueLearning";
import YourSubjects from "./components/YourSubjects";

export default function Learn() {
  return (
    <div className="w-full min-w-0 pb-8">
      <ContinueLearning />
      <YourSubjects />
    </div>
  );
}
