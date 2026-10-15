-- A bill can carry up to 5 slip images: they move from the single
-- tbl_bills.slip column into tbl_bill_slips (one row per image).
--
-- GORM AutoMigrate creates tbl_bill_slips, and the app runs the same
-- backfill (pkg/database/backfill.go) right after it on every start, so on
-- a normal deployment this file has nothing left to do. It is kept so the
-- change can also be applied by hand, and re-running it is safe.
--
-- Nothing is deleted or rewritten: every existing tbl_bills.slip is
-- COPIED into tbl_bill_slips, tbl_bills.slip keeps its value (the app keeps
-- it equal to the bill's first slip from now on), and the slip files in
-- storage are untouched — only their object keys are copied.
--
-- Idempotent: a bill that already has a row with its tbl_bills.slip key is
-- skipped.
BEGIN;

CREATE TABLE IF NOT EXISTS tbl_bill_slips (
	id bigserial PRIMARY KEY,
	created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
	bill_id bigint NOT NULL,
	object_key varchar(255) NOT NULL,
	uploaded_by bigint NULL
);

CREATE INDEX IF NOT EXISTS idx_tbl_bill_slips_bill_id ON tbl_bill_slips (bill_id);

DO $$ BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM pg_constraint WHERE conname = 'fk_tbl_bills_slips'
	) THEN
		ALTER TABLE tbl_bill_slips
			ADD CONSTRAINT fk_tbl_bills_slips FOREIGN KEY (bill_id)
			REFERENCES tbl_bills (id) ON UPDATE CASCADE ON DELETE CASCADE;
	END IF;
	IF NOT EXISTS (
		SELECT 1 FROM pg_constraint WHERE conname = 'fk_tbl_bill_slips_uploader'
	) THEN
		ALTER TABLE tbl_bill_slips
			ADD CONSTRAINT fk_tbl_bill_slips_uploader FOREIGN KEY (uploaded_by)
			REFERENCES tbl_users (id) ON UPDATE CASCADE ON DELETE SET NULL;
	END IF;
END $$;

INSERT INTO tbl_bill_slips (bill_id, object_key, created_at, updated_at)
SELECT b.id, b.slip,
	COALESCE(b.slip_uploaded_at, b.updated_at),
	COALESCE(b.slip_uploaded_at, b.updated_at)
FROM tbl_bills b
WHERE b.slip IS NOT NULL AND b.slip <> ''
	AND NOT EXISTS (
		SELECT 1 FROM tbl_bill_slips s
		WHERE s.bill_id = b.id AND s.object_key = b.slip
	)
ORDER BY b.id;

COMMIT;
