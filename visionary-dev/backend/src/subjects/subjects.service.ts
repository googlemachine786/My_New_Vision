import { Injectable } from "@nestjs/common";
import { PrismaService } from "../prisma/prisma.service";
import { AuthService } from "../auth/auth.service";

@Injectable()
export class SubjectsService {
  constructor(
    private prisma: PrismaService,
    private authService: AuthService,
  ) {}

  async getForUser(userId: string) {
    const user = await this.prisma.user.findUnique({
      where: { id: userId },
      select: { grade: true, board: true },
    });

    if (!user?.grade) return [];

    const select = { id: true, name: true, icon: true, sortOrder: true } as const;
    const orderBy = { sortOrder: "asc" } as const;

    // If user has a board, try board-specific subjects first
    if (user.board) {
      const boardSubjects = await this.prisma.subjectMaster.findMany({
        where: { gradeId: user.grade, boardId: user.board, isActive: true },
        orderBy,
        select,
      });
      if (boardSubjects.length > 0) return boardSubjects;
    }

    // Fall back to common subjects (boardId = null)
    return this.prisma.subjectMaster.findMany({
      where: { gradeId: user.grade, boardId: null, isActive: true },
      orderBy,
      select,
    });
  }

  async startSubject(userId: string, subjectId: string) {
    // Upsert subject_track only — isStarted is set on first chapter click
    await this.prisma.subjectTrack.upsert({
      where: { userId_subjectId: { userId, subjectId } },
      update: {},
      create: { userId, subjectId },
    });

    const user = await this.prisma.user.findUnique({ where: { id: userId } });
    return this.authService.issueToken(user);
  }

  async startChapter(userId: string, chapterId: string) {
    // Check if this is the user's first chapter click
    const existing = await this.prisma.chapterTrack.findUnique({
      where: { userId_chapterId: { userId, chapterId } },
    });

    // Upsert chapter_track
    await this.prisma.chapterTrack.upsert({
      where: { userId_chapterId: { userId, chapterId } },
      update: { status: "IN_PROGRESS", updatedAt: new Date() },
      create: { userId, chapterId, status: "IN_PROGRESS" },
    });

    // Set isStarted = true on first ever chapter click
    const user = await this.prisma.user.update({
      where: { id: userId },
      data: { isStarted: true },
    });

    return this.authService.issueToken(user);
  }

  async getChapterStats(userId: string) {
    const tracks = await this.prisma.chapterTrack.findMany({
      where: { userId },
      select: { status: true },
    });
    return {
      total: tracks.length,
      inProgress: tracks.filter((t) => t.status === "IN_PROGRESS").length,
      completed: tracks.filter((t) => t.status === "COMPLETED").length,
    };
  }

  async getAllTrackedChapters(userId: string) {
    return this._fetchTrackedChapters(userId);
  }

  async getRecentChapters(userId: string) {
    return this._fetchTrackedChapters(userId, 3);
  }

  private async _fetchTrackedChapters(userId: string, take?: number) {
    const tracks = await this.prisma.chapterTrack.findMany({
      where: { userId },
      orderBy: { updatedAt: "desc" },
      ...(take ? { take } : {}),
      select: {
        chapterId: true,
        progress: true,
        status: true,
        updatedAt: true,
        chapter: {
          select: {
            id: true,
            name: true,
            sortOrder: true,
            book: {
              select: {
                name: true,
                subject: {
                  select: { id: true, name: true, icon: true },
                },
              },
            },
          },
        },
      },
    });

    return tracks.map((t) => ({
      chapterId: t.chapterId,
      chapterName: t.chapter.name,
      sortOrder: t.chapter.sortOrder,
      bookName: t.chapter.book.name,
      subjectId: t.chapter.book.subject.id,
      subjectName: t.chapter.book.subject.name,
      subjectIcon: t.chapter.book.subject.icon,
      progress: t.progress,
      status: t.status,
    }));
  }

  async getStartedSubject(userId: string) {
    const track = await this.prisma.subjectTrack.findFirst({
      where: { userId },
      orderBy: { updatedAt: "desc" },
      select: {
        subjectId: true,
        subject: { select: { name: true, icon: true } },
      },
    });
    if (!track) return null;
    return { subjectId: track.subjectId, subjectName: track.subject.name, icon: track.subject.icon };
  }

  // Returns ALL started subjects (ordered by startedAt ASC = first started first)
  // with books + chapters + user progress merged in
  async getAllStartedChapters(userId: string) {
    const tracks = await this.prisma.subjectTrack.findMany({
      where: { userId },
      orderBy: { startedAt: "asc" },
      select: {
        subjectId: true,
        startedAt: true,
        subject: { select: { name: true, icon: true } },
      },
    });

    const result = [];
    for (const track of tracks) {
      const books = await this.getChapters(userId, track.subjectId);
      result.push({
        subjectId: track.subjectId,
        subjectName: track.subject.name,
        icon: track.subject.icon,
        startedAt: track.startedAt,
        books,
      });
    }
    return result;
  }

  async getChapterTopics(chapterId: string) {
    return this.prisma.topicMaster.findMany({
      where: { chapterId, status: "active" },
      orderBy: { sortOrder: "asc" },
      select: {
        id: true,
        pointNumber: true,
        title: true,
        sortOrder: true,
        subtopics: {
          where: { status: "active" },
          orderBy: { sortOrder: "asc" },
          select: {
            id: true,
            pointNumber: true,
            title: true,
            sortOrder: true,
          },
        },
      },
    });
  }

  async getChapters(userId: string, subjectId: string) {
    const books = await this.prisma.bookMaster.findMany({
      where: { subjectId, isActive: true },
      orderBy: { sortOrder: "asc" },
      select: {
        id: true,
        name: true,
        sortOrder: true,
        totalParts: true,
        chapters: {
          where: { isActive: true },
          orderBy: { sortOrder: "asc" },
          select: { id: true, name: true, sortOrder: true, part: true },
        },
      },
    });

    // Fetch user's chapter tracks for this subject
    const chapterIds = books.flatMap((b) => b.chapters.map((c) => c.id));
    const tracks = await this.prisma.chapterTrack.findMany({
      where: { userId, chapterId: { in: chapterIds } },
      select: { chapterId: true, progress: true, status: true },
    });
    const trackMap = new Map(tracks.map((t) => [t.chapterId, t]));

    // Merge progress into chapters
    return books.map((book) => ({
      ...book,
      chapters: book.chapters.map((chapter) => {
        const t = trackMap.get(chapter.id);
        return {
          ...chapter,
          progress: t?.progress ?? 0,
          status: t?.status ?? "NOT_STARTED",
        };
      }),
    }));
  }
}
