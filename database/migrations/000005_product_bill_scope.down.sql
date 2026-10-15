-- Reverts 000005: drops bill_scope and puts back 000001's unit check
-- ('กก.', 'ขวด' only — NOT VALID, so products already saved with another
-- unit stay as they are; only new ones are restricted again).
--
-- Dropping bill_scope: Products typed in on payment bills
-- stay in tbl_products (bills reference them) but lose their 'payment'
-- marker, so they'd show up on receipt bills too. No bill data is touched.
--
-- Note: while the app's models still declare BillScope, GORM AutoMigrate
-- re-adds the column (as 'all' for every product) on the next start.
BEGIN;

ALTER TABLE tbl_products DROP COLUMN IF EXISTS bill_scope;

ALTER TABLE tbl_products DROP CONSTRAINT IF EXISTS tbl_products_unit_chk;
ALTER TABLE tbl_products ADD CONSTRAINT tbl_products_unit_chk
	CHECK (unit IN ('กก.', 'ขวด')) NOT VALID;

COMMIT;
