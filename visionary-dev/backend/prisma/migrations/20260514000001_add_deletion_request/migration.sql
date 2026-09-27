ALTER TABLE "users" ADD COLUMN IF NOT EXISTS "isDeleted" BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS "deletion_request" (
    "id"        TEXT         NOT NULL,
    "userId"    TEXT         NOT NULL,
    "reason"    TEXT,
    "status"    TEXT         NOT NULL DEFAULT 'pending',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    CONSTRAINT "deletion_request_pkey" PRIMARY KEY ("id")
);

ALTER TABLE "deletion_request"
    ADD CONSTRAINT "deletion_request_userId_fkey"
    FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE;
