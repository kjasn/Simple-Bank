CREATE TYPE "user_role" AS ENUM (
  'depositor',
  'banker'
);

ALTER TABLE "users" ADD COLUMN "role" "user_role" NOT NULL DEFAULT 'depositor';