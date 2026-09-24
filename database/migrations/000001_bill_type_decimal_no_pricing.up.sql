-- Refactor: drop promotions/per-store pricing, move money+quantity to
-- numeric, add bills.type + tbl_bill_sequences, drop label from customer
-- addresses/phones.
--
-- Safe to run against a database that already has real bill data:
--   * every ADD COLUMN / DROP COLUMN / DROP TABLE is IF [NOT] EXISTS
--   * new CHECK constraints are added NOT VALID so the migration cannot be
--     aborted by old rows that predate the rule — run VALIDATE CONSTRAINT
--     later, per table, once you've confirmed there's nothing to clean up
--   * no bills/bill_items/customers/products rows are deleted or truncated
--
-- Idempotent: re-running this file after a successful run is a no-op.
BEGIN;

-- 1) Drop promotions / per-store pricing -------------------------------------------------

ALTER TABLE tbl_bill_items DROP COLUMN IF EXISTS promotion_id;
DROP TABLE IF EXISTS tbl_promotion_prices;
DROP TABLE IF EXISTS tbl_promotions;
DROP TABLE IF EXISTS tbl_store_product_prices;

-- 2) products.unit ------------------------------------------------------------------------

ALTER TABLE tbl_products ADD COLUMN IF NOT EXISTS unit varchar(20);

UPDATE tbl_products
SET unit = CASE WHEN name = 'น้ำมะพร้าว' THEN 'ขวด' ELSE 'กก.' END
WHERE unit IS NULL;

ALTER TABLE tbl_products ALTER COLUMN unit SET NOT NULL;

DO $$ BEGIN
	ALTER TABLE tbl_products ADD CONSTRAINT tbl_products_unit_chk
		CHECK (unit IN ('กก.', 'ขวด')) NOT VALID;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- 3) bill_items: quantity/unit/price/subtotal ----------------------------------------------
-- Note: this codebase's bill_items column is already named "quantity" (not
-- "kilogram" as an older draft of the spec assumed), so there is nothing to
-- rename here.

ALTER TABLE tbl_bill_items ADD COLUMN IF NOT EXISTS unit varchar(20);

UPDATE tbl_bill_items bi
SET unit = p.unit
FROM tbl_products p
WHERE p.id = bi.product_id AND bi.unit IS NULL;

ALTER TABLE tbl_bill_items ALTER COLUMN unit SET NOT NULL;

ALTER TABLE tbl_bill_items
	ALTER COLUMN quantity TYPE numeric(10, 3) USING quantity::numeric(10, 3),
	ALTER COLUMN price TYPE numeric(10, 2) USING price::numeric(10, 2),
	ALTER COLUMN subtotal TYPE numeric(12, 2) USING subtotal::numeric(12, 2);

DO $$ BEGIN
	ALTER TABLE tbl_bill_items ADD CONSTRAINT tbl_bill_items_quantity_pos_chk
		CHECK (quantity > 0) NOT VALID;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
	ALTER TABLE tbl_bill_items ADD CONSTRAINT tbl_bill_items_price_pos_chk
		CHECK (price > 0) NOT VALID;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
	ALTER TABLE tbl_bill_items ADD CONSTRAINT tbl_bill_items_bottle_int_chk
		CHECK (unit <> 'ขวด' OR quantity = trunc(quantity)) NOT VALID;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- 4) bills: money to numeric, defaults, type -------------------------------------------------

ALTER TABLE tbl_bills
	ALTER COLUMN discount TYPE numeric(12, 2) USING discount::numeric(12, 2),
	ALTER COLUMN shipping_fee TYPE numeric(12, 2) USING shipping_fee::numeric(12, 2),
	ALTER COLUMN total TYPE numeric(12, 2) USING total::numeric(12, 2),
	ALTER COLUMN discount SET DEFAULT 0,
	ALTER COLUMN shipping_fee SET DEFAULT 0;

DO $$ BEGIN
	ALTER TABLE tbl_bills ADD CONSTRAINT tbl_bills_discount_nonneg_chk
		CHECK (discount >= 0) NOT VALID;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
	ALTER TABLE tbl_bills ADD CONSTRAINT tbl_bills_shipping_nonneg_chk
		CHECK (shipping_fee >= 0) NOT VALID;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

