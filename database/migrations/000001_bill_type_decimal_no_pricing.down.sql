-- Rollback for 000001_bill_type_decimal_no_pricing.up.sql
--
-- WARNING: promotions / promotion_prices / store_product_prices are dropped
-- tables — their data is gone and cannot be restored by this rollback. This
-- down migration only reverses the structural changes to tbl_bills,
-- tbl_bill_items, tbl_products, tbl_customer_addresses, and
-- tbl_customer_phones, and removes tbl_bill_sequences. Bill data itself
-- (tbl_bills, tbl_bill_items rows) is preserved throughout.
BEGIN;

DROP INDEX IF EXISTS tbl_bills_store_type_created_idx;
DROP INDEX IF EXISTS tbl_bills_store_type_no_uq;

DROP TABLE IF EXISTS tbl_bill_sequences;

ALTER TABLE tbl_customer_phones ADD COLUMN IF NOT EXISTS label varchar(100);
ALTER TABLE tbl_customer_addresses ADD COLUMN IF NOT EXISTS label varchar(100);

ALTER TABLE tbl_bills DROP CONSTRAINT IF EXISTS tbl_bills_type_chk;
ALTER TABLE tbl_bills DROP COLUMN IF EXISTS type;

ALTER TABLE tbl_bills DROP CONSTRAINT IF EXISTS tbl_bills_shipping_nonneg_chk;
ALTER TABLE tbl_bills DROP CONSTRAINT IF EXISTS tbl_bills_discount_nonneg_chk;
ALTER TABLE tbl_bills
	ALTER COLUMN discount TYPE double precision USING discount::double precision,
	ALTER COLUMN shipping_fee TYPE double precision USING shipping_fee::double precision,
	ALTER COLUMN total TYPE double precision USING total::double precision,
	ALTER COLUMN discount DROP DEFAULT,
	ALTER COLUMN shipping_fee DROP DEFAULT;
-- slip stays nullable — the pre-migration "" sentinel is not restored.

ALTER TABLE tbl_bill_items DROP CONSTRAINT IF EXISTS tbl_bill_items_bottle_int_chk;
ALTER TABLE tbl_bill_items DROP CONSTRAINT IF EXISTS tbl_bill_items_price_pos_chk;
ALTER TABLE tbl_bill_items DROP CONSTRAINT IF EXISTS tbl_bill_items_quantity_pos_chk;
ALTER TABLE tbl_bill_items
	ALTER COLUMN quantity TYPE double precision USING quantity::double precision,
	ALTER COLUMN price TYPE double precision USING price::double precision,
	ALTER COLUMN subtotal TYPE double precision USING subtotal::double precision;
ALTER TABLE tbl_bill_items DROP COLUMN IF EXISTS unit;
ALTER TABLE tbl_bill_items ADD COLUMN IF NOT EXISTS promotion_id int;

ALTER TABLE tbl_products DROP CONSTRAINT IF EXISTS tbl_products_unit_chk;
ALTER TABLE tbl_products ALTER COLUMN unit DROP NOT NULL;

-- tbl_promotions / tbl_promotion_prices / tbl_store_product_prices are not
-- recreated — restore their DDL from source control (pre-refactor
-- src/models/promotion_model.go, promotion_price_model.go,
-- store_product_price_model.go) and AutoMigrate if you actually need them
-- back, since their data cannot be recovered here regardless.

COMMIT;
