-- Add boardId and icon columns to subject_master
ALTER TABLE "subject_master" ADD COLUMN "boardId" TEXT;
ALTER TABLE "subject_master" ADD COLUMN "icon"    TEXT;

-- Drop old unique constraint (name + gradeId only)
DROP INDEX IF EXISTS "subject_master_name_gradeId_key";

-- Add FK: boardId → board_master.id (SET NULL on delete)
ALTER TABLE "subject_master"
  ADD CONSTRAINT "subject_master_boardId_fkey"
  FOREIGN KEY ("boardId") REFERENCES "board_master"("id")
  ON DELETE SET NULL ON UPDATE CASCADE;
