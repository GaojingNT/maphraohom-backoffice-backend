-- Reverts 000003: drops the three audit columns (and the created_by foreign
-- key with them). Bills themselves are untouched; only the audit history
-- in these columns is lost.
--
-- Note: while the app's models still declare these fields, GORM
-- AutoMigrate re-adds the (empty) columns on the next start.
BEGIN;

ALTER TABLE tbl_bills
	DROP CONSTRAINT IF EXISTS fk_tbl_bills_creator,
	DROP COLUMN IF EXISTS created_by,
	DROP COLUMN IF EXISTS edited_at,
	DROP COLUMN IF EXISTS slip_uploaded_at;

COMMIT;
