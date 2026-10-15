-- Payment bills can carry items the user types in, saved as products.
--
-- 1. Products get a bill_scope: 'all' (the existing coconut catalog — usable on
-- any bill) or 'payment' (items users type in on payment bills, saved as
-- products so they're suggested next time; payment bills only).
--
-- 2. tbl_products_unit_chk (000001: only 'กก.' and 'ขวด') is widened to the
--    units a typed-in item can have. NOT VALID, as before.
--
-- The app does both on start (GORM AutoMigrate for the column,
-- pkg/database/product_unit_check.go for the constraint), so on a normal
-- deployment this file is a no-op; it is kept for applying the change by
-- hand. Every existing product becomes 'all' through the column default;
-- no bill, item, product or customer row is changed or deleted.
--
-- Idempotent.
BEGIN;

ALTER TABLE tbl_products
	ADD COLUMN IF NOT EXISTS bill_scope varchar(20) NOT NULL DEFAULT 'all';

ALTER TABLE tbl_products DROP CONSTRAINT IF EXISTS tbl_products_unit_chk;
ALTER TABLE tbl_products ADD CONSTRAINT tbl_products_unit_chk
	CHECK (unit IN ('รายการ', 'ครั้ง', 'ชิ้น', 'ถุง', 'กก.', 'ขวด')) NOT VALID;

COMMIT;
