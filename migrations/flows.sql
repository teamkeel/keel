CREATE TABLE IF NOT EXISTS keel_flow_run (
	"id" text NOT NULL DEFAULT ksuid() PRIMARY KEY,
	"name" TEXT NOT NULL,
	"status" TEXT NOT NULL,
	"input" JSONB DEFAULT NULL,
	"created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
	"updated_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE OR REPLACE TRIGGER keel_flow_run_updated_at BEFORE UPDATE ON "keel_flow_run" FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TABLE IF NOT EXISTS keel_flow_step (
	"id" text NOT NULL DEFAULT ksuid() PRIMARY KEY,
	"run_id" text NOT NULL REFERENCES "keel_flow_run" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
	"name" TEXT NOT NULL,
	"status" TEXT NOT NULL,
	"type" TEXT NOT NULL,
	"value" JSONB DEFAULT NULL,
	"created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
	"updated_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE OR REPLACE TRIGGER keel_flow_step_updated_at BEFORE UPDATE ON "keel_flow_step" FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

