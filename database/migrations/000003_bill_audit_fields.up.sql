-- Audit fields on bills: who created it, when it was last edited through
-- PUT /bills/:id, and when its slip was last attached/replaced.
--
-- GORM AutoMigrate adds the three columns (and the created_by foreign key,
-- same constraint name as below) when the app starts, so on a normal
-- deployment the only thing this file adds is the slip_uploaded_at
-- backfill. Run it once after deploying.
--
-- created_by / edited_at are NOT backfilled — that history was never
-- recorded, so they stay NULL (the API reports null) for existing bills.
-- slip_uploaded_at is estimated as updated_at for bills that already have a
-- slip: an approximation, but better than empty.
--
-- Idempotent: re-running after a successful run is a no-op (the backfill
-- only fills rows still NULL).
BEGIN;

ALTER TABLE tbl_bills
	ADD COLUMN IF NOT EXISTS created_by bigint NULL,
	ADD COLUMN IF NOT EXISTS edited_at timestamptz NULL,
	ADD COLUMN IF NOT EXISTS slip_uploaded_at timestamptz NULL;

DO $$ BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM pg_constraint WHERE conname = 'fk_tbl_bills_creator'
	) THEN
		ALTER TABLE tbl_bills
			ADD CONSTRAINT fk_tbl_bills_creator FOREIGN KEY (created_by)
			REFERENCES tbl_users (id) ON UPDATE CASCADE ON DELETE SET NULL;
	END IF;
END $$;

UPDATE tbl_bills
SET slip_uploaded_at = updated_at
WHERE slip IS NOT NULL AND slip_uploaded_at IS NULL;

COMMIT;