ALTER TABLE tbl_bills ADD COLUMN IF NOT EXISTS type varchar(20) NOT NULL DEFAULT 'receipt';

DO $$ BEGIN
	ALTER TABLE tbl_bills ADD CONSTRAINT tbl_bills_type_chk
		CHECK (type IN ('receipt', 'payment')) NOT VALID;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- slip previously used '' as its "no slip" sentinel; the app now treats the
-- column as nullable, so normalize existing empty strings to NULL.
UPDATE tbl_bills SET slip = NULL WHERE slip = '';
ALTER TABLE tbl_bills ALTER COLUMN slip DROP NOT NULL;

-- 5) drop label from customer addresses/phones -----------------------------------------------

ALTER TABLE tbl_customer_addresses DROP COLUMN IF EXISTS label;
ALTER TABLE tbl_customer_phones DROP COLUMN IF EXISTS label;

-- 6) bill numbering: tbl_bill_sequences per (store_id, type) ----------------------------------
-- Preserves the app's existing numbering rule (cap at 50 receipts per book,
-- reset to book 1 / receipt 1 every calendar year) — see
-- bill_module.Repository.nextBookReceipt, which this table's rows feed via
-- an atomic UPDATE ... RETURNING at bill-creation time. This table did not
-- exist before (numbering was computed live from tbl_bills), so there is no
-- old counter to migrate — only tbl_bills itself to read for a starting
-- point.

CREATE TABLE IF NOT EXISTS tbl_bill_sequences (
	store_id int NOT NULL REFERENCES tbl_stores (id),
	type varchar(20) NOT NULL,
	last_book_no int NOT NULL DEFAULT 0,
	last_receipt_no int NOT NULL DEFAULT 0,
	updated_at timestamp NOT NULL DEFAULT now(),
	PRIMARY KEY (store_id, type)
);

DO $$ BEGIN
	ALTER TABLE tbl_bill_sequences ADD CONSTRAINT tbl_bill_sequences_type_chk
		CHECK (type IN ('receipt', 'payment'));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- receipt: seed from the most recent bill issued this calendar year per
-- store (every pre-migration bill is type='receipt' by definition — the
-- type column above just defaulted them all). updated_at is backdated to
-- that bill's created_at (not "now") so a future year-rollover check compares
-- against the real last-issue date, not the migration's run date.
INSERT INTO tbl_bill_sequences (store_id, type, last_book_no, last_receipt_no, updated_at)
SELECT DISTINCT ON (b.store_id)
	b.store_id, 'receipt', b.book_no, b.receipt_no, b.created_at
FROM tbl_bills b
WHERE b.type = 'receipt'
	AND b.created_at >= date_trunc('year', now())
ORDER BY b.store_id, b.id DESC
ON CONFLICT (store_id, type) DO NOTHING;

-- Any store with no receipt bill yet this year (including stores with none
-- at all) still needs a row, starting from zero.
INSERT INTO tbl_bill_sequences (store_id, type)
SELECT s.id, 'receipt' FROM tbl_stores s
ON CONFLICT (store_id, type) DO NOTHING;

-- payment: brand new bill type, always starts from zero.
INSERT INTO tbl_bill_sequences (store_id, type)
SELECT s.id, 'payment' FROM tbl_stores s
ON CONFLICT (store_id, type) DO NOTHING;

-- 7) bill numbering uniqueness + report index --------------------------------------------------
-- No prior unique constraint existed on (store_id, book_no, receipt_no) —
-- uniqueness was only guaranteed by application-level locking — so this is
-- a new addition, not a replacement.

CREATE UNIQUE INDEX IF NOT EXISTS tbl_bills_store_type_no_uq
	ON tbl_bills (store_id, type, book_no, receipt_no);

CREATE INDEX IF NOT EXISTS tbl_bills_store_type_created_idx
	ON tbl_bills (store_id, type, created_at);

COMMIT;
