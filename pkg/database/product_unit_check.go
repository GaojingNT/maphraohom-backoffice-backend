package database

import (
	"fmt"
	"strings"

	"maphraohom.app/maphraohom-backoffice/src/models"
)

// ProductUnitCheckSQL replaces migration 000001's tbl_products_unit_chk
// (which only allowed "กก." and "ขวด") with one allowing every unit in
// models.PaymentItemUnits, so items typed in on payment bills can be saved
// as products. NOT VALID like the original: existing rows aren't rechecked.
// Kept identical to database/migrations/000005_product_bill_scope.up.sql.
//
// Idempotent and data-safe: it only swaps a constraint, inside one
// transaction (a single multi-statement Exec), so there's never a moment
// without it.
var ProductUnitCheckSQL = fmt.Sprintf(`
BEGIN;
ALTER TABLE tbl_products DROP CONSTRAINT IF EXISTS tbl_products_unit_chk;
ALTER TABLE tbl_products ADD CONSTRAINT tbl_products_unit_chk
	CHECK (unit IN (%s)) NOT VALID;
COMMIT;
`, quotedUnits())

func quotedUnits() string {
	quoted := make([]string, 0, len(models.PaymentItemUnits))
	for _, unit := range models.PaymentItemUnits {
		quoted = append(quoted, "'"+strings.ReplaceAll(unit, "'", "''")+"'")
	}
	return strings.Join(quoted, ", ")
}
