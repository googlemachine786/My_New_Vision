#!/bin/sh
# Init-container entrypoint: runs Prisma migrations and conditionally seeds master data.
#
# Behavior:
#   - Always runs `prisma migrate deploy` (idempotent).
#   - Runs the seed if subject_master is empty, OR if SEED_ON_BOOT=force.
#   - Skips the seed otherwise (the seed uses deleteMany on subjects/books/chapters,
#     which cascades to subject_track/chapter_track and would wipe user progress).
set -e

echo "[migrate-and-seed] Running prisma migrate deploy..."
npx prisma migrate deploy

echo "[migrate-and-seed] Checking whether to seed..."
SHOULD_SEED=$(node -e '
const { PrismaClient } = require("@prisma/client");
const p = new PrismaClient();
const force = process.env.SEED_ON_BOOT === "force";
p.subjectMaster.count()
  .then((n) => { p.$disconnect(); console.log(force || n === 0 ? "yes" : "no"); })
  .catch((e) => { console.error("[seed-check] failed:", e.message); process.exit(1); });
')

if [ "$SHOULD_SEED" = "yes" ]; then
  if [ "$SEED_ON_BOOT" = "force" ]; then
    if [ "$NODE_ENV" = "production" ]; then
      echo "[migrate-and-seed] Refusing force seed: NODE_ENV=production." >&2
      exit 1
    fi
    case "$DATABASE_URL" in
      *supabase.co*|*supabase.com*)
        echo "[migrate-and-seed] Refusing force seed: DATABASE_URL points at Supabase." >&2
        exit 1
        ;;
    esac
    if [ "$I_KNOW_WHAT_IM_DOING" != "1" ]; then
      echo "[migrate-and-seed] Refusing force seed: set I_KNOW_WHAT_IM_DOING=1 to confirm." >&2
      exit 1
    fi
    echo "[migrate-and-seed] force seed confirmed (I_KNOW_WHAT_IM_DOING=1)."
  fi
  echo "[migrate-and-seed] Seeding master data..."
  npx ts-node prisma/seed.ts
else
  echo "[migrate-and-seed] Master data already populated; skipping seed."
  echo "                   Set SEED_ON_BOOT=force AND I_KNOW_WHAT_IM_DOING=1 in .env to re-seed."
fi

echo "[migrate-and-seed] Done."
