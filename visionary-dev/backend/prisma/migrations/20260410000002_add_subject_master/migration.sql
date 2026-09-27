-- CreateTable: subject_master
CREATE TABLE "subject_master" (
    "id"        TEXT NOT NULL,
    "name"      TEXT NOT NULL,
    "gradeId"   TEXT NOT NULL,
    "isActive"  BOOLEAN NOT NULL DEFAULT true,
    "sortOrder" INTEGER NOT NULL DEFAULT 0,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "subject_master_pkey" PRIMARY KEY ("id")
);

-- CreateIndex (unique per grade)
CREATE UNIQUE INDEX "subject_master_name_gradeId_key" ON "subject_master"("name", "gradeId");

-- AddForeignKey
ALTER TABLE "subject_master" ADD CONSTRAINT "subject_master_gradeId_fkey"
    FOREIGN KEY ("gradeId") REFERENCES "grade_master"("id")
    ON DELETE CASCADE ON UPDATE CASCADE;
