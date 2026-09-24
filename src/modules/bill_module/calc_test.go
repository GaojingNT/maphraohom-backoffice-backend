package bill_module

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func d(s string) decimal.Decimal {
	v, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return v
}

func TestComputeSubtotal(t *testing.T) {
	tests := []struct {
		name     string
		quantity string
		price    string
		want     string
	}{
		{"whole numbers", "10", "25", "250"},
		{"fractional kilograms", "20.5", "80", "1640"},
		// 0.1 + 0.2 style float traps must not leak into money math.
		{"rounds half up at 2dp", "3", "0.335", "1.01"},
		{"rounds down at 2dp", "3", "0.334", "1.00"},
		{"three decimal quantity", "1.234", "10", "12.34"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeSubtotal(d(tt.quantity), d(tt.price))
			if !got.Equal(d(tt.want)) {
				t.Errorf("computeSubtotal(%s, %s) = %s, want %s", tt.quantity, tt.price, got, tt.want)
			}
		})
	}
}

func TestComputeTotal(t *testing.T) {
	tests := []struct {
		name        string
		subtotals   []string
		discount    string
		shippingFee string
		want        string
	}{
		{"single item, no discount/shipping", []string{"1640"}, "0", "0", "1640"},
		{
			name:        "two items with discount and shipping",
			subtotals:   []string{"1640", "250"},
			discount:    "100",
			shippingFee: "50",
			want:        "1840",
		},
		{"no items", nil, "0", "0", "0"},
		{"discount exceeds subtotal but shipping covers it", []string{"100"}, "150", "60", "10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subtotals := make([]decimal.Decimal, 0, len(tt.subtotals))
			for _, s := range tt.subtotals {
				subtotals = append(subtotals, d(s))
			}
			got := computeTotal(subtotals, d(tt.discount), d(tt.shippingFee))
			if !got.Equal(d(tt.want)) {
				t.Errorf("computeTotal(%v, %s, %s) = %s, want %s", tt.subtotals, tt.discount, tt.shippingFee, got, tt.want)
			}
		})
	}
}

func TestComputeTotal_CanBeNegative(t *testing.T) {
	// computeTotal itself does not clamp — the caller (Service.validateAndPrice)
	// is responsible for rejecting a negative result with a 400.
	got := computeTotal([]decimal.Decimal{d("100")}, d("500"), d("0"))
	if !got.Equal(d("-400")) {
		t.Errorf("computeTotal(...) = %s, want -400", got)
	}
	if !got.IsNegative() {
		t.Errorf("expected IsNegative() to be true for %s", got)
	}
}

func TestNextBookReceipt(t *testing.T) {
	y2025 := time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC)
	y2026 := time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		lastBookNo    int
		lastReceiptNo int
		lastUpdatedAt time.Time
		now           time.Time
		wantBookNo    int
		wantReceiptNo int
	}{
		{
			name:          "brand new sequence (never used) starts at book 1 receipt 1",
			lastBookNo:    0,
			lastReceiptNo: 0,
			lastUpdatedAt: y2025,
			now:           y2025,
			wantBookNo:    1,
			wantReceiptNo: 1,
		},
		{
			name:          "normal increment within the same book",
			lastBookNo:    1,
			lastReceiptNo: 1,
			lastUpdatedAt: y2025,
			now:           y2025,
			wantBookNo:    1,
			wantReceiptNo: 2,
		},
		{
			name:          "receipt 49 -> 50 stays in the same book",
			lastBookNo:    1,
			lastReceiptNo: 49,
			lastUpdatedAt: y2025,
			now:           y2025,
			wantBookNo:    1,
			wantReceiptNo: 50,
		},
		{
			name:          "hitting the 50-receipt cap rolls to a new book",
			lastBookNo:    1,
			lastReceiptNo: 50,
			lastUpdatedAt: y2025,
			now:           y2025,
			wantBookNo:    2,
			wantReceiptNo: 1,
		},
		{
			name:          "past the cap (defensive) still rolls to a new book",
			lastBookNo:    2,
			lastReceiptNo: 63,
			lastUpdatedAt: y2025,
			now:           y2025,
			wantBookNo:    3,
			wantReceiptNo: 1,
		},
		{
			name:          "calendar year rollover resets to book 1 receipt 1 regardless of prior book",
			lastBookNo:    5,
			lastReceiptNo: 23,
			lastUpdatedAt: y2025,
			now:           y2026,
			wantBookNo:    1,
			wantReceiptNo: 1,
		},
		{
			name:          "year rollover takes priority even exactly at the receipt cap",
			lastBookNo:    5,
			lastReceiptNo: 50,
			lastUpdatedAt: y2025,
			now:           y2026,
			wantBookNo:    1,
			wantReceiptNo: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBook, gotReceipt := nextBookReceipt(tt.lastBookNo, tt.lastReceiptNo, tt.lastUpdatedAt, tt.now)
			if gotBook != tt.wantBookNo || gotReceipt != tt.wantReceiptNo {
				t.Errorf("nextBookReceipt(%d, %d, %s, %s) = (%d, %d), want (%d, %d)",
					tt.lastBookNo, tt.lastReceiptNo, tt.lastUpdatedAt, tt.now,
					gotBook, gotReceipt, tt.wantBookNo, tt.wantReceiptNo)
			}
		})
	}
}

func TestNextBookReceipt_SequentialCallsNeverRepeatANumber(t *testing.T) {
	// Simulates 120 bills issued one after another within the same year —
	// every (bookNo, receiptNo) pair returned must be unique, exercising the
	// cap-then-rollover rule across several books.
	now := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)

	bookNo, receiptNo := 0, 0
	lastUpdatedAt := now
	seen := make(map[[2]int]bool)

	for i := 0; i < 120; i++ {
		bookNo, receiptNo = nextBookReceipt(bookNo, receiptNo, lastUpdatedAt, now)
		lastUpdatedAt = now

		key := [2]int{bookNo, receiptNo}
		if seen[key] {
			t.Fatalf("iteration %d: (book %d, receipt %d) was already issued", i, bookNo, receiptNo)
		}
		seen[key] = true

		if receiptNo < 1 || receiptNo > MaxReceiptNoPerBook {
			t.Fatalf("iteration %d: receiptNo %d out of range [1, %d]", i, receiptNo, MaxReceiptNoPerBook)
		}
	}

	// 120 receipts at 50/book means book 3 should be in progress.
	if bookNo != 3 {
		t.Errorf("after 120 receipts, expected book 3, got book %d", bookNo)
	}
}
