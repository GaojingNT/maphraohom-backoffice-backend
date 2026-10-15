package database

// BackfillBillSlipsSQL copies each bill's single pre-multi-slip bills.slip
// into tbl_bill_slips. It only inserts — bills.slip itself and the slip files
// in storage are left untouched. Kept identical to
// database/migrations/000004_bill_slips.up.sql.
//
// Idempotent: a bill already having a tbl_bill_slips row with that same key
// is skipped. Every slip change keeps bills.slip equal to one of the bill's
// rows (or NULL), so re-running never duplicates a slip.
const BackfillBillSlipsSQL = `
INSERT INTO tbl_bill_slips (bill_id, object_key, created_at, updated_at)
SELECT b.id, b.slip,
	COALESCE(b.slip_uploaded_at, b.updated_at),
	COALESCE(b.slip_uploaded_at, b.updated_at)
FROM tbl_bills b
WHERE b.slip IS NOT NULL AND b.slip <> ''
	AND NOT EXISTS (
		SELECT 1 FROM tbl_bill_slips s
		WHERE s.bill_id = b.id AND s.object_key = b.slip
	)
ORDER BY b.id
`
