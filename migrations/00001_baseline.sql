-- +goose Up
CREATE SCHEMA IF NOT EXISTS "public";

CREATE TYPE "public"."effect_permission" AS ENUM ('grant', 'deny');
CREATE TYPE "public"."role_type" AS ENUM ('system', 'domain');

CREATE TABLE "public"."roles" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "role_name" varchar(50) NOT NULL UNIQUE,
    "role_code" varchar(50) NOT NULL UNIQUE,
    "role_category" role_type NOT NULL,
    "is_active" boolean NOT NULL DEFAULT true,
    "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp,
    PRIMARY KEY ("id")
);

CREATE TABLE "public"."user_roles" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "role_id" uuid NOT NULL,
    "user_id" uuid NOT NULL,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" timestamp,
    PRIMARY KEY ("id")
);
-- Indexes
CREATE UNIQUE INDEX "user_roles_index_2" ON "public"."user_roles" ("user_id", "role_id");

CREATE TABLE "public"."role_permissions" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "permission_id" uuid NOT NULL,
    "role_id" uuid NOT NULL,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" timestamp,
    PRIMARY KEY ("id")
);
-- Indexes
CREATE UNIQUE INDEX "role_permissions_index_2" ON "public"."role_permissions" ("role_id", "permission_id");

CREATE TABLE "public"."permissions" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "permission_name" varchar(50) NOT NULL,
    "module" varchar(50) NOT NULL,
    "act" varchar(50) NOT NULL,
    "is_active" boolean NOT NULL DEFAULT true,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id")
);
-- Indexes
CREATE UNIQUE INDEX "permissions_index_2" ON "public"."permissions" ("module", "act");

CREATE TABLE "public"."users" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "fullname" varchar(150) NOT NULL,
    "username" varchar(50) NOT NULL UNIQUE,
    "email" varchar(255) NOT NULL UNIQUE,
    "verified_email" boolean NOT NULL DEFAULT false,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp,
    "deleted_at" timestamp,
    "last_login_at" timestamp,
    PRIMARY KEY ("id")
);

CREATE TABLE "public"."user_permissions" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "user_id" uuid NOT NULL,
    "permission_id" uuid NOT NULL,
    "effect" effect_permission NOT NULL,
    "reasons" text,
    "created_by" uuid NOT NULL,
    "expires_at" timestamp,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" timestamp,
    PRIMARY KEY ("id")
);
-- Indexes
CREATE UNIQUE INDEX "user_permissions_index_2" ON "public"."user_permissions" ("user_id", "permission_id");

CREATE TABLE "public"."user_accounts" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "user_id" uuid NOT NULL,
    "password_hash" varchar(255),
    "provider" varchar(250) NOT NULL DEFAULT 'credentials',
    "provider_account_id" varchar(250),
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp,
    PRIMARY KEY ("id")
);
-- Indexes
CREATE UNIQUE INDEX "user_accounts_index_2" ON "public"."user_accounts" ("provider", "provider_account_id");
CREATE UNIQUE INDEX "user_accounts_index_3" ON "public"."user_accounts" ("user_id", "provider");

CREATE TABLE "public"."user_sessions" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "user_id" uuid NOT NULL,
    "active_role" uuid,
    "token_hash" varchar(255) NOT NULL,
    "csrf_token" varchar(255),
    "ip_address" inet NOT NULL,
    "user_agent" text NOT NULL,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "expires_at" timestamp NOT NULL,
    PRIMARY KEY ("id")
);

-- Foreign key constraints
-- Schema: public
ALTER TABLE "public"."user_accounts" ADD CONSTRAINT "fk_user_accounts_user_id_users_id" FOREIGN KEY("user_id") REFERENCES "public"."users"("id");
ALTER TABLE "public"."role_permissions" ADD CONSTRAINT "fk_role_permissions_permission_id_permissions_id" FOREIGN KEY("permission_id") REFERENCES "public"."permissions"("id");
ALTER TABLE "public"."role_permissions" ADD CONSTRAINT "fk_role_permissions_role_id_roles_id" FOREIGN KEY("role_id") REFERENCES "public"."roles"("id");
ALTER TABLE "public"."user_sessions" ADD CONSTRAINT "fk_user_sessions_user_id_users_id" FOREIGN KEY("user_id") REFERENCES "public"."users"("id");
ALTER TABLE "public"."user_permissions" ADD CONSTRAINT "fk_user_permissions_permission_id_permissions_id" FOREIGN KEY("permission_id") REFERENCES "public"."permissions"("id");
ALTER TABLE "public"."user_permissions" ADD CONSTRAINT "fk_user_permissions_user_id_users_id" FOREIGN KEY("user_id") REFERENCES "public"."users"("id");
ALTER TABLE "public"."user_roles" ADD CONSTRAINT "fk_user_roles_role_id_roles_id" FOREIGN KEY("role_id") REFERENCES "public"."roles"("id");
ALTER TABLE "public"."user_roles" ADD CONSTRAINT "fk_user_roles_user_id_users_id" FOREIGN KEY("user_id") REFERENCES "public"."users"("id");
ALTER TABLE "public"."user_sessions" ADD CONSTRAINT "fk_user_sessions_active_role_roles_id" FOREIGN KEY("active_role") REFERENCES "public"."roles"("id");
-- +goose Down
DROP TABLE IF EXISTS "public"."user_permissions" CASCADE;
DROP TABLE IF EXISTS "public"."role_permissions" CASCADE;
DROP TABLE IF EXISTS "public"."user_roles" CASCADE;
DROP TABLE IF EXISTS "public"."permissions" CASCADE;
DROP TABLE IF EXISTS "public"."sessions" CASCADE;
DROP TABLE IF EXISTS "public"."accounts" CASCADE;
DROP TABLE IF EXISTS "public"."roles" CASCADE;
DROP TABLE IF EXISTS "public"."users" CASCADE;
DROP TABLE IF EXISTS "public"."user_sessions" CASCADE;
DROP TABLE IF EXISTS "public"."user_accounts" CASCADE;
