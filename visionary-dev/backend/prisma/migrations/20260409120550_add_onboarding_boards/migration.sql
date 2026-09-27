-- AlterTable
ALTER TABLE "users" ADD COLUMN     "board" TEXT,
ADD COLUMN     "category" TEXT,
ADD COLUMN     "fullName" TEXT,
ADD COLUMN     "grade" TEXT,
ADD COLUMN     "onboardingCompleted" BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN     "organizationName" TEXT,
ADD COLUMN     "organizationType" TEXT,
ADD COLUMN     "subject" TEXT;

-- CreateTable
CREATE TABLE "board_master" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "isActive" BOOLEAN NOT NULL DEFAULT true,
    "sortOrder" INTEGER NOT NULL DEFAULT 0,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "board_master_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "board_master_name_key" ON "board_master"("name");
