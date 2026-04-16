CREATE SCHEMA IF NOT EXISTS "public";

CREATE TABLE "public"."permissions" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "permission_name" varchar NOT NULL,
    "module" varchar NOT NULL,
    "act" varchar NOT NULL,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" timestamp,
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

CREATE TABLE "public"."accounts" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "user_id" uuid NOT NULL,
    "password_hash" varchar NOT NULL,
    "provider" varchar,
    "provider_account_id" varchar,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp,
    PRIMARY KEY ("id")
);

CREATE TABLE "public"."users" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "fullname" varchar NOT NULL,
    "username" varchar NOT NULL,
    "email" varchar NOT NULL,
    "verified_email" boolean NOT NULL DEFAULT false,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp,
    "deleted_at" timestamp,
    PRIMARY KEY ("id")
);

CREATE TABLE "public"."sessions" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "user_id" uuid NOT NULL,
    "active_role" uuid,
    "token_hash" varchar NOT NULL,
    "csrf_token" varchar NOT NULL,
    "ip_address" varchar NOT NULL,
    "user_agent" varchar NOT NULL,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "expired_at" timestamp,
    PRIMARY KEY ("id")
);

CREATE TABLE "public"."roles" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "role_name" varchar NOT NULL,
    "role_code" varchar,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp,
    "deleted_at" timestamp,
    PRIMARY KEY ("id")
);

CREATE TABLE "public"."role_permissions" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "permission_id" uuid NOT NULL,
    "role_id" uuid NOT NULL,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" timestamp,
    PRIMARY KEY ("id")
);

CREATE TABLE "public"."user_permissions" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "user_id" uuid NOT NULL,
    "permission_id" uuid NOT NULL,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" timestamp,
    PRIMARY KEY ("id")
);

-- Foreign key constraints
-- Schema: public
ALTER TABLE "public"."role_permissions" ADD CONSTRAINT "fk_role_permissions_permission_id_permissions_id" FOREIGN KEY("permission_id") REFERENCES "public"."permissions"("id");
ALTER TABLE "public"."user_permissions" ADD CONSTRAINT "fk_user_permissions_permission_id_permissions_id" FOREIGN KEY("permission_id") REFERENCES "public"."permissions"("id");
ALTER TABLE "public"."user_roles" ADD CONSTRAINT "fk_user_roles_role_id_roles_id" FOREIGN KEY("role_id") REFERENCES "public"."roles"("id");
ALTER TABLE "public"."role_permissions" ADD CONSTRAINT "fk_role_permissions_role_id_roles_id" FOREIGN KEY("role_id") REFERENCES "public"."roles"("id");
ALTER TABLE "public"."accounts" ADD CONSTRAINT "fk_accounts_user_id_users_id" FOREIGN KEY("user_id") REFERENCES "public"."users"("id");
ALTER TABLE "public"."user_roles" ADD CONSTRAINT "fk_user_roles_user_id_users_id" FOREIGN KEY("user_id") REFERENCES "public"."users"("id");
ALTER TABLE "public"."user_permissions" ADD CONSTRAINT "fk_user_permissions_user_id_users_id" FOREIGN KEY("user_id") REFERENCES "public"."users"("id");
ALTER TABLE "public"."sessions" ADD CONSTRAINT "fk_sessions_user_id_users_id" FOREIGN KEY("user_id") REFERENCES "public"."users"("id");