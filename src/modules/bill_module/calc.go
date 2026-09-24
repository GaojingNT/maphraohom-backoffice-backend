package bill_module

import (
	"time"

	"github.com/shopspring/decimal"
)

// MaxReceiptNoPerBook is how many receipts ("ใบเสร็จ") fit in one physical
// book ("เล่ม") before rolling over to the next book. Numbering resets to
// book 1 / receipt 1 at the start of every calendar year. This rule is
// scoped per (store, type) — see tbl_bill_sequences.
const MaxReceiptNoPerBook = 50

// computeSubtotal rounds quantity*price to 2 decimal places — the money
// scale every bills/bill_items numeric(_, 2) column uses.
func computeSubtotal(quantity decimal.Decimal, price decimal.Decimal) decimal.Decimal {
	return quantity.Mul(price).Round(2)
}

// computeTotal sums every line item's subtotal, then applies discount and
// shipping fee — the same formula for both receipt and payment bills.
func computeTotal(subtotals []decimal.Decimal, discount decimal.Decimal, shippingFee decimal.Decimal) decimal.Decimal {
	total := decimal.Zero
	for _, subtotal := range subtotals {
		total = total.Add(subtotal)
	}
	return total.Sub(discount).Add(shippingFee)
}

// nextBookReceipt is the pure form of the numbering rule applied by the
// "UPDATE tbl_bill_sequences ... RETURNING" statement that
// Repository.CreateBill runs inside its transaction (see the CASE
// expressions there — they must stay in sync with this function). It exists
// so the rule itself — cap at MaxReceiptNoPerBook per book, reset to book 1 /
// receipt 1 every calendar year — has a plain unit test; production
// numbering always goes through the SQL statement, never this function,
// because only a single atomically-locked UPDATE is safe under concurrent
// bill creation.
func nextBookReceipt(lastBookNo int, lastReceiptNo int, lastUpdatedAt time.Time, now time.Time) (bookNo int, receiptNo int) {
	if lastBookNo == 0 || lastUpdatedAt.Year() != now.Year() {
		return 1, 1
	}
	if lastReceiptNo >= MaxReceiptNoPerBook {
		return lastBookNo + 1, 1
	}
	return lastBookNo, lastReceiptNo + 1
}
