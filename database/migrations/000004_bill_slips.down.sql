-- Reverts 000004 WITHOUT losing any slip: tbl_bills.slip is re-synced to
-- each bill's first slip (what the single-slip app shows), and
-- tbl_bill_slips is deliberately NOT dropped — bills attached with more
-- than one slip would lose the extra images. Drop it by hand only after
-- taking a backup and confirming nothing in it is still needed:
--
--   DROP TABLE tbl_bill_slips;
--
-- Note: while the app's models still declare BillSlip, GORM AutoMigrate
-- recreates the table on the next start anyway.
BEGIN;

UPDATE tbl_bills b
SET slip = (
		SELECT s.object_key FROM tbl_bill_slips s
		WHERE s.bill_id = b.id ORDER BY s.id LIMIT 1
	),
	slip_uploaded_at = (
		SELECT MAX(s.created_at) FROM tbl_bill_slips s WHERE s.bill_id = b.id
	)
WHERE EXISTS (SELECT 1 FROM tbl_bill_slips s WHERE s.bill_id = b.id);

COMMIT;
