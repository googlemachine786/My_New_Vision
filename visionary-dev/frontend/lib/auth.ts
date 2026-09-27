import api from "./api";
import type { BoardMaster, BookWithChapters, GradeMaster, OnboardingPayload, RecentChapter, StartedSubject, StartedSubjectTrack, SubjectMaster, Topic, User, UserProfile } from "@/types";

interface AuthResponse {
  accessToken: string;
  user: {
    id: string;
    email: string;
    name: string;
    onboardingCompleted: boolean;
    plan?: string;
  };
}

function mapResponse(data: AuthResponse): { token: string; user: User } {
  return {
    token: data.accessToken,
    user: {
      id: data.user.id,
      email: data.user.email,
      fullName: data.user.name,
      onboardingCompleted: data.user.onboardingCompleted,
      isStarted: (data.user as any).isStarted ?? false,
    },
  };
}

export async function loginWithEmail(email: string, password: string) {
  const { data } = await api.post<AuthResponse>("/auth/login", { email, password });
  return mapResponse(data);
}

export async function signupWithEmail(email: string, password: string) {
  const { data } = await api.post<{ message: string; email: string }>("/auth/register", { email, password });
  return data;
}

export async function verifyOtp(email: string, otp: string) {
  const { data } = await api.post<AuthResponse>("/auth/verify-otp", { email, otp });
  return mapResponse(data);
}

export async function resendOtp(email: string) {
  const { data } = await api.post<{ message: string }>("/auth/resend-otp", { email });
  return data;
}

export async function completeOnboarding(payload: OnboardingPayload) {
  const { data } = await api.post<AuthResponse>("/onboarding/complete", payload);
  return mapResponse(data);
}

export async function getOnboardingStatus() {
  const { data } = await api.get<User>("/onboarding/status");
  return data;
}

export async function getBoards(): Promise<BoardMaster[]> {
  const { data } = await api.get<BoardMaster[]>("/boards");
  return data;
}

export async function getGrades(): Promise<GradeMaster[]> {
  const { data } = await api.get<GradeMaster[]>("/grades");
  return data;
}

export async function getMySubjects(): Promise<SubjectMaster[]> {
  const { data } = await api.get<SubjectMaster[]>("/subjects/me");
  return data;
}

export async function startSubject(subjectId: string): Promise<{ token: string; user: User }> {
  const { data } = await api.post<AuthResponse>(`/subjects/${subjectId}/start`);
  return mapResponse(data);
}

export async function getStartedSubject(): Promise<StartedSubject | null> {
  const { data } = await api.get<StartedSubject | null>("/subjects/started");
  return data;
}

export async function getSubjectChapters(subjectId: string): Promise<BookWithChapters[]> {
  const { data } = await api.get<BookWithChapters[]>(`/subjects/${subjectId}/chapters`);
  return data;
}

export async function getAllStartedChapters(): Promise<StartedSubjectTrack[]> {
  const { data } = await api.get<StartedSubjectTrack[]>("/subjects/my-tracks");
  return data;
}

export async function startChapter(chapterId: string): Promise<{ token: string; user: User }> {
  const { data } = await api.post<AuthResponse>(`/subjects/chapters/${chapterId}/start`);
  return mapResponse(data);
}

export async function getChapterStats(): Promise<{ total: number; inProgress: number; completed: number }> {
  const { data } = await api.get("/subjects/chapter-stats");
  return data;
}

export async function getAllTrackedChapters(): Promise<RecentChapter[]> {
  const { data } = await api.get<RecentChapter[]>("/subjects/all-tracked-chapters");
  return data;
}

export async function getRecentChapters(): Promise<RecentChapter[]> {
  const { data } = await api.get<RecentChapter[]>("/subjects/recent-chapters");
  return data;
}

export async function getUserProfile(): Promise<UserProfile> {
  const { data } = await api.get<UserProfile>("/auth/me");
  return data;
}

export async function updateUserProfile(payload: {
  fullName?: string;
  schoolName?: string;
  phone?: string;
}): Promise<UserProfile> {
  const { data } = await api.patch<UserProfile>("/auth/profile", payload);
  return data;
}

export async function changePassword(currentPassword: string, newPassword: string): Promise<{ message: string }> {
  const { data } = await api.patch<{ message: string }>("/auth/password", { currentPassword, newPassword });
  return data;
}

export async function requestAccountDeletion(reason?: string): Promise<{ message: string }> {
  const { data } = await api.post<{ message: string }>("/auth/delete-account", { reason });
  return data;
}

export async function getChapterTopics(chapterId: string): Promise<Topic[]> {
  const { data } = await api.get<Topic[]>(`/subjects/chapters/${chapterId}/topics`);
  return data;
}

export async function getMe(): Promise<User> {
  const { data } = await api.get<{
    id: string;
    email: string;
    name: string;
    fullName?: string | null;
    category?: string | null;
    onboardingCompleted: boolean;
    isStarted: boolean;
  }>("/auth/me");
  return {
    id: data.id,
    email: data.email,
    fullName: data.fullName ?? data.name,
    category: (data.category ?? undefined) as User["category"],
    onboardingCompleted: data.onboardingCompleted,
    isStarted: data.isStarted,
  };
}

export async function logout(): Promise<void> {
  try {
    await api.post("/auth/logout");
  } catch {
    // ignore — caller still clears local state
  }
}

// Same-origin path; Next.js rewrites /api/* to the backend so the OAuth
// callback returns through the frontend origin and the auth_token cookie
// lands where middleware can read it.
export function getGoogleAuthUrl(): string {
  return "/api/auth/google";
}
