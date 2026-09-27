export interface User {
  id: string;
  email: string;
  fullName?: string;
  schoolName?: string;
  phone?: string;
  category?: "student" | "teacher" | "organization";
  grade?: string;
  board?: string;
  subject?: string;
  organizationName?: string;
  organizationType?: string;
  onboardingCompleted: boolean;
  isStarted?: boolean;
}

export interface UserProfile {
  id: string;
  email: string;
  fullName: string | null;
  schoolName: string | null;
  phone: string | null;
  grade: string | null;
  board: string | null;
  gradeName: string | null;
  boardName: string | null;
}

export interface ChapterProgress {
  id: string;
  name: string;
  sortOrder: number;
  part: number | null;
  progress: number;
  status: "NOT_STARTED" | "IN_PROGRESS" | "COMPLETED";
}

export interface BookWithChapters {
  id: string;
  name: string;
  sortOrder: number;
  totalParts: number | null;
  chapters: ChapterProgress[];
}

export interface RecentChapter {
  chapterId: string;
  chapterName: string;
  sortOrder: number;
  bookName: string;
  subjectId: string;
  subjectName: string;
  subjectIcon: string | null;
  progress: number;
  status: "NOT_STARTED" | "IN_PROGRESS" | "COMPLETED";
}

export interface StartedSubject {
  subjectId: string;
  subjectName: string;
  icon: string | null;
}

export interface StartedSubjectTrack {
  subjectId: string;
  subjectName: string;
  icon: string | null;
  startedAt: string;
  books: BookWithChapters[];
}

export interface AuthState {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
  setAuth: (token: string, user: User) => void;
  setUserFromCookie: (user: User) => void;
  clear: () => Promise<void>;
}

export interface BoardMaster {
  id: string;
  name: string;
  isActive: boolean;
  sortOrder: number;
}

export interface GradeMaster {
  id: string;
  name: string;
  sortOrder: number;
}

export interface SubjectMaster {
  id: string;
  name: string;
  icon: string | null;
  sortOrder: number;
}

export interface OnboardingPayload {
  category: "student" | "teacher" | "organization";
  fullName: string;
  grade?: string;
  board?: string;
  subject?: string;
  organizationName?: string;
  organizationType?: string;
}

export interface Subtopic {
  id: string;
  pointNumber: string;
  title: string;
  sortOrder: number;
}

export interface Topic {
  id: string;
  pointNumber: string;
  title: string;
  sortOrder: number;
  subtopics: Subtopic[];
}

export interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  createdAt: string;
}

export interface Chat {
  id: string;
  title: string;
  createdAt: string;
  messages: ChatMessage[];
}
