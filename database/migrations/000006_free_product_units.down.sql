-- Reverts 000006 by putting back 000005's fixed unit list. NOT VALID, so
-- products already saved with another unit (e.g. ลัง) stay as they are;
-- only new or updated rows are restricted again. No row is touched.
BEGIN;

ALTER TABLE tbl_products DROP CONSTRAINT IF EXISTS tbl_products_unit_chk;
ALTER TABLE tbl_products ADD CONSTRAINT tbl_products_unit_chk
	CHECK (unit IN ('รายการ', 'ครั้ง', 'ชิ้น', 'ถุง', 'กก.', 'ขวด')) NOT VALID;

COMMIT;
