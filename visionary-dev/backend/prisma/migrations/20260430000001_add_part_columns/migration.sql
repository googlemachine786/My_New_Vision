-- Add totalParts to book_master (null = no parts/tabs)
ALTER TABLE "book_master" ADD COLUMN "totalParts" INTEGER;

-- Add part to chapter_master (null = no part grouping)
ALTER TABLE "chapter_master" ADD COLUMN "part" INTEGER;
