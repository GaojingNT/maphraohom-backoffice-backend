package database

// ProductUnitCheckSQL replaces tbl_products_unit_chk (000001: "กก."/"ขวด"
// only; 000005: a fixed list of six) with "any non-blank unit", since
// payment-bill items can now be billed in any unit the user types. The
// varchar(20) column already caps the length. NOT VALID like before:
// existing rows aren't rechecked. Kept identical to
// database/migrations/000006_free_product_units.up.sql.
//
// Idempotent and data-safe: it only swaps a constraint, inside one
// transaction (a single multi-statement Exec), so there's never a moment
// without it.
const ProductUnitCheckSQL = `
BEGIN;
ALTER TABLE tbl_products DROP CONSTRAINT IF EXISTS tbl_products_unit_chk;
ALTER TABLE tbl_products ADD CONSTRAINT tbl_products_unit_chk
	CHECK (btrim(unit) <> '') NOT VALID;
COMMIT;
`
