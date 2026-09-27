-- CreateTable: topic_master
CREATE TABLE "topic_master" (
    "id" TEXT NOT NULL,
    "chapterId" TEXT NOT NULL,
    "pointNumber" TEXT NOT NULL,
    "title" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'active',
    "sortOrder" INTEGER NOT NULL DEFAULT 0,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    CONSTRAINT "topic_master_pkey" PRIMARY KEY ("id")
);

-- CreateTable: subtopic_master
CREATE TABLE "subtopic_master" (
    "id" TEXT NOT NULL,
    "topicId" TEXT NOT NULL,
    "pointNumber" TEXT NOT NULL,
    "title" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'active',
    "sortOrder" INTEGER NOT NULL DEFAULT 0,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    CONSTRAINT "subtopic_master_pkey" PRIMARY KEY ("id")
);

-- AddForeignKey
ALTER TABLE "topic_master" ADD CONSTRAINT "topic_master_chapterId_fkey"
    FOREIGN KEY ("chapterId") REFERENCES "chapter_master"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "subtopic_master" ADD CONSTRAINT "subtopic_master_topicId_fkey"
    FOREIGN KEY ("topicId") REFERENCES "topic_master"("id") ON DELETE CASCADE ON UPDATE CASCADE;
