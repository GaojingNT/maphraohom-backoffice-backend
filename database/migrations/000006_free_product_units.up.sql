-- Payment-bill items can be billed in any unit the user types (ลัง, กล่อง,
-- ...), so tbl_products_unit_chk goes from a fixed list (000005) to "any
-- non-blank unit". varchar(20) already caps the length. NOT VALID, as
-- before.
--
-- The app swaps the constraint on start (pkg/database/product_unit_check.go),
-- so on a normal deployment this file is a no-op; it is kept for applying
-- the change by hand. Only a constraint changes — no row is touched.
--
-- Idempotent.
BEGIN;

ALTER TABLE tbl_products DROP CONSTRAINT IF EXISTS tbl_products_unit_chk;
ALTER TABLE tbl_products ADD CONSTRAINT tbl_products_unit_chk
	CHECK (btrim(unit) <> '') NOT VALID;

COMMIT;
