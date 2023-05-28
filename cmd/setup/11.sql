BEGIN;
-- create table with empty created_at
ALTER TABLE eventstore.events ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NULL;
COMMIT;

BEGIN;
-- backfill created_at
UPDATE eventstore.events SET created_at = creation_date WHERE created_at IS NULL;
COMMIT;

BEGIN;
CREATE INDEX events_ca ON eventstore.events (created_at);
-- set column rules
ALTER TABLE eventstore.events ALTER COLUMN created_at SET DEFAULT clock_timestamp();
ALTER TABLE eventstore.events ALTER COLUMN created_at SET NOT NULL;
COMMIT;