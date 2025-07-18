CREATE SCHEMA IF NOT EXISTS "keel";

CREATE TABLE IF NOT EXISTS "keel"."flow_run" (
	"id" text NOT NULL DEFAULT ksuid() PRIMARY KEY,
	"name" TEXT NOT NULL,
	"trace_id" TEXT,
	"status" TEXT NOT NULL,
	"input" JSONB DEFAULT NULL,
	"created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
	"updated_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE OR REPLACE TRIGGER "keel_flow_run_updated_at" BEFORE UPDATE ON "keel"."flow_run" FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TABLE IF NOT EXISTS "keel"."flow_step" (
	"id" text NOT NULL DEFAULT ksuid() PRIMARY KEY,
	"run_id" text NOT NULL REFERENCES "keel"."flow_run" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
	"name" TEXT NOT NULL,
	"stage" TEXT NULL,
	"status" TEXT NOT NULL,
	"type" TEXT NOT NULL,
	"value" JSONB DEFAULT NULL,
	"error" TEXT DEFAULT NULL,
	"span_id" TEXT,
	"start_time" TIMESTAMPTZ,
	"end_time" TIMESTAMPTZ,
	"created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
	"updated_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE OR REPLACE TRIGGER "keel_flow_step_updated_at" BEFORE UPDATE ON "keel"."flow_step" FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

ALTER TABLE "keel"."flow_run" ADD COLUMN IF NOT EXISTS "traceparent" TEXT;
ALTER TABLE "keel"."flow_run" ADD COLUMN IF NOT EXISTS "started_by" TEXT;
ALTER TABLE "keel"."flow_run" ADD COLUMN IF NOT EXISTS "data" JSONB DEFAULT NULL;
ALTER TABLE "keel"."flow_run" ADD COLUMN IF NOT EXISTS "config" JSONB DEFAULT NULL;

ALTER TABLE "keel"."flow_step" ADD COLUMN IF NOT EXISTS "action" TEXT;
ALTER TABLE "keel"."flow_step" ADD COLUMN IF NOT EXISTS "ui" JSONB DEFAULT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE indexname = 'flow_run_started_by_idx' 
        AND schemaname = 'keel'
    ) THEN
        CREATE INDEX "flow_run_started_by_idx" ON "keel"."flow_run" USING BTREE ("started_by");
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE indexname = 'flow_step_run_id_name_idx' 
        AND schemaname = 'keel'
    ) THEN
        CREATE INDEX "flow_step_run_id_name_idx" ON "keel"."flow_step" USING BTREE ("run_id", "name");
    END IF;
END $$;
