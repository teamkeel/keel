CREATE SCHEMA IF NOT EXISTS "keel";

CREATE TABLE IF NOT EXISTS "keel"."task" (
	"id" text NOT NULL DEFAULT ksuid() PRIMARY KEY,
	"name" TEXT NOT NULL,
	"flow_run_id" TEXT NULL REFERENCES "keel"."flow_run" ("id") ON UPDATE CASCADE ON DELETE SET NULL,
	"status" TEXT NOT NULL,
	"created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
	"updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
	"assigned_to" TEXT REFERENCES "public"."identity" ("id") ON UPDATE CASCADE ON DELETE SET NULL,
	"assigned_at" TIMESTAMPTZ,
	"resolved_at" TIMESTAMPTZ,
	"deferred_until" TIMESTAMPTZ
);
CREATE OR REPLACE TRIGGER "keel_tasks_updated_at" BEFORE UPDATE ON "keel"."task" FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TABLE IF NOT EXISTS "keel"."task_status" (
	"id" text NOT NULL DEFAULT ksuid() PRIMARY KEY,
	"keel_task_id" TEXT NOT NULL REFERENCES "keel"."task" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
	"status" TEXT NOT NULL,
	"flow_run_id" TEXT NULL REFERENCES "keel"."flow_run" ("id") ON UPDATE CASCADE ON DELETE SET NULL,
	"assigned_to" TEXT REFERENCES "public"."identity" ("id") ON UPDATE CASCADE ON DELETE SET NULL,
	"created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
	"set_by" TEXT NOT NULL REFERENCES "public"."identity" ("id") ON UPDATE CASCADE ON DELETE RESTRICT
);

DO $$
BEGIN
	-- Add flow_run_id column to task_status if it doesn't exist
	IF NOT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema = 'keel'
		AND table_name = 'task_status'
		AND column_name = 'flow_run_id'
	) THEN
		ALTER TABLE "keel"."task_status"
		ADD COLUMN "flow_run_id" TEXT NULL REFERENCES "keel"."flow_run" ("id") ON UPDATE CASCADE ON DELETE SET NULL;
	END IF;

	IF NOT EXISTS (
		SELECT 1 FROM pg_indexes
		WHERE indexname = 'tasks_flow_run_id_idx'
		AND schemaname = 'keel'
	) THEN
		CREATE INDEX "tasks_flow_run_id_idx" ON "keel"."task" USING BTREE ("flow_run_id");
	END IF;

	IF NOT EXISTS (
		SELECT 1 FROM pg_indexes 
		WHERE indexname = 'tasks_status_idx' 
		AND schemaname = 'keel'
	) THEN
		CREATE INDEX "tasks_status_idx" ON "keel"."task" USING BTREE ("status");
	END IF;

	IF NOT EXISTS (
		SELECT 1 FROM pg_indexes 
		WHERE indexname = 'tasks_assigned_to_idx' 
		AND schemaname = 'keel'
	) THEN
		CREATE INDEX "tasks_assigned_to_idx" ON "keel"."task" USING BTREE ("assigned_to");
	END IF;

	IF NOT EXISTS (
		SELECT 1 FROM pg_indexes 
		WHERE indexname = 'tasks_deferred_until_idx' 
		AND schemaname = 'keel'
	) THEN
		CREATE INDEX "tasks_deferred_until_idx" ON "keel"."task" USING BTREE ("deferred_until");
	END IF;

	IF NOT EXISTS (
		SELECT 1 FROM pg_indexes 
		WHERE indexname = 'task_status_task_id_created_at_idx' 
		AND schemaname = 'keel'
	) THEN
		CREATE INDEX "task_status_task_id_created_at_idx" ON "keel"."task_status" USING BTREE ("keel_task_id", "created_at");
	END IF;
END $$;
