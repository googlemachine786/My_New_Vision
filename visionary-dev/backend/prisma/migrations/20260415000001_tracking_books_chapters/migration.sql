-- Add isStarted to users
ALTER TABLE "users" ADD COLUMN "isStarted" BOOLEAN NOT NULL DEFAULT false;

-- book_master
CREATE TABLE "book_master" (
  "id"        TEXT NOT NULL,
  "name"      TEXT NOT NULL,
  "subjectId" TEXT NOT NULL,
  "sortOrder" INTEGER NOT NULL DEFAULT 0,
  "isActive"  BOOLEAN NOT NULL DEFAULT true,
  "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT "book_master_pkey" PRIMARY KEY ("id"),
  CONSTRAINT "book_master_subjectId_fkey" FOREIGN KEY ("subjectId")
    REFERENCES "subject_master"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

-- chapter_master
CREATE TABLE "chapter_master" (
  "id"        TEXT NOT NULL,
  "name"      TEXT NOT NULL,
  "bookId"    TEXT NOT NULL,
  "sortOrder" INTEGER NOT NULL DEFAULT 0,
  "isActive"  BOOLEAN NOT NULL DEFAULT true,
  "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT "chapter_master_pkey" PRIMARY KEY ("id"),
  CONSTRAINT "chapter_master_bookId_fkey" FOREIGN KEY ("bookId")
    REFERENCES "book_master"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

-- subject_track
CREATE TABLE "subject_track" (
  "id"        TEXT NOT NULL,
  "userId"    TEXT NOT NULL,
  "subjectId" TEXT NOT NULL,
  "startedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updatedAt" TIMESTAMP(3) NOT NULL,
  CONSTRAINT "subject_track_pkey" PRIMARY KEY ("id"),
  CONSTRAINT "subject_track_userId_fkey" FOREIGN KEY ("userId")
    REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT "subject_track_subjectId_fkey" FOREIGN KEY ("subjectId")
    REFERENCES "subject_master"("id") ON DELETE CASCADE ON UPDATE CASCADE
);
CREATE UNIQUE INDEX "subject_track_userId_subjectId_key" ON "subject_track"("userId", "subjectId");

-- chapter_track
CREATE TYPE "TrackStatus" AS ENUM ('NOT_STARTED', 'IN_PROGRESS', 'COMPLETED');

CREATE TABLE "chapter_track" (
  "id"          TEXT NOT NULL,
  "userId"      TEXT NOT NULL,
  "chapterId"   TEXT NOT NULL,
  "progress"    INTEGER NOT NULL DEFAULT 0,
  "status"      "TrackStatus" NOT NULL DEFAULT 'NOT_STARTED',
  "startedAt"   TIMESTAMP(3),
  "completedAt" TIMESTAMP(3),
  "updatedAt"   TIMESTAMP(3) NOT NULL,
  CONSTRAINT "chapter_track_pkey" PRIMARY KEY ("id"),
  CONSTRAINT "chapter_track_userId_fkey" FOREIGN KEY ("userId")
    REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT "chapter_track_chapterId_fkey" FOREIGN KEY ("chapterId")
    REFERENCES "chapter_master"("id") ON DELETE CASCADE ON UPDATE CASCADE
);
CREATE UNIQUE INDEX "chapter_track_userId_chapterId_key" ON "chapter_track"("userId", "chapterId");
