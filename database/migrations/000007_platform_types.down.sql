-- Reverts 000007. WARNING: this drops every bill's platform tag
-- (tbl_bills.platform_type_id) — the bills themselves are kept. Back up
-- first if any bill has a platform set.
BEGIN;

ALTER TABLE tbl_bills DROP CONSTRAINT IF EXISTS fk_tbl_bills_platform_type;
ALTER TABLE tbl_bills DROP COLUMN IF EXISTS platform_type_id;
DROP TABLE IF EXISTS tbl_platform_types;

COMMIT;
